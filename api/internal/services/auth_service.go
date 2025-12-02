package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/repository"
	"github.com/humfurie/pulpulitiko/api/pkg/email"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo     *repository.UserRepository
	roleRepo     *repository.RoleRepository
	authorRepo   *repository.AuthorRepository
	emailService *email.EmailService
	jwtSecret    []byte
}

func NewAuthService(userRepo *repository.UserRepository, roleRepo *repository.RoleRepository, authorRepo *repository.AuthorRepository, emailService *email.EmailService, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		authorRepo:   authorRepo,
		emailService: emailService,
		jwtSecret:    []byte(jwtSecret),
	}
}

type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	RoleID string `json:"role_id"`
	Role   string `json:"role"` // Role slug for backwards compatibility
	jwt.RegisteredClaims
}

func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Get user permissions
	var permissions []string
	if user.RoleID != nil {
		permissions, _ = s.roleRepo.GetPermissionSlugsByRoleID(ctx, *user.RoleID)
	}

	return &models.LoginResponse{
		Token:       token,
		User:        *user,
		Permissions: permissions,
	}, nil
}

// GetPermissionsByRoleID returns permission slugs for a role
func (s *AuthService) GetPermissionsByRoleID(ctx context.Context, roleID uuid.UUID) ([]string, error) {
	return s.roleRepo.GetPermissionSlugsByRoleID(ctx, roleID)
}

func (s *AuthService) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("user with this email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return nil, fmt.Errorf("invalid role ID: %w", err)
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Name:         req.Name,
		RoleID:       &roleID,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Register creates a new user with the "user" role (public registration)
func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.LoginResponse, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("user with this email already exists")
	}

	// Get the "user" role
	userRole, err := s.roleRepo.GetBySlug(ctx, "user")
	if err != nil {
		return nil, fmt.Errorf("failed to get user role: %w", err)
	}
	if userRole == nil {
		return nil, fmt.Errorf("user role not found in system")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user with "user" role
	user := &models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Name:         req.Name,
		RoleID:       &userRole.ID,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create corresponding author record for profile functionality
	slug := s.generateSlug(req.Name)
	author := &models.Author{
		Name:   req.Name,
		Slug:   slug,
		Email:  &req.Email,
		RoleID: &userRole.ID,
	}
	if err := s.authorRepo.Create(ctx, author); err != nil {
		// Log but don't fail - user is created, author profile can be created later
		fmt.Printf("Warning: failed to create author profile for user %s: %v\n", req.Email, err)
	}

	// Fetch the user again to get the role slug from the join
	user, err = s.userRepo.GetByID(ctx, user.ID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("failed to fetch created user: %w", err)
	}

	// Generate token for immediate login
	token, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Get user permissions
	var permissions []string
	if user.RoleID != nil {
		permissions, _ = s.roleRepo.GetPermissionSlugsByRoleID(ctx, *user.RoleID)
	}

	return &models.LoginResponse{
		Token:       token,
		User:        *user,
		Permissions: permissions,
	}, nil
}

// generateSlug creates a URL-friendly slug from a name
func (s *AuthService) generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove special characters, keeping only alphanumeric and hyphens
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	slug = result.String()
	// Add timestamp suffix to ensure uniqueness
	slug = fmt.Sprintf("%s-%d", slug, time.Now().UnixNano()%100000)
	return slug
}

func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func (s *AuthService) generateToken(user *models.User) (string, error) {
	roleID := ""
	if user.RoleID != nil {
		roleID = user.RoleID.String()
	}

	claims := &JWTClaims{
		UserID: user.ID.String(),
		Email:  user.Email,
		RoleID: roleID,
		Role:   user.RoleSlug,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "pulpulitiko",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// ForgotPassword initiates password reset by sending an email with reset link
func (s *AuthService) ForgotPassword(ctx context.Context, req *models.ForgotPasswordRequest) error {
	// Check if email service is configured
	if s.emailService == nil || !s.emailService.IsConfigured() {
		return fmt.Errorf("email service not configured")
	}

	// Find user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return fmt.Errorf("failed to check user: %w", err)
	}

	// Always return success to prevent email enumeration attacks
	if user == nil {
		return nil
	}

	// Invalidate any existing tokens for this user
	if err := s.userRepo.InvalidateUserPasswordResetTokens(ctx, user.ID); err != nil {
		return fmt.Errorf("failed to invalidate existing tokens: %w", err)
	}

	// Generate a secure random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	// Create password reset token (expires in 1 hour)
	resetToken := &models.PasswordResetToken{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	if err := s.userRepo.CreatePasswordResetToken(ctx, resetToken); err != nil {
		return fmt.Errorf("failed to create reset token: %w", err)
	}

	// Send email
	if err := s.emailService.SendPasswordReset(user.Email, token); err != nil {
		return fmt.Errorf("failed to send reset email: %w", err)
	}

	return nil
}

// ResetPassword resets the user's password using a valid token
func (s *AuthService) ResetPassword(ctx context.Context, req *models.ResetPasswordRequest) error {
	// Get the reset token
	resetToken, err := s.userRepo.GetPasswordResetToken(ctx, req.Token)
	if err != nil {
		return fmt.Errorf("failed to verify token: %w", err)
	}

	if resetToken == nil {
		return fmt.Errorf("invalid or expired reset token")
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update the password
	if err := s.userRepo.UpdatePassword(ctx, resetToken.UserID, string(hashedPassword)); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark the token as used
	if err := s.userRepo.MarkPasswordResetTokenUsed(ctx, resetToken.ID); err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	// Invalidate all other tokens for this user
	if err := s.userRepo.InvalidateUserPasswordResetTokens(ctx, resetToken.UserID); err != nil {
		return fmt.Errorf("failed to invalidate other tokens: %w", err)
	}

	return nil
}
