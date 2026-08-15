package request

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"bogmater/weekscale-web/internal/assert"
)

type testDecodeFormTarget struct {
	Name  string `form:"name"`
	Age   int    `form:"age"`
	Email string `form:"email"`
}

func TestDecodePostForm(t *testing.T) {
	t.Run("Decode valid POST form data successfully", func(t *testing.T) {
		form := url.Values{
			"name":  []string{"Jane"},
			"age":   []string{"25"},
			"email": []string{"jane@example.com"},
		}

		req := httptest.NewRequest("POST", "/test", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var target testDecodeFormTarget
		err := DecodePostForm(httptest.NewRecorder(), req, &target)
		assert.Nil(t, err)
		assert.Equal(t, target.Name, "Jane")
		assert.Equal(t, target.Age, 25)
		assert.Equal(t, target.Email, "jane@example.com")
	})

	t.Run("Return error when form parsing fails", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test", strings.NewReader("invalid%form%data"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var target testDecodeFormTarget
		err := DecodePostForm(httptest.NewRecorder(), req, &target)
		assert.NotNil(t, err)
	})

	t.Run("Ignore query string parameters", func(t *testing.T) {
		form := url.Values{
			"name": []string{"David"},
		}

		req := httptest.NewRequest("POST", "/test?name=Karen&age=63", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var target testDecodeFormTarget
		err := DecodePostForm(httptest.NewRecorder(), req, &target)
		assert.Nil(t, err)
		assert.Equal(t, target.Name, "David")
		assert.Equal(t, target.Age, 0)
	})
}

func TestDecodeQueryString(t *testing.T) {
	t.Run("Decode valid query string successfully", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test?name=Bob&age=35&email=bob@example.com", nil)

		var target testDecodeFormTarget
		err := DecodeQueryString(req, &target)
		assert.Nil(t, err)
		assert.Equal(t, target.Name, "Bob")
		assert.Equal(t, target.Age, 35)
		assert.Equal(t, target.Email, "bob@example.com")
	})

	t.Run("Decode empty query string to zero values", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)

		var target testDecodeFormTarget
		err := DecodeQueryString(req, &target)
		assert.Nil(t, err)
		assert.Equal(t, target.Name, "")
		assert.Equal(t, target.Age, 0)
		assert.Equal(t, target.Email, "")
	})

	t.Run("Handle URL encoded query parameters", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test?name=John%20Doe&email=john%2Bdoe%40example.com", nil)

		var target testDecodeFormTarget
		err := DecodeQueryString(req, &target)
		assert.Nil(t, err)
		assert.Equal(t, target.Name, "John Doe")
		assert.Equal(t, target.Email, "john+doe@example.com")
	})
}
