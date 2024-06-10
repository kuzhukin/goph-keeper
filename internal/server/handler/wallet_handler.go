package handler

import (
	"context"
	"net/http"

	"github.com/kuzhukin/goph-keeper/internal/zlog"
)

type CardData struct {
	Number string
	Data   string
}

//go:generate mockgen -source=wallet_handler.go -destination=./mock_wallet_handler.go -package=handler
type WalletStorage interface {
	CreateCard(ctx context.Context, u *User, d *CardData) error
	ListCard(ctx context.Context, u *User) ([]*CardData, error)
	DeleteCard(ctx context.Context, u *User, d *CardData) error
}

type WalletHandler struct {
	storage WalletStorage
}

func NewWalletHandler(storage WalletStorage) *WalletHandler {
	return &WalletHandler{storage: storage}
}

func (h *WalletHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	switch r.Method {
	case http.MethodGet:
		err = h.handleGetData(w, r)
	case http.MethodPut:
		err = h.handleSaveData(w, r)
	case http.MethodDelete:
		err = h.handleDeleteData(w, r)
	default:
		zlog.Logger().Infof("unhandled method %s", r.Method)

		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	if err != nil {
		zlog.Logger().Infof("handle error: %s", err)
	}
}

type GetCardsResponse struct {
	Cards []*CardData `json:"cards"`
}

func (h *WalletHandler) handleGetData(w http.ResponseWriter, r *http.Request) error {
	user := getUserFromRequestContext(r)

	cards, err := h.storage.ListCard(r.Context(), user)
	if err != nil {
		responseErrorCardDataRequest(w, err)

		return err
	}

	response := GetCardsResponse{
		Cards: cards,
	}

	if err := writeResponse(w, response); err != nil {
		return err
	}

	return nil
}

type SaveCardDataRequest struct {
	User       string `json:"user"`
	CardNumber string `json:"number"`
	CardData   string `json:"data"`
}

func (r *SaveCardDataRequest) Validate() bool {
	return len(r.User) > 0 && len(r.CardNumber) > 0 && len(r.CardData) > 0
}

func (h *WalletHandler) handleSaveData(w http.ResponseWriter, r *http.Request) error {
	req, err := readRequest[*SaveCardDataRequest](r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return err
	}

	user := getUserFromRequestContext(r)

	card := &CardData{Number: req.CardNumber, Data: req.CardData}

	if err := h.storage.CreateCard(r.Context(), user, card); err != nil {

		responseErrorCardDataRequest(w, err)

		return err
	}

	w.WriteHeader(http.StatusOK)

	return nil
}

type DeleteCardDataRequest struct {
	User       string `json:"user"`
	CardNumber string `json:"card_number"`
}

func (r *DeleteCardDataRequest) Validate() bool {
	return len(r.User) > 0 && len(r.CardNumber) > 0
}

func (h *WalletHandler) handleDeleteData(w http.ResponseWriter, r *http.Request) error {
	req, err := readRequest[*DeleteCardDataRequest](r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return err
	}

	user := getUserFromRequestContext(r)

	data := &CardData{Number: req.CardNumber}

	if err := h.storage.DeleteCard(r.Context(), user, data); err != nil {
		responseErrorCardDataRequest(w, err)

		return err
	}

	w.WriteHeader(http.StatusOK)

	return nil
}

func responseErrorCardDataRequest(w http.ResponseWriter, err error) {
	zlog.Logger().Infof("storage err=%s", err)

	w.WriteHeader(http.StatusInternalServerError)
}
