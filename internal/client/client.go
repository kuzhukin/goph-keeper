package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/kuzhukin/goph-keeper/internal/client/config"
	"github.com/kuzhukin/goph-keeper/internal/client/storage"
	"github.com/kuzhukin/goph-keeper/internal/server"
	"github.com/kuzhukin/goph-keeper/internal/server/handler"
)

type Client struct {
	userName string
	password string
	hostport string
	done     chan error
}

func newClient(userName string, password string, config *config.Config) *Client {
	cl := &Client{
		userName: userName,
		password: password,
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

	return c.doSaveRequest(filename, base64Data)
}

func (c *Client) doSaveRequest(filename string, filedata string) error {
	saveDataRequest := handler.SaveDataRequest{
		User: c.userName,
		Key:  filename,
		Data: string(filedata),
	}

	uri := makeDataCtrlURI(c.hostport, server.DataEndpoint)
	headers := map[string]string{
		"Password": c.password,
	}

	return request(uri, http.MethodPost, headers, saveDataRequest)
}

func (c *Client) GetFile(filename string) (*storage.File, error) {
	uri := makeDataCtrlURI(c.hostport, server.DataEndpoint)
	headers := map[string]string{
		"Password": c.password,
	}

	getDataRequest := &handler.GetDataRequest{
		User: c.userName,
		Key:  filename,
	}

	resp, err := requestAndParse[handler.GetDataResponse](uri, http.MethodGet, headers, getDataRequest)
	if err != nil {
		return nil, err
	}

	decodedData, err := base64.RawStdEncoding.DecodeString(resp.Data)

	if err != nil {
		return nil, err
	}

	resp.Data = string(decodedData)

	return &storage.File{Name: filename, Data: string(decodedData), Revision: resp.Revision}, nil
}

func request(uri string, method string, headers map[string]string, request any) error {
	req, err := makeRequest(uri, method, headers, request)
	if err != nil {
		return err
	}

	resp, err := doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func requestAndParse[T any](uri string, method string, headers map[string]string, request any) (*T, error) {
	req, err := makeRequest(uri, method, headers, request)
	if err != nil {
		return nil, err
	}

	resp, err := doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse[T](resp)
}

func parseResponse[T any](resp *http.Response) (*T, error) {
	bin, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	obj := new(T)

	err = json.Unmarshal(bin, obj)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func makeDataCtrlURI(hostport string, endpoint string) string {
	return hostport + endpoint
}

func makeRequest(
	uri string,
	method string,
	additionalHeaders map[string]string,
	msg any,
) (*http.Request, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, uri, bytes.NewReader(data))
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

func doRequest(req *http.Request) (*http.Response, error) {
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
			return resp, nil
		}
	}

	return nil, fmt.Errorf("request error: %w", err)
}
