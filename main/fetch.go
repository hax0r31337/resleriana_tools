package main

import (
	"net/http"
	"net/url"
	"path/filepath"
)

var client = &http.Client{
	Transport: &http.Transport{
		DisableKeepAlives: true,
		ForceAttemptHTTP2: true,
		Proxy:             http.ProxyFromEnvironment,
	},
}

func fetch(path string) (*http.Response, error) {
	u, err := url.Parse(BASE_URL)
	if err != nil {
		return nil, err
	}
	u.Path = filepath.Join(u.Path, path)

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", "BestHTTP/2 v2.5.4")

	return client.Do(req)
}
