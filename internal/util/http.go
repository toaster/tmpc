package util

import (
	"fmt"
	"net/http"

	"golang.org/x/net/html"
)

// HTTPGet performs an HTTP GET with the TMPC user agent.
func HTTPGet(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "tmpc/0.1")
	return http.DefaultClient.Do(req)
}

// HTTPGetHTML performs an HTTP GET with the TMPC user agent and returns the response as parsed HTML.
func HTTPGetHTML(url string) (*html.Node, error) {
	res, err := HTTPGet(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP GET failed from %s: %w", url, err)
	}
	defer res.Body.Close()
	doc, err := html.Parse(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML answer from %s: %w", url, err)
	}
	return doc, nil
}
