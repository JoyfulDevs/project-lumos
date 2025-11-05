package timestamp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	timestampFile = ".last_sync_timestamp"
	timeFormat    = "2006-01-02 15:04" // Jira JQL format
)

// Manager handles timestamp persistence for incremental sync
type Manager struct {
	stateDir string
}

// NewManager creates a new timestamp manager
func NewManager(stateDir string) *Manager {
	return &Manager{
		stateDir: stateDir,
	}
}

// GetLastSync returns the last sync timestamp, empty string if not exists
func (m *Manager) GetLastSync() (string, error) {
	path := filepath.Join(m.stateDir, timestampFile)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to read timestamp file: %w", err)
	}

	return strings.TrimSpace(string(data)), nil
}

// SaveLastSync saves the current timestamp
func (m *Manager) SaveLastSync(timestamp string) error {
	// Create state directory if not exists
	if err := os.MkdirAll(m.stateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	path := filepath.Join(m.stateDir, timestampFile)

	if err := os.WriteFile(path, []byte(timestamp+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to write timestamp file: %w", err)
	}

	return nil
}

// GetCurrentTimestamp returns current time in Jira format
func (m *Manager) GetCurrentTimestamp() string {
	return time.Now().UTC().Format(timeFormat)
}

// SaveNow saves the current timestamp
func (m *Manager) SaveNow() error {
	return m.SaveLastSync(m.GetCurrentTimestamp())
}
