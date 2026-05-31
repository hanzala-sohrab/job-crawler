package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func main() {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.WindowSize(1920, 1080),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var requestID network.RequestID
	var responseBody string

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventResponseReceived:
			if strings.Contains(ev.Response.URL, "/jobapi/v3/search") {
				requestID = ev.RequestID
			}
		case *network.EventLoadingFinished:
			if ev.RequestID == requestID {
				go func() {
					c := chromedp.FromContext(ctx)
					body, err := network.GetResponseBody(ev.RequestID).Do(cdp.WithExecutor(ctx, c.Target))
					if err == nil {
						responseBody = string(body)
					}
				}()
			}
		}
	})

	err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.naukri.com/software-engineer-jobs"),
		chromedp.Sleep(5*time.Second),
	)

	if err != nil {
		log.Fatal(err)
	}

	if responseBody != "" {
		fmt.Printf("✅ Success! Extracted %d bytes of JSON\n", len(responseBody))
		fmt.Println(responseBody[:200] + "...")
	} else {
		fmt.Println("❌ Failed to intercept JSON")
	}
}
