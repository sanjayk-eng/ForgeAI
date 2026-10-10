package github

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestGitHubAPIErrorResponse(t *testing.T) {
	tests := []struct {
		name       string
		operation  string
		upstream   int
		wantStatus int
		wantPart   string
	}{
		{
			name:       "repository permission denied",
			operation:  "repository creation",
			upstream:   http.StatusForbidden,
			wantStatus: http.StatusForbidden,
			wantPart:   "Repository creation permission",
		},
		{
			name:       "repository name rejected",
			operation:  "repository creation",
			upstream:   http.StatusUnprocessableEntity,
			wantStatus: http.StatusConflict,
			wantPart:   "name or visibility",
		},
		{
			name:       "installation token credentials rejected",
			operation:  "installation token request",
			upstream:   http.StatusUnauthorized,
			wantStatus: http.StatusBadGateway,
			wantPart:   "app credentials",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := fmt.Errorf("create export: %w", &githubAPIError{
				operation: test.operation,
				status:    test.upstream,
			})
			var githubErr *githubAPIError
			if !errors.As(err, &githubErr) {
				t.Fatal("wrapped GitHub API error was not detected")
			}
			status, message := githubAPIErrorResponse(githubErr)
			if status != test.wantStatus {
				t.Fatalf("status = %d, want %d", status, test.wantStatus)
			}
			if !strings.Contains(message, test.wantPart) {
				t.Fatalf("message = %q, want it to contain %q", message, test.wantPart)
			}
		})
	}
}
