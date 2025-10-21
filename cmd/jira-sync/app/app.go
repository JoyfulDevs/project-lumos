package app

import (
	"errors"
	"os"

	"github.com/joyfuldevs/project-lumos/cmd/jira-sync/app/domain"
	"github.com/joyfuldevs/project-lumos/cmd/jira-sync/app/service"
)

// RunFull executes a full synchronization
func RunFull(dataDir, stateDir string) error {
	config, err := loadConfig(dataDir, stateDir)
	if err != nil {
		return err
	}

	svc := service.NewSyncService(config)
	_, err = svc.RunFull()
	return err
}

// RunIncremental executes an incremental synchronization
func RunIncremental(dataDir, stateDir string) error {
	config, err := loadConfig(dataDir, stateDir)
	if err != nil {
		return err
	}

	svc := service.NewSyncService(config)
	_, err = svc.RunIncremental()
	return err
}

func loadConfig(dataDir, stateDir string) (*domain.SyncConfig, error) {
	// Jira configuration
	jiraToken, ok := os.LookupEnv("JIRA_API_TOKEN")
	if !ok {
		return nil, errors.New("JIRA_API_TOKEN is not set")
	}

	jiraServer, ok := os.LookupEnv("JIRA_SERVER")
	if !ok {
		return nil, errors.New("JIRA_SERVER is not set")
	}

	jiraProject, ok := os.LookupEnv("JIRA_PROJECT_KEY")
	if !ok {
		return nil, errors.New("JIRA_PROJECT_KEY is not set")
	}

	// Embedding configuration
	embeddingURL, ok := os.LookupEnv("EMBEDDING_API_URL")
	if !ok {
		return nil, errors.New("EMBEDDING_API_URL is not set")
	}

	// Qdrant configuration
	qdrantHost, ok := os.LookupEnv("QDRANT_HOST")
	if !ok {
		return nil, errors.New("QDRANT_HOST is not set")
	}

	qdrantPort, ok := os.LookupEnv("QDRANT_PORT")
	if !ok {
		return nil, errors.New("QDRANT_PORT is not set")
	}

	collectionName := os.Getenv("COLLECTION_NAME")
	if collectionName == "" {
		collectionName = "jira_issues"
	}

	bm42Collection := os.Getenv("BM42_COLLECTION")
	if bm42Collection == "" {
		bm42Collection = "jira_bm42_full"
	}

	return &domain.SyncConfig{
		DataDir:          dataDir,
		StateDir:         stateDir,
		JiraToken:        jiraToken,
		JiraServer:       jiraServer,
		JiraProjectKey:   jiraProject,
		EmbeddingAPIURL:  embeddingURL,
		QdrantHost:       qdrantHost,
		QdrantPort:       qdrantPort,
		CollectionName:   collectionName,
		BM42Collection:   bm42Collection,
	}, nil
}
