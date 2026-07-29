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

Requires Go 1.26+ (see `go.mod`).

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
*   The application can be containerised using the provided `Dockerfile` (multi-stage build on `golang:1.26`, running as non-root on `gcr.io/distroless/static-debian13:nonroot`).
*   The CI/CD pipeline is defined in `.github/workflows/google-cloudrun-docker.yml` (build, lint, unit tests, gosec, govulncheck, Trivy, deploy to Cloud Run, post-deploy health check).
