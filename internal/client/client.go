package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/kuzhukin/goph-keeper/internal/client/config"
	"github.com/kuzhukin/goph-keeper/internal/server"
	"github.com/kuzhukin/goph-keeper/internal/server/handler"
)

type Client struct {
	userName string
	password string
	hostport string
	done     chan error
}

func newClient(config *config.Config) *Client {
	cl := &Client{
		hostport: config.Hostport,
		done:     make(chan error),
	}

	return cl
}

func (c *Client) CreateFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	base64Data := base64.RawStdEncoding.EncodeToString(data)

	return c.send(filename, base64Data)
}

func (c *Client) send(filename string, filedata string) error {
	saveDataRequest := handler.SaveDataRequest{
		User: c.userName,
		Key:  filename,
		Data: string(filedata),
	}

	uri := createURI(c.hostport, server.DataEndpoint)
	headers := map[string]string{
		"Password": c.password,
	}

	req, err := makeRequest(uri, headers, saveDataRequest)
	if err != nil {
		return err
	}

	return doRequest(req)
}

func createURI(hostport string, endpoint string) string {
	return hostport + endpoint
}

func makeRequest(
	uri string,
	additionalHeaders map[string]string,
	msg any,
) (*http.Request, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	for header, v := range additionalHeaders {
		req.Header.Set(header, v)
	}

	return req, nil
}

var tryingIntervals []time.Duration = []time.Duration{
	time.Millisecond * 100,
	time.Millisecond * 300,
	time.Millisecond * 500,
}

func doRequest(req *http.Request) error {
	maxTryingsNum := len(tryingIntervals)

	var err error

	for trying := 0; trying <= maxTryingsNum; trying++ {
		var resp *http.Response

		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			if trying < maxTryingsNum {
				time.Sleep(tryingIntervals[trying])
			}
		} else {
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("request failed with statusCode=%d", resp.StatusCode)
			}

			return nil
		}
	}

	return fmt.Errorf("request error: %w", err)
}
