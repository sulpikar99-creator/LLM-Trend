package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sulpikar99-creator/LLM-Trend/auth"
	"golang.org/x/crypto/bcrypt"
)

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	BetaCode string `json:"beta_code"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// handleRegister handles user registration
func (s *Server) handleRegister(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Validate beta code if provided
	betaCodeMgr := s.app.Config.BetaCodeManager
	if req.BetaCode != "" {
		valid, msg := betaCodeMgr.ValidateBetaCode(req.BetaCode)
		if !valid {
			errorResponse(c, http.StatusBadRequest, "Beta code validation failed: "+msg)
			return
		}
	}

	// Check if username or email already exists
	var exists bool
	err := s.app.Database.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM users WHERE username = ? OR email = ?)",
		req.Username, req.Email,
	).Scan(&exists)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}
	if exists {
		errorResponse(c, http.StatusConflict, "Username or email already exists")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Create user
	userID := uuid.New().String()
	var betaCodeParam interface{}
	if req.BetaCode != "" {
		betaCodeParam = req.BetaCode
	} else {
		betaCodeParam = nil
	}

	_, err = s.app.Database.DB.Exec(
		`INSERT INTO users (id, username, email, password_hash, role, beta_code)
		VALUES (?, ?, ?, ?, ?, ?)`,
		userID, req.Username, req.Email, string(hashedPassword), "user", betaCodeParam,
	)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Use beta code if provided
	if req.BetaCode != "" {
		if err := betaCodeMgr.UseBetaCode(req.BetaCode); err != nil {
			// Log error but don't fail registration
			// User is already created
			// TODO: Add proper logging
		}
	}

	// Create default user config
	userConfigMgr := s.app.Config.UserConfigManager
	if err := userConfigMgr.CreateDefaultConfig(userID); err != nil {
		// Log error but don't fail registration
		// User can configure later
	}

	// Generate token
	token, err := auth.GenerateToken(userID, req.Username, "user")
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	successResponse(c, AuthResponse{
		Token:    token,
		UserID:   userID,
		Username: req.Username,
		Role:     "user",
	})
}

// handleLogin handles user login
func (s *Server) handleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Get user from database
	var userID, username, email, passwordHash, role string
	err := s.app.Database.DB.QueryRow(
		`SELECT id, username, email, password_hash, role FROM users WHERE username = ?`,
		req.Username,
	).Scan(&userID, &username, &email, &passwordHash, &role)
	if err == sql.ErrNoRows {
		errorResponse(c, http.StatusUnauthorized, "Invalid username or password")
		return
	}
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		errorResponse(c, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	// Generate token
	token, err := auth.GenerateToken(userID, username, role)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	successResponse(c, AuthResponse{
		Token:    token,
		UserID:   userID,
		Username: username,
		Role:     role,
	})
}

// handleGetProfile handles get user profile
func (s *Server) handleGetProfile(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var username, email, role string
	err := s.app.Database.DB.QueryRow(
		`SELECT username, email, role FROM users WHERE id = ?`,
		userID,
	).Scan(&username, &email, &role)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to get user profile")
		return
	}

	successResponse(c, gin.H{
		"user_id":  userID,
		"username": username,
		"email":    email,
		"role":     role,
	})
}

// handleUpdateProfile handles update user profile
func (s *Server) handleUpdateProfile(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req struct {
		Email string `json:"email" binding:"omitempty,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	if req.Email != "" {
		_, err := s.app.Database.DB.Exec(
			`UPDATE users SET email = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
			req.Email, userID,
		)
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, "Failed to update profile")
			return
		}
	}

	successResponse(c, gin.H{"message": "Profile updated successfully"})
}
