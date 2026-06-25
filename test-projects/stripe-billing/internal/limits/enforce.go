package limits

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5"
)

// EnforceUserLimit checks whether the workspace has reached its plan's user limit.
// It runs the count inside a SERIALIZABLE transaction to prevent TOCTOU races
// when multiple users are being added concurrently.
//
// Return values:
//   - (0, nil, nil)  → within limit; caller should proceed.
//   - (status, body, nil) → limit exceeded; caller should write status+body and return.
//   - (0, nil, err)  → unexpected error; caller should return 500.
func EnforceUserLimit(ctx context.Context, db StoreIface, limitsCache *Cache, workspaceID string) (int, []byte, error) {
	pl, err := limitsCache.Get(ctx, workspaceID)
	if err != nil {
		return 0, nil, err
	}

	var currentCount int
	txErr := db.WithTx(ctx, func(tx pgx.Tx) error {
		n, err := db.CountUsersInWorkspaceTx(ctx, tx, workspaceID)
		if err != nil {
			return err
		}
		currentCount = n
		return nil
	})
	if txErr != nil {
		return 0, nil, txErr
	}

	if pl.UserLimit > 0 && currentCount >= pl.UserLimit {
		body, _ := json.Marshal(map[string]interface{}{
			"error":   "user_limit_exceeded",
			"message": "workspace has reached its user limit for the current plan",
			"limit":   pl.UserLimit,
			"current": currentCount,
			"plan":    pl.Plan,
		})
		return http.StatusForbidden, body, nil
	}

	return 0, nil, nil
}
