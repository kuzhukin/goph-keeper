package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/kuzhukin/goph-keeper/internal/server/handler"
)

type AuthModdleware struct {
	checker UserChecker
}

func NewAuthMiddleware(c UserChecker) *AuthModdleware {
	return &AuthModdleware{
		checker: c,
	}
}

type UserChecker interface {
	Check(ctx context.Context, login, password string) error
}

func (a *AuthModdleware) Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		login := r.Header.Get("login")
		if len(login) == 0 {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		password := r.Header.Get("password")
		if len(password) == 0 {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		if err := a.checker.Check(r.Context(), login, password); err != nil {
			if errors.Is(err, handler.ErrUnknownUser) {
				w.WriteHeader(http.StatusBadRequest)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}

			return
		}

		ctxWithAuthInfo := context.WithValue(r.Context(), "auth", &handler.User{Login: login, Password: password})
		r = r.WithContext(ctxWithAuthInfo)

		h.ServeHTTP(w, r)
	})
}
