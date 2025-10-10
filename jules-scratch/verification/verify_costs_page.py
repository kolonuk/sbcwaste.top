from playwright.sync_api import sync_playwright, expect

def run_verification():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        try:
            page.goto("http://localhost:8080/static/costs.html", wait_until="networkidle")

            # Wait for the "Loading costs..." message to disappear
            loading_cell = page.locator("td", has_text="Loading costs...")
            expect(loading_cell).to_have_count(0, timeout=15000) # Increased timeout for API call

            # Take a screenshot of the final state
            page.screenshot(path="jules-scratch/verification/verification-costs-dynamic.png")

        except Exception as e:
            print(f"An error occurred: {e}")
            # In case of error, still take a screenshot for debugging
            page.screenshot(path="jules-scratch/verification/verification-costs-error.png")

        finally:
            browser.close()

if __name__ == "__main__":
    run_verification()