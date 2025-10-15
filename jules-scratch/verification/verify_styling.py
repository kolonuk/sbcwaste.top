import asyncio
from playwright.async_api import async_playwright

async def main():
    async with async_playwright() as p:
        browser = await p.chromium.launch()
        page = await browser.new_page()

        try:
            # Verify the main page
            await page.goto("http://localhost:8080")
            await page.screenshot(path="jules-scratch/verification/main_page.png")

            # Verify the costs page
            await page.goto("http://localhost:8080/costs.html")
            await page.screenshot(path="jules-scratch/verification/costs_page.png")

        finally:
            await browser.close()

if __name__ == "__main__":
    asyncio.run(main())