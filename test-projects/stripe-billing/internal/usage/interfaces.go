package usage

import (
	"context"
	"time"
)

// StoreIface is the store dependency interface for the usage package.
type StoreIface interface {
	UpsertUsage(ctx context.Context, workspaceID string, periodStart, periodEnd time.Time, delta int64) error
}
