package service

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	"commerce-platform/internal/domain/social"
	"commerce-platform/internal/pkg/config"
)

func TestSocialOAuthBuildsPKCEAuthorizationURLs(t *testing.T) {
	service := &SocialOAuthService{
		config: config.SocialOAuthConfig{
			X: config.SocialOAuthProviderConfig{
				ClientID:    "x-client",
				RedirectURL: "https://admin.example.com/api/admin/social/oauth/callback",
			},
			YouTube: config.SocialOAuthProviderConfig{
				ClientID:    "youtube-client",
				RedirectURL: "https://admin.example.com/api/admin/social/oauth/callback",
			},
		},
	}

	for _, provider := range []string{social.ProviderX, social.ProviderYouTube} {
		authorizationURL, err := service.buildAuthorizationURL(provider, "state-value", "verifier-value")
		if err != nil {
			t.Fatalf("buildAuthorizationURL(%q) returned error: %v", provider, err)
		}
		parsed, err := url.Parse(authorizationURL)
		if err != nil {
			t.Fatalf("parse authorization URL: %v", err)
		}
		query := parsed.Query()
		if query.Get("state") != "state-value" || query.Get("code_challenge_method") != "S256" {
			t.Fatalf("authorization query = %v, want state and S256 PKCE", query)
		}
		if query.Get("code_challenge") != pkceChallenge("verifier-value") {
			t.Fatalf("authorization query code_challenge = %q, want PKCE challenge", query.Get("code_challenge"))
		}
	}
}

func TestSocialOAuthReturnPathRejectsExternalURLs(t *testing.T) {
	if got := normalizeSocialOAuthReturnPath("https://evil.example/steal", "https://admin.example.com/social"); got != "/social" {
		t.Fatalf("external return path = %q, want /social", got)
	}
	if got := normalizeSocialOAuthReturnPath("/social/meta?locale=zh-CN", ""); got != "/social/meta?locale=zh-CN" {
		t.Fatalf("relative return path = %q, want preserved path and query", got)
	}
}

func TestSocialOAuthConnectionViewDoesNotExposeTokens(t *testing.T) {
	service := &SocialOAuthService{config: config.SocialOAuthConfig{}}
	connection := &social.OAuthConnection{
		Provider:              social.ProviderX,
		Status:                social.OAuthStatusConnected,
		AccessTokenEncrypted:  "encrypted-access-token",
		RefreshTokenEncrypted: "encrypted-refresh-token",
		ProviderAccountName:   "brand",
		ProviderResources:     []byte(`{"username":"brand"}`),
	}

	view := service.connectionView(connection)
	encoded, err := json.Marshal(view)
	if err != nil {
		t.Fatalf("marshal connection view: %v", err)
	}
	if strings.Contains(string(encoded), "encrypted-access-token") || strings.Contains(string(encoded), "encrypted-refresh-token") {
		t.Fatalf("connection view leaked encrypted token fields: %s", encoded)
	}
	if view.ProviderAccountName != "brand" || !view.Connected {
		t.Fatalf("connection view = %#v, want connected account without token fields", view)
	}
}
