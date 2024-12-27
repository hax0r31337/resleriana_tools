package main

import (
	"net/http"
	"net/url"
	"path/filepath"
	"fmt"
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
	encodedPath := url.QueryEscape(path)
	u.Path = filepath.Join(u.Path, encodedPath)

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", "BestHTTP/2 v2.5.4")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("client Do Request: %s\n", resp)
		return nil, err
	}


	return resp, nil

}
