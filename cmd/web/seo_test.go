package main

import (
	"net/http"
	"strings"
	"testing"

	"bogmater/weekscale-web/internal/assert"
)

func TestSEOMetadata(t *testing.T) {
	t.Run("home has canonical social and application metadata", func(t *testing.T) {
		app := newTestApplication(t)
		req := newTestRequest(t, http.MethodGet, "/")

		res := send(t, req, app.routes())

		assert.Equal(t, res.StatusCode, http.StatusOK)
		assert.Equal(t, res.Header.Get("Content-Type"), "text/html; charset=utf-8")
		assert.True(t, containsHTMLNode(t, res.Body, `link[rel="canonical"][href="https://www.weekscale.net/"]`))
		assert.True(t, containsHTMLNode(t, res.Body, `meta[property="og:image"][content="https://www.weekscale.net/static/img/weekscale-social.png"]`))
		assert.True(t, containsHTMLNode(t, res.Body, `[itemtype="https://schema.org/SoftwareApplication"]`))
		assert.True(t, containsHTMLNode(t, res.Body, `[itemtype="https://schema.org/Offer"] [itemprop="priceCurrency"][content="USD"]`))
		assert.True(t, containsHTMLNode(t, res.Body, `[itemtype="https://schema.org/Offer"] [itemprop="priceCurrency"][content="EUR"]`))
	})

	t.Run("canonical excludes support success query", func(t *testing.T) {
		app := newTestApplication(t)
		req := newTestRequest(t, http.MethodGet, "/support?sent=1")

		res := send(t, req, app.routes())

		assert.True(t, containsHTMLNode(t, res.Body, `link[rel="canonical"][href="https://www.weekscale.net/support"]`))
		assert.True(t, containsHTMLNode(t, res.Body, `meta[name="robots"][content="noindex,follow"]`))
	})

	t.Run("canonical excludes beta success query", func(t *testing.T) {
		app := newTestApplication(t)
		req := newTestRequest(t, http.MethodGet, "/?beta=sent")

		res := send(t, req, app.routes())

		assert.True(t, containsHTMLNode(t, res.Body, `link[rel="canonical"][href="https://www.weekscale.net/"]`))
		assert.True(t, containsHTMLNode(t, res.Body, `meta[name="robots"][content="noindex,follow"]`))
		assert.True(t, containsHTMLNode(t, res.Body, `.beta-success`))
	})

	t.Run("FAQ exposes question metadata", func(t *testing.T) {
		app := newTestApplication(t)
		req := newTestRequest(t, http.MethodGet, "/faq")

		res := send(t, req, app.routes())

		assert.True(t, containsHTMLNode(t, res.Body, `[itemtype="https://schema.org/FAQPage"]`))
		assert.True(t, containsHTMLNode(t, res.Body, `[itemtype="https://schema.org/Question"] [itemtype="https://schema.org/Answer"] [itemprop="text"]`))
	})

	t.Run("error pages are not indexable", func(t *testing.T) {
		app := newTestApplication(t)
		req := newTestRequest(t, http.MethodGet, "/missing")

		res := send(t, req, app.routes())

		assert.Equal(t, res.StatusCode, http.StatusNotFound)
		assert.True(t, containsHTMLNode(t, res.Body, `meta[name="robots"][content="noindex,follow"]`))
	})
}

func TestRobots(t *testing.T) {
	app := newTestApplication(t)
	req := newTestRequest(t, http.MethodGet, "/robots.txt")

	res := send(t, req, app.routes())

	assert.Equal(t, res.StatusCode, http.StatusOK)
	assert.Equal(t, res.Header.Get("Content-Type"), "text/plain; charset=utf-8")
	assert.True(t, strings.Contains(res.Body, "User-agent: *\nAllow: /"))
	assert.True(t, strings.Contains(res.Body, "Sitemap: https://www.weekscale.net/sitemap.xml"))
}

func TestSitemap(t *testing.T) {
	app := newTestApplication(t)
	req := newTestRequest(t, http.MethodGet, "/sitemap.xml")

	res := send(t, req, app.routes())

	assert.Equal(t, res.StatusCode, http.StatusOK)
	assert.Equal(t, res.Header.Get("Content-Type"), "application/xml; charset=utf-8")
	assert.True(t, strings.Contains(res.Body, `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`))
	for _, location := range []string{
		"https://www.weekscale.net/",
		"https://www.weekscale.net/faq",
		"https://www.weekscale.net/private-weight-tracker",
		"https://www.weekscale.net/privacy",
		"https://www.weekscale.net/support",
		"https://www.weekscale.net/weekly-average-weight",
	} {
		assert.True(t, strings.Contains(res.Body, "<loc>"+location+"</loc>"))
	}
}
