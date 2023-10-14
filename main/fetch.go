package main

import (
	"net/http"
	"net/url"
	"path/filepath"
	"time"
)

var client = &http.Client{
	Timeout: 5 * time.Second,
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

	return http.DefaultClient.Do(req)
}
