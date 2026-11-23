package temporal

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

type WorkflowStarter struct {
	Client client.Client
}

func (s WorkflowStarter) StartLimitExtraction(ctx context.Context, cardID, pan, bankName, cardLast4, email string) (string, string, error) {
	if s.Client == nil {
		return "", "", fmt.Errorf("temporal client is not configured")
	}
	run, err := s.Client.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        fmt.Sprintf("credit-limit-%d", time.Now().UnixNano()),
		TaskQueue: CreditLimitTaskQueue,
	}, LimitExtractionWorkflow, SyncWorkflowInput{CardID: cardID, PAN: pan, BankName: bankName, CardLast4: cardLast4, Email: email})
	if err != nil {
		return "", "", err
	}
	return run.GetID(), run.GetRunID(), nil
}

// WorkerConfig contains the Temporal client and the dependencies used by the
// activity methods. The Temporal client should be created by the worker
// executable, not by workflow code.
type WorkerConfig struct {
	Client     client.Client
	Activities *JobActivities
	Options    worker.Options
}

// NewWorker creates a worker with the workflow and all activity methods
// registered. The activity names match the constants used by the workflow.
func NewWorker(config WorkerConfig) (worker.Worker, error) {
	if config.Client == nil {
		return nil, fmt.Errorf("temporal client is not configured")
	}
	if config.Activities == nil {
		return nil, fmt.Errorf("job activities are not configured")
	}

	w := worker.New(config.Client, CreditLimitTaskQueue, config.Options)
	w.RegisterWorkflow(LimitExtractionWorkflow)
	w.RegisterActivity(config.Activities)
	return w, nil
}
