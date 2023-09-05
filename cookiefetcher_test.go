package cookiefetcher_test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/ory/dockertest/v3"
	cookiefetcher "github.com/ras0q/traq-oauth2client-cookiefetcher"
)

func TestFetchJar(t *testing.T) {
	var (
		user = mustLookupEnv(t, "USERNAME")
		pass = mustLookupEnv(t, "PASSWORD")
	)

	type args struct {
		url       *url.URL
		cookieKey string
		opt       *cookiefetcher.Option
	}
	tests := map[string]struct {
		args    args
		isError bool
	}{
		"knoq.trap.jp": {
			args: args{
				url:       mustURL(t, "http://knoq.trap.jp"),
				cookieKey: "session",
				opt: &cookiefetcher.Option{
					Context: []chromedp.ContextOption{chromedp.WithLogf(t.Logf)},
				},
			},
		},
	}

	port, err := setupBrowser(t)
	if err != nil {
		t.Fatal(err)
	}

	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			ctx, cancel := chromedp.NewRemoteAllocator(context.Background(), fmt.Sprintf("ws://localhost:%s", port))
			t.Cleanup(cancel)

			a := tt.args
			jar, err := cookiefetcher.FetchJar(ctx, a.url, a.cookieKey, user, pass, a.opt)
			if (err != nil) != tt.isError {
				t.Fatalf("wantError is %t, but got %+v", tt.isError, err)
			}

			cookies := jar.Cookies(a.url)
			if len(cookies) == 0 {
				t.Fatalf("cookies is empty")
			}
		})
	}
}

func mustLookupEnv(t *testing.T, key string) string {
	t.Helper()

	v, ok := os.LookupEnv(key)
	if !ok {
		t.Fatal("missing environment variable: " + key)
	}

	return v
}

func mustURL(t *testing.T, rawurl string) *url.URL {
	t.Helper()

	u, err := url.Parse(rawurl)
	if err != nil {
		t.Fatal(err)
	}

	return u
}

func setupBrowser(t *testing.T) (string, error) {
	t.Helper()

	pool, err := dockertest.NewPool("")
	if err != nil {
		return "", err
	}

	if err := pool.Client.Ping(); err != nil {
		return "", err
	}

	resources, err := pool.Run("chromedp/headless-shell", "latest", nil)
	if err != nil {
		return "", err
	}

	var port string
	if err := pool.Retry(func() error {
		port = resources.GetPort("9222/tcp")
		res, err := http.Get(fmt.Sprintf("http://localhost:%s/json/version", port))
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return err
		}
		return nil
	}); err != nil {
		return "", err
	}

	t.Cleanup(func() {
		if err := pool.Purge(resources); err != nil {
			t.Fatal(err)
		}
	})

	return port, nil
}
