package adapter

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/joyfuldevs/project-lumos/cmd/jira-sync/app/domain"
)

const (
	maxResults = 100
	timeout    = 300 * time.Second
	maxRetries = 3
)

type jiraResponse struct {
	Issues []domain.Issue `json:"issues"`
	Total  int            `json:"total"`
}

// JiraClient handles communication with Jira API
type JiraClient struct {
	baseURL string
	token   string
	project string
	client  *http.Client
}

// NewJiraClient creates a new Jira client
func NewJiraClient(baseURL, token, project string) *JiraClient {
	return &JiraClient{
		baseURL: baseURL,
		token:   token,
		project: project,
		client:  &http.Client{Timeout: timeout},
	}
}

// FetchIssues fetches issues from Jira, optionally filtering by update time
func (j *JiraClient) FetchIssues(since string) ([]domain.Issue, error) {
	jql := fmt.Sprintf("project=%s", j.project)
	if since != "" {
		slog.Info("fetching issues updated since", slog.String("since", since))
		jql = fmt.Sprintf("%s AND updated >= \"%s\"", jql, since)
	} else {
		slog.Info("fetching all issues")
	}
	jql = fmt.Sprintf("%s ORDER BY updated DESC", jql)

	var allIssues []domain.Issue
	startAt := 0

	for {
		issues, err := j.fetchPage(jql, startAt)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch page at offset %d: %w", startAt, err)
		}

		allIssues = append(allIssues, issues...)
		slog.Info("collected issues",
			slog.Int("count", len(issues)),
			slog.Int("total", len(allIssues)))

		if len(issues) < maxResults {
			break
		}

		startAt += maxResults
	}

	return allIssues, nil
}

func (j *JiraClient) fetchPage(jql string, startAt int) ([]domain.Issue, error) {
	u, err := url.Parse(j.baseURL)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("jql", jql)
	q.Set("startAt", strconv.Itoa(startAt))
	q.Set("maxResults", strconv.Itoa(maxResults))
	q.Set("fields", "*all")
	u.RawQuery = q.Encode()

	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "Bearer "+j.token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := j.client.Do(req)
		if err != nil {
			slog.Warn("request failed",
				slog.Int("attempt", attempt),
				slog.Any("error", err))
			if attempt == maxRetries {
				return nil, err
			}
			time.Sleep(3 * time.Second)
			continue
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				slog.Warn("failed to close response body", slog.Any("error", err))
			}
		}()

		if resp.StatusCode != http.StatusOK {
			slog.Warn("HTTP error",
				slog.Int("attempt", attempt),
				slog.Int("status", resp.StatusCode))
			if attempt == maxRetries {
				return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
			}
			time.Sleep(3 * time.Second)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.Warn("failed to read response",
				slog.Int("attempt", attempt),
				slog.Any("error", err))
			if attempt == maxRetries {
				return nil, err
			}
			time.Sleep(3 * time.Second)
			continue
		}

		var jiraResp jiraResponse
		if err := json.Unmarshal(body, &jiraResp); err != nil {
			slog.Warn("JSON parse error",
				slog.Int("attempt", attempt),
				slog.Any("error", err))
			if attempt == maxRetries {
				return nil, err
			}
			time.Sleep(3 * time.Second)
			continue
		}

		return jiraResp.Issues, nil
	}

	return nil, fmt.Errorf("max retries exceeded")
}
