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
	hostport string
	done     chan error
}

func newClient(config *config.Config) *Client {
	return &Client{
		hostport: config.Hostport,
		done:     make(chan error),
	}
}

func (c *Client) CreateFile(username, password, filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	base64Data := base64.RawStdEncoding.EncodeToString(data)

	return c.doSaveRequest(username, password, filename, base64Data)
}

func (c *Client) doSaveRequest(username string, password string, filename string, filedata string) error {
	saveDataRequest := handler.SaveDataRequest{
		User: username,
		Key:  filename,
		Data: string(filedata),
	}

	uri := makeURI(c.hostport, server.BinaryDataEndpoint)
	headers := map[string]string{
		"Password": password,
	}

	return request(uri, http.MethodPost, headers, saveDataRequest)
}

func (c *Client) GetFile(username string, password string, filename string) (*storage.File, error) {
	uri := makeURI(c.hostport, server.BinaryDataEndpoint)
	headers := map[string]string{
		"Password": password,
	}

	getDataRequest := &handler.GetDataRequest{
		User: username,
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

func (c *Client) Register(login, password string) error {
	uri := makeURI(c.hostport, server.RegisterEndpoint)

	regstrationRequest := &handler.RegistrationRequest{
		User:     login,
		Password: password,
	}

	return request(uri, http.MethodPut, map[string]string{}, regstrationRequest)
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

func makeURI(hostport string, endpoint string) string {
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
