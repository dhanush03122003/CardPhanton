package temporal

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func LimitExtractionWorkflow(ctx workflow.Context, input SyncWorkflowInput) (err error) {
	imapOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 3},
	}
	llmOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy:         &temporal.RetryPolicy{MaximumInterval: time.Minute, MaximumAttempts: 4},
	}
	dbOptions := workflow.ActivityOptions{StartToCloseTimeout: 10 * time.Second}


	if err := workflow.ExecuteActivity(workflow.WithActivityOptions(ctx, dbOptions), CreateCardSyncRunActivityName, input).Get(ctx, nil); err != nil {
		return fmt.Errorf("create card sync run: %w", err)
	}
	defer func() {
		status := "COMPLETED"
		if err != nil {
			status = "FAILED"
		}
		// Create a disconnected context so the DB update still runs if ctx was canceled
		disconnectedCtx, _ := workflow.NewDisconnectedContext(ctx)
		_ = workflow.ExecuteActivity(
			workflow.WithActivityOptions(disconnectedCtx, dbOptions),
			UpdateCardSyncRunActivityName,
			CardSyncRunStatusUpdate{
				RunID:  workflow.GetInfo(ctx).WorkflowExecution.RunID,
				Status: status,
			},
		).Get(disconnectedCtx, nil)
	}()
	
	var metadata []EmailMetadata
	if err := workflow.ExecuteActivity(workflow.WithActivityOptions(ctx, imapOptions), FetchEmailMetadataActivityName, input).Get(ctx, &metadata); err != nil {
		return fmt.Errorf("fetch email metadata: %w", err)
	}
	if len(metadata) == 0 {
		return nil
	}

	var indices []int
	if err := workflow.ExecuteActivity(workflow.WithActivityOptions(ctx, llmOptions), RouteEmailsLLMActivityName, metadata).Get(ctx, &indices); err != nil {
		return fmt.Errorf("route email metadata: %w", err)
	}
	for _, index := range indices {
		if index < 0 || index >= len(metadata) {
			return fmt.Errorf("llm returned invalid email index %d", index)
		}
		var event *CardEvent
		if err := workflow.ExecuteActivity(workflow.WithActivityOptions(ctx, llmOptions), ExtractEmailLimitsActivityName, input.Email, metadata[index].UID).Get(ctx, &event); err != nil {
			return fmt.Errorf("extract email %d: %w", index, err)
		}
		if event == nil {
			continue
		}
		event.CardID = input.CardID
		if err := workflow.ExecuteActivity(workflow.WithActivityOptions(ctx, dbOptions), SaveCardEventActivityName, event).Get(ctx, nil); err != nil {
			return fmt.Errorf("save email %d event: %w", index, err)
		}
	}
	return nil
}
