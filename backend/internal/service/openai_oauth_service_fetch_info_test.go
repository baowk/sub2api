//go:build unit

package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/imroc/req/v3"
	"github.com/stretchr/testify/require"
)

func newOpenAIAccountInfoTestClient(t *testing.T, serverURL string) *req.Client {
	t.Helper()

	target, err := url.Parse(serverURL)
	require.NoError(t, err)

	client := req.C()
	client.GetTransport().WrapRoundTripFunc(func(rt http.RoundTripper) req.HttpRoundTripFunc {
		return func(r *http.Request) (*http.Response, error) {
			cloned := r.Clone(r.Context())
			cloned.URL.Scheme = target.Scheme
			cloned.URL.Host = target.Host
			cloned.Host = target.Host
			return rt.RoundTrip(cloned)
		}
	})
	return client
}

func TestOpenAIOAuthService_FetchAccountInfo(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/backend-api/accounts/check/v4-2023-04-27", r.URL.Path)
		require.Equal(t, "Bearer existing-access-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"accounts": {
				"org-123": {
					"account": {
						"plan_type": "plus",
						"is_default": true
					},
					"entitlement": {
						"expires_at": "2026-05-01T00:00:00Z"
					}
				}
			}
		}`))
	}))
	defer server.Close()

	svc := NewOpenAIOAuthService(nil, nil)
	svc.SetPrivacyClientFactory(func(proxyURL string) (*req.Client, error) {
		return newOpenAIAccountInfoTestClient(t, server.URL), nil
	})

	account := &Account{
		ID:       77,
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token":    "existing-access-token",
			"organization_id": "org-123",
			"expires_at":      time.Now().Add(30 * time.Minute).UTC().Format(time.RFC3339),
			"email":           "old@example.com",
		},
	}

	info, err := svc.FetchAccountInfo(context.Background(), account)
	require.NoError(t, err)
	require.NotNil(t, info)
	require.Equal(t, "existing-access-token", info.AccessToken)
	require.Equal(t, "plus", info.PlanType)
	require.Equal(t, "2026-05-01T00:00:00Z", info.SubscriptionExpiresAt)
	require.Equal(t, "old@example.com", info.Email)
}

func TestOpenAIOAuthService_FetchSupportedModels(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/backend-api/codex/models", r.URL.Path)
		require.Equal(t, "Bearer existing-access-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"models": [
				{"slug": "gpt-5.5"},
				{"slug": "gpt-5.4"},
				{"slug": "gpt-5.5"}
			]
		}`))
	}))
	defer server.Close()

	svc := NewOpenAIOAuthService(nil, nil)
	svc.SetPrivacyClientFactory(func(proxyURL string) (*req.Client, error) {
		return newOpenAIAccountInfoTestClient(t, server.URL), nil
	})

	account := &Account{
		ID:       77,
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token": "existing-access-token",
		},
	}

	models, err := svc.FetchSupportedModels(context.Background(), account)
	require.NoError(t, err)
	require.Equal(t, []string{"gpt-5.4", "gpt-5.5"}, models)
}
