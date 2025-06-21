package repository

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/hexabase/hexabase-ai/api/internal/auth/domain"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestOAuthRepository_getGithubUserInfo(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	repo := NewOAuthRepository(map[string]*ProviderConfig{}, logger)

	client := &http.Client{}
	httpmock.ActivateNonDefault(client)
	defer httpmock.DeactivateAndReset()

	t.Run("should get GitHub user info successfully", func(t *testing.T) {
		// Mock GitHub User API
		githubUser := map[string]interface{}{
			"id":         12345,
			"login":      "testuser",
			"avatar_url": "http://example.com/avatar.png",
		}
		httpmock.RegisterResponder("GET", "https://api.github.com/user",
			httpmock.NewJsonResponderOrPanic(200, githubUser))

		// Execute
		userInfo, err := repo.(*oauthRepository).getGithubUserInfo(context.Background(), client)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, userInfo)
		expected := &domain.UserInfo{
			ID:       "12345",
			Name:     "testuser",
			Picture:  "http://example.com/avatar.png",
			Provider: "github",
		}
		assert.Equal(t, expected, userInfo)
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