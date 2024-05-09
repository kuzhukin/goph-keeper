package handler

import (
	"errors"
	"net/http"

	"github.com/kuzhukin/goph-keeper/internal/zlog"
)

var ErrDataAlreadyExist = errors.New("data alreadt exist")
var ErrInternalProblem = errors.New("holder internal error")
var ErrDataNotFound = errors.New("holder doesn't have data")
var ErrBadRevision = errors.New("bad revision error")

type DataHolder interface {
	Save(key []byte, data []byte) error
	Update(key []byte, data []byte, revision int) error
	Load(key []byte) ([]byte, int, error)
	Delete(key []byte) error
}

type DataHandler struct {
	holder DataHolder
}

func NewDataHandler(holder DataHolder) *DataHandler {
	return &DataHandler{holder: holder}
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

	data, revision, err := h.holder.Load(req.Key)
	if err != nil {
		responseholderError(w, err)
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

	if err := h.holder.Save(req.Key, req.Data); err != nil {
		responseholderError(w, err)
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

	if err := h.holder.Update(req.Key, req.Data, req.Revision); err != nil {
		responseholderError(w, err)
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

	if err := h.holder.Delete(req.Key); err != nil {
		responseholderError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func responseholderError(w http.ResponseWriter, err error) {
	zlog.Logger().Infof("holder err=%s", err)

	if errors.Is(err, ErrBadRevision) {
		w.WriteHeader(http.StatusConflict)
	} else if errors.Is(err, ErrDataNotFound) {
		w.WriteHeader(http.StatusNotFound)
	} else if errors.Is(err, ErrDataAlreadyExist) {
		w.WriteHeader(http.StatusConflict)
	}

	w.WriteHeader(http.StatusInternalServerError)
}
