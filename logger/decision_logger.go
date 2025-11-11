package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// DecisionLogger handles logging of trading decisions
type DecisionLogger struct {
	baseDir string
	mu      sync.Mutex
}

// NewDecisionLogger creates a new decision logger
func NewDecisionLogger(baseDir string) *DecisionLogger {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		fmt.Printf("Warning: failed to create decision logs directory: %v\n", err)
	}

	return &DecisionLogger{
		baseDir: baseDir,
	}
}

// LogDecision logs a decision record to a JSON file
func (dl *DecisionLogger) LogDecision(traderID string, record interface{}) error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	// Create trader-specific directory
	traderDir := filepath.Join(dl.baseDir, traderID)
	if err := os.MkdirAll(traderDir, 0755); err != nil {
		return fmt.Errorf("failed to create trader directory: %w", err)
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("decision_%s.json", timestamp)
	filePath := filepath.Join(traderDir, filename)

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal decision: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write decision file: %w", err)
	}

	return nil
}

// GetLatestDecisions retrieves the latest N decisions for a trader
func (dl *DecisionLogger) GetLatestDecisions(traderID string, count int) ([]map[string]interface{}, error) {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	traderDir := filepath.Join(dl.baseDir, traderID)

	// Check if directory exists
	if _, err := os.Stat(traderDir); os.IsNotExist(err) {
		return []map[string]interface{}{}, nil
	}

	// Read all decision files
	files, err := os.ReadDir(traderDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read trader directory: %w", err)
	}

	// Filter JSON files and sort by name (which includes timestamp)
	var jsonFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			jsonFiles = append(jsonFiles, file.Name())
		}
	}

	// Sort in descending order (newest first)
	sort.Sort(sort.Reverse(sort.StringSlice(jsonFiles)))

	// Limit to requested count
	if len(jsonFiles) > count {
		jsonFiles = jsonFiles[:count]
	}

	// Read and parse each file
	decisions := []map[string]interface{}{}
	for _, filename := range jsonFiles {
		filePath := filepath.Join(traderDir, filename)

		data, err := os.ReadFile(filePath)
		if err != nil {
			continue // Skip files that can't be read
		}

		var decision map[string]interface{}
		if err := json.Unmarshal(data, &decision); err != nil {
			continue // Skip files that can't be parsed
		}

		decisions = append(decisions, decision)
	}

	return decisions, nil
}

// GetDecisionByDate retrieves decisions for a specific date
func (dl *DecisionLogger) GetDecisionByDate(traderID string, date time.Time) ([]map[string]interface{}, error) {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	traderDir := filepath.Join(dl.baseDir, traderID)

	// Check if directory exists
	if _, err := os.Stat(traderDir); os.IsNotExist(err) {
		return []map[string]interface{}{}, nil
	}

	// Format date prefix for filtering
	datePrefix := "decision_" + date.Format("20060102")

	// Read all decision files
	files, err := os.ReadDir(traderDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read trader directory: %w", err)
	}

	// Filter files by date
	decisions := []map[string]interface{}{}
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), datePrefix) {
			filePath := filepath.Join(traderDir, file.Name())

			data, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}

			var decision map[string]interface{}
			if err := json.Unmarshal(data, &decision); err != nil {
				continue
			}

			decisions = append(decisions, decision)
		}
	}

	// Sort by filename (chronological)
	sort.Slice(decisions, func(i, j int) bool {
		// Extract timestamps from decisions if available
		ti, _ := time.Parse(time.RFC3339, getStringValue(decisions[i], "timestamp"))
		tj, _ := time.Parse(time.RFC3339, getStringValue(decisions[j], "timestamp"))
		return ti.Before(tj)
	})

	return decisions, nil
}

// DeleteOldDecisions deletes decision logs older than the specified days
func (dl *DecisionLogger) DeleteOldDecisions(traderID string, olderThanDays int) (int, error) {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	traderDir := filepath.Join(dl.baseDir, traderID)

	// Check if directory exists
	if _, err := os.Stat(traderDir); os.IsNotExist(err) {
		return 0, nil
	}

	cutoffTime := time.Now().AddDate(0, 0, -olderThanDays)
	deletedCount := 0

	// Read all decision files
	files, err := os.ReadDir(traderDir)
	if err != nil {
		return 0, fmt.Errorf("failed to read trader directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(traderDir, file.Name())

		// Get file info
		info, err := os.Stat(filePath)
		if err != nil {
			continue
		}

		// Delete if older than cutoff
		if info.ModTime().Before(cutoffTime) {
			if err := os.Remove(filePath); err != nil {
				fmt.Printf("Warning: failed to delete old decision file %s: %v\n", filePath, err)
			} else {
				deletedCount++
			}
		}
	}

	return deletedCount, nil
}

// GetDecisionCount returns the total number of decisions for a trader
func (dl *DecisionLogger) GetDecisionCount(traderID string) (int, error) {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	traderDir := filepath.Join(dl.baseDir, traderID)

	// Check if directory exists
	if _, err := os.Stat(traderDir); os.IsNotExist(err) {
		return 0, nil
	}

	// Count JSON files
	files, err := os.ReadDir(traderDir)
	if err != nil {
		return 0, fmt.Errorf("failed to read trader directory: %w", err)
	}

	count := 0
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			count++
		}
	}

	return count, nil
}

// CleanupTraderLogs removes all logs for a trader
func (dl *DecisionLogger) CleanupTraderLogs(traderID string) error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	traderDir := filepath.Join(dl.baseDir, traderID)

	// Check if directory exists
	if _, err := os.Stat(traderDir); os.IsNotExist(err) {
		return nil // Nothing to clean
	}

	// Remove entire trader directory
	if err := os.RemoveAll(traderDir); err != nil {
		return fmt.Errorf("failed to cleanup trader logs: %w", err)
	}

	return nil
}

// getStringValue safely extracts string value from map
func getStringValue(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
