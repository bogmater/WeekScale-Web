package main

import (
	"net/http"

	"bogmater/weekscale-web/assets"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()
	mux.NotFound(app.notFound)

	mux.Use(app.logRequest)
	mux.Use(app.recoverPanic)
	mux.Use(app.securityHeaders)
	mux.Use(app.canonicalHost)
	mux.Use(middleware.Compress(5))
	mux.Use(middleware.GetHead)

	fileServer := http.FileServer(http.FS(assets.EmbeddedFiles))
	mux.Handle("/static/*", app.cacheStaticFiles(fileServer))

	mux.Get("/", app.home)
	mux.Post("/beta", app.submitBeta)
	mux.Get("/faq", app.faq)
	mux.Get("/healthz", app.healthcheck)
	mux.Get("/privacy", app.privacy)
	mux.Get("/robots.txt", app.robots)
	mux.Get("/sitemap.xml", app.sitemap)
	mux.Get("/support", app.support)
	mux.Post("/support", app.submitSupport)

	return mux
}
