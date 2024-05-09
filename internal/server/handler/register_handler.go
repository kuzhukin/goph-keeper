package handler

import (
	"context"
	"net/http"
)

type Registrator interface {
	Register(ctx context.Context, user string, password string) error
}

type RegisterHandler struct {
	registrator Registrator
}

func NewRegistrationHandler(registrator Registrator) *RegisterHandler {
	return &RegisterHandler{
		registrator: registrator,
	}
}

type RegistrationRequest struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

func (r *RegistrationRequest) Validate() bool {
	return len(r.User) > 0 && len(r.Password) > 0
}

func (h *RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	req, err := readRequest[*RegistrationRequest](r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.registrator.Register(r.Context(), req.User, req.Password); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}
