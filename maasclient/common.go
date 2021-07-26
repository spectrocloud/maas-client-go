package maasclient

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Controller struct {
	client  Client
	apiPath string
	params  Params
}

func unMarshalJson(data HTTPResponse, v interface{}) error {
	switch {
	case statusAcceptable(data.status):
		// responses are either string
		// or properly formatted json
		if v != nil {
			json.Unmarshal(data.body, v)
		}
		return nil
	case data.status >= 300:
		return fmt.Errorf("status: %d, message: %s", data.status, data.body)
	}
	return fmt.Errorf("status: %d, message: %s", data.status, data.body)
}

func statusAcceptable(status int) bool {
	return status == http.StatusOK ||
		status == http.StatusCreated ||
		status == http.StatusAccepted ||
		status == http.StatusNoContent
}
