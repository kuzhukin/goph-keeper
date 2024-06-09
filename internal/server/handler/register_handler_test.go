package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/kuzhukin/goph-keeper/internal/server/endpoint"
	"github.com/stretchr/testify/require"
)

func TestRegisterHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRegistrator := NewMockRegistrator(ctrl)
	h := NewRegistrationHandler(mockRegistrator)

	req := RegistrationRequest{User: "user", Password: "1234"}
	data, err := json.Marshal(req)
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodPut, endpoint.RegisterEndpoint, bytes.NewBuffer(data))
	w := httptest.NewRecorder()

	mockRegistrator.EXPECT().Register(gomock.Any(), req.User, req.Password).Return(nil)

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)
}
