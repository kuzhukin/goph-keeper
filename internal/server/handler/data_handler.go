package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/kuzhukin/goph-keeper/internal/zlog"
)

var ErrDataAlreadyExist = errors.New("data alreadt exist")
var ErrInternalProblem = errors.New("storage internal error")
var ErrDataNotFound = errors.New("storage doesn't have data")
var ErrBadRevision = errors.New("bad revision error")

type Storage interface {
	Save(ctx context.Context, user, key string, data string) error
	Update(ctx context.Context, user, key string, data string, revision uint64) error
	Load(ctx context.Context, user, key string) (string, uint64, error)
	Delete(ctx context.Context, user, key string) error
}

type DataHandler struct {
	storage Storage
}

func NewDataHandler(storage Storage) *DataHandler {
	return &DataHandler{storage: storage}
}

func (h *DataHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	switch r.Method {
	case http.MethodGet:
		err = h.handleGetData(w, r)
	case http.MethodPost:
		err = h.handleSaveData(w, r)
	case http.MethodPut:
		err = h.handleUpdateData(w, r)
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

type GetDataRequest struct {
	User string `json:"user"`
	Key  string `json:"key"`
}

func (r *GetDataRequest) Validate() bool {
	return len(r.User) > 0 && len(r.Key) > 0
}

type GetDataResponse struct {
	Key      string `json:"key"`
	Data     string `json:"data"`
	Revision uint64 `json:"revision"`
}

func (h *DataHandler) handleGetData(w http.ResponseWriter, r *http.Request) error {
	req, err := readRequest[*GetDataRequest](r)
	if err != nil || len(req.Key) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	zlog.Logger().Debugf("111 %v", req)

	data, revision, err := h.storage.Load(r.Context(), req.User, req.Key)
	if err != nil {
		responsestorageError(w, err)
		return err
	}

	response := GetDataResponse{
		Key:      req.Key,
		Data:     data,
		Revision: revision,
	}

	if err := writeResponse(w, response); err != nil {
		return err
	}

	return nil
}

type SaveDataRequest struct {
	User string `json:"user"`
	Key  string `json:"key"`
	Data string `json:"data"`
}

func (r *SaveDataRequest) Validate() bool {
	return len(r.User) > 0 && len(r.Key) > 0 && len(r.Data) > 0
}

func (h *DataHandler) handleSaveData(w http.ResponseWriter, r *http.Request) error {
	req, err := readRequest[*SaveDataRequest](r)
	if err != nil || len(req.Key) == 0 || len(req.Data) == 0 {
		w.WriteHeader(http.StatusBadRequest)

		return err
	}

	zlog.Logger().Infof("handleSaveData req=%v", req)

	if err := h.storage.Save(r.Context(), req.User, req.Key, req.Data); err != nil {
		responsestorageError(w, err)

		return err
	}

	w.WriteHeader(http.StatusOK)

	return nil
}

type UpdateDataMode int

const (
	SyncUpdateDataMode  UpdateDataMode = 1
	ForceUpdateDataMode UpdateDataMode = 2
)

type UpdateDataRequest struct {
	User     string         `json:"user"`
	Key      string         `json:"key"`
	Data     string         `json:"data"`
	Revision uint64         `json:"revision"`
	Mode     UpdateDataMode `json:"updateDataMode,omitempty"`
}

func (r *UpdateDataRequest) Validate() bool {
	switch r.Mode {
	case SyncUpdateDataMode, ForceUpdateDataMode:
	default:
		r.Mode = SyncUpdateDataMode
	}

	return len(r.Key) > 0 && len(r.Data) > 0 && r.Revision != 0
}

func (h *DataHandler) handleUpdateData(w http.ResponseWriter, r *http.Request) error {
	req, err := readRequest[*UpdateDataRequest](r)
	if err != nil || len(req.Key) == 0 || len(req.Data) == 0 {
		w.WriteHeader(http.StatusBadRequest)

		return err
	}

	if err := h.storage.Update(r.Context(), req.User, req.Key, req.Data, req.Revision); err != nil {
		responsestorageError(w, err)

		return err
	}

	w.WriteHeader(http.StatusOK)

	return nil
}

type DeleteDataRequest struct {
	User string `json:"user"`
	Key  string `json:"key"`
}

func (r *DeleteDataRequest) Validate() bool {
	return len(r.Key) > 0
}

func (h *DataHandler) handleDeleteData(w http.ResponseWriter, r *http.Request) error {
	req, err := readRequest[*DeleteDataRequest](r)
	if err != nil || len(req.Key) == 0 {
		w.WriteHeader(http.StatusBadRequest)

		return err
	}

	if err := h.storage.Delete(r.Context(), req.User, req.Key); err != nil {
		responsestorageError(w, err)

		return err
	}

	w.WriteHeader(http.StatusOK)

	return nil
}

func responsestorageError(w http.ResponseWriter, err error) {
	zlog.Logger().Infof("storage err=%s", err)

	if errors.Is(err, ErrBadRevision) {
		w.WriteHeader(http.StatusConflict)
	} else if errors.Is(err, ErrDataNotFound) {
		w.WriteHeader(http.StatusNotFound)
	} else if errors.Is(err, ErrDataAlreadyExist) {
		w.WriteHeader(http.StatusConflict)
	}

	w.WriteHeader(http.StatusInternalServerError)
}
