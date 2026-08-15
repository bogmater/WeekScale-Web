package main

import (
	"bytes"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"bogmater/weekscale-web/internal/assert"
)

func TestSecurityHeaders(t *testing.T) {
	t.Run("Sets appropriate security headers", func(t *testing.T) {
		app := newTestApplication(t)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		req := newTestRequest(t, http.MethodGet, "/test")

		res := send(t, req, app.securityHeaders(next))
		assert.Equal(t, res.StatusCode, http.StatusTeapot)
		assert.Equal(t, res.Header.Get("Referrer-Policy"), "origin-when-cross-origin")
		assert.Equal(t, res.Header.Get("X-Content-Type-Options"), "nosniff")
		assert.Equal(t, res.Header.Get("X-Frame-Options"), "deny")
	})
}

func TestRecoverPanic(t *testing.T) {
	t.Run("Allows normal requests to proceed", func(t *testing.T) {
		app := newTestApplication(t)
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		req := newTestRequest(t, http.MethodGet, "/test")

		res := send(t, req, app.recoverPanic(next))
		assert.Equal(t, res.StatusCode, http.StatusTeapot)
	})

	t.Run("Recovers from panic and renders the 500 error page", func(t *testing.T) {
		var buf bytes.Buffer
		app := newTestApplication(t)
		app.logger = slog.New(slog.NewTextHandler(&buf, nil))

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("something went wrong")
		})

		req := newTestRequest(t, http.MethodGet, "/test")

		res := send(t, req, app.recoverPanic(next))
		assert.Equal(t, res.StatusCode, http.StatusInternalServerError)
		assert.True(t, containsPageTag(t, res.Body, "errors/500"))
		assert.True(t, strings.Contains(buf.String(), "level=ERROR"))
		assert.True(t, strings.Contains(buf.String(), `msg="something went wrong"`))
	})
}

func TestLogRequest(t *testing.T) {
	t.Run("Logs the request and response details", func(t *testing.T) {
		var buf bytes.Buffer
		app := newTestApplication(t)
		app.logger = slog.New(slog.NewTextHandler(&buf, nil))

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
			w.Write([]byte("I'm a test teapot"))
		})

		req := newTestRequest(t, http.MethodGet, "/test")

		res := send(t, req, app.logRequest(next))
		assert.Equal(t, res.StatusCode, http.StatusTeapot)
		assert.True(t, strings.Contains(buf.String(), "level=INFO"))
		assert.True(t, strings.Contains(buf.String(), "msg=request"))
		assert.True(t, strings.Contains(buf.String(), "request.method=GET"))
		assert.True(t, strings.Contains(buf.String(), "request.url=/test"))
		assert.True(t, strings.Contains(buf.String(), "response.status=418"))
		assert.True(t, strings.Contains(buf.String(), "response.size=17"))
	})
}
