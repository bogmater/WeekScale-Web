package main

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"bogmater/weekscale-web/internal/assert"
)

func TestHome(t *testing.T) {
	t.Run("GET renders the home page", func(t *testing.T) {
		app := newTestApplication(t)

		req := newTestRequest(t, http.MethodGet, "/")

		res := send(t, req, app.routes())
		assert.Equal(t, res.StatusCode, http.StatusOK)
		assert.True(t, containsPageTag(t, res.Body, "home"))
		assert.True(t, containsHTMLNode(t, res.Body, `.cf-turnstile[data-action="beta"][data-sitekey="test-site-key"]`))
		assert.True(t, !strings.Contains(res.Body, "test-secret-key"))
	})
}

func TestHealthcheck(t *testing.T) {
	app := newTestApplication(t)
	req := newTestRequest(t, http.MethodGet, "/healthz")

	res := send(t, req, app.routes())

	assert.Equal(t, res.StatusCode, http.StatusOK)
	assert.Equal(t, res.Header.Get("Content-Type"), "text/plain; charset=utf-8")
	assert.Equal(t, res.Header.Get("Cache-Control"), "no-store")
	assert.Equal(t, res.Body, "ok")
}

func TestContentPages(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		pageTag string
	}{
		{name: "FAQ", path: "/faq", pageTag: "faq"},
		{name: "privacy", path: "/privacy", pageTag: "privacy"},
		{name: "support", path: "/support", pageTag: "support"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestApplication(t)
			req := newTestRequest(t, http.MethodGet, tt.path)

			res := send(t, req, app.routes())
			assert.Equal(t, res.StatusCode, http.StatusOK)
			assert.True(t, containsPageTag(t, res.Body, tt.pageTag))
			assert.True(t, containsHTMLNode(t, res.Body, `nav a[aria-current="page"]`))
			if tt.path == "/support" {
				assert.True(t, containsHTMLNode(t, res.Body, `.cf-turnstile[data-action="support"][data-sitekey="test-site-key"]`))
			}
		})
	}
}

func TestSubmitBeta(t *testing.T) {
	validForm := func() url.Values {
		return url.Values{
			"Email":                 {"tester@example.com"},
			"Platform":              {"ios"},
			"cf-turnstile-response": {"beta"},
		}
	}

	t.Run("sends a valid beta signup", func(t *testing.T) {
		app := newTestApplication(t)
		req := newTestRequest(t, http.MethodPost, "/beta")
		req.PostForm = validForm()

		res := send(t, req, app.routes())
		app.wg.Wait()

		assert.Equal(t, res.StatusCode, http.StatusSeeOther)
		assert.Equal(t, res.Header.Get("Location"), "/?beta=sent#beta")
		assert.Equal(t, len(app.mailer.SentMessages), 1)
		assert.True(t, strings.Contains(app.mailer.SentMessages[0], "tester@example.com"))
		assert.True(t, strings.Contains(app.mailer.SentMessages[0], "iOS"))
	})

	t.Run("rejects a missing Turnstile token", func(t *testing.T) {
		app := newTestApplication(t)
		req := newTestRequest(t, http.MethodPost, "/beta")
		req.PostForm = validForm()
		req.PostForm.Del("cf-turnstile-response")

		res := send(t, req, app.routes())

		assert.Equal(t, res.StatusCode, http.StatusUnprocessableEntity)
		assert.True(t, strings.Contains(res.Body, "Complete the security verification"))
		assert.Equal(t, len(app.mailer.SentMessages), 0)
	})

	t.Run("rejects a mismatched Turnstile action", func(t *testing.T) {
		app := newTestApplication(t)
		req := newTestRequest(t, http.MethodPost, "/beta")
		req.PostForm = validForm()
		req.PostForm.Set("cf-turnstile-response", "support")

		res := send(t, req, app.routes())

		assert.Equal(t, res.StatusCode, http.StatusUnprocessableEntity)
		assert.True(t, strings.Contains(res.Body, "Security verification failed"))
		assert.Equal(t, len(app.mailer.SentMessages), 0)
	})

	t.Run("renders beta validation errors", func(t *testing.T) {
		app := newTestApplication(t)
		req := newTestRequest(t, http.MethodPost, "/beta")
		req.PostForm = url.Values{
			"Email":    {"not-an-email"},
			"Platform": {"windows"},
		}

		res := send(t, req, app.routes())

		assert.Equal(t, res.StatusCode, http.StatusUnprocessableEntity)
		assert.True(t, containsHTMLNode(t, res.Body, `input[name="Email"][aria-invalid="true"]`))
		assert.True(t, strings.Contains(res.Body, "Choose Android or iOS"))
		assert.Equal(t, len(app.mailer.SentMessages), 0)
	})

	t.Run("silently accepts beta honeypot submissions", func(t *testing.T) {
		app := newTestApplication(t)
		req := newTestRequest(t, http.MethodPost, "/beta")
		req.PostForm = validForm()
		req.PostForm.Set("Company", "Spam Company")

		res := send(t, req, app.routes())

		assert.Equal(t, res.StatusCode, http.StatusSeeOther)
		assert.Equal(t, len(app.mailer.SentMessages), 0)
	})

	t.Run("reports unavailable beta configuration", func(t *testing.T) {
		app := newTestApplication(t)
		app.config.beta.email = ""
		req := newTestRequest(t, http.MethodPost, "/beta")
		req.PostForm = validForm()

		res := send(t, req, app.routes())

		assert.Equal(t, res.StatusCode, http.StatusServiceUnavailable)
		assert.True(t, strings.Contains(res.Body, "temporarily unavailable"))
	})

	t.Run("fails closed without Turnstile configuration", func(t *testing.T) {
		app := newTestApplication(t)
		app.config.turnstile.secretKey = ""
		req := newTestRequest(t, http.MethodPost, "/beta")
		req.PostForm = validForm()

		res := send(t, req, app.routes())

		assert.Equal(t, res.StatusCode, http.StatusServiceUnavailable)
		assert.True(t, strings.Contains(res.Body, "Security verification is temporarily unavailable"))
		assert.Equal(t, len(app.mailer.SentMessages), 0)
	})
}

func TestSubmitSupport(t *testing.T) {
	validForm := func() url.Values {
		return url.Values{
			"Name":                  {"Alex Example"},
			"Email":                 {"alex@example.com"},
			"Message":               {"I need some help with WeekScale."},
			"cf-turnstile-response": {"support"},
		}
	}

	t.Run("sends a valid message", func(t *testing.T) {
		app := newTestApplication(t)
		req := newTestRequest(t, http.MethodPost, "/support")
		req.PostForm = validForm()

		res := send(t, req, app.routes())
		app.wg.Wait()

		assert.Equal(t, res.StatusCode, http.StatusSeeOther)
		assert.Equal(t, res.Header.Get("Location"), "/support?sent=1")
		assert.Equal(t, len(app.mailer.SentMessages), 1)
		assert.True(t, strings.Contains(app.mailer.SentMessages[0], "alex@example.com"))
	})

	t.Run("renders validation errors", func(t *testing.T) {
		app := newTestApplication(t)
		req := newTestRequest(t, http.MethodPost, "/support")
		req.PostForm = url.Values{
			"Name":    {""},
			"Email":   {"not-an-email"},
			"Message": {"short"},
		}

		res := send(t, req, app.routes())

		assert.Equal(t, res.StatusCode, http.StatusUnprocessableEntity)
		assert.True(t, containsHTMLNode(t, res.Body, `input[name="Email"][aria-invalid="true"]`))
		assert.Equal(t, len(app.mailer.SentMessages), 0)
	})

	t.Run("silently accepts honeypot submissions", func(t *testing.T) {
		app := newTestApplication(t)
		req := newTestRequest(t, http.MethodPost, "/support")
		req.PostForm = validForm()
		req.PostForm.Set("Website", "https://spam.example")

		res := send(t, req, app.routes())

		assert.Equal(t, res.StatusCode, http.StatusSeeOther)
		assert.Equal(t, len(app.mailer.SentMessages), 0)
	})

	t.Run("rejects cross-site browser submissions", func(t *testing.T) {
		app := newTestApplication(t)
		req := newTestRequest(t, http.MethodPost, "/support")
		req.Header.Set("Sec-Fetch-Site", "cross-site")
		req.PostForm = validForm()

		res := send(t, req, app.routes())

		assert.Equal(t, res.StatusCode, http.StatusBadRequest)
		assert.Equal(t, len(app.mailer.SentMessages), 0)
	})

	t.Run("reports unavailable support configuration", func(t *testing.T) {
		app := newTestApplication(t)
		app.config.support.email = ""
		req := newTestRequest(t, http.MethodPost, "/support")
		req.PostForm = validForm()

		res := send(t, req, app.routes())

		assert.Equal(t, res.StatusCode, http.StatusServiceUnavailable)
		assert.True(t, strings.Contains(res.Body, "temporarily unavailable"))
	})
}

func TestFormSubmissionRateLimit(t *testing.T) {
	app := newTestApplication(t)
	now := time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)

	for range formSubmissionLimit {
		assert.True(t, app.allowFormSubmission("support:192.0.2.1", now))
	}
	assert.True(t, !app.allowFormSubmission("support:192.0.2.1", now))
	assert.True(t, app.allowFormSubmission("beta:192.0.2.1", now))
	assert.True(t, app.allowFormSubmission("support:192.0.2.1", now.Add(formSubmissionWindow+time.Second)))
}
