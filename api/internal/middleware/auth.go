package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/services"
)

type contextKey string

const UserContextKey contextKey = "user"

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
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(UserContextKey).(*services.JWTClaims)
		if !ok {
			http.Error(w, `{"success":false,"error":{"code":"UNAUTHORIZED","message":"unauthorized"}}`, http.StatusUnauthorized)
			return
		}

		if claims.Role != string(models.UserRoleAdmin) {
			http.Error(w, `{"success":false,"error":{"code":"FORBIDDEN","message":"admin access required"}}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func GetUserClaims(ctx context.Context) *services.JWTClaims {
	claims, ok := ctx.Value(UserContextKey).(*services.JWTClaims)
	if !ok {
		return nil
	}
	return claims
}
