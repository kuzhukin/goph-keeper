package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/kuzhukin/goph-keeper/internal/server/endpoint"
	"github.com/stretchr/testify/require"
)

func TestSecretHandlerCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSecretStorage := NewMockSecretStorage(ctrl)
	h := NewSecretDataHandler(mockSecretStorage)

	req := SaveSecretRequest{Key: "key", Value: "value"}
	data, err := json.Marshal(req)
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodPut, endpoint.WalletEndpoint, bytes.NewBuffer(data))
	r = r.WithContext(context.WithValue(r.Context(), "auth", &User{Login: "user", Password: "1234"}))
	w := httptest.NewRecorder()

	user := &User{Login: "user", Password: "1234"}
	secret := &Secret{Key: "key", Value: "value"}
	mockSecretStorage.EXPECT().CreateSecret(gomock.Any(), user, secret).Return(nil)

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestSecretHandlerGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSecretStorage := NewMockSecretStorage(ctrl)
	h := NewSecretDataHandler(mockSecretStorage)

	req := SaveSecretRequest{Key: "key", Value: "value"}
	data, err := json.Marshal(req)
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodGet, endpoint.WalletEndpoint, bytes.NewBuffer(data))
	r = r.WithContext(context.WithValue(r.Context(), "auth", &User{Login: "user", Password: "1234"}))
	w := httptest.NewRecorder()

	user := &User{Login: "user", Password: "1234"}
	secret := &Secret{Key: "key", Value: "value"}
	mockSecretStorage.EXPECT().GetSecret(gomock.Any(), user, "key").Return(secret, nil)

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)

	answer := w.Body.Bytes()
	resp := &GetSecretResponse{}
	err = json.Unmarshal(answer, resp)
	require.NoError(t, err)

	require.Equal(t, "key", resp.Key)
	require.Equal(t, "value", resp.Data)
}

func TestSecretHandlerDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSecretStorage := NewMockSecretStorage(ctrl)
	h := NewSecretDataHandler(mockSecretStorage)

	req := DeleteSecretRequest{Key: "key"}
	data, err := json.Marshal(req)
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodDelete, endpoint.WalletEndpoint, bytes.NewBuffer(data))
	r = r.WithContext(context.WithValue(r.Context(), "auth", &User{Login: "user", Password: "1234"}))
	w := httptest.NewRecorder()

	user := &User{Login: "user", Password: "1234"}
	mockSecretStorage.EXPECT().DeleteSecret(gomock.Any(), user, "key").Return(nil)

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)
}
