package admin

import (
	"context"
	"time"
)

type RefreshRun struct {
	ID         string    `json:"id"`
	CardID     string    `json:"card_id"`
	PAN        string    `json:"pan"`
	WorkflowID string    `json:"workflow_id"`
	RunID      string    `json:"run_id"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (r *SQLiteRepository) GetRefreshRuns(ctx context.Context) ([]RefreshRun, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT id, card_id, PAN, workflow_id, run_id, status, created_at, updated_at
		FROM card_sync_runs
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []RefreshRun
	for rows.Next() {
		var run RefreshRun
		if err := rows.Scan(
			&run.ID,
			&run.CardID,
			&run.PAN,
			&run.WorkflowID,
			&run.RunID,
			&run.Status,
			&run.CreatedAt,
			&run.UpdatedAt,
		); err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return runs, nil
}
