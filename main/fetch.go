package main

import (
    "fmt"
    "net/http"
    "net/url"
    "path"
)

func fetch(pathSegment string) (*http.Response, error) {
    u, err := url.Parse(BASE_URL)
    if err != nil {
        return nil, err
    }
    u.Path = path.Join(u.Path, pathSegment)
    
    fmt.Printf("u.String(): %s\n", u.String())

    req, err := http.NewRequest("GET", u.String(), nil)
    if err != nil {
        return nil, err
    }

    req.Header.Add("User-Agent", "BestHTTP/2 v2.5.4")

    client := &http.Client{}
    return client.Do(req)
}
