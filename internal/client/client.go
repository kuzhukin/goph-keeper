package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

func (c *Client) Upload(u *storage.User, r *storage.Record) error {
	saveDataRequest := handler.SaveDataRequest{
		User: u.Login,
		Key:  r.Name,
		Data: string(r.Data),
	}

	uri := makeURI(c.hostport, server.BinaryDataEndpoint)
	headers := map[string]string{
		"Password": u.Password,
	}

	return request(uri, http.MethodPost, headers, saveDataRequest)
}

func (c *Client) Download(u *storage.User, filename string) (*storage.Record, error) {
	uri := makeURI(c.hostport, server.BinaryDataEndpoint)
	headers := map[string]string{
		"Password": u.Password,
	}

	getDataRequest := &handler.GetDataRequest{
		User: u.Login,
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

	return &storage.Record{Name: filename, Data: string(decodedData), Revision: resp.Revision}, nil
}

func (c *Client) Register(login, password string) error {
	uri := makeURI(c.hostport, server.RegisterEndpoint)

	regstrationRequest := &handler.RegistrationRequest{
		User:     login,
		Password: password,
	}

	return request(uri, http.MethodPut, map[string]string{}, regstrationRequest)
}

func (c *Client) Delete(login, password, dataKey string) error {
	uri := makeURI(c.hostport, server.BinaryDataEndpoint)

	deleteRequest := &handler.DeleteDataRequest{
		User: login,
		Key:  dataKey,
	}

	return request(uri, http.MethodDelete, map[string]string{password: password}, deleteRequest)
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

	// TODO: add user friendly errors printing
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request errors, code=%v", resp.StatusCode)
	}

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
