package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {
	// create allocator options
	allocatorOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoSandbox,
	)

	// create context
	allocatorCtx, cancel := chromedp.NewExecAllocator(context.Background(), allocatorOpts...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocatorCtx)
	defer cancel()

	// create a timeout
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	log.Println("Navigating to page...")
	// navigate to a page, wait for an element, and retrieve the outer HTML
	var html string
	err := chromedp.Run(ctx,
		chromedp.Navigate(`https://www.swindon.gov.uk/info/20122/rubbish_and_recycling_collection_days?addressList=10008541132&uprnSubmit=Yes`),
		// wait for the next collection date to be visible, which indicates the dynamic content has loaded
		chromedp.WaitVisible(`span.nextCollectionDate`, chromedp.ByQuery),
		// retrieve the full HTML of the page
		chromedp.OuterHTML("html", &html),
	)
	if err != nil {
		log.Fatal(err)
	}

	// write the HTML to a file
	err = os.WriteFile("testdata/sbc_response.html", []byte(html), 0644)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Successfully captured HTML to testdata/sbc_response.html")
}