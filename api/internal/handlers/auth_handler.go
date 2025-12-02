package handlers

import (
	"net/http"

	"github.com/humfurie/pulpulitiko/api/internal/middleware"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/services"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// POST /api/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteValidationError(w, err)
		return
	}

	response, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid credentials")
		return
	}

	WriteSuccess(w, response)
}

// GET /api/auth/me
func (h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	WriteSuccess(w, map[string]interface{}{
		"user_id": claims.UserID,
		"email":   claims.Email,
		"role":    claims.Role,
	})
}

// POST /api/admin/users
func (h *AuthHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteValidationError(w, err)
		return
	}

	user, err := h.authService.CreateUser(r.Context(), &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteCreated(w, user)
}

// POST /api/auth/register - Public user registration (always gets "user" role)
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteValidationError(w, err)
		return
	}

	response, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		// Check if it's a duplicate email error
		if err.Error() == "user with this email already exists" {
			WriteError(w, http.StatusConflict, "EMAIL_EXISTS", "A user with this email already exists")
			return
		}
		WriteInternalError(w, err.Error())
		return
	}

	WriteCreated(w, response)
}

// POST /api/auth/forgot-password - Request password reset email
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req models.ForgotPasswordRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteValidationError(w, err)
		return
	}

	err := h.authService.ForgotPassword(r.Context(), &req)
	if err != nil {
		// Check if it's a configuration error
		if err.Error() == "email service not configured" {
			WriteError(w, http.StatusServiceUnavailable, "EMAIL_NOT_CONFIGURED", "Password reset is temporarily unavailable")
			return
		}
		WriteInternalError(w, err.Error())
		return
	}

	// Always return success to prevent email enumeration
	WriteSuccess(w, map[string]string{
		"message": "If an account exists with this email, you will receive a password reset link",
	})
}

// POST /api/auth/reset-password - Reset password with token
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req models.ResetPasswordRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteValidationError(w, err)
		return
	}

	err := h.authService.ResetPassword(r.Context(), &req)
	if err != nil {
		if err.Error() == "invalid or expired reset token" {
			WriteError(w, http.StatusBadRequest, "INVALID_TOKEN", "Invalid or expired reset token")
			return
		}
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, map[string]string{
		"message": "Password has been reset successfully",
	})
}
