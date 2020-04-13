package http

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"github.com/dbalduini/darko/shard"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

var transport = &http.Transport{
	TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 100,
	IdleConnTimeout:     60 * time.Second,
	DisableCompression:  true,
}

var tuneHTTPClient = &http.Client{Transport: transport}

// PostCallback calls the callback url with the payload
func PostCallback(job shard.Job) (int, error) {
	url := os.Getenv("CALLBACK_URL")

	buf, err := json.Marshal(&job)
	if err != nil {
		return -1, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(buf))
	if err != nil {
		return -1, err
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := tuneHTTPClient.Do(req)
	if err != nil {
		return -1, err
	}

	defer res.Body.Close()
	io.Copy(ioutil.Discard, res.Body)
	return res.StatusCode, nil
}
