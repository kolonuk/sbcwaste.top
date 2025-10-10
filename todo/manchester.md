# Manchester City Council Bin Collection Integration

This document outlines the steps and considerations for integrating Manchester City Council's bin collection schedule into the `sbcwaste` application.

## 1. Council Website

*   **Bin Collection Page:** [https://www.manchester.gov.uk/bincollections](https://www.manchester.gov.uk/bincollections)
*   **Portal URL:** The above link redirects to a Verint/Empro portal: `https://manchester.portal.uk.empro.verintcloudservices.com/site/myaccount/request/sr_bin_coll_day_checker`

## 2. Address Lookup Process

The process is interactive and cannot be deep-linked directly to a results page. It requires user interaction, which will need to be automated.

1.  **Enter Postcode:** The user must first enter a postcode.
    *   *Example Postcode:* `M2 5DB` (Manchester Town Hall)
2.  **Select Address:** After submitting the postcode, a list of addresses is presented. The user must select their specific address from this list.
3.  **View Collections:** The bin collection schedule is then displayed for the selected address.

Due to this multi-step process, a direct link to an example collection schedule is not possible.

## 3. Suggested Application Changes

Integrating Manchester City Council will require significant changes to the existing application architecture.

### a. Council Determination

A mechanism is needed to determine the responsible council from a given postcode. This could be achieved through:

*   An external API that provides council information based on a postcode.
*   A local database or lookup file that maps postcodes or postcode prefixes to councils.

### b. Data Scraping

The current `http.Get` approach is insufficient. A headless browser solution will be necessary to automate the address lookup process.

*   **Recommended Go Library:** `chromedp` is a good candidate for this.
*   **Scraping Steps:**
    1.  Navigate to the portal URL.
    2.  Input the postcode into the search field and trigger the search.
    3.  Wait for the address selection dropdown/list to be populated.
    4.  Select the correct address from the list. This may require presenting the options to the user if a UPRN is not sufficient to uniquely identify an address.
    5.  Once the address is selected, wait for the collection information to be displayed.
    6.  Parse the resulting HTML to extract the bin types, collection dates, and any other relevant information.

### c. Code Refactoring for Multi-Council Support

To support multiple councils, the application should be refactored to use an interface-based design.

1.  **`Council` Interface:** Define an interface that abstracts the collection retrieval process.

    ```go
    package main

    // Council represents a local council's bin collection service.
    type Council interface {
        // GetCollections retrieves bin collection data for a given address identifier (e.g., UPRN).
        GetCollections(identifier string) ([]Collection, error)
        // CouncilName returns the name of the council.
        CouncilName() string
    }
    ```

2.  **`ManchesterCouncil` Implementation:** Create a struct that implements the `Council` interface for Manchester.

    ```go
    package main

    import (
        // ... other imports
        "github.com/chromedp/chromedp"
    )

    // ManchesterCouncil implements the Council interface for Manchester City Council.
    type ManchesterCouncil struct {
        // ... fields for headless browser context, etc.
    }

    func (m *ManchesterCouncil) GetCollections(uprn string) ([]Collection, error) {
        // Implement the headless browser logic here.
        // 1. Navigate to the portal.
        // 2. Enter postcode derived from UPRN or another identifier.
        // 3. Select address.
        // 4. Scrape data.
        // 5. Return []Collection or an error.
        return nil, nil // Placeholder
    }

    func (m *ManchesterCouncil) CouncilName() string {
        return "Manchester City Council"
    }
    ```

3.  **Council Factory/Registry:** Create a mechanism to select the correct `Council` implementation based on the postcode.

    ```go
    package main

    // GetCouncilForPostcode returns the appropriate Council implementation for a given postcode.
    func GetCouncilForPostcode(postcode string) (Council, error) {
        // In a real implementation, this would use a lookup service.
        if strings.HasPrefix(postcode, "M") {
            return &ManchesterCouncil{}, nil
        }
        // Add other councils here...
        // ... else if strings.HasPrefix(postcode, "BS") { ... }

        // Fallback to the default sbcwaste implementation if no other council matches.
        return &SBCWasteCouncil{}, nil // Assuming the original logic is refactored into its own Council implementation.
    }
    ```

This approach will allow for the addition of new councils without modifying the core application logic, simply by adding a new `Council` implementation and updating the factory function.