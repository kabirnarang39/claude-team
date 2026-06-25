package httpx

import "context"

// StoreIface is the store dependency interface for the httpx package.
type StoreIface interface {
	IsUserInWorkspace(ctx context.Context, userID, workspaceID string) (bool, error)
}
