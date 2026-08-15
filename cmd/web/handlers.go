package main

import (
	"net/http"
	"strings"
	"time"

	"bogmater/weekscale-web/internal/request"
	"bogmater/weekscale-web/internal/response"
	"bogmater/weekscale-web/internal/validator"

	"github.com/tomasen/realip"
)

const (
	formSubmissionLimit  = 3
	formSubmissionWindow = time.Hour
)

type supportForm struct {
	Name      string
	Email     string
	Message   string
	Website   string
	Validator validator.Validator `form:"-"`
}

type betaForm struct {
	Email     string
	Platform  string
	Company   string
	Validator validator.Validator `form:"-"`
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data["BetaForm"] = betaForm{}
	data["BetaSent"] = r.URL.Query().Get("beta") == "sent"

	err := response.Page(w, http.StatusOK, data, "pages/home.tmpl")
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) submitBeta(w http.ResponseWriter, r *http.Request) {
	if site := r.Header.Get("Sec-Fetch-Site"); site != "" && site != "same-origin" && site != "same-site" {
		app.badRequest(w, r, errCrossSiteForm)
		return
	}

	var form betaForm
	err := request.DecodePostForm(w, r, &form)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	if form.Company != "" {
		http.Redirect(w, r, "/?beta=sent#beta", http.StatusSeeOther)
		return
	}

	form.Email = strings.TrimSpace(form.Email)
	form.Platform = strings.ToLower(strings.TrimSpace(form.Platform))

	form.Validator.CheckField(validator.IsEmail(form.Email), "Email", "Enter a valid email address")
	form.Validator.CheckField(validator.In(form.Platform, "android", "ios"), "Platform", "Choose Android or iOS")

	if form.Validator.HasErrors() {
		app.renderBetaForm(w, r, http.StatusUnprocessableEntity, form)
		return
	}

	if app.config.beta.email == "" {
		form.Validator.AddError("Beta signup is temporarily unavailable. Please try again later.")
		app.renderBetaForm(w, r, http.StatusServiceUnavailable, form)
		return
	}

	if !app.allowFormSubmission("beta:"+realip.FromRequest(r), time.Now()) {
		form.Validator.AddError("Too many signup attempts have been sent from this connection. Please try again later.")
		app.renderBetaForm(w, r, http.StatusTooManyRequests, form)
		return
	}

	emailData := app.newEmailData()
	emailData["Email"] = form.Email
	emailData["Platform"] = form.Platform

	app.backgroundTask(r, func() error {
		return app.mailer.Send(app.config.beta.email, emailData, "beta-signup.tmpl")
	})

	http.Redirect(w, r, "/?beta=sent#beta", http.StatusSeeOther)
}

func (app *application) renderBetaForm(w http.ResponseWriter, r *http.Request, status int, form betaForm) {
	data := app.newTemplateData(r)
	data["BetaForm"] = form
	data["BetaSent"] = false

	err := response.Page(w, status, data, "pages/home.tmpl")
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) healthcheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok\n"))
}

func (app *application) faq(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	err := response.Page(w, http.StatusOK, data, "pages/faq.tmpl")
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) privacy(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	err := response.Page(w, http.StatusOK, data, "pages/privacy.tmpl")
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) support(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data["Form"] = supportForm{}
	data["Sent"] = r.URL.Query().Get("sent") == "1"

	err := response.Page(w, http.StatusOK, data, "pages/support.tmpl")
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) submitSupport(w http.ResponseWriter, r *http.Request) {
	if site := r.Header.Get("Sec-Fetch-Site"); site != "" && site != "same-origin" && site != "same-site" {
		app.badRequest(w, r, errCrossSiteForm)
		return
	}

	var form supportForm
	err := request.DecodePostForm(w, r, &form)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	// Give simple form-filling bots the same response as a real submission.
	if form.Website != "" {
		http.Redirect(w, r, "/support?sent=1", http.StatusSeeOther)
		return
	}

	form.Name = strings.TrimSpace(form.Name)
	form.Email = strings.TrimSpace(form.Email)
	form.Message = strings.TrimSpace(form.Message)

	form.Validator.CheckField(validator.NotBlank(form.Name), "Name", "Enter your name")
	form.Validator.CheckField(validator.MaxRunes(form.Name, 100), "Name", "Name must be 100 characters or fewer")
	form.Validator.CheckField(validator.IsEmail(form.Email), "Email", "Enter a valid email address")
	form.Validator.CheckField(validator.MinRunes(form.Message, 10), "Message", "Message must be at least 10 characters")
	form.Validator.CheckField(validator.MaxRunes(form.Message, 4000), "Message", "Message must be 4,000 characters or fewer")

	if form.Validator.HasErrors() {
		app.renderSupportForm(w, r, http.StatusUnprocessableEntity, form)
		return
	}

	if app.config.support.email == "" {
		form.Validator.AddError("Support messaging is temporarily unavailable. Please try again later.")
		app.renderSupportForm(w, r, http.StatusServiceUnavailable, form)
		return
	}

	if !app.allowFormSubmission("support:"+realip.FromRequest(r), time.Now()) {
		form.Validator.AddError("Too many messages have been sent from this connection. Please try again later.")
		app.renderSupportForm(w, r, http.StatusTooManyRequests, form)
		return
	}

	emailData := app.newEmailData()
	emailData["Name"] = form.Name
	emailData["Email"] = form.Email
	emailData["Message"] = form.Message

	app.backgroundTask(r, func() error {
		return app.mailer.Send(app.config.support.email, emailData, "support-message.tmpl")
	})

	http.Redirect(w, r, "/support?sent=1", http.StatusSeeOther)
}

func (app *application) renderSupportForm(w http.ResponseWriter, r *http.Request, status int, form supportForm) {
	data := app.newTemplateData(r)
	data["Form"] = form
	data["Sent"] = false

	err := response.Page(w, status, data, "pages/support.tmpl")
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) allowFormSubmission(key string, now time.Time) bool {
	app.formMu.Lock()
	defer app.formMu.Unlock()

	if app.formRequests == nil {
		app.formRequests = make(map[string][]time.Time)
	}

	cutoff := now.Add(-formSubmissionWindow)
	if app.formLastCleanup.IsZero() || now.Sub(app.formLastCleanup) >= formSubmissionWindow {
		for key, requests := range app.formRequests {
			if len(requests) == 0 || requests[len(requests)-1].Before(cutoff) {
				delete(app.formRequests, key)
			}
		}
		app.formLastCleanup = now
	}

	requests := app.formRequests[key]
	firstCurrent := 0
	for firstCurrent < len(requests) && requests[firstCurrent].Before(cutoff) {
		firstCurrent++
	}
	requests = requests[firstCurrent:]

	if len(requests) >= formSubmissionLimit {
		app.formRequests[key] = requests
		return false
	}

	app.formRequests[key] = append(requests, now)
	return true
}
