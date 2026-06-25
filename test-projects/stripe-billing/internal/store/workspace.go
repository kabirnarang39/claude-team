package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// IsUserInWorkspace returns true if userID is a member of workspaceID.
func (s *Store) IsUserInWorkspace(ctx context.Context, userID, workspaceID string) (bool, error) {
	const q = `
		SELECT EXISTS(
			SELECT 1 FROM workspace_users
			WHERE user_id = $1 AND workspace_id = $2
		)`
	var exists bool
	if err := s.pool.QueryRow(ctx, q, userID, workspaceID).Scan(&exists); err != nil {
		return false, fmt.Errorf("store: is user in workspace: %w", err)
	}
	return exists, nil
}

// CountUsersInWorkspaceTx counts the number of users in a workspace within
// the provided transaction. Using FOR SHARE prevents TOCTOU races when a
// concurrent request is also adding a user.
func (s *Store) CountUsersInWorkspaceTx(ctx context.Context, tx pgx.Tx, workspaceID string) (int, error) {
	const q = `
		SELECT COUNT(*) FROM workspace_users
		WHERE workspace_id = $1
		FOR SHARE`
	var count int
	if err := tx.QueryRow(ctx, q, workspaceID).Scan(&count); err != nil {
		return 0, fmt.Errorf("store: count users in workspace tx: %w", err)
	}
	return count, nil
}
