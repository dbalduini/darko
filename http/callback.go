package http

import (
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// transport is tuned for single host connection pool
var transport = &http.Transport{
	TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 100,
	IdleConnTimeout:     60 * time.Second,
	DisableCompression:  true,
}

var tuneHTTPClient = &http.Client{Transport: transport}

// PostCallback calls the callback url with the payload
func PostCallback(url string, body io.Reader) (int, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return 0, err
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := tuneHTTPClient.Do(req)
	if err != nil {
		return 0, err
	}

	defer res.Body.Close()
	io.Copy(ioutil.Discard, res.Body)
	return res.StatusCode, nil
}
