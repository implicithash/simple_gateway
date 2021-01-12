package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/implicithash/simple_gateway/src/domain/items"
	"github.com/implicithash/simple_gateway/src/utils/config"
	"github.com/implicithash/simple_gateway/src/utils/rest_errors"
	"net/http"
	"sync"
	"time"
)

var (
	// PayloadService contains a gateway logic
	PayloadService payloadServiceInterface = &payloadService{}
	// WorkerPool is a job queue
	WorkerPool *Worker
)

type payloadServiceInterface interface {
	DoRequest(ctx context.Context, items items.Request) <-chan Result
}

type payloadService struct {
	mu sync.Mutex
}

// Result is http result
type Result struct {
	Error    rest_errors.RestErr
	Response *items.Response
}

func (s *payloadService) apiRequest(ctx context.Context, item items.RequestItem, errChan chan error) *items.ResponseItem {
	request, err := http.NewRequest("GET", item.URL, nil)
	if err != nil {
		errChan <- err
	}
	request.Header.Set("Content-Type", "application/json")

	timeOut := time.Duration(config.Cfg.RequestTimeout) * time.Millisecond
	client := http.Client{Timeout: timeOut}
	resp, err := client.Do(request)
	if err != nil {
		errChan <- err
	}
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	jsonStr, err := json.Marshal(result)
	if err != nil {
		errChan <- err
	}
	response := &items.ResponseItem{Data: string(jsonStr)}

	return response
}

func (s *payloadService) DoRequest(ctx context.Context, request items.Request) <-chan Result {
	result := make(chan Result, 1)
	if len(request.Items) > config.Cfg.RequestPayload {
		respErr := rest_errors.BadRequestError(fmt.Sprintf("max request qty is %d", config.Cfg.RequestPayload))
		result <- Result{Response: nil, Error: respErr}
	}
	var reqItems []*items.ResponseItem
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	errChan := make(chan error, 1)

	wg := &sync.WaitGroup{}
	wg.Add(len(request.Items))
	var restErr rest_errors.RestErr
	for _, req := range request.Items {
		go func() {
			defer wg.Done()
			result := s.apiRequest(ctx, req, errChan)
			s.mu.Lock()
			reqItems = append(reqItems, result)
			s.mu.Unlock()
			select {
			case err := <-errChan:
				if err != nil {
					restErr = rest_errors.InternalServerError(fmt.Sprintf("error when trying to perform a request"), err)
				}
				cancel()
				return
			case <-ctx.Done():
				errChan <- errors.New("HTTP request cancelled")
				return
			default:
			}
		}()
	}
	wg.Wait()

	response := &items.Response{Items: reqItems}
	result <- Result{Response: response, Error: restErr}
	return result
}
