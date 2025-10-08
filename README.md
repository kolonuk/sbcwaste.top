# Swindon Borough Council Waste Collection API

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

When deployed to the Google Cloud environment, the service uses Memorystore for Memcached for caching. The Memcached instance is provisioned automatically as part of the GitHub Actions deployment workflow.

### Cache Expiry

The cache expiry time is configurable. By default, it is set to 3 days (259200 seconds). You can change this value by modifying the `CACHE_EXPIRY_SECONDS` input in the GitHub Actions workflow when running it manually.

### Clearing the Cache

**Local (SQLite):**

To clear the local cache, simply delete the `sbcwaste.db` file from the project directory.

**Cloud (Memcached):**

To clear the Memcached cache in the cloud environment, you can use the `gcloud` command-line tool. There is currently no option to do this from the Google Cloud Console GUI.

1.  Make sure you have the `gcloud` CLI installed and authenticated.
2.  Run the following command:
    ```sh
    gcloud memcache instances update [YOUR_MEMCACHED_INSTANCE_NAME] --project=[YOUR_PROJECT_ID] --region=[YOUR_REGION] --clear-cache
    ```
