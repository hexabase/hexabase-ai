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

func TestOAuthRepository_getGithubVerifiedPrimaryEmail(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	repo := NewOAuthRepository(map[string]*ProviderConfig{}, logger)

	client := &http.Client{}
	httpmock.ActivateNonDefault(client)
	defer httpmock.DeactivateAndReset()

	t.Run("should return verified primary email", func(t *testing.T) {
		// Mock GitHub API
		emails := []map[string]interface{}{
			{"email": "secondary@example.com", "primary": false, "verified": true},
			{"email": "primary@example.com", "primary": true, "verified": true},
		}
		httpmock.RegisterResponder("GET", "https://api.github.com/user/emails",
			httpmock.NewJsonResponderOrPanic(200, emails))

		// Execute
		email, err := repo.(*oauthRepository).getGithubVerifiedPrimaryEmail(context.Background(), client)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "primary@example.com", email)
	})

	t.Run("should return empty string if no primary email exists", func(t *testing.T) {
		// Mock GitHub API
		emails := []map[string]interface{}{
			{"email": "secondary@example.com", "primary": false, "verified": true},
			{"email": "unverified@example.com", "primary": true, "verified": false},
		}
		httpmock.RegisterResponder("GET", "https://api.github.com/user/emails",
			httpmock.NewJsonResponderOrPanic(200, emails))

		// Execute
		email, err := repo.(*oauthRepository).getGithubVerifiedPrimaryEmail(context.Background(), client)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "", email)
	})

	t.Run("should return error on API failure", func(t *testing.T) {
		// Mock GitHub API
		httpmock.RegisterResponder("GET", "https://api.github.com/user/emails",
			httpmock.NewStringResponder(500, "Internal Server Error"))

		// Execute
		email, err := repo.(*oauthRepository).getGithubVerifiedPrimaryEmail(context.Background(), client)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, email)
	})
}

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