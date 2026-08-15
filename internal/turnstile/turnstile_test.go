package turnstile

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bogmater/weekscale-web/internal/assert"
)

func TestVerify(t *testing.T) {
	t.Run("returns successful verification", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			assert.Nil(t, err)
			assert.Equal(t, r.Form.Get("secret"), "test-secret")
			assert.Equal(t, r.Form.Get("response"), "test-token")
			assert.Equal(t, r.Form.Get("remoteip"), "192.0.2.1")
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"success":true,"hostname":"www.weekscale.net","action":"beta"}`))
		}))
		defer server.Close()

		client := NewClient("test-secret")
		client.endpoint = server.URL
		result, err := client.Verify(context.Background(), "test-token", "192.0.2.1")

		assert.Nil(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, result.Action, "beta")
	})

	t.Run("rejects empty and oversized tokens without a request", func(t *testing.T) {
		client := NewClient("test-secret")

		result, err := client.Verify(context.Background(), "", "")
		assert.Nil(t, err)
		assert.True(t, !result.Success)

		result, err = client.Verify(context.Background(), strings.Repeat("x", 2049), "")
		assert.Nil(t, err)
		assert.True(t, !result.Success)
	})

	t.Run("returns an error for a failed service request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		}))
		defer server.Close()

		client := NewClient("test-secret")
		client.endpoint = server.URL
		_, err := client.Verify(context.Background(), "test-token", "")

		assert.NotNil(t, err)
	})
}
