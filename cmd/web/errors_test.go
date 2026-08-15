package main

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"bogmater/weekscale-web/internal/assert"
)

func TestReportServerError(t *testing.T) {
	t.Run("Logs error with correct details", func(t *testing.T) {
		var buf bytes.Buffer
		app := newTestApplication(t)
		app.logger = slog.New(slog.NewTextHandler(&buf, nil))

		req := newTestRequest(t, http.MethodGet, "/test")

		app.reportServerError(req, errors.New("this is a test error"))
		assert.True(t, strings.Contains(buf.String(), "level=ERROR"))
		assert.True(t, strings.Contains(buf.String(), `msg="this is a test error"`))
		assert.True(t, strings.Contains(buf.String(), "request.method=GET"))
		assert.True(t, strings.Contains(buf.String(), "request.url=/test"))
	})

	t.Run("Does not send notification email when disabled", func(t *testing.T) {
		app := newTestApplication(t)
		app.config.notifications.email = ""

		req := newTestRequest(t, http.MethodGet, "/test")

		send(t, req, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			app.serverError(w, r, errors.New("this is a test error"))
		}))
		assert.Equal(t, len(app.mailer.SentMessages), 0)
	})

	t.Run("Sends notification email when enabled", func(t *testing.T) {
		app := newTestApplication(t)
		app.config.notifications.email = "zoe@example.com"

		req := newTestRequest(t, http.MethodGet, "/test")

		send(t, req, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			app.serverError(w, r, errors.New("this is a test error"))
		}))
		assert.Equal(t, len(app.mailer.SentMessages), 1)
		assert.True(t, strings.Contains(app.mailer.SentMessages[0], "To: <zoe@example.com>"))
		assert.True(t, strings.Contains(app.mailer.SentMessages[0], "Error message: this is a test error"))
		assert.True(t, strings.Contains(app.mailer.SentMessages[0], "Request method: GET"))
		assert.True(t, strings.Contains(app.mailer.SentMessages[0], "Request URL: /test"))
	})

}

func TestServerError(t *testing.T) {
	t.Run("Logs error and renders the 500 error page without exposing error details", func(t *testing.T) {
		var buf bytes.Buffer
		app := newTestApplication(t)
		app.logger = slog.New(slog.NewTextHandler(&buf, nil))

		req := newTestRequest(t, http.MethodGet, "/test")

		res := send(t, req, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			app.serverError(w, r, errors.New("this is a test error"))
		}))
		assert.Equal(t, res.StatusCode, http.StatusInternalServerError)

		assert.True(t, containsPageTag(t, res.Body, "errors/500"))
		assert.False(t, strings.Contains(res.Body, "this is a test error"))

		assert.True(t, strings.Contains(buf.String(), "level=ERROR"))
		assert.True(t, strings.Contains(buf.String(), `msg="this is a test error"`))
		assert.True(t, strings.Contains(buf.String(), "request.method=GET"))
		assert.True(t, strings.Contains(buf.String(), "request.url=/test"))
	})

	t.Run("Does not send notification email when disabled", func(t *testing.T) {
		app := newTestApplication(t)
		app.config.notifications.email = ""

		req := newTestRequest(t, http.MethodGet, "/test")

		send(t, req, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			app.serverError(w, r, errors.New("this is a test error"))
		}))
		assert.Equal(t, len(app.mailer.SentMessages), 0)
	})

	t.Run("Sends notification email when enabled", func(t *testing.T) {
		app := newTestApplication(t)
		app.config.notifications.email = "zoe@example.com"

		req := newTestRequest(t, http.MethodGet, "/test")

		send(t, req, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			app.serverError(w, r, errors.New("this is a test error"))
		}))
		assert.Equal(t, len(app.mailer.SentMessages), 1)
		assert.True(t, strings.Contains(app.mailer.SentMessages[0], "To: <zoe@example.com>"))
		assert.True(t, strings.Contains(app.mailer.SentMessages[0], "Error message: this is a test error"))
		assert.True(t, strings.Contains(app.mailer.SentMessages[0], "Request method: GET"))
		assert.True(t, strings.Contains(app.mailer.SentMessages[0], "Request URL: /test"))
	})

}

func TestNotFound(t *testing.T) {
	t.Run("Renders the 404 error page", func(t *testing.T) {
		app := newTestApplication(t)

		req := newTestRequest(t, http.MethodGet, "/test")

		res := send(t, req, http.HandlerFunc(app.notFound))
		assert.Equal(t, res.StatusCode, http.StatusNotFound)

		assert.True(t, containsPageTag(t, res.Body, "errors/404"))

	})
}

func TestBadRequest(t *testing.T) {
	t.Run("Renders the 400 error page including the error message", func(t *testing.T) {
		app := newTestApplication(t)

		req := newTestRequest(t, http.MethodGet, "/test")

		res := send(t, req, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			app.badRequest(w, r, errors.New("this is a baaaad request"))
		}))
		assert.Equal(t, res.StatusCode, http.StatusBadRequest)

		assert.True(t, containsPageTag(t, res.Body, "errors/400"))
		assert.True(t, strings.Contains(res.Body, "this is a baaaad request"))

	})
}
