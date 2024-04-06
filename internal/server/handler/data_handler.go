package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/kuzhukin/goph-keeper/internal/zlog"
)

var ErrDataAlreadyExist = errors.New("data alreadt exist")
var ErrInternalProblem = errors.New("storage internal error")
var ErrDataNotFound = errors.New("storage doesn't have data")
var ErrBadRevision = errors.New("bad revision error")

type Storage interface {
	Save(key []byte, data []byte) error
	Update(key []byte, data []byte, revision int) error
	Load(key []byte) ([]byte, int, error)
	Delete(key []byte) error
}

type DataHandler struct {
	storage Storage
}

func NewDataHandler(storage Storage) *DataHandler {
	return &DataHandler{storage: storage}
}

func (h *DataHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGetData(w, r)
	case http.MethodPost:
		h.handleSaveData(w, r)
	case http.MethodPut:
		h.handleUpdateData(w, r)
	case http.MethodDelete:
		h.handleDeleteData(w, r)
	default:
		zlog.Logger().Infof("unhandled method %s", r.Method)

		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

type GetDataRequest struct {
	Key []byte `json:"key"`
}

type GetDataResponse struct {
	Key      []byte `json:"key"`
	Data     []byte `json:"data"`
	Revision int    `json:"revision"`
}

func (h *DataHandler) handleGetData(w http.ResponseWriter, r *http.Request) {
	req, err := readRequest[GetDataRequest](r)
	if err != nil || len(req.Key) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data, revision, err := h.storage.Load(req.Key)
	if err != nil {
		responseStorageError(w, err)
		return
	}

	response := GetDataResponse{
		Key:      req.Key,
		Data:     data,
		Revision: revision,
	}

	if err := writeResponse[GetDataResponse](w, response); err != nil {
		zlog.Logger().Errorf("write response=%+v err=%s", response, err)
	}
}

type SaveDataRequest struct {
	Key  []byte `json:"key"`
	Data []byte `json:"data"`
}

func (h *DataHandler) handleSaveData(w http.ResponseWriter, r *http.Request) {
	req, err := readRequest[SaveDataRequest](r)
	if err != nil || len(req.Key) == 0 || len(req.Data) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.storage.Save(req.Key, req.Data); err != nil {
		responseStorageError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type UpdateDataMode int

const (
	SyncUpdateDataMode  UpdateDataMode = 1
	ForceUpdateDataMode UpdateDataMode = 2
)

type UpdateDataRequest struct {
	Key      []byte         `json:"key"`
	Data     []byte         `json:"data"`
	Revision int            `json:"revision"`
	Mode     UpdateDataMode `json:"updateDataMode,omitempty"`
}

func (h *DataHandler) handleUpdateData(w http.ResponseWriter, r *http.Request) {
	req, err := readRequest[UpdateDataRequest](r)
	if err != nil || len(req.Key) == 0 || len(req.Data) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.storage.Update(req.Key, req.Data, req.Revision); err != nil {
		responseStorageError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type DeleteDataRequest struct {
	Key []byte `json:"key"`
}

func (h *DataHandler) handleDeleteData(w http.ResponseWriter, r *http.Request) {
	req, err := readRequest[DeleteDataRequest](r)
	if err != nil || len(req.Key) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.storage.Delete(req.Key); err != nil {
		responseStorageError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func responseStorageError(w http.ResponseWriter, err error) {
	zlog.Logger().Infof("Storage err=%s", err)

	if errors.Is(err, ErrBadRevision) {
		w.WriteHeader(http.StatusConflict)
	} else if errors.Is(err, ErrDataNotFound) {
		w.WriteHeader(http.StatusNotFound)
	} else if errors.Is(err, ErrDataAlreadyExist) {
		w.WriteHeader(http.StatusConflict)
	}

	w.WriteHeader(http.StatusInternalServerError)
}

func readRequest[T any](r *http.Request) (T, error) {
	var parsedRequest T

	data, err := io.ReadAll(r.Body)
	if err != nil {
		return parsedRequest, fmt.Errorf("read all err=%w", err)
	}

	if err := json.Unmarshal(data, &parsedRequest); err != nil {
		return parsedRequest, fmt.Errorf("unmarshal data=%s err=%w", string(data), err)
	}

	return parsedRequest, nil
}

func writeResponse[T any](w http.ResponseWriter, response T) error {
	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("marshal err=%w", err)
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
	if err != nil {
		return fmt.Errorf("write data  err=%w", err)
	}

	return nil
}
