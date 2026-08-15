package request

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/go-playground/form/v4"
)

var decoder = form.NewDecoder()

func DecodePostForm(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1_048_576)

	err := r.ParseForm()
	if err != nil {
		return err
	}

	return decodeURLValues(r.PostForm, dst)
}

func DecodeQueryString(r *http.Request, dst any) error {
	return decodeURLValues(r.URL.Query(), dst)
}

func decodeURLValues(v url.Values, dst any) error {
	err := decoder.Decode(dst, v)
	if err != nil {
		var invalidDecoderError *form.InvalidDecoderError

		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}
	}

	return err
}
