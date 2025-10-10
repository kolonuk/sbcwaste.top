# Bristol City Council Bin Collection Integration

This document outlines the steps and considerations for integrating Bristol City Council's bin collection schedule into the `sbcwaste` application.

## 1. Council Website

*   **Bin Collection Information Page:** [https://www.bristol.gov.uk/residents/bins-and-recycling/bins-and-recycling-collection-dates](https://www.bristol.gov.uk/residents/bins-and-recycling/bins-and-recycling-collection-dates)
*   **Collection Day Finder:** [https://waste.bristol.gov.uk/](https://waste.bristol.gov.uk/)

## 2. Address Lookup Process

The process is interactive and requires automation to retrieve collection schedules.

1.  **Enter Postcode:** The user must enter a postcode on the collection day finder page.
    *   *Example Postcode:* `BS1 4RW`
2.  **Select Address:** After submitting the postcode, a list of addresses is presented in a dropdown menu. The user must select their specific address.
3.  **View Collections:** The bin collection schedule, including a downloadable calendar, is then displayed for the selected address.

Because of this multi-step, interactive process, a direct link to a collection schedule page is not feasible. The closest would be a link to the results page after an address has been selected, which would look something like this (this is a hypothetical example as the session is required): `https://waste.bristol.gov.uk/collections/10001093128` where `10001093128` is likely a UPRN or internal property identifier.

## 3. Suggested Application Changes

The changes required for Bristol are structurally identical to those needed for Manchester City Council, reinforcing the need for a flexible, multi-council architecture.

### a. Council Determination

An external API or a local lookup mechanism is needed to identify Bristol City Council as the responsible authority for postcodes starting with "BS".

### b. Data Scraping

A headless browser solution is required to automate the address lookup.

*   **Recommended Go Library:** `chromedp`
*   **Scraping Steps:**
    1.  Navigate to `https://waste.bristol.gov.uk/`.
    2.  Enter the postcode into the input field and submit the form.
    3.  Wait for the address selection dropdown to be populated.
    4.  Select the correct address from the dropdown list based on a UPRN or other property identifier.
    5.  The page will then display the collection schedule.
    6.  Parse the HTML of the results page to extract bin types, collection dates, and the link to the downloadable calendar.

### c. Code Refactoring for Multi-Council Support

The `Council` interface proposed in the Manchester documentation should be used.

1.  **`BristolCouncil` Implementation:** Create a new struct that implements the `Council` interface for Bristol.

    ```go
    package main

    import (
        // ... other imports
        "github.com/chromedp/chromedp"
    )

    // BristolCouncil implements the Council interface for Bristol City Council.
    type BristolCouncil struct {
        // ... fields for headless browser context, etc.
    }

    func (b *BristolCouncil) GetCollections(uprn string) ([]Collection, error) {
        // Implement the headless browser logic here for Bristol's website.
        // 1. Navigate to the collection finder.
        // 2. Enter postcode.
        // 3. Select address by UPRN.
        // 4. Scrape data.
        // 5. Return []Collection or an error.
        return nil, nil // Placeholder
    }

    func (b *BristolCouncil) CouncilName() string {
        return "Bristol City Council"
    }
    ```

2.  **Update Council Factory/Registry:** The factory function should be updated to include Bristol.

    ```go
    package main

    import "strings"

    // GetCouncilForPostcode returns the appropriate Council implementation for a given postcode.
    func GetCouncilForPostcode(postcode string) (Council, error) {
        // This is a simplified lookup. A more robust solution is needed for production.
        if strings.HasPrefix(postcode, "M") {
            return &ManchesterCouncil{}, nil
        }
        if strings.HasPrefix(postcode, "BS") {
            return &BristolCouncil{}, nil
        }

        // Fallback to the default sbcwaste implementation.
        return &SBCWasteCouncil{}, nil
    }
    ```

By following this pattern, the application can be extended to support numerous councils, each with its own unique scraping logic encapsulated within its `Council` implementation.