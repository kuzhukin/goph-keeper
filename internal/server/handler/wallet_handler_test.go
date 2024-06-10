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

const (
	testCardNumber = "1234 1234 1234 1234"
	testCardData   = "some-cryptd-data"
)

func TestWalletHandlerAddCard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWalletStorage := NewMockWalletStorage(ctrl)
	h := NewWalletHandler(mockWalletStorage)

	req := SaveCardDataRequest{User: "user", CardNumber: testCardNumber, CardData: testCardData}
	data, err := json.Marshal(req)
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodPut, endpoint.WalletEndpoint, bytes.NewBuffer(data))
	r = r.WithContext(context.WithValue(r.Context(), "auth", &User{Login: "user", Password: "1234"}))
	w := httptest.NewRecorder()

	user := &User{Login: req.User, Password: "1234"}
	card := &CardData{Number: req.CardNumber, Data: req.CardData}
	mockWalletStorage.EXPECT().CreateCard(gomock.Any(), user, card).Return(nil)

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestWalletHandlerListCard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWalletStorage := NewMockWalletStorage(ctrl)
	h := NewWalletHandler(mockWalletStorage)

	r := httptest.NewRequest(http.MethodGet, endpoint.WalletEndpoint, nil)
	r = r.WithContext(context.WithValue(r.Context(), "auth", &User{Login: "user", Password: "1234"}))
	w := httptest.NewRecorder()

	user := &User{Login: "user", Password: "1234"}
	expectedList := []*CardData{{"1234", "xxxx"}, {"1234", "yyyy"}}
	mockWalletStorage.EXPECT().ListCard(gomock.Any(), user).Return(expectedList, nil)

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)

	resp := &GetCardsResponse{}
	err := json.Unmarshal(w.Body.Bytes(), resp)
	require.NoError(t, err)

	require.Equal(t, expectedList, resp.Cards)
}

func TestWalletHandlerDeleteCard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWalletStorage := NewMockWalletStorage(ctrl)
	h := NewWalletHandler(mockWalletStorage)

	req := DeleteCardDataRequest{User: "user", CardNumber: "1234"}
	data, err := json.Marshal(req)
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodDelete, endpoint.WalletEndpoint, bytes.NewBuffer(data))
	r = r.WithContext(context.WithValue(r.Context(), "auth", &User{Login: "user", Password: "1234"}))
	w := httptest.NewRecorder()

	user := &User{Login: req.User, Password: "1234"}
	deletedCard := &CardData{Number: "1234"}
	mockWalletStorage.EXPECT().DeleteCard(gomock.Any(), user, deletedCard).Return(nil)

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)
}
