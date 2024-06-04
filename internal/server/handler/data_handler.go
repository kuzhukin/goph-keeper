package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/kuzhukin/goph-keeper/internal/zlog"
)

var (
	ErrDataAlreadyExist = errors.New("data alreadt exist")
	ErrInternalProblem  = errors.New("storage internal error")
	ErrDataNotFound     = errors.New("storage doesn't have data")
	ErrBadRevision      = errors.New("bad revision error")
	ErrUnknownUser      = errors.New("unknown user")
	ErrBadPassword      = errors.New("bad password")
)

type Storage interface {
	Save(ctx context.Context, u *User, d *Data) error
	Update(ctx context.Context, u *User, d *Data) error
	Load(ctx context.Context, u *User, d string) (*Data, error)
	Delete(ctx context.Context, u *User, d *Data) error
}

type User struct {
	Login    string
	Password string
}

type Data struct {
	Key      string
	Data     string
	Revision uint64
	Metainfo string
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

	password := r.Header.Get("password")
	if len(password) == 0 {
		w.WriteHeader(http.StatusBadRequest)

		return nil
	}

	user := &User{Login: req.User, Password: password}

	data, err := h.storage.Load(r.Context(), user, req.Key)
	if err != nil {
		responsestorageError(w, err)
		return err
	}

	response := GetDataResponse{
		Key:      data.Key,
		Data:     data.Data,
		Revision: data.Revision,
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

	password := r.Header.Get("password")
	if len(password) == 0 {
		w.WriteHeader(http.StatusBadRequest)

		return nil
	}

	user := &User{Login: req.User, Password: password}
	data := &Data{Key: req.Key, Data: req.Data}

	if err := h.storage.Save(r.Context(), user, data); err != nil {

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

	password := r.Header.Get("password")
	if len(password) == 0 {
		w.WriteHeader(http.StatusBadRequest)

		return nil
	}

	user := &User{Login: req.User, Password: password}
	data := &Data{Key: req.Key, Data: req.Data, Revision: req.Revision}

	if err := h.storage.Update(r.Context(), user, data); err != nil {
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

	password := r.Header.Get("password")
	if len(password) == 0 {
		w.WriteHeader(http.StatusBadRequest)

		return nil
	}

	user := &User{Login: req.User, Password: password}
	data := &Data{Key: req.Key}

	if err := h.storage.Delete(r.Context(), user, data); err != nil {
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
