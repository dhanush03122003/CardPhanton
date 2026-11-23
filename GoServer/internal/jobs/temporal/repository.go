package temporal

import (
	"context"
	"database/sql"
	"fmt"

	"webauthn-server/internal/timeutil"

	"github.com/google/uuid"
)

// SQLCardEventRepository updates only values present in a CardEvent. It accepts
// database/sql so it can be backed by PostgreSQL or another SQL driver.
type SQLCardEventRepository struct {
	DB *sql.DB
}

func (r *SQLCardEventRepository) UpdateCardLimits(ctx context.Context, event *CardEvent) error {
	if r == nil || r.DB == nil {
		return fmt.Errorf("sql database is not configured")
	}
	if event == nil {
		return fmt.Errorf("card event is nil")
	}
	if event.CardID == "" {
		return fmt.Errorf("card ID is required")
	}

	sets := make([]string, 0, 2)
	args := make([]any, 0, 4)
	if event.TotalLimit != nil {
		sets = append(sets, "total_limit = ?")
		args = append(args, *event.TotalLimit)
	}
	if event.AvailableLimit != nil {
		sets = append(sets, "available_limit = ?")
		args = append(args, *event.AvailableLimit)
	}
	if len(sets) == 0 {
		return nil
	}

	args = append(args, event.CardID)
	query := fmt.Sprintf("UPDATE cards SET %s WHERE id = ?", joinSQL(sets, ", "))
	_, err := r.DB.ExecContext(ctx, query, args...)
	return err
}

func (r *SQLCardEventRepository) CreateCardSyncRun(ctx context.Context, run *CardSyncRun) error {
	if r == nil || r.DB == nil {
		return fmt.Errorf("sql database is not configured")
	}
	if run == nil {
		return fmt.Errorf("card sync run is nil")
	}
	now := timeutil.Now()
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO card_sync_runs (id, card_id, PAN, workflow_id, run_id, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, uuid.NewString(), run.CardID, run.PAN, run.WorkflowID, run.RunID, run.Status, now, now)
	return err
}

func (r *SQLCardEventRepository) UpdateCardSyncRunStatus(ctx context.Context, update CardSyncRunStatusUpdate) error {
	if r == nil || r.DB == nil {
		return fmt.Errorf("sql database is not configured")
	}
	_, err := r.DB.ExecContext(ctx,
		"UPDATE card_sync_runs SET status = ?, updated_at = ? WHERE run_id = ?",
		update.Status,
		timeutil.Now(),
		update.RunID,
	)
	return err
}

func joinSQL(values []string, separator string) string {
	result := ""
	for index, value := range values {
		if index > 0 {
			result += separator
		}
		result += value
	}
	return result
}
