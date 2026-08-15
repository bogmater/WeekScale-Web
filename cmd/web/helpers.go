package main

import (
	"fmt"
	"net/http"

	"bogmater/weekscale-web/internal/version"
)

func (app *application) newTemplateData(r *http.Request) map[string]any {
	siteURL := app.siteURL()
	assetVersion := version.GetRevision()
	if assetVersion == "unavailable" {
		assetVersion = version.Get()
	}

	data := map[string]any{
		"CanonicalURL":   siteURL + r.URL.Path,
		"CurrentPath":    r.URL.Path,
		"SiteURL":        siteURL,
		"SocialImageURL": siteURL + "/static/img/weekscale-social.png",
		"Version":        assetVersion,
	}

	return data
}

func (app *application) newEmailData() map[string]any {
	data := map[string]any{
		"BaseURL": app.config.baseURL,
	}

	return data
}

func (app *application) backgroundTask(r *http.Request, fn func() error) {
	app.wg.Go(func() {
		defer func() {
			pv := recover()
			if pv != nil {
				app.reportServerError(r, fmt.Errorf("%v", pv))
			}
		}()

		err := fn()
		if err != nil {
			app.reportServerError(r, err)
		}
	})
}
