package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
)

type sitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Location string `xml:"loc"`
}

func (app *application) robots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=86400")

	fmt.Fprintf(w, "User-agent: *\nAllow: /\n\nSitemap: %s/sitemap.xml\n", app.siteURL())
}

func (app *application) sitemap(w http.ResponseWriter, r *http.Request) {
	paths := []string{"/", "/about", "/faq", "/happy-scale-alternative", "/libra-alternative-ios", "/offline-weight-tracker-no-account", "/private-weight-tracker", "/privacy", "/support", "/weekly-average-weight", "/weight-tracker-without-subscription"}
	urls := make([]sitemapURL, 0, len(paths))
	for _, path := range paths {
		urls = append(urls, sitemapURL{Location: app.siteURL() + path})
	}

	data, err := xml.MarshalIndent(sitemapURLSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}, "", "  ")
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write([]byte(xml.Header))
	w.Write(data)
}

func (app *application) siteURL() string {
	return strings.TrimRight(app.config.baseURL, "/")
}
