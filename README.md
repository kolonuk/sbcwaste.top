# Swindon Borough Council Waste Collection API

[![Go Tests](https://github.com/418-teapot/sbc-waste-api-go/actions/workflows/google-cloudrun-docker.yml/badge.svg?branch=main&event=push&job=build)](https://github.com/418-teapot/sbc-waste-api-go/actions/workflows/google-cloudrun-docker.yml)
[![Gosec Security Scanner](https://github.com/418-teapot/sbc-waste-api-go/actions/workflows/google-cloudrun-docker.yml/badge.svg?branch=main&event=push&job=build)](https://github.com/418-teapot/sbc-waste-api-go/actions/workflows/google-cloudrun-docker.yml)
[![Trivy Vulnerability Scanner](https://github.com/418-teapot/sbc-waste-api-go/actions/workflows/google-cloudrun-docker.yml/badge.svg?branch=main&event=push&job=build)](https://github.com/418-teapot/sbc-waste-api-go/actions/workflows/google-cloudrun-docker.yml)
[![HTML Validator](https://github.com/418-teapot/sbc-waste-api-go/actions/workflows/google-cloudrun-docker.yml/badge.svg?branch=main&event=push&job=validate-html)](https://github.com/418-teapot/sbc-waste-api-go/actions/workflows/google-cloudrun-docker.yml)
[![CSS Validator](https://github.com/418-teapot/sbc-waste-api-go/actions/workflows/google-cloudrun-docker.yml/badge.svg?branch=main&event=push&job=validate-css)](https://github.com/418-teapot/sbc-waste-api-go/actions/workflows/google-cloudrun-docker.yml)
[![JavaScript Validator](https://github.com/418-teapot/sbc-waste-api-go/actions/workflows/google-cloudrun-docker.yml/badge.svg?branch=main&event=push&job=validate-js)](https://github.com/418-teapot/sbc-waste-api-go/actions/workflows/google-cloudrun-docker.yml)

This service provides waste collection dates for properties in Swindon, UK. It scrapes the Swindon Borough Council website to get the latest collection information.

## Usage

The API endpoint is `/[UPRN]/[format]`.

-   `UPRN`: The Unique Property Reference Number for your address.
-   `format`: The output format. Can be `json` (default) or `ics` (for iCalendar).

**Optional Parameters:**

-   `?debug=yes`: Enable debug logging.
-   `?icons=yes`: Include icon data in the JSON output.

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
3.  **Run tests:** `go test ./...`

## Development Environment

For a consistent development environment, you can use the provided development Docker container. This container includes all the necessary dependencies and tools for working on this project.

**Building the Development Container:**

```bash
docker build -f Dockerfile.dev -t sbc-waste-api-dev .
```

**Running the Development Container:**

```bash
docker run -it -v $(pwd):/app sbc-waste-api-dev
```

This will give you an interactive shell inside the container, with the project directory mounted at `/app`. You can then run any commands you need, such as `go test ./...` or `go run ./src`.

## Notes

*   The application uses a local SQLite database (`sbcwaste.db`) for caching in development. This file is git-ignored.
*   The compiled binary (`sbcwaste`) is also git-ignored.
*   The application uses `chromedp` for web scraping. You may need to have chromium installed. On ubuntu, you can use `sudo apt install -y chromium-browser`
*   The application can be containerised using the provided `Dockerfile`.
*   The CI/CD pipeline is defined in `.github/workflows/google-cloudrun-docker.yml`.

## Adding a new council

To add a new council, you will need to:

1.  Create a new scraper function for the council's website.
2.  Implement the `Council` interface for the new council.
3.  Add the new council to the `CouncilFactory` so that it can be selected based on postcode or other identifier.
4.  Add tests for the new council's scraper.
5.  Update the documentation to include the new council.
6.  Consider using a headless browser library like `chromedp` for interactive web scraping.

Remember to add tests for any new functionality to prevent regressions.
