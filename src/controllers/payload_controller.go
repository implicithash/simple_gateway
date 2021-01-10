package controllers

import (
	"encoding/json"
	"github.com/implicithash/simple_gateway/src/domain/items"
	"github.com/implicithash/simple_gateway/src/services"
	"github.com/implicithash/simple_gateway/src/utils/http_utils"
	"github.com/implicithash/simple_gateway/src/utils/rest_errors"
	"io/ioutil"
	"net/http"
)

var (
	PayloadController payloadControllerInterface = &payloadController{}
)

type payloadControllerInterface interface {
	DoRequest(w http.ResponseWriter, r *http.Request)
}

type payloadController struct {
}

func (p *payloadController) DoRequest(w http.ResponseWriter, r *http.Request) {
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respErr := rest_errors.BadRequestError("invalid request body")
		http_utils.RespondError(w, respErr)
		return
	}
	defer r.Body.Close()

	var payloadRequest items.Request
	if err := json.Unmarshal(requestBody, &payloadRequest); err != nil {
		respErr := rest_errors.BadRequestError("invalid item json body")
		http_utils.RespondError(w, respErr)
		return
	}

	result, respErr := services.PayloadService.DoRequest(r.Context(), payloadRequest)
	if respErr != nil {
		http_utils.RespondError(w, respErr)
		return
	}
	http_utils.RespondJson(w, http.StatusCreated, result)
}
