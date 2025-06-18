package wire

import (
	"reflect"
	"testing"

	authRepo "github.com/hexabase/hexabase-ai/api/internal/repository/auth"
	"github.com/hexabase/hexabase-ai/api/internal/shared/config"
)

func TestProvideOAuthProviderConfigs(t *testing.T) {
	testCases := []struct {
		name     string
		cfg      *config.Config
		expected map[string]*authRepo.ProviderConfig
	}{
		{
			name: "should create provider configs from config with one provider",
			cfg: &config.Config{
				Auth: config.AuthConfig{
					ExternalProviders: map[string]config.OAuthProvider{
						"github": {
							ClientID:     "test_client_id",
							ClientSecret: "test_client_secret",
							RedirectURL:  "test_redirect_url",
							Scopes:       []string{"email", "profile"},
							AuthURL:      "https://github.com/login/oauth/authorize",
							TokenURL:     "https://github.com/login/oauth/access_token",
						},
					},
				},
			},
			expected: map[string]*authRepo.ProviderConfig{
				"github": {
					ClientID:     "test_client_id",
					ClientSecret: "test_client_secret",
					RedirectURL:  "test_redirect_url",
					Scopes:       []string{"email", "profile"},
					AuthURL:      "https://github.com/login/oauth/authorize",
					TokenURL:     "https://github.com/login/oauth/access_token",
				},
			},
		},
		{
			name: "should return empty map if no providers in config",
			cfg: &config.Config{
				Auth: config.AuthConfig{
					ExternalProviders: map[string]config.OAuthProvider{},
				},
			},
			expected: map[string]*authRepo.ProviderConfig{},
		},
		{
			name: "should handle multiple providers",
			cfg: &config.Config{
				Auth: config.AuthConfig{
					ExternalProviders: map[string]config.OAuthProvider{
						"github": {
							ClientID:     "github_id",
							ClientSecret: "github_secret",
							RedirectURL:  "github_redirect",
							Scopes:       []string{"email"},
							AuthURL:      "https://github.com/login/oauth/authorize",
							TokenURL:     "https://github.com/login/oauth/access_token",
						},
						"google": {
							ClientID:     "google_id",
							ClientSecret: "google_secret",
							RedirectURL:  "google_redirect",
							Scopes:       []string{"openid", "profile"},
							AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
							TokenURL:     "https://oauth2.googleapis.com/token",
						},
					},
				},
			},
			expected: map[string]*authRepo.ProviderConfig{
				"github": {
					ClientID:     "github_id",
					ClientSecret: "github_secret",
					RedirectURL:  "github_redirect",
					Scopes:       []string{"email"},
					AuthURL:      "https://github.com/login/oauth/authorize",
					TokenURL:     "https://github.com/login/oauth/access_token",
				},
				"google": {
					ClientID:     "google_id",
					ClientSecret: "google_secret",
					RedirectURL:  "google_redirect",
					Scopes:       []string{"openid", "profile"},
					AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
					TokenURL:     "https://oauth2.googleapis.com/token",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := ProvideOAuthProviderConfigs(tc.cfg)
			if len(actual) == 0 && len(tc.expected) == 0 {
				return
			}
			if !reflect.DeepEqual(tc.expected, actual) {
				t.Errorf("expected: %+v, actual: %+v", tc.expected, actual)
			}
		})
	}
} 