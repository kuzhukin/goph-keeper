package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kuzhukin/goph-keeper/internal/server/config"
	"github.com/kuzhukin/goph-keeper/internal/server/controller"
	"github.com/kuzhukin/goph-keeper/internal/server/handler"
	"github.com/kuzhukin/goph-keeper/internal/server/middleware"
	"github.com/kuzhukin/goph-keeper/internal/zlog"
)

// endpoints
const (
	// POST - registrer new user
	RegisterEndpoint = "/api/user/register"
	// POST - auth user
	AuthEndpoint = "/api/user/auth"
	// PUT - load new data to storage
	// POST - update binary data to storage
	// GET - get binary data from storage
	// DELETE - delete item from storage
	DataEndpoint = "/api/data"
)

type Server struct {
	httpServer http.Server
	controller *controller.Controller

	wait chan struct{}
}

func StartNew(config *config.Config) (*Server, error) {
	controller, err := controller.StartNewController(config.DataSourceName)
	if err != nil {
		return nil, fmt.Errorf("start db controller, err=%w", err)
	}

	router := chi.NewRouter()

	router.Use(middleware.LoggingHTTPHandler)

	router.Handle(DataEndpoint, handler.NewDataHandler(controller))
	router.Handle(RegisterEndpoint, handler.NewRegistrationHandler(controller))
	router.Handle(AuthEndpoint, handler.NewAuthenticationHandler(controller))

	server := &Server{
		httpServer: http.Server{Addr: config.Hostport, Handler: router},
		controller: controller,
		wait:       make(chan struct{}),
	}

	server.start()

	return server, nil
}

func (s *Server) start() {
	go func() {
		defer close(s.wait)

		err := s.httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			zlog.Logger().Errorf("server failed with err=%s", err)

			return
		}
	}()
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	return s.httpServer.Shutdown(ctx)
}

func (s *Server) WaitStop() <-chan struct{} {
	return s.wait
}
