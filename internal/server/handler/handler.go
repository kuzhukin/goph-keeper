package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

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
