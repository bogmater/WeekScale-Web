package main

import (
	"net/http"
	"strings"
	"testing"

	"bogmater/weekscale-web/internal/assert"
)

func TestRoutes(t *testing.T) {
	t.Run("Serves CSS file with appropriate headers and content", func(t *testing.T) {
		app := newTestApplication(t)

		req := newTestRequest(t, http.MethodGet, "/static/css/main.css")

		res := send(t, req, app.routes())
		assert.Equal(t, res.StatusCode, http.StatusOK)
		assert.Equal(t, res.Header.Get("Content-Type"), "text/css; charset=utf-8")
		assert.Equal(t, res.Header.Get("Cache-Control"), "public, max-age=31536000, immutable")
		assert.True(t, len(res.Body) > 0)
	})

	t.Run("Serves the platform switcher script", func(t *testing.T) {
		app := newTestApplication(t)

		req := newTestRequest(t, http.MethodGet, "/static/js/platform-switcher.js")

		res := send(t, req, app.routes())
		assert.Equal(t, res.StatusCode, http.StatusOK)
		assert.Equal(t, res.Header.Get("Cache-Control"), "public, max-age=31536000, immutable")
		assert.True(t, strings.Contains(res.Body, "platform-switcher"))
	})

	t.Run("Compresses HTML when requested", func(t *testing.T) {
		app := newTestApplication(t)
		req := newTestRequest(t, http.MethodGet, "/")
		req.Header.Set("Accept-Encoding", "gzip")

		res := send(t, req, app.routes())
		assert.Equal(t, res.StatusCode, http.StatusOK)
		assert.Equal(t, res.Header.Get("Content-Encoding"), "gzip")
		assert.Equal(t, res.Header.Get("Vary"), "Accept-Encoding")
	})

	t.Run("Serves HEAD requests for public pages", func(t *testing.T) {
		app := newTestApplication(t)
		req := newTestRequest(t, http.MethodHead, "/")

		res := send(t, req, app.routes())
		assert.Equal(t, res.StatusCode, http.StatusOK)
		assert.Equal(t, res.Header.Get("Content-Type"), "text/html; charset=utf-8")
	})

	t.Run("Redirects the bare production domain to www", func(t *testing.T) {
		app := newTestApplication(t)
		req := newTestRequest(t, http.MethodGet, "/faq?from=apex")
		req.Host = "weekscale.net"
		req.Header.Set("X-Forwarded-Proto", "http")

		res := send(t, req, app.routes())
		assert.Equal(t, res.StatusCode, http.StatusPermanentRedirect)
		assert.Equal(t, res.Header.Get("Location"), "https://www.weekscale.net/faq?from=apex")
	})

	t.Run("Redirects HTTPS on the bare production domain to www", func(t *testing.T) {
		app := newTestApplication(t)
		req := newTestRequest(t, http.MethodGet, "/faq?from=secure-apex")
		req.Host = "weekscale.net"
		req.Header.Set("X-Forwarded-Proto", "https")

		res := send(t, req, app.routes())
		assert.Equal(t, res.StatusCode, http.StatusPermanentRedirect)
		assert.Equal(t, res.Header.Get("Location"), "https://www.weekscale.net/faq?from=secure-apex")
	})

	t.Run("Redirects HTTP on the canonical host to HTTPS", func(t *testing.T) {
		app := newTestApplication(t)
		req := newTestRequest(t, http.MethodGet, "/privacy?from=http")
		req.Host = "www.weekscale.net"
		req.Header.Set("X-Forwarded-Proto", "http")

		res := send(t, req, app.routes())
		assert.Equal(t, res.StatusCode, http.StatusPermanentRedirect)
		assert.Equal(t, res.Header.Get("Location"), "https://www.weekscale.net/privacy?from=http")
	})

	t.Run("Serves HTTPS requests on the canonical host", func(t *testing.T) {
		app := newTestApplication(t)
		req := newTestRequest(t, http.MethodGet, "/faq")
		req.Host = "www.weekscale.net"
		req.Header.Set("X-Forwarded-Proto", "https")

		res := send(t, req, app.routes())
		assert.Equal(t, res.StatusCode, http.StatusOK)
	})

	t.Run("Serves the brand mark", func(t *testing.T) {
		app := newTestApplication(t)
		req := newTestRequest(t, http.MethodGet, "/static/img/weekscale-mark.svg")

		res := send(t, req, app.routes())
		assert.Equal(t, res.StatusCode, http.StatusOK)
		assert.Equal(t, res.Header.Get("Content-Type"), "image/svg+xml")
	})

	t.Run("Serves social and manifest assets", func(t *testing.T) {
		app := newTestApplication(t)

		for _, path := range []string{"/static/img/weekscale-social.png", "/static/site.webmanifest"} {
			req := newTestRequest(t, http.MethodGet, path)
			res := send(t, req, app.routes())

			assert.Equal(t, res.StatusCode, http.StatusOK)
			assert.Equal(t, res.Header.Get("Cache-Control"), "public, max-age=31536000, immutable")
			assert.True(t, len(res.Body) > 0)
		}
	})

	t.Run("Renders the 404 error page for non-existent routes", func(t *testing.T) {
		app := newTestApplication(t)

		req := newTestRequest(t, http.MethodGet, "/nonexistent")

		res := send(t, req, app.routes())
		assert.Equal(t, res.StatusCode, http.StatusNotFound)
		assert.True(t, containsPageTag(t, res.Body, "errors/404"))
	})

	t.Run("Sends a 405 response for routes with a matching route pattern but no matching HTTP method", func(t *testing.T) {
		app := newTestApplication(t)

		req := newTestRequest(t, http.MethodTrace, "/")

		res := send(t, req, app.routes())
		assert.Equal(t, res.StatusCode, http.StatusMethodNotAllowed)
	})
}
