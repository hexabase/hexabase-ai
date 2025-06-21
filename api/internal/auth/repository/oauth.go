package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/hexabase/hexabase-ai/api/internal/auth/domain"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type oauthRepository struct {
	providers map[string]*oauth2.Config
	logger    *slog.Logger
}

// NewOAuthRepository creates a new OAuth repository
func NewOAuthRepository(configs map[string]*ProviderConfig, logger *slog.Logger) domain.OAuthRepository {
	providers := make(map[string]*oauth2.Config)

	for provider, config := range configs {
		var endpoint oauth2.Endpoint

		switch provider {
		case "google":
			endpoint = google.Endpoint
		default:
			endpoint = oauth2.Endpoint{
				AuthURL:  config.AuthURL,
				TokenURL: config.TokenURL,
			}
		}

		providers[provider] = &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			RedirectURL:  config.RedirectURL,
			Scopes:       config.Scopes,
			Endpoint:     endpoint,
		}
	}

	return &oauthRepository{
		providers: providers,
		logger:    logger,
	}
}

func (r *oauthRepository) GetProviderConfig(provider string) (*domain.ProviderConfig, error) {
	config, ok := r.providers[provider]
	if !ok {
		return nil, fmt.Errorf("provider %s not configured", provider)
	}

	return &domain.ProviderConfig{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       config.Scopes,
		AuthURL:      config.Endpoint.AuthURL,
		TokenURL:     config.Endpoint.TokenURL,
	}, nil
}

func (r *oauthRepository) GetAuthURL(provider, state string, params map[string]string) (string, error) {
	config, ok := r.providers[provider]
	if !ok {
		return "", fmt.Errorf("provider %s not configured", provider)
	}

	// Build auth URL with additional parameters
	options := []oauth2.AuthCodeOption{}
	for key, value := range params {
		options = append(options, oauth2.SetAuthURLParam(key, value))
	}

	authURL := config.AuthCodeURL(state, options...)
	return authURL, nil
}

func (r *oauthRepository) ExchangeCode(ctx context.Context, provider, code string) (*domain.OAuthToken, error) {
	config, ok := r.providers[provider]
	if !ok {
		return nil, fmt.Errorf("provider %s not configured", provider)
	}

	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	return &domain.OAuthToken{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresIn:    int(token.Expiry.Sub(token.Expiry).Seconds()),
	}, nil
}

func (r *oauthRepository) GetUserInfo(ctx context.Context, provider string, token *domain.OAuthToken) (*domain.UserInfo, error) {
	config, ok := r.providers[provider]
	if !ok {
		return nil, fmt.Errorf("provider %s not configured", provider)
	}

	// Create HTTP client with token
	oauth2Token := &oauth2.Token{
		AccessToken: token.AccessToken,
		TokenType:   token.TokenType,
	}
	client := config.Client(ctx, oauth2Token)

	// Get user info based on provider
	switch provider {
	case "google":
		return r.getGoogleUserInfo(ctx, client)
	case "github":
		u, err := r.getGithubUserInfo(ctx, client)
		if err != nil {
			return nil, err
		}
		u.Email, err = r.getGithubVerifiedPrimaryEmail(ctx, client)
		if err != nil {
			r.logger.Error("failed to get github verified primary email", "error", err)
			return nil, err
		}
		if u.Email == "" {
			r.logger.Warn("github verified primary email is empty")
		}
		return u, nil
	default:
		return nil, fmt.Errorf("user info not implemented for provider %s", provider)
	}
}

func (r *oauthRepository) RefreshOAuthToken(ctx context.Context, provider string, refreshToken string) (*domain.OAuthToken, error) {
	config, ok := r.providers[provider]
	if !ok {
		return nil, fmt.Errorf("provider %s not configured", provider)
	}

	// Create token source with refresh token
	tokenSource := config.TokenSource(ctx, &oauth2.Token{
		RefreshToken: refreshToken,
	})

	// Get new token
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return &domain.OAuthToken{
		AccessToken:  newToken.AccessToken,
		RefreshToken: newToken.RefreshToken,
		TokenType:    newToken.TokenType,
		ExpiresIn:    int(newToken.Expiry.Sub(newToken.Expiry).Seconds()),
	}, nil
}

// Provider-specific implementations

func (r *oauthRepository) getGoogleUserInfo(ctx context.Context, client *http.Client) (*domain.UserInfo, error) {
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	var googleUser struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &domain.UserInfo{
		ID:       googleUser.ID,
		Email:    googleUser.Email,
		Name:     googleUser.Name,
		Picture:  googleUser.Picture,
		Provider: "google",
	}, nil
}

func (r *oauthRepository) getGithubUserInfo(ctx context.Context, client *http.Client) (*domain.UserInfo, error) {
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	var githubUser struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &domain.UserInfo{
		ID:       fmt.Sprintf("%d", githubUser.ID),
		Name:     githubUser.Login,
		Picture:  githubUser.AvatarURL,
		Provider: "github",
	}, nil
}

func (r *oauthRepository) getGithubVerifiedPrimaryEmail(ctx context.Context, client *http.Client) (string, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", fmt.Errorf("failed to get user emails: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get user emails: status %d", resp.StatusCode)
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", fmt.Errorf("failed to decode user emails: %w", err)
	}

	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}

	return "", nil
}

// Helper types

type ProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
}