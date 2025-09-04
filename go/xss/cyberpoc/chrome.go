package cyberpoc

import (
	"context"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"log"
	"time"
)

func RunChrome(ticker *time.Ticker, targetUrl string, cookies map[string]string) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	for range ticker.C {
		err := chromedp.Run(ctx,
			setCookies(targetUrl, cookies),
		)
		if err != nil {
			log.Println("chrome access error", err.Error())
			continue
		}
		log.Println("chrome access success", targetUrl)
	}
}

func setCookies(url string, cookies map[string]string) chromedp.Tasks {

	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			// create cookie expiration
			expr := cdp.TimeSinceEpoch(time.Now().Add(180 * 24 * time.Hour))
			// add cookies to chrome
			for key, value := range cookies {
				err := network.SetCookie(key, value).
					WithExpires(&expr).
					WithDomain("localhost").
					Do(ctx)
				if err != nil {
					return err
				}
			}
			return nil
		}),
		// navigate to site
		chromedp.Navigate(url),
	}
}
