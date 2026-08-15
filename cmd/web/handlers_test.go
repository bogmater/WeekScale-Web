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
		})
	}
}

func TestSubmitSupport(t *testing.T) {
	validForm := func() url.Values {
		return url.Values{
			"Name":    {"Alex Example"},
			"Email":   {"alex@example.com"},
			"Message": {"I need some help with WeekScale."},
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

func TestSupportRateLimit(t *testing.T) {
	app := newTestApplication(t)
	now := time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)

	for range supportRequestLimit {
		assert.True(t, app.allowSupportRequest("192.0.2.1", now))
	}
	assert.True(t, !app.allowSupportRequest("192.0.2.1", now))
	assert.True(t, app.allowSupportRequest("192.0.2.1", now.Add(supportRequestWindow+time.Second)))
}
