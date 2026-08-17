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
	mux.Get("/about", app.about)
	mux.Post("/beta", app.submitBeta)
	mux.Get("/faq", app.faq)
	mux.Get("/happy-scale-alternative", app.happyScaleAlternative)
	mux.Get("/healthz", app.healthcheck)
	mux.Get("/libra-alternative-ios", app.libraAlternativeIOS)
	mux.Get("/offline-weight-tracker-no-account", app.offlineWeightTrackerNoAccount)
	mux.Get("/private-weight-tracker", app.privateWeightTracker)
	mux.Get("/privacy", app.privacy)
	mux.Get("/robots.txt", app.robots)
	mux.Get("/sitemap.xml", app.sitemap)
	mux.Get("/support", app.support)
	mux.Post("/support", app.submitSupport)
	mux.Get("/weekly-average-weight", app.weeklyAverageWeight)
	mux.Get("/weight-tracker-without-subscription", app.weightTrackerWithoutSubscription)

	return mux
}
