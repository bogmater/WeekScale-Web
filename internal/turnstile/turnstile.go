package turnstile

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const siteverifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

type Result struct {
	Success  bool     `json:"success"`
	Hostname string   `json:"hostname"`
	Action   string   `json:"action"`
	Errors   []string `json:"error-codes"`
}

type Client struct {
	secret     string
	endpoint   string
	httpClient *http.Client
}

func NewClient(secret string) *Client {
	return &Client{
		secret:   secret,
		endpoint: siteverifyURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *Client) Verify(ctx context.Context, token, remoteIP string) (Result, error) {
	if token == "" || len(token) > 2048 {
		return Result{}, nil
	}

	values := url.Values{
		"secret":   {c.secret},
		"response": {token},
	}
	if remoteIP != "" {
		values.Set("remoteip", remoteIP)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return Result{}, fmt.Errorf("create Turnstile request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("verify Turnstile response: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("verify Turnstile response: unexpected status %d", res.StatusCode)
	}

	var result Result
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return Result{}, fmt.Errorf("decode Turnstile response: %w", err)
	}

	return result, nil
}
