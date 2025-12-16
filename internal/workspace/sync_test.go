package workspace

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyncCalculateSummary(t *testing.T) {
	tests := []struct {
		name     string
		results  []SyncResult
		expected SyncSummary
	}{
		{
			name:    "empty results",
			results: []SyncResult{},
			expected: SyncSummary{
				Total:   0,
				Synced:  0,
				Updated: 0,
				Failed:  0,
				Skipped: 0,
				Pushed:  0,
			},
		},
		{
			name: "all synced",
			results: []SyncResult{
				{RepoName: "repo1", Status: "synced"},
				{RepoName: "repo2", Status: "up-to-date"},
			},
			expected: SyncSummary{
				Total:   2,
				Synced:  2,
				Updated: 0,
				Failed:  0,
				Skipped: 0,
				Pushed:  0,
			},
		},
		{
			name: "mixed statuses",
			results: []SyncResult{
				{RepoName: "repo1", Status: "synced"},
				{RepoName: "repo2", Status: "updated"},
				{RepoName: "repo3", Status: "failed"},
				{RepoName: "repo4", Status: "skipped"},
				{RepoName: "repo5", Status: "pushed"},
			},
			expected: SyncSummary{
				Total:   5,
				Synced:  1,
				Updated: 1,
				Failed:  1,
				Skipped: 1,
				Pushed:  1,
			},
		},
		{
			name: "multiple failures",
			results: []SyncResult{
				{RepoName: "repo1", Status: "failed"},
				{RepoName: "repo2", Status: "failed"},
				{RepoName: "repo3", Status: "synced"},
			},
			expected: SyncSummary{
				Total:   3,
				Synced:  1,
				Updated: 0,
				Failed:  2,
				Skipped: 0,
				Pushed:  0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := CalculateSummary(tt.results)
			assert.Equal(t, tt.expected, summary)
		})
	}
}

func TestFormatSyncResults(t *testing.T) {
	tests := []struct {
		name      string
		results   []SyncResult
		operation string
		contains  []string
	}{
		{
			name:      "empty results",
			results:   []SyncResult{},
			operation: "sync",
			contains:  []string{},
		},
		{
			name: "successful sync",
			results: []SyncResult{
				{RepoName: "repo1", Status: "synced"},
				{RepoName: "repo2", Status: "up-to-date"},
			},
			operation: "sync",
			contains:  []string{"✓ repo1: synced", "✓ repo2: up-to-date"},
		},
		{
			name: "failed sync",
			results: []SyncResult{
				{RepoName: "repo1", Status: "failed", Error: errors.New("connection timeout")},
			},
			operation: "sync",
			contains:  []string{"✗ repo1: failed", "connection timeout"},
		},
		{
			name: "skipped sync",
			results: []SyncResult{
				{RepoName: "repo1", Status: "skipped"},
			},
			operation: "sync",
			contains:  []string{"⊘ repo1: skipped"},
		},
		{
			name: "mixed results",
			results: []SyncResult{
				{RepoName: "repo1", Status: "synced"},
				{RepoName: "repo2", Status: "failed", Error: errors.New("error")},
				{RepoName: "repo3", Status: "skipped"},
				{RepoName: "repo4", Status: "pushed"},
			},
			operation: "sync",
			contains: []string{
				"✓ repo1: synced",
				"✗ repo2: failed",
				"⊘ repo3: skipped",
				"✓ repo4: pushed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := FormatSyncResults(tt.results, tt.operation)
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestFormatSyncResults_MultilineOutput(t *testing.T) {
	results := []SyncResult{
		{RepoName: "repo1", Status: "synced"},
		{RepoName: "repo2", Status: "updated"},
	}

	output := FormatSyncResults(results, "sync")
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Len(t, lines, 2, "Should have one line per result")
}
