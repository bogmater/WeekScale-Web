package response

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bogmater/weekscale-web/internal/assert"
)

func TestNamedTemplate(t *testing.T) {
	t.Run("Write valid HTML response the correct status code", func(t *testing.T) {
		w := httptest.NewRecorder()

		err := NamedTemplate(w, http.StatusTeapot, "this is a test", "test:data", "testdata/test.tmpl")
		assert.Nil(t, err)
		assert.Equal(t, w.Code, http.StatusTeapot)
		assert.Equal(t, strings.TrimSpace(w.Body.String()), "<strong>this is a test</strong>")
	})
}

func TestNamedTemplateWithHeaders(t *testing.T) {
	t.Run("Write valid HTML response the correct status code", func(t *testing.T) {
		w := httptest.NewRecorder()

		err := NamedTemplateWithHeaders(w, http.StatusTeapot, "this is a test", nil, "test:data", "testdata/test.tmpl")
		assert.Nil(t, err)
		assert.Equal(t, w.Code, http.StatusTeapot)
		assert.Equal(t, strings.TrimSpace(w.Body.String()), "<strong>this is a test</strong>")
	})

	t.Run("Write valid HTML response with custom headers", func(t *testing.T) {
		w := httptest.NewRecorder()

		headers := http.Header{
			"X-Custom-Header": []string{"custom-value"},
			"X-Request-ID":    []string{"12345"},
			"X-Multiple":      []string{"value1", "value2", "value3"},
		}

		err := NamedTemplateWithHeaders(w, http.StatusTeapot, "this is a test", headers, "test:data", "testdata/test.tmpl")
		assert.Nil(t, err)
		assert.Equal(t, w.Code, http.StatusTeapot)
		assert.Equal(t, strings.TrimSpace(w.Body.String()), "<strong>this is a test</strong>")
		assert.Equal(t, w.Header().Get("X-Custom-Header"), "custom-value")
		assert.Equal(t, w.Header().Get("X-Request-ID"), "12345")
		assert.Equal(t, w.Header().Values("X-Multiple"), []string{"value1", "value2", "value3"})
	})

	t.Run("Check functions are available to the templates", func(t *testing.T) {
		w := httptest.NewRecorder()

		err := NamedTemplateWithHeaders(w, http.StatusTeapot, nil, nil, "test:function", "testdata/test.tmpl")
		assert.Nil(t, err)
		assert.Equal(t, strings.TrimSpace(w.Body.String()), "<strong>THIS IS ANOTHER TEST</strong>")
	})

	t.Run("Returns error for non-existent template name", func(t *testing.T) {
		w := httptest.NewRecorder()

		err := NamedTemplateWithHeaders(w, http.StatusTeapot, nil, nil, "test:non-existent-template", "testdata/test.tmpl")
		assert.NotNil(t, err)
	})

	t.Run("Returns error for non-existent template pattern", func(t *testing.T) {
		w := httptest.NewRecorder()

		err := NamedTemplateWithHeaders(w, http.StatusTeapot, nil, nil, "test:data", "testdata/non-existent-file.tmpl")
		assert.NotNil(t, err)
	})
}
