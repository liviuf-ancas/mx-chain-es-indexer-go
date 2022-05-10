package rest

import (
	"encoding/json"
	"net/http"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("restClient")

type genericAPIResponse struct {
	Data  json.RawMessage `json:"data"`
	Error string          `json:"error"`
	Code  string          `json:"code"`
}

type restClient struct {
	httpClient *http.Client
	url        string
}

// NewRestClient will create a new instance of restClient
func NewRestClient(url string) (*restClient, error) {
	c := http.DefaultClient

	return &restClient{
		httpClient: c,
		url:        url,
	}, nil
}

// CallGetRestEndPoint calls an external end point (sends a get request)
func (rc *restClient) CallGetRestEndPoint(
	path string,
	value interface{},
) error {
	req, err := http.NewRequest("GET", rc.url+path, nil)
	if err != nil {
		return err
	}

	userAgent := "Accounts manager>"
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)

TryAgain:
	resp, err := rc.httpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode == 429 {
		_ = resp.Body.Close()
		time.Sleep(100 * time.Millisecond)
		goto TryAgain
	}

	defer func() {
		errNotCritical := resp.Body.Close()
		if errNotCritical != nil {
			log.Warn("restClient.CallGetRestEndPoint: close body", "error", errNotCritical.Error())
		}
	}()

	err = json.NewDecoder(resp.Body).Decode(value)
	if err != nil {
		return err
	}

	return nil
}
