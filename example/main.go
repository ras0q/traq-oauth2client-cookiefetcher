package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/chromedp/chromedp"
	cookiefetcher "github.com/ras0q/traq-oauth2client-cookiefetcher"
)

const serverBaseURL = "http://knoq.trap.jp"

var (
	username = mustLookupEnv("USERNAME")
	password = mustLookupEnv("PASSWORD")
)

func main() {
	serverURL, err := url.Parse(serverBaseURL)
	panicOnError(err)

	ctx, cancel := chromedp.NewRemoteAllocator(context.Background(), "ws://browser:9222")
	defer cancel()

	log.Println("fetching cookiejar...")
	jar, err := cookiefetcher.FetchJar(
		ctx,
		serverURL,
		"session",
		username,
		password,
		nil,
		[]chromedp.ContextOption{
			chromedp.WithLogf(log.Printf),
		},
	)
	panicOnError(err)
	log.Println("cookiejar fetched successfully", jar)

	client := http.Client{
		Jar: jar,
	}
	req, err := http.NewRequest(http.MethodGet, serverBaseURL+"/me", nil)
	panicOnError(err)

	resp, err := client.Do(req)
	panicOnError(err)

	if resp.StatusCode != http.StatusOK {
		panic("unexpected status code: " + resp.Status)
	}
}

func mustLookupEnv(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		panic("missing environment variable: " + key)
	}

	return v
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
