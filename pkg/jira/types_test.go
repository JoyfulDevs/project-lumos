package jira_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/joyfuldevs/project-lumos/pkg/jira"
)

const (
	sample_json_1 = `{ "fields": { "assignee": { "displayName": "Display Name", "emailAddress": "email@example.com", "name": "User Name" }, "attachment": [], "comment": { "comments": [ { "author": { "displayName": "Display Name", "emailAddress": "email@example.com", "name": "User Name" }, "body": "Comment Body", "created": "2025-07-25T13:57:57.000+0900", "id": "1044263", "updated": "2025-07-25T13:57:57.000+0900" } ], "maxResults": 1, "startAt": 0, "total": 1 }, "created": "2025-07-25T11:20:55.000+0900", "creator": { "displayName": "Display Name", "emailAddress": "email@example.com", "name": "User Name" }, "description": "Issue Description", "issuelinks": [], "labels": [], "status": { "name": "Resolved" }, "subtasks": [], "summary": "Issue Summary = Title", "updated": "2025-07-25T13:59:32.000+0900" }, "id": "185730", "key": "AA-12345" }`
	sample_json_2 = `{ "fields": { "assignee": { "displayName": "Display Name", "emailAddress": "email@example.com", "name": "User Name" }, "comment": { "comments": [ { "author": { "displayName": "Display Name", "emailAddress": "email@example.com", "name": "User Name" }, "body": "Comment Body", "created": "2025-07-24T14:06:47.000+0900", "id": "1043845", "updated": "2025-07-24T14:06:47.000+0900" }, { "author": { "displayName": "Display Name", "emailAddress": "email@example.com", "name": "User Name" }, "body": "Comment Body", "created": "2025-07-24T15:42:32.000+0900", "id": "1043914", "updated": "2025-07-24T15:42:32.000+0900" }, { "author": { "displayName": "Display Name", "emailAddress": "email@example.com", "name": "User Name" }, "body": "Comment Body", "created": "2025-07-24T15:59:49.000+0900", "id": "1043933", "updated": "2025-07-24T15:59:49.000+0900" }, { "author": { "displayName": "Display Name", "emailAddress": "email@example.com", "name": "User Name" }, "body": "Comment Body", "created": "2025-07-24T16:48:26.000+0900", "id": "1043978", "updated": "2025-07-24T16:48:26.000+0900" }, { "author": { "displayName": "Display Name", "emailAddress": "email@example.com", "name": "User Name" }, "body": "Comment Body", "created": "2025-07-24T18:22:13.000+0900", "id": "1044028", "updated": "2025-07-24T18:22:13.000+0900" } ], "maxResults": 5, "startAt": 0, "total": 5 }, "created": "2025-07-24T13:02:16.000+0900", "creator": { "displayName": "Display Name", "emailAddress": "email@example.com", "name": "User Name" }, "description": "Issue Description", "issuelinks": [], "labels": [ "SomeLabel" ], "status": { "name": "In Progress" }, "subtasks": [], "summary": "Issue Summary = Title", "updated": "2025-07-24T18:26:17.000+0900" }, "id": "185660", "key": "AA-23456" }`
)

func TestUnmarshalIssue(t *testing.T) {
	testCases := []struct {
		desc  string
		value string
	}{
		{
			desc:  "sample_json_1",
			value: sample_json_1,
		},
		{
			desc:  "sample_json_2",
			value: sample_json_2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			issue := jira.Issue{}
			if err := json.Unmarshal([]byte(tc.value), &issue); err != nil {
				t.Fatalf("failed to unmarshal issue: %v", err)
			}
			if issue.ID == "" {
				t.Error("expected issue ID to be set, got empty string")
			}
			if issue.Key == "" {
				t.Error("expected issue Key to be set, got empty string")
			}
			if issue.Fields.Title == "" {
				t.Error("expected issue Fields Title to be set, got empty string")
			}
			if issue.Fields.Content == "" {
				t.Error("expected issue Fields Content to be set, got empty string")
			}
			if issue.Fields.Creator.ID == "" {
				t.Error("expected issue Fields Creator ID to be set, got empty string")
			}
			if issue.Fields.Creator.Name == "" {
				t.Error("expected issue Fields Creator Name to be set, got empty string")
			}
			if issue.Fields.Creator.EmailAddress == "" {
				t.Error("expected issue Fields Creator EmailAddress to be set, got empty string")
			}
			if issue.Fields.Status.Name == "" {
				t.Error("expected issue Fields Status Name to be set, got empty string")
			}
			if issue.Fields.Created == "" {
				t.Error("expected issue Fields Created to be set, got empty string")
			}
			if _, err := time.Parse(jira.TimeFormat, issue.Fields.Created); err != nil {
				t.Errorf("expected issue Fields Created to be a valid time, got %s: %v", issue.Fields.Created, err)
			}
			for _, comment := range issue.Fields.CommentInfo.Comments {
				if comment.ID == "" {
					t.Error("expected comment ID to be set, got empty string")
				}
				if comment.Author.ID == "" {
					t.Error("expected comment Author ID to be set, got empty string")
				}
				if comment.Author.Name == "" {
					t.Error("expected comment Author Name to be set, got empty string")
				}
				if comment.Author.EmailAddress == "" {
					t.Error("expected comment Author EmailAddress to be set, got empty string")
				}
				if comment.Body == "" {
					t.Error("expected comment Body to be set, got empty string")
				}
				if comment.Created == "" {
					t.Error("expected comment Created to be set, got empty string")
				}
				if _, err := time.Parse(jira.TimeFormat, comment.Created); err != nil {
					t.Errorf("expected comment Created to be a valid time, got %s: %v", comment.Created, err)
				}
				if comment.Updated == "" {
					t.Error("expected comment Updated to be set, got empty string")
				}
				if _, err := time.Parse(jira.TimeFormat, comment.Updated); err != nil {
					t.Errorf("expected comment Updated to be a valid time, got %s: %v", comment.Updated, err)
				}
			}
		})
	}
}
