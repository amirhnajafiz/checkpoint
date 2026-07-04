package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const googleUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"

// Google wraps the OAuth2 client used for the "Login with Google" flow.
type GoogleOAuth struct {
	cfg *oauth2.Config
}

// NewGoogle builds the Google OAuth2 client from the app's credentials.
func NewGoogleOAuth(clientID, clientSecret, redirectURL string) *GoogleOAuth {
	return &GoogleOAuth{
		cfg: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
			Endpoint:     google.Endpoint,
		},
	}
}

// AuthCodeURL returns the Google consent URL for the given CSRF state.
func (g *GoogleOAuth) AuthCodeURL(state string) string {
	return g.cfg.AuthCodeURL(state)
}

// Exchange trades an authorization code for an OAuth token.
func (g *GoogleOAuth) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return g.cfg.Exchange(ctx, code)
}

// Email fetches the authenticated account's email from Google's userinfo
// endpoint.
func (g *GoogleOAuth) Email(ctx context.Context, tok *oauth2.Token) (string, error) {
	resp, err := g.cfg.Client(ctx, tok).Get(googleUserInfoURL)
	if err != nil {
		return "", fmt.Errorf("reach google userinfo: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("google userinfo returned status %d", resp.StatusCode)
	}

	var info struct {
		Email    string `json:"email"`
		Verified bool   `json:"verified_email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", fmt.Errorf("decode google userinfo: %w", err)
	}
	if info.Email == "" {
		return "", fmt.Errorf("google account has no email")
	}

	return info.Email, nil
}
