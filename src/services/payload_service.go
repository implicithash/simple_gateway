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
	PayloadService payloadServiceInterface = &payloadService{
		//Cond: sync.NewCond(&sync.Mutex{}),
	}
	// WorkerPool is a job queue
	WorkerPool *Worker
	Limiter    *RateLimiter
)

type payloadServiceInterface interface {
	DoRequest(ctx context.Context, items items.Request) <-chan Result
}

type payloadService struct {
	mu sync.Mutex
	Cond *sync.Cond
}

// Result is http result
type Result struct {
	Error    rest_errors.RestErr
	Response *items.Response
}

func (s *payloadService) apiRequest(ctx context.Context, item items.RequestItem, errChan chan error) *items.ResponseItem {
	request, err := http.NewRequest(http.MethodGet, item.URL, nil)
	if err != nil {
		errChan <- err
	}
	request.Header.Set("Content-Type", "application/json")

	timeOut := time.Duration(config.Cfg.RequestTimeout) * time.Millisecond
	client := http.Client{Timeout: timeOut}
	resp, err := client.Do(request)
	if err != nil {
		errChan <- err
		//return nil
	}
	if resp == nil {
		errChan <- errors.New("HTTP request timeout")
		//return nil
	}
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)

	jsonStr, err := json.Marshal(result)
	if err != nil {
		errChan <- err
		//return nil
	}
	response := &items.ResponseItem{Data: string(jsonStr)}

	return response
}

// DoRequest performs a payload request
func (s *payloadService) DoRequest(ctx context.Context, request items.Request) <-chan Result {
	resultChan := make(chan Result, 1)
	if len(request.Items) > config.Cfg.RequestPayload {
		respErr := rest_errors.BadRequestError(fmt.Sprintf("max request qty is %d", config.Cfg.RequestPayload))
		resultChan <- Result{Response: nil, Error: respErr}
	}
	var reqItems []*items.ResponseItem
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	errChan := make(chan error, 1)

	// One incoming request gives a green light to four outgoing ones
	Limiter.IncomingQueue <- struct{}{}

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
				//s.Cond.L.Lock()
				//resultChan <- Result{Response: nil, Error: restErr}
				//s.Cond.L.Unlock()
				//s.Cond.Signal()
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
	resultChan <- Result{Response: response, Error: restErr}
	return resultChan
}
