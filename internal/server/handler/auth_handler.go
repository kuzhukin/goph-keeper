package handler

import "net/http"

type Authenticator interface {
	Authenticate(user string, password string) error
}

type AuthenticationHandler struct {
	authenticator Authenticator
}

func NewAuthenticationHandler(authenticator Authenticator) *AuthenticationHandler {
	return &AuthenticationHandler{
		authenticator: authenticator,
	}
}

type AuthenticationRequest struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

func (h *AuthenticationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	req, err := readRequest[RegistrationRequest](r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.authenticator.Authenticate(req.User, req.Password); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}
