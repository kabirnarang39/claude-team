package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// UpsertUsage atomically increments the call_count for the given workspace and
// billing period by delta. If no row exists yet it is inserted with call_count
// set to delta.
func (s *Store) UpsertUsage(ctx context.Context, workspaceID string, periodStart, periodEnd time.Time, delta int64) error {
	const q = `
		INSERT INTO api_usage (workspace_id, billing_period_start, billing_period_end, call_count)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (workspace_id, billing_period_start) DO UPDATE
			SET call_count  = api_usage.call_count + EXCLUDED.call_count,
			    updated_at  = NOW()`

	_, err := s.pool.Exec(ctx, q, workspaceID, periodStart, periodEnd, delta)
	if err != nil {
		return fmt.Errorf("store: upsert api usage: %w", err)
	}
	return nil
}

// GetUsage returns the api_usage row for the given workspace and period start.
// Returns pgx.ErrNoRows (wrapped) if no row exists yet.
func (s *Store) GetUsage(ctx context.Context, workspaceID string, periodStart time.Time) (*APIUsage, error) {
	const q = `
		SELECT id, workspace_id, billing_period_start, billing_period_end,
		       call_count, updated_at
		FROM api_usage
		WHERE workspace_id = $1
		  AND billing_period_start = $2`

	u := &APIUsage{}
	err := s.pool.QueryRow(ctx, q, workspaceID, periodStart).Scan(
		&u.ID,
		&u.WorkspaceID,
		&u.BillingPeriodStart,
		&u.BillingPeriodEnd,
		&u.CallCount,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("store: api usage not found for workspace %s period %s: %w",
				workspaceID, periodStart.Format(time.RFC3339), pgx.ErrNoRows)
		}
		return nil, fmt.Errorf("store: get api usage: %w", err)
	}
	return u, nil
}
