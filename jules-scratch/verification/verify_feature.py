import os
from playwright.sync_api import sync_playwright, expect

def run():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        # Listen for all console events and print them
        page.on("console", lambda msg: print(f"Browser console: {msg.text}"))

        file_path = os.path.abspath('static/index.html')
        page.goto(f'file://{file_path}')

        uprn_input = page.locator('#uprn-ics-input')

        # Fill the input field
        uprn_input.fill('100010943963')

        # Explicitly click the generate button to trigger the link generation
        page.click('#generate-ics-btn')

        # Add a small delay to see if it helps
        page.wait_for_timeout(1000)

        try:
            # Use web-first assertions to wait for the elements to be visible
            copy_button = page.locator('#copy-ics-btn')
            expect(copy_button).to_be_visible(timeout=5000)

            google_button = page.locator('.calendar-btn.google-btn')
            expect(google_button).to_be_visible()

            # Take the screenshot
            page.screenshot(path='jules-scratch/verification/verification.png')
            print("Screenshot taken successfully.")
        except Exception as e:
            print(f"An error occurred: {e}")
            print("Page content:")
            print(page.content())

        browser.close()

if __name__ == '__main__':
    run()