package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/YHVCorp/gitdocai-cli/schemas"
)

type RemoteClient struct {
	BaseURL        string
	AuthHeaders    map[string]string
	httpClient     *http.Client
}

func NewRemoteClient(baseURL string, auth schemas.Authentication) *RemoteClient {
	headers := map[string]string{
		"X-Organization-ID": auth.OrganizationID,
		"X-Project-ID":      auth.ProjectID,
		"X-Session-Token":   auth.SessionToken,
		"Content-Type":      "application/json",
	}

	return &RemoteClient{
		BaseURL:     baseURL,
		AuthHeaders: headers,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (rc *RemoteClient) Post(endpoint string, body interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", rc.BaseURL, endpoint), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	for k, v := range rc.AuthHeaders {
		req.Header.Set(k, v)
	}

	return rc.httpClient.Do(req)
}

func (rc *RemoteClient) Get(endpoint string) (*http.Response, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", rc.BaseURL, endpoint), nil)
	if err != nil {
		return nil, err
	}

	for k, v := range rc.AuthHeaders {
		req.Header.Set(k, v)
	}

	return rc.httpClient.Do(req)
}

func ReadBody(res *http.Response) ([]byte, error) {
	defer res.Body.Close()
	return io.ReadAll(res.Body)
}

