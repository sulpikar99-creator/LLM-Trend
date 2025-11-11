package config

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
)

// BetaCode represents a beta access code
type BetaCode struct {
	Code        string    `json:"code"`
	Description string    `json:"description"`
	MaxUses     int       `json:"max_uses"`
	CurrentUses int       `json:"current_uses"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedBy   string    `json:"created_by,omitempty"`
}

// BetaCodeManager manages beta codes
type BetaCodeManager struct {
	db *sql.DB
}

// NewBetaCodeManager creates a new beta code manager
func NewBetaCodeManager(db *sql.DB) *BetaCodeManager {
	return &BetaCodeManager{db: db}
}

// GenerateCode generates a random beta code
func GenerateCode() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreateBetaCode creates a new beta code
func (bcm *BetaCodeManager) CreateBetaCode(description string, maxUses int, expiresAt *time.Time, createdBy string) (*BetaCode, error) {
	code, err := GenerateCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate code: %w", err)
	}

	query := `
		INSERT INTO beta_codes (code, description, max_uses, current_uses, is_active, expires_at, created_by)
		VALUES (?, ?, ?, 0, 1, ?, ?)
	`

	_, err = bcm.db.Exec(query, code, description, maxUses, expiresAt, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create beta code: %w", err)
	}

	return &BetaCode{
		Code:        code,
		Description: description,
		MaxUses:     maxUses,
		CurrentUses: 0,
		IsActive:    true,
		CreatedAt:   time.Now(),
		ExpiresAt:   expiresAt,
		CreatedBy:   createdBy,
	}, nil
}

// ValidateBetaCode validates and uses a beta code
func (bcm *BetaCodeManager) ValidateBetaCode(code string) (bool, string) {
	// Get beta code
	var betaCode BetaCode
	var expiresAt sql.NullTime

	query := `
		SELECT code, description, max_uses, current_uses, is_active, created_at, expires_at, created_by
		FROM beta_codes
		WHERE code = ?
	`

	err := bcm.db.QueryRow(query, code).Scan(
		&betaCode.Code,
		&betaCode.Description,
		&betaCode.MaxUses,
		&betaCode.CurrentUses,
		&betaCode.IsActive,
		&betaCode.CreatedAt,
		&expiresAt,
		&betaCode.CreatedBy,
	)

	if err == sql.ErrNoRows {
		return false, "Invalid beta code"
	}
	if err != nil {
		return false, "Error validating beta code"
	}

	if expiresAt.Valid {
		betaCode.ExpiresAt = &expiresAt.Time
	}

	// Check if active
	if !betaCode.IsActive {
		return false, "Beta code has been deactivated"
	}

	// Check if expired
	if betaCode.ExpiresAt != nil && time.Now().After(*betaCode.ExpiresAt) {
		return false, "Beta code has expired"
	}

	// Check if max uses reached
	if betaCode.CurrentUses >= betaCode.MaxUses {
		return false, "Beta code has reached maximum uses"
	}

	return true, ""
}

// UseBetaCode increments the usage count of a beta code
func (bcm *BetaCodeManager) UseBetaCode(code string) error {
	query := `
		UPDATE beta_codes
		SET current_uses = current_uses + 1
		WHERE code = ?
	`

	result, err := bcm.db.Exec(query, code)
	if err != nil {
		return fmt.Errorf("failed to update beta code: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("beta code not found")
	}

	return nil
}

// GetBetaCode retrieves a beta code
func (bcm *BetaCodeManager) GetBetaCode(code string) (*BetaCode, error) {
	var betaCode BetaCode
	var expiresAt sql.NullTime
	var createdBy sql.NullString

	query := `
		SELECT code, description, max_uses, current_uses, is_active, created_at, expires_at, created_by
		FROM beta_codes
		WHERE code = ?
	`

	err := bcm.db.QueryRow(query, code).Scan(
		&betaCode.Code,
		&betaCode.Description,
		&betaCode.MaxUses,
		&betaCode.CurrentUses,
		&betaCode.IsActive,
		&betaCode.CreatedAt,
		&expiresAt,
		&createdBy,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("beta code not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get beta code: %w", err)
	}

	if expiresAt.Valid {
		betaCode.ExpiresAt = &expiresAt.Time
	}
	if createdBy.Valid {
		betaCode.CreatedBy = createdBy.String
	}

	return &betaCode, nil
}

// ListBetaCodes lists all beta codes (admin only)
func (bcm *BetaCodeManager) ListBetaCodes(activeOnly bool) ([]*BetaCode, error) {
	query := `
		SELECT code, description, max_uses, current_uses, is_active, created_at, expires_at, created_by
		FROM beta_codes
	`

	if activeOnly {
		query += ` WHERE is_active = 1`
	}

	query += ` ORDER BY created_at DESC`

	rows, err := bcm.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list beta codes: %w", err)
	}
	defer rows.Close()

	var codes []*BetaCode
	for rows.Next() {
		var code BetaCode
		var expiresAt sql.NullTime
		var createdBy sql.NullString

		err := rows.Scan(
			&code.Code,
			&code.Description,
			&code.MaxUses,
			&code.CurrentUses,
			&code.IsActive,
			&code.CreatedAt,
			&expiresAt,
			&createdBy,
		)
		if err != nil {
			continue
		}

		if expiresAt.Valid {
			code.ExpiresAt = &expiresAt.Time
		}
		if createdBy.Valid {
			code.CreatedBy = createdBy.String
		}

		codes = append(codes, &code)
	}

	return codes, nil
}

// DeactivateBetaCode deactivates a beta code
func (bcm *BetaCodeManager) DeactivateBetaCode(code string) error {
	query := `UPDATE beta_codes SET is_active = 0 WHERE code = ?`

	result, err := bcm.db.Exec(query, code)
	if err != nil {
		return fmt.Errorf("failed to deactivate beta code: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("beta code not found")
	}

	return nil
}

// ReactivateBetaCode reactivates a beta code
func (bcm *BetaCodeManager) ReactivateBetaCode(code string) error {
	query := `UPDATE beta_codes SET is_active = 1 WHERE code = ?`

	result, err := bcm.db.Exec(query, code)
	if err != nil {
		return fmt.Errorf("failed to reactivate beta code: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("beta code not found")
	}

	return nil
}

// DeleteBetaCode deletes a beta code (only if unused)
func (bcm *BetaCodeManager) DeleteBetaCode(code string) error {
	// Check if code has been used
	var currentUses int
	err := bcm.db.QueryRow(`SELECT current_uses FROM beta_codes WHERE code = ?`, code).Scan(&currentUses)
	if err == sql.ErrNoRows {
		return fmt.Errorf("beta code not found")
	}
	if err != nil {
		return fmt.Errorf("failed to check beta code: %w", err)
	}

	if currentUses > 0 {
		return fmt.Errorf("cannot delete beta code that has been used")
	}

	query := `DELETE FROM beta_codes WHERE code = ?`

	result, err := bcm.db.Exec(query, code)
	if err != nil {
		return fmt.Errorf("failed to delete beta code: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("beta code not found")
	}

	return nil
}

// GetBetaCodeStats returns statistics about beta codes
func (bcm *BetaCodeManager) GetBetaCodeStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total codes
	var totalCodes int
	err := bcm.db.QueryRow(`SELECT COUNT(*) FROM beta_codes`).Scan(&totalCodes)
	if err != nil {
		return nil, err
	}
	stats["total_codes"] = totalCodes

	// Active codes
	var activeCodes int
	err = bcm.db.QueryRow(`SELECT COUNT(*) FROM beta_codes WHERE is_active = 1`).Scan(&activeCodes)
	if err != nil {
		return nil, err
	}
	stats["active_codes"] = activeCodes

	// Total uses
	var totalUses int
	err = bcm.db.QueryRow(`SELECT COALESCE(SUM(current_uses), 0) FROM beta_codes`).Scan(&totalUses)
	if err != nil {
		return nil, err
	}
	stats["total_uses"] = totalUses

	// Available capacity
	var availableCapacity int
	err = bcm.db.QueryRow(`
		SELECT COALESCE(SUM(max_uses - current_uses), 0)
		FROM beta_codes
		WHERE is_active = 1 AND (expires_at IS NULL OR expires_at > datetime('now'))
	`).Scan(&availableCapacity)
	if err != nil {
		return nil, err
	}
	stats["available_capacity"] = availableCapacity

	return stats, nil
}
