package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/services"
)

type contextKey string

const UserContextKey contextKey = "user"
const PermissionsContextKey contextKey = "permissions"

type AuthMiddleware struct {
	authService *services.AuthService
}

func NewAuthMiddleware(authService *services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"success":false,"error":{"code":"UNAUTHORIZED","message":"missing authorization header"}}`, http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"success":false,"error":{"code":"UNAUTHORIZED","message":"invalid authorization header format"}}`, http.StatusUnauthorized)
			return
		}

		claims, err := m.authService.ValidateToken(parts[1])
		if err != nil {
			http.Error(w, `{"success":false,"error":{"code":"UNAUTHORIZED","message":"invalid or expired token"}}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, claims)

		// Load permissions for the user's role
		if claims.RoleID != "" {
			roleID, err := uuid.Parse(claims.RoleID)
			if err == nil {
				permissions, _ := m.authService.GetPermissionsByRoleID(r.Context(), roleID)
				ctx = context.WithValue(ctx, PermissionsContextKey, permissions)
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth extracts user claims if a valid token is present, but doesn't require authentication
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// No token - continue without user context
			next.ServeHTTP(w, r)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Invalid format - continue without user context
			next.ServeHTTP(w, r)
			return
		}

		claims, err := m.authService.ValidateToken(parts[1])
		if err != nil {
			// Invalid token - continue without user context
			next.ServeHTTP(w, r)
			return
		}

		// Valid token - add user context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)

		// Load permissions for the user's role
		if claims.RoleID != "" {
			roleID, err := uuid.Parse(claims.RoleID)
			if err == nil {
				permissions, _ := m.authService.GetPermissionsByRoleID(r.Context(), roleID)
				ctx = context.WithValue(ctx, PermissionsContextKey, permissions)
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAdmin checks if user has admin role
func (m *AuthMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(UserContextKey).(*services.JWTClaims)
		if !ok {
			http.Error(w, `{"success":false,"error":{"code":"UNAUTHORIZED","message":"unauthorized"}}`, http.StatusUnauthorized)
			return
		}

		if claims.Role != "admin" {
			http.Error(w, `{"success":false,"error":{"code":"FORBIDDEN","message":"admin access required"}}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequirePermission creates middleware that checks for a specific permission
func (m *AuthMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			permissions, ok := r.Context().Value(PermissionsContextKey).([]string)
			if !ok {
				http.Error(w, `{"success":false,"error":{"code":"FORBIDDEN","message":"permission denied"}}`, http.StatusForbidden)
				return
			}

			hasPermission := false
			for _, p := range permissions {
				if p == permission {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				errMsg := fmt.Sprintf(`{"success":false,"error":{"code":"FORBIDDEN","message":"permission '%s' required"}}`, permission)
				http.Error(w, errMsg, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission creates middleware that checks for any of the given permissions
func (m *AuthMiddleware) RequireAnyPermission(requiredPermissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			permissions, ok := r.Context().Value(PermissionsContextKey).([]string)
			if !ok {
				http.Error(w, `{"success":false,"error":{"code":"FORBIDDEN","message":"permission denied"}}`, http.StatusForbidden)
				return
			}

			hasPermission := false
			for _, userPerm := range permissions {
				for _, reqPerm := range requiredPermissions {
					if userPerm == reqPerm {
						hasPermission = true
						break
					}
				}
				if hasPermission {
					break
				}
			}

			if !hasPermission {
				http.Error(w, `{"success":false,"error":{"code":"FORBIDDEN","message":"insufficient permissions"}}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetUserClaims(ctx context.Context) *services.JWTClaims {
	claims, ok := ctx.Value(UserContextKey).(*services.JWTClaims)
	if !ok {
		return nil
	}
	return claims
}

func GetUserPermissions(ctx context.Context) []string {
	permissions, ok := ctx.Value(PermissionsContextKey).([]string)
	if !ok {
		return nil
	}
	return permissions
}

// HasPermission checks if the current user has a specific permission
func HasPermission(ctx context.Context, permission string) bool {
	permissions := GetUserPermissions(ctx)
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}
