package oauthServices

import (
	"net/http"
)

func newHTTPClientWithToken(token string) *http.Client {
	return &http.Client{
		Transport: &bearerTransport{token: token, base: http.DefaultTransport},
	}
}

func newDiscordAPIRequest(token string) (*http.Request, error) {
	req, err := http.NewRequest("GET", "https://discord.com/api/users/@me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return req, nil
}

type bearerTransport struct {
	token string
	base  http.RoundTripper
}

func (t *bearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.Header.Set("Authorization", "Bearer "+t.token)
	return t.base.RoundTrip(clone)
}
