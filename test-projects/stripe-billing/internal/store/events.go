package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// InsertEvent records a Stripe event for idempotency tracking. It must be
// called inside the same transaction as the business-logic update so that a
// duplicate event never partially processes.
//
// Returns inserted=true when the event is new, inserted=false when it already
// exists (duplicate delivery). The caller should skip processing and return
// HTTP 200 when inserted=false.
func (s *Store) InsertEvent(ctx context.Context, tx pgx.Tx, eventID, eventType string) (bool, error) {
	const q = `
		INSERT INTO stripe_events (event_id, event_type)
		VALUES ($1, $2)
		ON CONFLICT (event_id) DO NOTHING`

	tag, err := tx.Exec(ctx, q, eventID, eventType)
	if err != nil {
		return false, fmt.Errorf("store: insert stripe event: %w", err)
	}
	return tag.RowsAffected() == 1, nil
}
