package limits

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/kabirnarang39/stripe-billing/internal/store"
)

// StoreIface is the store dependency interface for the limits package.
type StoreIface interface {
	GetByWorkspace(ctx context.Context, workspaceID string) (*store.Subscription, error)
	WithTx(ctx context.Context, fn func(pgx.Tx) error) error
	CountUsersInWorkspaceTx(ctx context.Context, tx pgx.Tx, workspaceID string) (int, error)
}
