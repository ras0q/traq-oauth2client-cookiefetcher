package cookiefetcher

import (
	"context"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func FetchJar(
	ctxWithAllocator context.Context,
	url *url.URL,
	cookieKey string,
	username string,
	password string,
	cookieOpts *cookiejar.Options,
	contextOpts []chromedp.ContextOption,
) (*cookiejar.Jar, error) {
	// ctxWithAllocator must have allocator info
	if pc := chromedp.FromContext(ctxWithAllocator); pc == nil {
		return nil, chromedp.ErrInvalidContext
	}

	ctxWithAllocator, cancel := chromedp.NewContext(ctxWithAllocator, contextOpts...)
	defer cancel()

	j, err := cookiejar.New(cookieOpts)
	if err != nil {
		return nil, err
	}

	chromedp.ListenTarget(ctxWithAllocator, func(ev interface{}) {
		switch ev := ev.(type) {
		case *page.EventJavascriptDialogOpening:
			go func(ev *page.EventJavascriptDialogOpening) {
				log.Println("accepting dialog...", ev.Message)
				if err := chromedp.Run(ctxWithAllocator, page.HandleJavaScriptDialog(true)); err != nil {
					log.Println(err)
				}
			}(ev)
		}
	})

	if err := chromedp.Run(
		ctxWithAllocator,
		chromedp.Navigate(url.String()),
		chromedp.SendKeys("input[autocomplete=\"username\"]", username, chromedp.NodeVisible),
		chromedp.SendKeys("input[autocomplete=\"current-password\"]", password, chromedp.NodeVisible),
		chromedp.Click("button[type=\"submit\"]", chromedp.NodeVisible),
	); err != nil {
		return nil, err
	}

	if _, err := chromedp.RunResponse(
		ctxWithAllocator,
		chromedp.Click("button[data-type=\"primary\"][type=\"button\"]", chromedp.NodeVisible),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var count int

		Retry:
			cookies, err := network.GetCookies().Do(ctx)
			if err != nil {
				return err
			}

			for _, cookie := range cookies {
				if cookie.Name == cookieKey {
					j.SetCookies(url, []*http.Cookie{
						{
							Name:   cookie.Name,
							Value:  cookie.Value,
							Domain: cookie.Domain,
							Path:   cookie.Path,
						},
					})
				}
			}

			if len(j.Cookies(url)) == 0 {
				count++
				if count > 10 {
					return chromedp.ErrNoResults
				}

				time.Sleep(1 * time.Second)
				log.Println("retrying...", count)
				goto Retry
			}

			return nil
		}),
	); err != nil {
		return nil, err
	}

	return j, nil
}
