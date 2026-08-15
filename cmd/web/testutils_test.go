package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"bogmater/weekscale-web/internal/smtp"
	"bogmater/weekscale-web/internal/turnstile"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

type mockTurnstileVerifier struct {
	result turnstile.Result
	err    error
}

func (m *mockTurnstileVerifier) Verify(ctx context.Context, token, remoteIP string) (turnstile.Result, error) {
	if m.err != nil || m.result.Success {
		return m.result, m.err
	}

	return turnstile.Result{
		Success:  true,
		Hostname: "www.weekscale.net",
		Action:   token,
	}, nil
}

func newTestApplication(t *testing.T) *application {
	app := new(application)

	app.logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	app.mailer = smtp.NewMockMailer("test@example.com")
	app.config.baseURL = "https://www.weekscale.net"
	app.config.beta.email = "beta@example.com"
	app.config.support.email = "support@example.com"
	app.config.turnstile.siteKey = "test-site-key"
	app.config.turnstile.secretKey = "test-secret-key"
	app.turnstile = &mockTurnstileVerifier{}

	return app
}

func newTestRequest(t *testing.T, method, path string) *http.Request {
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Form = url.Values{}
	req.PostForm = url.Values{}

	req.Header.Set("Sec-Fetch-Site", "same-origin")
	return req
}

type testResponse struct {
	*http.Response
	Body string
}

func send(t *testing.T, req *http.Request, h http.Handler) testResponse {
	if len(req.PostForm) > 0 {
		body := req.PostForm.Encode()
		req.Body = io.NopCloser(strings.NewReader(body))
		req.ContentLength = int64(len(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	res := rec.Result()

	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	return testResponse{
		Response: res,
		Body:     strings.TrimSpace(string(resBody)),
	}
}

func containsPageTag(t *testing.T, htmlBody string, tag string) bool {
	return containsHTMLNode(t, htmlBody, fmt.Sprintf(`meta[name="page"][content="%s"]`, tag))
}

func containsHTMLNode(t *testing.T, htmlBody string, cssSelector string) bool {
	_, found := getHTMLNode(t, htmlBody, cssSelector)
	return found
}

func getHTMLNode(t *testing.T, htmlBody string, cssSelector string) (*html.Node, bool) {
	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		t.Fatal(err)
	}

	selector, err := cascadia.Compile(cssSelector)
	if err != nil {
		t.Fatal(err)
	}

	node := cascadia.Query(doc, selector)
	if node == nil {
		return nil, false
	}

	return node, true
}
