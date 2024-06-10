package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kuzhukin/goph-keeper/internal/client/config"
	"github.com/kuzhukin/goph-keeper/internal/client/storage"
	"github.com/kuzhukin/goph-keeper/internal/server/endpoint"
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

func (c *Client) UploadBinaryData(u *storage.User, r *storage.Record) error {
	saveDataRequest := handler.SaveDataRequest{
		User: u.Login,
		Key:  r.Name,
		Data: string(r.Data),
	}

	uri := makeURI(c.hostport, endpoint.BinaryDataEndpoint)
	headers := map[string]string{
		"Password": u.Password,
	}

	return requestAndHandle(uri, http.MethodPost, headers, saveDataRequest, func(r *http.Response) error {
		if r.StatusCode != http.StatusOK && r.StatusCode != http.StatusAccepted {
			return statusCodeToError(r.StatusCode)
		}

		return nil
	})
}

func (c *Client) DownloadBinaryData(u *storage.User, dataKey string) (*storage.Record, error) {
	uri := makeURI(c.hostport, endpoint.BinaryDataEndpoint)
	headers := map[string]string{
		"Password": u.Password,
	}

	getDataRequest := &handler.GetDataRequest{
		User: u.Login,
		Key:  dataKey,
	}

	resp, err := requestAndParse[handler.GetDataResponse](uri, http.MethodGet, headers, getDataRequest)
	if err != nil {
		return nil, err
	}

	return &storage.Record{Name: dataKey, Data: resp.Data, Revision: resp.Revision}, nil
}

func (c *Client) RegisterUser(login, password string) error {
	uri := makeURI(c.hostport, endpoint.RegisterEndpoint)

	regstrationRequest := &handler.RegistrationRequest{
		User:     login,
		Password: password,
	}

	return request(uri, http.MethodPut, map[string]string{}, regstrationRequest)
}

func (c *Client) DeleteBinaryData(login, password, dataKey string) error {
	uri := makeURI(c.hostport, endpoint.BinaryDataEndpoint)

	deleteRequest := &handler.DeleteDataRequest{
		User: login,
		Key:  dataKey,
	}

	return request(uri, http.MethodDelete, map[string]string{password: password}, deleteRequest)
}

func (c *Client) CreateCardData(login, password string, cardNumber string, cardData string) error {
	uri := makeURI(c.hostport, endpoint.WalletEndpoint)

	saveRequest := &handler.SaveCardDataRequest{User: login, CardNumber: cardNumber, CardData: cardData}

	return request(uri, http.MethodPut, map[string]string{password: password}, saveRequest)
}

func (c *Client) DeleteCardData(login, password string, cardNumber string) error {
	uri := makeURI(c.hostport, endpoint.WalletEndpoint)

	deleteRequest := &handler.DeleteCardDataRequest{User: login, CardNumber: cardNumber}

	return request(uri, http.MethodDelete, map[string]string{password: password}, deleteRequest)
}

func (c *Client) ListCardData(login, password string) ([]*handler.CardData, error) {
	uri := makeURI(c.hostport, endpoint.WalletEndpoint)

	getRequest := &handler.GetDataRequest{User: login}

	resp, err := requestAndParse[handler.GetCardsResponse](uri, http.MethodGet, map[string]string{password: password}, getRequest)
	if err != nil {
		return nil, err
	}

	return resp.Cards, nil
}

func (c *Client) CreateSecret(login, password string, secretName string, secretData string) error {
	uri := makeURI(c.hostport, endpoint.SecretEndpoint)

	saveRequest := &handler.SaveSecretRequest{User: login, Key: secretName, Value: secretData}

	return request(uri, http.MethodPut, map[string]string{password: password}, saveRequest)
}

func (c *Client) DeleteSecret(login, password string, secretKey string) error {
	uri := makeURI(c.hostport, endpoint.SecretEndpoint)

	saveRequest := &handler.DeleteSecretRequest{User: login, Key: secretKey}

	return request(uri, http.MethodPut, map[string]string{password: password}, saveRequest)
}

func (c *Client) GetSecret(login, password string) (*storage.Secret, error) {
	uri := makeURI(c.hostport, endpoint.SecretEndpoint)

	getRequest := &handler.GetSecretDataRequest{User: login}

	resp, err := requestAndParse[handler.GetSecretResponse](uri, http.MethodGet, map[string]string{password: password}, getRequest)
	if err != nil {
		return nil, err
	}

	return &storage.Secret{Key: resp.Key, Value: resp.Data}, nil
}

type httpResponseHandler func(*http.Response) error

func defaultHttpResponseHandler(r *http.Response) error {
	switch r.StatusCode {
	case http.StatusOK:
		return nil
	default:
		return statusCodeToError(r.StatusCode)
	}
}

func request(uri string, method string, headers map[string]string, request any) error {
	return requestAndHandle(uri, method, headers, request, defaultHttpResponseHandler)
}

func statusCodeToError(statusCode int) error {
	return fmt.Errorf("request failed code=%d error=%s", statusCode, http.StatusText(statusCode))
}

func requestAndHandle(uri string, method string, headers map[string]string, request any, handler httpResponseHandler) error {
	req, err := makeRequest(uri, method, headers, request)
	if err != nil {
		return err
	}

	resp, err := doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := handler(resp); err != nil {
		return fmt.Errorf("handler err=%w", err)
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
