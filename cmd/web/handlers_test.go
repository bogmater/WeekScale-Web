package main

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"bogmater/weekscale-web/internal/assert"

	"golang.org/x/net/html"
)

func TestHome(t *testing.T) {
	t.Run("GET renders the home page", func(t *testing.T) {
		app := newTestApplication(t)

		req := newTestRequest(t, http.MethodGet, "/")

		res := send(t, req, app.routes())
		assert.Equal(t, res.StatusCode, http.StatusOK)
		assert.True(t, containsPageTag(t, res.Body, "home"))
		assert.True(t, containsHTMLNode(t, res.Body, `.cf-turnstile[data-action="beta"][data-sitekey="test-site-key"]`))
		assert.True(t, strings.Contains(res.Body, "test-secret-key") == false)
		assert.True(t, containsHTMLNode(t, res.Body, `[class="platform-switcher"]`))
		assert.True(t, strings.Contains(res.Body, "/static/js/platform-switcher.js"))
		assert.True(t, strings.Contains(res.Body, `data-ios-src="/static/img/weekscale-ios-dashboard-720.webp`))
		assert.True(t, strings.Contains(res.Body, "weekscale-ios-entry-1080.webp"))
		assert.True(t, strings.Contains(res.Body, "weekscale-ios-trend-720.webp"))
		assert.True(t, strings.Contains(res.Body, "I built the tracker I wanted to use."))
		assert.True(t, strings.Contains(res.Body, "A small app, with a deliberate plan."))
		assert.True(t, strings.Contains(res.Body, "A weight tracker that doesn't want your data"))
		assert.True(t, strings.Contains(res.Body, "On Android, WeekScale isn't even granted internet permission."))
		assert.True(t, strings.Contains(res.Body, "A finished week stays finished."))
		assert.True(t, strings.Contains(res.Body, "WeekScale will never have a goal weight."))
		assert.True(t, strings.Contains(res.Body, `href="/why-calendar-weeks"`))
		assert.True(t, strings.Contains(res.Body, `href="/about"`))
	})
}

func TestAboutPage(t *testing.T) {
	t.Run("renders the story, values, and roadmap", func(t *testing.T) {
		app := newTestApplication(t)
		req := newTestRequest(t, http.MethodGet, "/about")

		res := send(t, req, app.routes())
		assert.Equal(t, res.StatusCode, http.StatusOK)
		assert.True(t, containsPageTag(t, res.Body, "about"))
		assert.True(t, strings.Contains(res.Body, "I built the tracker I wanted to use."))
		assert.True(t, strings.Contains(res.Body, "No data sharing"))
		assert.True(t, strings.Contains(res.Body, "The only thing WeekScale rewards"))
		assert.True(t, strings.Contains(res.Body, "green card"))
		assert.True(t, strings.Contains(res.Body, "UI-focused refinements"))
		assert.True(t, strings.Contains(res.Body, "Apple Health on iOS"))
		assert.True(t, strings.Contains(res.Body, "Health Connect on Android"))
		assert.True(t, strings.Contains(res.Body, "nothing is uploaded"))
		assert.True(t, strings.Contains(res.Body, `href="/support"`))
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
		{name: "about", path: "/about", pageTag: "about"},
		{name: "happy scale alternative", path: "/happy-scale-alternative", pageTag: "happy-scale-alternative"},
		{name: "libra alternative ios", path: "/libra-alternative-ios", pageTag: "libra-alternative-ios"},
		{name: "private weight tracker", path: "/private-weight-tracker", pageTag: "private-weight-tracker"},
		{name: "privacy", path: "/privacy", pageTag: "privacy"},
		{name: "support", path: "/support", pageTag: "support"},
		{name: "weekly average weight", path: "/weekly-average-weight", pageTag: "weekly-average-weight"},
		{name: "weight tracker without subscription", path: "/weight-tracker-without-subscription", pageTag: "weight-tracker-without-subscription"},
		{name: "why calendar weeks", path: "/why-calendar-weeks", pageTag: "why-calendar-weeks"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestApplication(t)
			req := newTestRequest(t, http.MethodGet, tt.path)

			res := send(t, req, app.routes())
			assert.Equal(t, res.StatusCode, http.StatusOK)
			assert.True(t, containsPageTag(t, res.Body, tt.pageTag))
			if tt.path == "/faq" || tt.path == "/privacy" || tt.path == "/support" || tt.path == "/about" {
				assert.True(t, containsHTMLNode(t, res.Body, `nav a[aria-current="page"]`))
			}
			if tt.path == "/support" {
				assert.True(t, containsHTMLNode(t, res.Body, `.cf-turnstile[data-action="support"][data-sitekey="test-site-key"]`))
			}
		})
	}
}

func TestOfflineWeightTrackerRedirect(t *testing.T) {
	app := newTestApplication(t)
	req := newTestRequest(t, http.MethodGet, "/offline-weight-tracker-no-account")

	res := send(t, req, app.routes())

	assert.Equal(t, res.StatusCode, http.StatusMovedPermanently)
	assert.Equal(t, res.Header.Get("Location"), "/private-weight-tracker")
}

func TestInternalLinksResolveWithoutRedirects(t *testing.T) {
	app := newTestApplication(t)
	handler := app.routes()
	paths := []string{"/", "/about", "/faq", "/happy-scale-alternative", "/libra-alternative-ios", "/private-weight-tracker", "/privacy", "/support", "/weekly-average-weight", "/weight-tracker-without-subscription", "/why-calendar-weeks"}

	for _, path := range paths {
		res := send(t, newTestRequest(t, http.MethodGet, path), handler)
		doc, err := html.Parse(strings.NewReader(res.Body))
		if err != nil {
			t.Fatal(err)
		}

		var visit func(*html.Node)
		visit = func(node *html.Node) {
			if node.Type == html.ElementNode && node.Data == "a" {
				for _, attr := range node.Attr {
					if attr.Key != "href" || !strings.HasPrefix(attr.Val, "/") || strings.HasPrefix(attr.Val, "//") {
						continue
					}

					link, err := url.Parse(attr.Val)
					if err != nil {
						t.Errorf("%s contains invalid internal link %q: %v", path, attr.Val, err)
						continue
					}
					linkedRes := send(t, newTestRequest(t, http.MethodGet, link.Path), handler)
					if linkedRes.StatusCode >= http.StatusMultipleChoices && linkedRes.StatusCode < http.StatusBadRequest {
						t.Errorf("%s links to redirect %s", path, attr.Val)
					} else if linkedRes.StatusCode == http.StatusNotFound {
						t.Errorf("%s links to missing page %s", path, attr.Val)
					}
				}
			}
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				visit(child)
			}
		}
		visit(doc)
	}
}

func TestIOSPrivacyDisclosures(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{path: "/", want: "the iOS app makes no direct network requests"},
		{path: "/faq", want: "Apple-managed device or iCloud Backup"},
		{path: "/privacy", want: "CloudKit app synchronization is disabled on iOS"},
		{path: "/private-weight-tracker", want: "The iOS app stores entries with SwiftData"},
	}

	app := newTestApplication(t)
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := newTestRequest(t, http.MethodGet, tt.path)
			res := send(t, req, app.routes())

			assert.Equal(t, res.StatusCode, http.StatusOK)
			assert.True(t, strings.Contains(res.Body, tt.want))
			assert.True(t, !strings.Contains(res.Body, "iOS version is still in development"))
		})
	}
}

func TestSubmitBeta(t *testing.T) {
	validForm := func() url.Values {
		return url.Values{
			"Email":                 {"tester@example.com"},
			"Platform":              {"ios"},
			"cf-turnstile-response": {"beta"},
		}
	}

	t.Run("sends a valid beta signup", func(t *testing.T) {
		app := newTestApplication(t)
		req := newTestRequest(t, http.MethodPost, "/beta")
		req.PostForm = validForm()

		res := send(t, req, app.routes())
		app.wg.Wait()

		assert.Equal(t, res.StatusCode, http.StatusSeeOther)
		assert.Equal(t, res.Header.Get("Location"), "/?beta=sent#beta")
		assert.Equal(t, len(app.mailer.SentMessages), 1)
		assert.True(t, strings.Contains(app.mailer.SentMessages[0], "tester@example.com"))
		assert.True(t, strings.Contains(app.mailer.SentMessages[0], "iOS"))
	})

	t.Run("rejects a missing Turnstile token", func(t *testing.T) {
		app := newTestApplication(t)
		req := newTestRequest(t, http.MethodPost, "/beta")
		req.PostForm = validForm()
		req.PostForm.Del("cf-turnstile-response")

		res := send(t, req, app.routes())

		assert.Equal(t, res.StatusCode, http.StatusUnprocessableEntity)
		assert.True(t, strings.Contains(res.Body, "Complete the security verification"))
		assert.Equal(t, len(app.mailer.SentMessages), 0)
	})

	t.Run("rejects a mismatched Turnstile action", func(t *testing.T) {
		app := newTestApplication(t)
		req := newTestRequest(t, http.MethodPost, "/beta")
		req.PostForm = validForm()
		req.PostForm.Set("cf-turnstile-response", "support")

		res := send(t, req, app.routes())

		assert.Equal(t, res.StatusCode, http.StatusUnprocessableEntity)
		assert.True(t, strings.Contains(res.Body, "Security verification failed"))
		assert.Equal(t, len(app.mailer.SentMessages), 0)
	})

	t.Run("renders beta validation errors", func(t *testing.T) {
		app := newTestApplication(t)
		req := newTestRequest(t, http.MethodPost, "/beta")
		req.PostForm = url.Values{
			"Email":    {"not-an-email"},
			"Platform": {"windows"},
		}

		res := send(t, req, app.routes())

		assert.Equal(t, res.StatusCode, http.StatusUnprocessableEntity)
		assert.True(t, containsHTMLNode(t, res.Body, `input[name="Email"][aria-invalid="true"]`))
		assert.True(t, strings.Contains(res.Body, "Choose Android or iOS"))
		assert.Equal(t, len(app.mailer.SentMessages), 0)
	})

	t.Run("silently accepts beta honeypot submissions", func(t *testing.T) {
		app := newTestApplication(t)
		req := newTestRequest(t, http.MethodPost, "/beta")
		req.PostForm = validForm()
		req.PostForm.Set("Company", "Spam Company")

		res := send(t, req, app.routes())

		assert.Equal(t, res.StatusCode, http.StatusSeeOther)
		assert.Equal(t, len(app.mailer.SentMessages), 0)
	})

	t.Run("reports unavailable beta configuration", func(t *testing.T) {
		app := newTestApplication(t)
		app.config.beta.email = ""
		req := newTestRequest(t, http.MethodPost, "/beta")
		req.PostForm = validForm()

		res := send(t, req, app.routes())

		assert.Equal(t, res.StatusCode, http.StatusServiceUnavailable)
		assert.True(t, strings.Contains(res.Body, "temporarily unavailable"))
	})

	t.Run("fails closed without Turnstile configuration", func(t *testing.T) {
		app := newTestApplication(t)
		app.config.turnstile.secretKey = ""
		req := newTestRequest(t, http.MethodPost, "/beta")
		req.PostForm = validForm()

		res := send(t, req, app.routes())

		assert.Equal(t, res.StatusCode, http.StatusServiceUnavailable)
		assert.True(t, strings.Contains(res.Body, "Security verification is temporarily unavailable"))
		assert.Equal(t, len(app.mailer.SentMessages), 0)
	})
}

func TestSubmitSupport(t *testing.T) {
	validForm := func() url.Values {
		return url.Values{
			"Name":                  {"Alex Example"},
			"Email":                 {"alex@example.com"},
			"Message":               {"I need some help with WeekScale."},
			"cf-turnstile-response": {"support"},
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

func TestFormSubmissionRateLimit(t *testing.T) {
	app := newTestApplication(t)
	now := time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)

	for range formSubmissionLimit {
		assert.True(t, app.allowFormSubmission("support:192.0.2.1", now))
	}
	assert.True(t, !app.allowFormSubmission("support:192.0.2.1", now))
	assert.True(t, app.allowFormSubmission("beta:192.0.2.1", now))
	assert.True(t, app.allowFormSubmission("support:192.0.2.1", now.Add(formSubmissionWindow+time.Second)))
}
