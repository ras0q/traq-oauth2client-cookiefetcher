package cookiefetcher

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func FetchJar(
	ctx context.Context,
	url *url.URL,
	cookieKey string,
	username string,
	password string,
	cookieOpts *cookiejar.Options,
	contextOpts []chromedp.ContextOption,
) (*cookiejar.Jar, error) {
	ctx, cancel := chromedp.NewExecAllocator(
		ctx,
		chromedp.ExecPath("./headless-shell"),
	)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx, contextOpts...)
	defer cancel()

	j, err := cookiejar.New(cookieOpts)
	if err != nil {
		return nil, err
	}

	if err := chromedp.Run(
		ctx,
		chromedp.Navigate(url.String()),
		chromedp.Text("input[autocomplete=\"username\"]", &username, chromedp.ByQuery),
		chromedp.Text("input[autocomplete=\"current-password\"]", &password, chromedp.ByQuery),
		chromedp.Click("button[type=\"submit\"]", chromedp.ByQuery),
		chromedp.Click("button[type=\"data-type\"]", chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
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

			return nil
		}),
	); err != nil {
		return nil, err
	}

	return j, nil
}
