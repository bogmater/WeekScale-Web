package main

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"

	"bogmater/weekscale-web/internal/response"

	"github.com/tomasen/realip"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			pv := recover()
			if pv != nil {
				app.serverError(w, r, fmt.Errorf("%v", pv))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'; img-src 'self'; style-src 'self'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")

		next.ServeHTTP(w, r)
	})
}

func (app *application) canonicalHost(next http.Handler) http.Handler {
	canonicalURL, err := url.Parse(app.siteURL())
	if err != nil || !strings.HasPrefix(canonicalURL.Hostname(), "www.") {
		return next
	}

	bareHost := strings.TrimPrefix(canonicalURL.Hostname(), "www.")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestHost := r.Host
		if host, _, err := net.SplitHostPort(r.Host); err == nil {
			requestHost = host
		}

		if strings.EqualFold(requestHost, bareHost) {
			target := canonicalURL.Scheme + "://" + canonicalURL.Host + r.URL.RequestURI()
			http.Redirect(w, r, target, http.StatusPermanentRedirect)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) cacheStaticFiles(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mw := response.NewMetricsResponseWriter(w)
		next.ServeHTTP(mw, r)

		var (
			ip     = realip.FromRequest(r)
			method = r.Method
			url    = r.URL.String()
			proto  = r.Proto
		)

		userAttrs := slog.Group("user", "ip", ip)
		requestAttrs := slog.Group("request", "method", method, "url", url, "proto", proto)
		responseAttrs := slog.Group("response", "status", mw.StatusCode, "size", mw.BytesCount)

		app.logger.Info("request", userAttrs, requestAttrs, responseAttrs)
	})
}
