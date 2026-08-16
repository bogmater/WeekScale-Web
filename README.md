# WeekScale Web

The marketing and support website for WeekScale, a local-first mobile weight tracker focused on weekly trends rather than day-to-day noise. Public iOS and Android releases are coming soon.

**Website:** [www.weekscale.net](https://www.weekscale.net)

The site is a server-rendered Go application. HTML templates, CSS, screenshots, and email templates are embedded into the application binary. It has no client-side framework, external fonts, analytics, or advertising scripts.

## Pages

- `/` - marketing landing page
- `/about` - why WeekScale exists and where it is headed
- `/faq` - product and privacy questions
- `/weekly-average-weight` - guide to calendar-week averages and rolling averages
- `/private-weight-tracker` - guide to local-first weight tracking and data privacy
- `/privacy` - iOS, Android, and website privacy details
- `/support` - spam-resistant support contact form

## Run locally

```bash
go mod tidy
go run ./cmd/web
```

Then open [http://localhost:3333](http://localhost:3333).

The support and beta forms can be viewed without SMTP configuration. Sending a message requires the SMTP and recipient variables below.

## Configuration

| Variable | Default | Purpose |
|---|---|---|
| `BASE_URL` | `https://www.weekscale.net` | Canonical origin used by metadata, sitemap, and email data |
| `HTTP_PORT` | `3333` | HTTP listener port |
| `SUPPORT_EMAIL` | empty | Private recipient for support messages |
| `BETA_EMAIL` | empty | Private recipient for beta signup messages |
| `TURNSTILE_SITE_KEY` | empty | Public Cloudflare Turnstile widget key |
| `TURNSTILE_SECRET_KEY` | empty | Private Cloudflare Turnstile verification key |
| `SMTP_HOST` | example value | SMTP server hostname |
| `SMTP_PORT` | `25` | SMTP server port |
| `SMTP_USERNAME` | example value | SMTP login username |
| `SMTP_PASSWORD` | example value | SMTP login password |
| `SMTP_FROM` | example value | Sender name and address |
| `NOTIFICATIONS_EMAIL` | empty | Optional recipient for application error reports |

Example:

```bash
export SUPPORT_EMAIL="support@example.com"
export BETA_EMAIL="beta@example.com"
export TURNSTILE_SITE_KEY="your-turnstile-site-key"
export TURNSTILE_SECRET_KEY="your-turnstile-secret-key"
export SMTP_HOST="smtp.resend.com"
export SMTP_PORT="587"
export SMTP_USERNAME="resend"
export SMTP_PASSWORD="re_your_resend_api_key"
export SMTP_FROM="WeekScale <hello@weekscale.net>"
go run ./cmd/web
```

Never commit production SMTP credentials.

## Form safeguards

- The recipient address remains server-side and is never rendered into public HTML.
- Input is length-limited, validated, and escaped by the email templates.
- Cross-site browser submissions are rejected.
- A hidden honeypot absorbs simple form-filling bots.
- Cloudflare Turnstile tokens are verified server-side for the expected hostname and form action.
- Support and beta submissions are independently limited to three per IP address per hour.
- Email delivery runs through the application's managed background-task mechanism.

The rate limit is held in process memory and applies independently to each running application instance. Add an edge-level rate limit or privacy-preserving challenge if public abuse requires stronger protection.

## Search and sharing

- Canonical URLs, social metadata, and sitemap entries use `BASE_URL`.
- `/robots.txt` allows public crawling and advertises `/sitemap.xml`.
- The landing page includes WebSite and SoftwareApplication metadata with one-time USD and EUR offers.
- The FAQ includes FAQPage, Question, and Answer metadata.
- Error pages and the post-submission support state are marked `noindex`.
- Static assets use versioned URLs and immutable caching.

After the production site is reachable, verify `https://www.weekscale.net` in Google Search Console and Bing Webmaster Tools, submit `/sitemap.xml`, and run Lighthouse against the deployed origin.

## Development

```bash
make test
make build
```

The complete quality gate is available as `make audit`. The app uses Go's standard templates and an embedded filesystem declared in `assets/efs.go`.

## Deploy with Coolify

Create an **Application** resource from the Git repository and select the **Dockerfile** build pack.

Use these settings:

| Coolify setting | Value |
|---|---|
| Base directory | `/` when this project is the repository root, otherwise `/WeekScale-Web` |
| Dockerfile location | `/Dockerfile` |
| Port exposes | `3333` |
| Domain | `https://www.weekscale.net` |
| Health check path | `/healthz` |
| Health check port | `3333` |

The Docker build uses Coolify's `SOURCE_COMMIT` when available. On Coolify versions that do not expose that option, it automatically derives a revision from the embedded assets instead. No additional build setting is required for static-asset cache invalidation.

Set these runtime environment variables:

```text
BASE_URL=https://www.weekscale.net
HTTP_PORT=3333
SUPPORT_EMAIL=your-private-support-address
BETA_EMAIL=your-private-beta-address
TURNSTILE_SITE_KEY=your-turnstile-site-key
TURNSTILE_SECRET_KEY=your-turnstile-secret-key
SMTP_HOST=smtp.resend.com
SMTP_PORT=587
SMTP_USERNAME=resend
SMTP_PASSWORD=re_your_resend_api_key
SMTP_FROM=WeekScale <hello@weekscale.net>
```

Keep SMTP variables and `TURNSTILE_SECRET_KEY` runtime-only in Coolify. `TURNSTILE_SITE_KEY` is intentionally public. `NOTIFICATIONS_EMAIL` is optional and receives server-error reports when configured. No persistent volume or database is required.

Create one Cloudflare Turnstile widget for `www.weekscale.net`. Both forms use that widget with separate `beta` and `support` action names, which are validated by the server. The forms fail closed with a temporary-unavailable response when either Turnstile key is missing.

Point both the `www` and bare DNS records to the deployment before issuing certificates for both hostnames. The application permanently redirects recognized HTTP and bare-domain requests directly to the `BASE_URL` origin while preserving the path and query string.

Coolify or another edge proxy may redirect HTTP to HTTPS before a request reaches the application. Do not combine a same-host HTTP redirect with a separate bare-to-`www` redirect, because that creates two hops. Configure the edge to send every non-canonical variant directly to `https://www.weekscale.net` with status `301` or `308`. If Cloudflare proxies the domain, create the bare-host redirect there before enabling its general HTTPS redirect; HTTP requests for `www` can use a direct permanent HTTPS redirect.
