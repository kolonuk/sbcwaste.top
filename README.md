# Swindon Borough Council Waste Collection API

This service provides waste collection dates for properties in Swindon, UK. It scrapes the Swindon Borough Council website to get the latest collection information.

## Usage

The API endpoint is `/[UPRN]/[format]`.

-   `UPRN`: The Unique Property Reference Number for your address.
-   `format`: The output format. Can be `json` (default) or `ics` (for iCalendar).

**Optional Parameters:**

-   `?debug=yes`: Enable debug logging.
-   `?icons=yes`: Include icon data in the JSON output.

## Deployment

This project uses a GitHub Actions workflow to automatically build and deploy the application to Google Cloud Run. The workflow is defined in `.github/workflows/google-cloudrun-docker.yml`. The workflow can also be triggered manually from the Actions tab in GitHub.

### Branches

-   **`dev` branch:** Deploys to the development environment at `dev.sbcwaste.top`.
-   **`main` branch:** Deploys to the production environment at `sbcwaste.top`.

### GitHub Secrets

To use the deployment workflow, you must configure the following secrets in your GitHub repository settings:

-   `WIF_PROVIDER`: The full identifier of the Workload Identity Provider.
-   `WIF_SERVICE_ACCOUNT`: The email address of the service account to use.

The `scripts/deploy-githubactions.sh` script can be used to set up the necessary Google Cloud resources and will output the values for `WIF_PROVIDER` and `WIF_SERVICE_ACCOUNT`.

### Verification

After deployment, the workflow performs two verification steps:

1.  **DNS Verification:** It checks if the domain's CNAME record points to `ghs.googlehosted.com.`.
2.  **Health Check:** It sends a request to the root of the application and checks for a 200 OK response and the presence of "sbcwaste - Swindon Borough Council Waste Collection API" in the response body.

The workflow will fail if these checks do not pass after several retries.
