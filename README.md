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
