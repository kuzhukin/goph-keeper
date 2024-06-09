package handler

import (
	"context"
	"net/http"

	"github.com/kuzhukin/goph-keeper/internal/zlog"
)

//go:generate mockgen -source=secret_handler.go -destination=./mock_secret_handler.go -package=handler
type SecretStorage interface {
	CreateSecret(ctx context.Context, user *User, secret *Secret) error
	GetSecret(ctx context.Context, user *User, secretKey string) (*Secret, error)
	DeleteSecret(ctx context.Context, user *User, secretKey string) error
}

type Secret struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SecretDataHandler struct {
	storage SecretStorage
}

func NewSecretDataHandler(secretStorage SecretStorage) *SecretDataHandler {
	return &SecretDataHandler{
		storage: secretStorage,
	}
}

func (h *SecretDataHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	switch r.Method {
	case http.MethodGet:
		err = h.handleGetData(w, r)
	case http.MethodPut:
		err = h.handleSaveSecret(w, r)
	case http.MethodDelete:
		err = h.handleDeleteSecret(w, r)
	default:
		zlog.Logger().Infof("unhandled method %s", r.Method)

		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	if err != nil {
		zlog.Logger().Infof("handle error: %s", err)
	}
}

type GetSecretDataRequest struct {
	User string `json:"user"`
	Key  string `json:"key"`
}

func (r *GetSecretDataRequest) Validate() bool {
	return len(r.User) > 0 && len(r.Key) > 0
}

type GetSecretResponse struct {
	Key  string `json:"key"`
	Data string `json:"data"`
}

func (h *SecretDataHandler) handleGetData(w http.ResponseWriter, r *http.Request) error {
	req, err := readRequest[*GetSecretDataRequest](r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	password := r.Header.Get("password")
	if len(password) == 0 {
		w.WriteHeader(http.StatusBadRequest)

		return nil
	}

	user := &User{Login: req.User, Password: password}

	secret, err := h.storage.GetSecret(r.Context(), user, req.Key)
	if err != nil {
		responsestorageError(w, err)
		return err
	}

	response := GetSecretResponse{
		Key:  secret.Key,
		Data: secret.Value,
	}

	if err := writeResponse(w, response); err != nil {
		return err
	}

	return nil
}

type SaveSecretRequest struct {
	User  string `json:"user"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (r *SaveSecretRequest) Validate() bool {
	return len(r.User) > 0 && len(r.Key) > 0 && len(r.Value) > 0
}

func (h *SecretDataHandler) handleSaveSecret(w http.ResponseWriter, r *http.Request) error {
	req, err := readRequest[*SaveSecretRequest](r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return err
	}

	password := r.Header.Get("password")
	if len(password) == 0 {
		w.WriteHeader(http.StatusBadRequest)

		return nil
	}

	user := &User{Login: req.User, Password: password}
	secret := &Secret{Key: req.Key, Value: req.Value}

	if err = h.storage.CreateSecret(r.Context(), user, secret); err != nil {

		responsestorageError(w, err)

		return err
	}

	w.WriteHeader(http.StatusOK)

	return nil
}

type DeleteSecretRequest struct {
	User string `json:"user"`
	Key  string `json:"key"`
}

func (r *DeleteSecretRequest) Validate() bool {
	return len(r.Key) > 0
}

func (h *SecretDataHandler) handleDeleteSecret(w http.ResponseWriter, r *http.Request) error {
	req, err := readRequest[*DeleteSecretRequest](r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return err
	}

	password := r.Header.Get("password")
	if len(password) == 0 {
		w.WriteHeader(http.StatusBadRequest)

		return nil
	}

	user := &User{Login: req.User, Password: password}

	if err := h.storage.DeleteSecret(r.Context(), user, req.Key); err != nil {
		responsestorageError(w, err)

		return err
	}

	w.WriteHeader(http.StatusOK)

	return nil
}
