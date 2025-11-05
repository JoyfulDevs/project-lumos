package domain

import "time"

// Issue represents a Jira issue
type Issue map[string]any

// Embedding represents an issue with its vector embedding
type Embedding struct {
	ID      string    `json:"id"`
	Vector  []float64 `json:"vector"`
	Payload any       `json:"payload"`
}

// SyncConfig holds configuration for synchronization
type SyncConfig struct {
	DataDir  string
	StateDir string

	// Jira configuration
	JiraToken      string
	JiraServer     string
	JiraProjectKey string

	// Embedding configuration
	EmbeddingAPIURL string

	// Qdrant configuration
	QdrantHost       string
	QdrantPort       string
	CollectionName   string
	BM42Collection   string
}

// SyncResult represents the result of a synchronization operation
type SyncResult struct {
	IssuesCollected int
	IssuesProcessed int
	StartTime       time.Time
	EndTime         time.Time
	LastSync        string
}
