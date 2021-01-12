package http_utils

import (
	"encoding/json"
	"github.com/implicithash/simple_gateway/src/utils/rest_errors"
	"net/http"
)

// RespondJSON is a response in json format
func RespondJSON(w http.ResponseWriter, statusCode int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(body)
}

// RespondError responds with an error
func RespondError(w http.ResponseWriter, err rest_errors.RestErr) {
	RespondJSON(w, err.Status(), err)
}
