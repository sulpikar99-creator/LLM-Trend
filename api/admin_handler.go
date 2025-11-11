package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Check if user is admin
func (s *Server) isAdmin(c *gin.Context) bool {
	userID := getUserID(c)
	if userID == "" {
		return false
	}

	var role string
	err := s.app.Database.DB.QueryRow("SELECT role FROM users WHERE id = ?", userID).Scan(&role)
	if err != nil {
		return false
	}

	return role == "admin"
}

// handleCreateBetaCode creates a new beta code (admin only)
func (s *Server) handleCreateBetaCode(c *gin.Context) {
	if !s.isAdmin(c) {
		errorResponse(c, http.StatusForbidden, "Admin access required")
		return
	}

	var req struct {
		Description string `json:"description"`
		MaxUses     int    `json:"max_uses" binding:"required,min=1"`
		ExpiresAt   string `json:"expires_at"` // ISO 8601 format
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Parse expiry date if provided
	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid expiry date format. Use ISO 8601 (RFC3339)")
			return
		}
		expiresAt = &t
	}

	// Create beta code
	userID := getUserID(c)
	betaCode, err := s.app.Config.BetaCodeManager.CreateBetaCode(req.Description, req.MaxUses, expiresAt, userID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to create beta code: "+err.Error())
		return
	}

	successResponse(c, betaCode)
}

// handleListBetaCodes lists all beta codes (admin only)
func (s *Server) handleListBetaCodes(c *gin.Context) {
	if !s.isAdmin(c) {
		errorResponse(c, http.StatusForbidden, "Admin access required")
		return
	}

	activeOnly := c.Query("active_only") == "true"

	codes, err := s.app.Config.BetaCodeManager.ListBetaCodes(activeOnly)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to list beta codes: "+err.Error())
		return
	}

	successResponse(c, gin.H{
		"beta_codes": codes,
		"count":      len(codes),
	})
}

// handleGetBetaCode gets a specific beta code (admin only)
func (s *Server) handleGetBetaCode(c *gin.Context) {
	if !s.isAdmin(c) {
		errorResponse(c, http.StatusForbidden, "Admin access required")
		return
	}

	code := c.Param("code")
	if code == "" {
		errorResponse(c, http.StatusBadRequest, "Beta code required")
		return
	}

	betaCode, err := s.app.Config.BetaCodeManager.GetBetaCode(code)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "Beta code not found")
		return
	}

	successResponse(c, betaCode)
}

// handleDeactivateBetaCode deactivates a beta code (admin only)
func (s *Server) handleDeactivateBetaCode(c *gin.Context) {
	if !s.isAdmin(c) {
		errorResponse(c, http.StatusForbidden, "Admin access required")
		return
	}

	code := c.Param("code")
	if code == "" {
		errorResponse(c, http.StatusBadRequest, "Beta code required")
		return
	}

	err := s.app.Config.BetaCodeManager.DeactivateBetaCode(code)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to deactivate beta code: "+err.Error())
		return
	}

	successResponse(c, gin.H{
		"message": "Beta code deactivated successfully",
		"code":    code,
	})
}

// handleReactivateBetaCode reactivates a beta code (admin only)
func (s *Server) handleReactivateBetaCode(c *gin.Context) {
	if !s.isAdmin(c) {
		errorResponse(c, http.StatusForbidden, "Admin access required")
		return
	}

	code := c.Param("code")
	if code == "" {
		errorResponse(c, http.StatusBadRequest, "Beta code required")
		return
	}

	err := s.app.Config.BetaCodeManager.ReactivateBetaCode(code)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to reactivate beta code: "+err.Error())
		return
	}

	successResponse(c, gin.H{
		"message": "Beta code reactivated successfully",
		"code":    code,
	})
}

// handleDeleteBetaCode deletes an unused beta code (admin only)
func (s *Server) handleDeleteBetaCode(c *gin.Context) {
	if !s.isAdmin(c) {
		errorResponse(c, http.StatusForbidden, "Admin access required")
		return
	}

	code := c.Param("code")
	if code == "" {
		errorResponse(c, http.StatusBadRequest, "Beta code required")
		return
	}

	err := s.app.Config.BetaCodeManager.DeleteBetaCode(code)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, gin.H{
		"message": "Beta code deleted successfully",
		"code":    code,
	})
}

// handleGetBetaCodeStats gets beta code statistics (admin only)
func (s *Server) handleGetBetaCodeStats(c *gin.Context) {
	if !s.isAdmin(c) {
		errorResponse(c, http.StatusForbidden, "Admin access required")
		return
	}

	stats, err := s.app.Config.BetaCodeManager.GetBetaCodeStats()
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to get stats: "+err.Error())
		return
	}

	successResponse(c, stats)
}

// handleListUsers lists all users (admin only)
func (s *Server) handleListUsers(c *gin.Context) {
	if !s.isAdmin(c) {
		errorResponse(c, http.StatusForbidden, "Admin access required")
		return
	}

	// Get pagination params
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	offset := (page - 1) * limit

	// Get total count
	var totalCount int
	err := s.app.Database.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalCount)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Get users
	rows, err := s.app.Database.DB.Query(`
		SELECT id, username, email, role, beta_code, is_active, created_at
		FROM users
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var id, username, email, role string
		var betaCode, isActive interface{}
		var createdAt time.Time

		err := rows.Scan(&id, &username, &email, &role, &betaCode, &isActive, &createdAt)
		if err != nil {
			continue
		}

		users = append(users, map[string]interface{}{
			"id":         id,
			"username":   username,
			"email":      email,
			"role":       role,
			"beta_code":  betaCode,
			"is_active":  isActive,
			"created_at": createdAt,
		})
	}

	successResponse(c, gin.H{
		"users":       users,
		"total_count": totalCount,
		"page":        page,
		"limit":       limit,
		"total_pages": (totalCount + limit - 1) / limit,
	})
}

// handleUpdateUserRole updates user role (admin only)
func (s *Server) handleUpdateUserRole(c *gin.Context) {
	if !s.isAdmin(c) {
		errorResponse(c, http.StatusForbidden, "Admin access required")
		return
	}

	userID := c.Param("user_id")
	if userID == "" {
		errorResponse(c, http.StatusBadRequest, "User ID required")
		return
	}

	var req struct {
		Role string `json:"role" binding:"required,oneof=user admin"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Don't allow changing own role
	if userID == getUserID(c) {
		errorResponse(c, http.StatusBadRequest, "Cannot change your own role")
		return
	}

	_, err := s.app.Database.DB.Exec(
		"UPDATE users SET role = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		req.Role, userID,
	)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to update role")
		return
	}

	successResponse(c, gin.H{
		"message": "User role updated successfully",
		"user_id": userID,
		"role":    req.Role,
	})
}

// handleDeactivateUser deactivates a user (admin only)
func (s *Server) handleDeactivateUser(c *gin.Context) {
	if !s.isAdmin(c) {
		errorResponse(c, http.StatusForbidden, "Admin access required")
		return
	}

	userID := c.Param("user_id")
	if userID == "" {
		errorResponse(c, http.StatusBadRequest, "User ID required")
		return
	}

	// Don't allow deactivating own account
	if userID == getUserID(c) {
		errorResponse(c, http.StatusBadRequest, "Cannot deactivate your own account")
		return
	}

	_, err := s.app.Database.DB.Exec(
		"UPDATE users SET is_active = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		userID,
	)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to deactivate user")
		return
	}

	successResponse(c, gin.H{
		"message": "User deactivated successfully",
		"user_id": userID,
	})
}

// handleReactivateUser reactivates a user (admin only)
func (s *Server) handleReactivateUser(c *gin.Context) {
	if !s.isAdmin(c) {
		errorResponse(c, http.StatusForbidden, "Admin access required")
		return
	}

	userID := c.Param("user_id")
	if userID == "" {
		errorResponse(c, http.StatusBadRequest, "User ID required")
		return
	}

	_, err := s.app.Database.DB.Exec(
		"UPDATE users SET is_active = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		userID,
	)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to reactivate user")
		return
	}

	successResponse(c, gin.H{
		"message": "User reactivated successfully",
		"user_id": userID,
	})
}
