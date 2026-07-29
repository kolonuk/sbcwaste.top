# Swindon Borough Council Waste Collection API

This service provides waste collection dates for properties in Swindon, UK. It scrapes the Swindon Borough Council website to get the latest collection information.

## Usage

The main API endpoint is `/[UPRN]/[format]`.

-   `UPRN`: The Unique Property Reference Number for your address.
-   `format`: The output format. Can be `json` (default), `ics` (iCalendar), `xml`, or `yaml`.

**Optional Parameters:**

-   `?debug=yes`: Enable debug logging (disabled when `APP_ENV=production`).
-   `?icons=yes`: Include bin icon data (`iconURL`/`iconDataURI`) in `json`/`xml`/`yaml` output. Not available for `ics`. **Currently broken in the deployed environment — see [Known Issues](#known-issues).**
-   `?skipcache=yes`: Bypass the cache and fetch fresh data (disabled when `APP_ENV=production`).
-   `?cachestats=yes`: Include cache source/age/expiry info in the response.

### Other endpoints

-   `GET /search-address?q=...`: Look up addresses/UPRNs by postcode or street name.
-   `GET /api/costs`: JSON billing/cost data backing the `costs.html` dashboard.
-   `GET /health`: Health check, returns `{"status": "ok"}`.
-   `GET /.well-known/security.txt`: Vulnerability disclosure contact info.

## Caching

To improve performance and reduce the load on the Swindon Borough Council website, this service implements a caching mechanism.

### Local Development

When running locally (`APP_ENV=development`), the service uses a local SQLite database (`sbcwaste.db`) for caching. The database file is created automatically in the root of the project directory.

### Cloud Environment

When deployed to the Google Cloud environment, the service uses Firestore for caching. The application will automatically create and use a collection named `sbcwaste_cache` in your project's default Firestore database.

### Cache Expiry

The cache expiry time is configurable. By default, it is set to 3 days (259200 seconds). You can change this value by modifying the `CACHE_EXPIRY_SECONDS` input in the GitHub Actions workflow when running it manually.

### Clearing the Cache

**Local (SQLite):**

To clear the local cache, simply delete the `sbcwaste.db` file from the project directory.

**Cloud (Firestore):**

To clear the Firestore cache in the cloud environment, you can delete the documents in the `sbcwaste_cache` collection using the Google Cloud Console:

1.  Navigate to the **Firestore** page in the Google Cloud Console.
2.  Select your project and you should see the `sbcwaste_cache` collection.
3.  You can manually delete individual documents (cached items) or delete the entire collection by clicking the three dots next to the collection name and selecting "Delete collection".

## Getting Started

1.  **Install dependencies:** `go mod tidy`
2.  **Run the application:** `go run ./src`
3.  **Run tests:** `go test ./src/...`

Requires Go 1.26+ (see `go.mod`). Set `APP_ENV=development` so the app uses
the local SQLite cache instead of trying to reach Firestore (see
[Local Development](#local-development)) — or use Docker (below), which sets
this for you.

Note: `github.com/mattn/go-sqlite3` requires cgo. Running `go run ./src` (or
`go build`) directly on the host works out of the box as long as your Go
toolchain has cgo enabled (the default) and a C compiler is available. If
`CGO_ENABLED=0` is set in your environment, SQLite calls will fail at
runtime — the production `Dockerfile` deliberately disables cgo for this
reason and only runs against Firestore (see [Docker](#docker) below).

## Docker

### Local development (SQLite)

The included `Dockerfile.dev` and `docker-compose.yml` run the app with cgo
enabled and `APP_ENV=development` set by default, so it always uses the
local SQLite cache — no GCP project or credentials required.

```sh
docker compose up --build
```

This:

-   Builds from `Dockerfile.dev` (a `golang:1.26` image with cgo enabled,
    running `go run ./src` — not the production distroless image).
-   Serves the app at [http://localhost:8080](http://localhost:8080).
-   Bind-mounts the repo into the container, so `sbcwaste.db` is created in
    the project root exactly as it would be with a plain `go run ./src` on
    the host, and code edits are picked up without rebuilding the image
    (just restart the container: `docker compose restart`).
-   Caches Go modules/build output in named volumes (`go-mod-cache`,
    `go-build-cache`) so repeated `docker compose up --build` runs don't
    re-download dependencies.

Stop it with `docker compose down` (add `-v` to also drop the module/build
cache volumes). To clear the SQLite cache, delete `sbcwaste.db` as described
above.

Note: the icon-scraping feature (`?icons=yes`) needs a Chrome/Chromium
binary that isn't installed in this image (see
[Known Issues](#known-issues)); requests without `?icons=yes` are
unaffected. BigQuery-backed billing data (`/api/costs`) also isn't
reachable without GCP credentials — it falls back to the historical CSV in
`static/data/billing_data.csv`.

### Production image

The production image (built from the root `Dockerfile`) is what CI builds
and deploys to Cloud Run, and is also mirrored to Docker Hub for `main` —
see [Notes](#notes). You generally shouldn't need to build it locally, but
`docker build -t sbcwaste .` works if you want to inspect it; it won't serve
useful data without `APP_ENV`/`PROJECT_ID`/Firestore credentials configured.

## Known Issues

-   **Bin icons (`?icons=yes`) do not work in the deployed environment.** `enrichCollectionsWithIcons` (in `src/sbcwaste.go`) shells out to a `google-chrome`/Chromium binary via `chromedp` to scrape icon URLs, but the production/dev/test container is built `FROM gcr.io/distroless/static-debian13:nonroot` (see `Dockerfile`), which intentionally has no shell, package manager, or browser binary — that base image was deliberately chosen to eliminate OS-level CVEs. As a result every icon request currently fails silently server-side:
    ```
    Could not enrich collections with icons: failed to get icon URLs via chromedp: exec: "google-chrome": executable file not found in $PATH
    ```
    (confirmed via Cloud Run logs on `sbcwaste-dev`, 2026-07-28). The response is still returned successfully, just without `iconURL`/`iconDataURI` fields.

    Fixing this requires a deliberate trade-off, not a one-line patch — options include:
    -   Switching the runtime base image to one that bundles Chromium/`chromium-headless-shell` (reintroduces a fuller OS and its CVE surface, larger image).
    -   Running icon scraping via a separate/remote headless-browser service instead of an in-process binary.
    -   Dropping the `chromedp`-based icon feature entirely (e.g. serving the icons statically, or pointing `iconURL` at swindon.gov.uk directly without scraping/re-encoding).

    This is left as a TODO pending a decision on which trade-off to make.

## Notes

*   The application uses a local SQLite database (`sbcwaste.db`) for caching in development. This file is git-ignored.
*   The compiled binary (`sbcwaste`) is also git-ignored.
*   The application uses `chromedp` for the icons feature. For **local development only**, you'll need Chrome/Chromium installed (e.g. `sudo apt install -y chromium-browser` on Ubuntu) — the deployed container does not have this, see [Known Issues](#known-issues).
*   The production image is built from the provided `Dockerfile` (multi-stage build on `golang:1.26`, running as non-root on `gcr.io/distroless/static-debian13:nonroot`). For local development, use `Dockerfile.dev` / `docker-compose.yml` instead — see [Docker](#docker).
*   The CI/CD pipeline is defined in `.github/workflows/google-cloudrun-docker.yml` (build, lint, unit tests, gosec, govulncheck, Trivy, deploy to Cloud Run, post-deploy health check, and — for `main` only, once deploy succeeds — mirroring the image to Docker Hub).
