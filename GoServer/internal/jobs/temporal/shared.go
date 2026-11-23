package temporal

import "time"

const CreditLimitTaskQueue = "credit-limit-task-queue"

const (
	FetchEmailMetadataActivityName = "FetchEmailMetadataActivity"
	RouteEmailsLLMActivityName     = "RouteEmailsLLMActivity"
	ExtractEmailLimitsActivityName = "ExtractEmailLimitsActivity"
	CreateCardSyncRunActivityName  = "CreateCardSyncRunActivity"
	SaveCardEventActivityName      = "SaveCardEventActivity"
	UpdateCardSyncRunActivityName  = "UpdateCardSyncRunActivity"
)

type SyncWorkflowInput struct {
	CardID    string
	PAN       string
	BankName  string
	CardLast4 string
	Email     string
}

type CardSyncRun struct {
	ID         string    `json:"id"`
	CardID     string    `json:"card_id"`
	PAN        string    `json:"pan"`
	WorkflowID string    `json:"workflow_id"`
	RunID      string    `json:"run_id"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CardSyncRunStatusUpdate struct {
	RunID  string `json:"run_id"`
	Status string `json:"status"`
}

type EmailMetadata struct {
	Index   int    `json:"index"`
	UID     uint32 `json:"uid"`
	Date    string `json:"date"`
	Subject string `json:"subject"`
	Snippet string `json:"snippet"`
}

type RouterOutput struct {
	SelectedIndices []int `json:"selected_indices"`
}

type CardEvent struct {
	CardID            string   `json:"card_id"`
	BankName          string   `json:"bank_name"`
	CardLastFour      string   `json:"card_last_four"`
	EventType         string   `json:"event_type"`
	TotalLimit        *float64 `json:"total_limit"`
	AvailableLimit    *float64 `json:"available_limit"`
	TransactionAmount *float64 `json:"transaction_amount"`
}
