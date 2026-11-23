package temporal

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
)

type EmailClient interface {
	FetchRecentMetadata(ctx context.Context, email, bankName, cardLast4 string) ([]EmailMetadata, error)
	FetchFullBodyByUID(ctx context.Context, email string, uid uint32) (string, error)
}

type LLMClient interface {
	RouteEmails(ctx context.Context, emails []EmailMetadata) ([]int, error)
	ExtractCardLimits(ctx context.Context, emailBody string) (*CardEvent, error)
}

type CardEventRepository interface {
	UpdateCardLimits(ctx context.Context, event *CardEvent) error
	CreateCardSyncRun(ctx context.Context, run *CardSyncRun) error
	UpdateCardSyncRunStatus(ctx context.Context, update CardSyncRunStatusUpdate) error
}

type JobActivities struct {
	Email EmailClient
	LLM  LLMClient
	Repo CardEventRepository
}

func (a *JobActivities) FetchEmailMetadataActivity(ctx context.Context, input SyncWorkflowInput) ([]EmailMetadata, error) {
	if a == nil || a.Email == nil {
		return nil, fmt.Errorf("imap client is not configured")
	}
	if input.Email == "" {
		return nil, fmt.Errorf("card email is required")
	}
	return a.Email.FetchRecentMetadata(ctx, input.Email, input.BankName, input.CardLast4)
}

func (a *JobActivities) RouteEmailsLLMActivity(ctx context.Context, metadata []EmailMetadata) ([]int, error) {
	if a == nil || a.LLM == nil {
		return nil, fmt.Errorf("llm client is not configured")
	}
	return a.LLM.RouteEmails(ctx, metadata)
}

func (a *JobActivities) ExtractEmailLimitsActivity(ctx context.Context, email string, uid uint32) (*CardEvent, error) {
	if a == nil || a.Email == nil || a.LLM == nil {
		return nil, fmt.Errorf("imap client and llm client are required")
	}
	body, err := a.Email.FetchFullBodyByUID(ctx, email, uid)
	if err != nil {
		return nil, err
	}
	return a.LLM.ExtractCardLimits(ctx, body)
}

func (a *JobActivities) SaveCardEventActivity(ctx context.Context, event *CardEvent) error {
	if a == nil || a.Repo == nil {
		return fmt.Errorf("card event repository is not configured")
	}
	if event == nil {
		return fmt.Errorf("card event is nil")
	}
	if event.CardID == "" {
		return fmt.Errorf("card ID is required")
	}
	return a.Repo.UpdateCardLimits(ctx, event)
}

func (a *JobActivities) CreateCardSyncRunActivity(ctx context.Context, input SyncWorkflowInput) error {
	if a == nil || a.Repo == nil {
		return fmt.Errorf("card event repository is not configured")
	}
	if input.CardID == "" || input.PAN == "" {
		return fmt.Errorf("card ID and PAN are required")
	}
	info := activity.GetInfo(ctx)
	return a.Repo.CreateCardSyncRun(ctx, &CardSyncRun{
		CardID:     input.CardID,
		PAN:        input.PAN,
		WorkflowID: info.WorkflowExecution.ID,
		RunID:      info.WorkflowExecution.RunID,
		Status:     "RUNNING",
	})
}

func (a *JobActivities) UpdateCardSyncRunActivity(ctx context.Context, update CardSyncRunStatusUpdate) error {
	if a == nil || a.Repo == nil {
		return fmt.Errorf("card event repository is not configured")
	}
	if update.RunID == "" {
		return fmt.Errorf("run ID is required")
	}
	if update.Status != "COMPLETED" && update.Status != "FAILED" {
		return fmt.Errorf("invalid card sync run status %q", update.Status)
	}
	return a.Repo.UpdateCardSyncRunStatus(ctx, update)
}
