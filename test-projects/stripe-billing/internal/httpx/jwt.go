package httpx

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	workspaceIDKey contextKey = "workspace_id"
	userIDKey      contextKey = "user_id"
)

// JWTMiddleware validates the Bearer token from the Authorization header and
// injects workspace_id and user_id into the request context.
//
// The JWT must contain at least "sub" (user_id) and "workspace_id" claims.
// Requests with missing or invalid tokens receive 401.
func JWTMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				Error(w, http.StatusUnauthorized, "unauthorized", "missing or malformed Authorization header")
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				Error(w, http.StatusUnauthorized, "unauthorized", "invalid token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				Error(w, http.StatusUnauthorized, "unauthorized", "invalid token claims")
				return
			}

			userID, _ := claims["sub"].(string)
			workspaceID, _ := claims["workspace_id"].(string)

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			ctx = context.WithValue(ctx, workspaceIDKey, workspaceID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// WorkspaceScopeMiddleware verifies that the authenticated user belongs to the
// workspace referenced in the request. The workspace_id is read from the URL
// parameter or request body (it must already be in the context from JWTMiddleware).
//
// Returns 403 if the user is not a member of the workspace.
func WorkspaceScopeMiddleware(db StoreIface) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := UserIDFromContext(r.Context())
			if !ok || userID == "" {
				Error(w, http.StatusUnauthorized, "unauthorized", "missing user_id in context")
				return
			}
			workspaceID, ok := WorkspaceIDFromContext(r.Context())
			if !ok || workspaceID == "" {
				Error(w, http.StatusUnauthorized, "unauthorized", "missing workspace_id in context")
				return
			}

			member, err := db.IsUserInWorkspace(r.Context(), userID, workspaceID)
			if err != nil {
				Error(w, http.StatusInternalServerError, "internal_error", "failed to verify workspace membership")
				return
			}
			if !member {
				Error(w, http.StatusForbidden, "forbidden", "user does not belong to this workspace")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// WorkspaceIDFromContext returns the workspace_id stored in the context by JWTMiddleware.
func WorkspaceIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(workspaceIDKey).(string)
	return v, ok && v != ""
}

// UserIDFromContext returns the user_id (JWT sub claim) stored in the context.
func UserIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userIDKey).(string)
	return v, ok && v != ""
}
