package controllers

import (
	"encoding/json"
	"github.com/implicithash/simple_gateway/src/domain/items"
	"github.com/implicithash/simple_gateway/src/services"
	"github.com/implicithash/simple_gateway/src/utils/http_utils"
	"github.com/implicithash/simple_gateway/src/utils/rest_errors"
	"io/ioutil"
	"net/http"
	"sync/atomic"
)

var (
	//PayloadController is a payload controller
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

	resultChan := make(chan services.Result)
	services.WorkerPool.Jobs <- func() {
		result := services.PayloadService.DoRequest(r.Context(), payloadRequest)
		resultChan <- <-result
	}
	msg := <-resultChan
	if msg.Error != nil {
		http_utils.RespondError(w, msg.Error)
		return
	}
	services.Limiter.Cond.L.Lock()
	//services.Limiter.Cond.Wait()
	<-services.Limiter.OutgoingQueue
	services.Limiter.Cond.L.Unlock()

	atomic.AddUint64(&services.Limiter.OutgoingCounter, 1)

	http_utils.RespondJSON(w, http.StatusCreated, msg.Response.Items)
}
