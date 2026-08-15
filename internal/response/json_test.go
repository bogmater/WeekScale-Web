package response

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bogmater/weekscale-web/internal/assert"
)

type testJSONSource struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

func TestJSON(t *testing.T) {
	t.Run("Write valid JSON response with formatting and the correct status code", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := testJSONSource{
			Name:  "John",
			Age:   30,
			Email: "john@example.com",
		}

		err := JSON(w, http.StatusTeapot, data)
		assert.Nil(t, err)
		assert.Equal(t, w.Code, http.StatusTeapot)
		assert.Equal(t, w.Header().Get("Content-Type"), "application/json")
		assert.Equal(t, w.Body.String(), "{\n\t\"name\": \"John\",\n\t\"age\": 30,\n\t\"email\": \"john@example.com\"\n}\n")
	})
}

func TestJSONWithHeaders(t *testing.T) {
	t.Run("Write valid JSON response with formatting and whitespace", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := testJSONSource{
			Name:  "John",
			Age:   30,
			Email: "john@example.com",
		}

		err := JSONWithHeaders(w, http.StatusTeapot, data, nil)
		assert.Nil(t, err)
		assert.Equal(t, w.Code, http.StatusTeapot)
		assert.Equal(t, w.Header().Get("Content-Type"), "application/json")
		assert.Equal(t, w.Body.String(), "{\n\t\"name\": \"John\",\n\t\"age\": 30,\n\t\"email\": \"john@example.com\"\n}\n")
	})

	t.Run("Return error for non-marshallable data", func(t *testing.T) {
		w := httptest.NewRecorder()

		err := JSONWithHeaders(w, http.StatusTeapot, make(chan int), nil)
		assert.NotNil(t, err)
	})

	t.Run("Write JSON response with custom headers", func(t *testing.T) {
		w := httptest.NewRecorder()

		headers := http.Header{
			"X-Custom-Header": []string{"custom-value"},
			"X-Request-ID":    []string{"12345"},
			"X-Multiple":      []string{"value1", "value2", "value3"},
		}

		err := JSONWithHeaders(w, http.StatusCreated, "test", headers)
		assert.Nil(t, err)
		assert.Equal(t, w.Header().Get("Content-Type"), "application/json")
		assert.Equal(t, w.Header().Get("X-Custom-Header"), "custom-value")
		assert.Equal(t, w.Header().Get("X-Request-ID"), "12345")
		assert.Equal(t, w.Header().Values("X-Multiple"), []string{"value1", "value2", "value3"})
	})
}
