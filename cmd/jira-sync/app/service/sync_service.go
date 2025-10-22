package service

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/joyfuldevs/project-lumos/cmd/jira-sync/app/adapter"
	"github.com/joyfuldevs/project-lumos/cmd/jira-sync/app/domain"
	"github.com/joyfuldevs/project-lumos/pkg/jira-sync/timestamp"
)

const (
	issuesFileName    = "issues.json"
	embeddingFileName = "embedding.json"
)

// SyncService handles the synchronization logic
type SyncService struct {
	jiraClient     *adapter.JiraClient
	embedder       *adapter.Embedder
	qdrantUploader *adapter.QdrantUploader
	timestampMgr   *timestamp.Manager
	config         *domain.SyncConfig
}

// NewSyncService creates a new sync service
func NewSyncService(config *domain.SyncConfig) *SyncService {
	return &SyncService{
		jiraClient:     adapter.NewJiraClient(config.JiraServer, config.JiraToken, config.JiraProjectKey),
		embedder:       adapter.NewEmbedder(config.EmbeddingAPIURL),
		qdrantUploader: adapter.NewQdrantUploader(config.QdrantHost, config.QdrantPort, config.CollectionName, config.BM42Collection),
		timestampMgr:   timestamp.NewManager(config.StateDir),
		config:         config,
	}
}

// RunFull performs a full synchronization
func (s *SyncService) RunFull() (*domain.SyncResult, error) {
	slog.Info("starting full synchronization")
	result := &domain.SyncResult{
		StartTime: time.Now(),
	}

	// Collect all issues
	slog.Info("collecting all Jira issues")
	issues, err := s.jiraClient.FetchIssues("")
	if err != nil {
		return nil, fmt.Errorf("collection failed: %w", err)
	}
	result.IssuesCollected = len(issues)

	if err := s.saveIssues(issues); err != nil {
		return nil, fmt.Errorf("failed to save issues: %w", err)
	}

	// Generate embeddings
	slog.Info("generating embeddings")
	if err := s.generateEmbeddings(); err != nil {
		return nil, fmt.Errorf("embedding generation failed: %w", err)
	}

	// Sync to Qdrant
	slog.Info("syncing to Qdrant")
	if err := s.syncToQdrant(); err != nil {
		return nil, fmt.Errorf("sync failed: %w", err)
	}
	result.IssuesProcessed = result.IssuesCollected

	// Save timestamp
	slog.Info("saving timestamp")
	if err := s.timestampMgr.SaveNow(); err != nil {
		return nil, fmt.Errorf("timestamp save failed: %w", err)
	}
	result.LastSync = s.timestampMgr.GetCurrentTimestamp()

	result.EndTime = time.Now()
	slog.Info("full synchronization completed successfully",
		slog.Int("issues", result.IssuesProcessed),
		slog.Duration("duration", result.EndTime.Sub(result.StartTime)))

	return result, nil
}

// RunIncremental performs an incremental synchronization
func (s *SyncService) RunIncremental() (*domain.SyncResult, error) {
	slog.Info("starting incremental synchronization")
	result := &domain.SyncResult{
		StartTime: time.Now(),
	}

	// Check last sync timestamp
	lastSync, err := s.timestampMgr.GetLastSync()
	if err != nil {
		return nil, fmt.Errorf("failed to get last sync time: %w", err)
	}

	if lastSync == "" {
		slog.Warn("no previous sync found, running full sync instead")
		return s.RunFull()
	}

	slog.Info("last sync found", slog.String("timestamp", lastSync))
	currentTime := s.timestampMgr.GetCurrentTimestamp()

	// Collect updated issues
	slog.Info("collecting updated issues")
	issues, err := s.jiraClient.FetchIssues(lastSync)
	if err != nil {
		return nil, fmt.Errorf("collection failed: %w", err)
	}
	result.IssuesCollected = len(issues)

	if result.IssuesCollected == 0 {
		slog.Info("no updated issues found, skipping sync")
		result.EndTime = time.Now()
		return result, nil
	}

	if err := s.saveIssues(issues); err != nil {
		return nil, fmt.Errorf("failed to save issues: %w", err)
	}

	// Generate embeddings
	slog.Info("generating embeddings")
	if err := s.generateEmbeddings(); err != nil {
		return nil, fmt.Errorf("embedding generation failed: %w", err)
	}

	// Upsert to Qdrant
	slog.Info("upserting to Qdrant")
	if err := s.syncToQdrant(); err != nil {
		return nil, fmt.Errorf("sync failed: %w", err)
	}
	result.IssuesProcessed = result.IssuesCollected

	// Update timestamp
	slog.Info("updating timestamp")
	if err := s.timestampMgr.SaveLastSync(currentTime); err != nil {
		return nil, fmt.Errorf("timestamp update failed: %w", err)
	}
	result.LastSync = currentTime

	result.EndTime = time.Now()
	slog.Info("incremental synchronization completed successfully",
		slog.Int("issues", result.IssuesProcessed),
		slog.Duration("duration", result.EndTime.Sub(result.StartTime)))

	return result, nil
}

func (s *SyncService) saveIssues(issues []domain.Issue) error {
	outputPath := filepath.Join(s.config.DataDir, issuesFileName)

	if err := os.MkdirAll(s.config.DataDir, 0755); err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			slog.Warn("failed to close file", slog.Any("error", err))
		}
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(issues); err != nil {
		return err
	}

	slog.Info("issues saved", slog.String("path", outputPath), slog.Int("count", len(issues)))
	return nil
}

func (s *SyncService) generateEmbeddings() error {
	inputPath := filepath.Join(s.config.DataDir, issuesFileName)
	outputPath := filepath.Join(s.config.DataDir, embeddingFileName)

	return s.embedder.GenerateEmbeddings(inputPath, outputPath)
}

func (s *SyncService) syncToQdrant() error {
	issuesPath := filepath.Join(s.config.DataDir, issuesFileName)
	embeddingPath := filepath.Join(s.config.DataDir, embeddingFileName)

	// Upsert dense vectors
	if err := s.qdrantUploader.UpsertDenseVectors(embeddingPath); err != nil {
		return err
	}

	// Upsert sparse vectors (BM42)
	if err := s.qdrantUploader.UpsertSparseVectors(issuesPath); err != nil {
		return err
	}

	return nil
}
