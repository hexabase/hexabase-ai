package auth

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestOAuthRepository_getGithubUserInfo(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	repo := NewOAuthRepository(map[string]*ProviderConfig{}, logger)

	client := &http.Client{}
	httpmock.ActivateNonDefault(client)
	defer httpmock.DeactivateAndReset()

	t.Run("should log warning when no verified primary email is found", func(t *testing.T) {
		buf.Reset()
		// Mock GitHub User API
		githubUser := map[string]interface{}{
			"id":         12345,
			"login":      "testuser",
			"avatar_url": "http://example.com/avatar.png",
		}
		httpmock.RegisterResponder("GET", "https://api.github.com/user",
			httpmock.NewJsonResponderOrPanic(200, githubUser))

		// Mock GitHub Emails API
		emails := []map[string]interface{}{
			{"email": "secondary@example.com", "primary": false, "verified": true},
		}
		httpmock.RegisterResponder("GET", "https://api.github.com/user/emails",
			httpmock.NewJsonResponderOrPanic(200, emails))

		// Execute
		userInfo, err := repo.(*oauthRepository).getGithubUserInfo(context.Background(), client)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, userInfo)
		assert.Equal(t, "", userInfo.Email)
		assert.Contains(t, buf.String(), "level=WARN msg=\"github verified primary email not found\"")
	})
}

func TestOAuthRepository_getGithubVerifiedPrimaryEmail(t *testing.T) {
	// t.Parallel()

	tests := []struct {
		name          string
		responder     httpmock.Responder
		wantErr       bool
		expectedEmail string
	}{
		{
			name: "Success: should return the primary and verified email",
			responder: httpmock.NewJsonResponderOrPanic(200, []map[string]interface{}{
				{"email": "secondary@example.com", "primary": false, "verified": true},
				{"email": "primary@example.com", "primary": true, "verified": true},
			}),
			wantErr:       false,
			expectedEmail: "primary@example.com",
		},
		{
			name: "Success: should return an empty string when no primary and verified email exists",
			responder: httpmock.NewJsonResponderOrPanic(200, []map[string]interface{}{
				{"email": "secondary@example.com", "primary": false, "verified": true},
				{"email": "unverified@example.com", "primary": true, "verified": false},
			}),
			wantErr:       false,
			expectedEmail: "",
		},
		{
			name:          "Success: should return an empty string for an empty email list",
			responder:     httpmock.NewJsonResponderOrPanic(200, []map[string]interface{}{}),
			wantErr:       false,
			expectedEmail: "",
		},
		{
			name:          "Error: should return an error when the API returns a 500 status",
			responder:     httpmock.NewStringResponder(500, "Internal Server Error"),
			wantErr:       true,
			expectedEmail: "",
		},
		{
			name:          "Error: should return an error for invalid JSON response",
			responder:     httpmock.NewStringResponder(200, "not a json"),
			wantErr:       true,
			expectedEmail: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()

			logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
			repo := NewOAuthRepository(map[string]*ProviderConfig{}, logger)
			client := &http.Client{}
			httpmock.ActivateNonDefault(client)
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder("GET", "https://api.github.com/user/emails", tt.responder)

			email, err := repo.(*oauthRepository).getGithubVerifiedPrimaryEmail(context.Background(), client)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedEmail, email)
		})
	}
}