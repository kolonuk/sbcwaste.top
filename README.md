# Swindon Borough Council Waste Collection API

This service provides waste collection dates for properties in Swindon, UK. It scrapes the Swindon Borough Council website to get the latest collection information.

## Usage

The API endpoint is `/[UPRN]/[format]`.

-   `UPRN`: The Unique Property Reference Number for your address.
-   `format`: The output format. Can be `json` (default) or `ics` (for iCalendar).

**Optional Parameters:**

-   `?debug=yes`: Enable debug logging.
-   `?icons=yes`: Include icon data in the JSON output.
