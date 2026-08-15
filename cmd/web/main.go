package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"bogmater/weekscale-web/internal/env"
	"bogmater/weekscale-web/internal/smtp"
	"bogmater/weekscale-web/internal/version"

	"github.com/lmittmann/tint"
)

func main() {
	logger := slog.New(tint.NewTextHandler(os.Stdout, &tint.Options{Level: slog.LevelDebug}))

	err := run(logger)
	if err != nil {
		trace := string(debug.Stack())
		logger.Error(err.Error(), "trace", trace)
		os.Exit(1)
	}
}

type config struct {
	baseURL       string
	httpPort      int
	notifications struct {
		email string
	}
	support struct {
		email string
	}
	beta struct {
		email string
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		from     string
	}
}

type application struct {
	config          config
	logger          *slog.Logger
	mailer          *smtp.Mailer
	wg              sync.WaitGroup
	formMu          sync.Mutex
	formRequests    map[string][]time.Time
	formLastCleanup time.Time
}

func run(logger *slog.Logger) error {
	var cfg config

	cfg.baseURL = env.GetString("BASE_URL", "https://www.weekscale.net")
	cfg.httpPort = env.GetInt("HTTP_PORT", 3333)
	cfg.notifications.email = env.GetString("NOTIFICATIONS_EMAIL", "")
	cfg.support.email = env.GetString("SUPPORT_EMAIL", "")
	cfg.beta.email = env.GetString("BETA_EMAIL", "")
	cfg.smtp.host = env.GetString("SMTP_HOST", "example.smtp.host")
	cfg.smtp.port = env.GetInt("SMTP_PORT", 25)
	cfg.smtp.username = env.GetString("SMTP_USERNAME", "example_username")
	cfg.smtp.password = env.GetString("SMTP_PASSWORD", "pa55word")
	cfg.smtp.from = env.GetString("SMTP_FROM", "Example Name <no_reply@example.org>")

	showVersion := flag.Bool("version", false, "display version and exit")

	flag.Parse()

	if *showVersion {
		fmt.Printf("version: %s\n", version.Get())
		return nil
	}

	mailer, err := smtp.NewMailer(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.from)
	if err != nil {
		return err
	}

	app := &application{
		config: cfg,
		logger: logger,
		mailer: mailer,
	}

	return app.serveHTTP()
}
