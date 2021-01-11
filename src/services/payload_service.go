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
	PayloadService payloadServiceInterface = &payloadService{}
)

type payloadServiceInterface interface {
	DoRequest(ctx context.Context, items items.Request) (*items.Response, rest_errors.RestErr)
}

type payloadService struct {
	mu sync.Mutex
}

func (s *payloadService) MakeRequest(ctx context.Context, item items.RequestItem, errChan chan error) *items.ResponseItem {
	//timeOut := time.Duration(config.Cfg.RequestTimeout)
	request, err := http.NewRequest("GET", item.Url, nil)
	if err != nil {
		errChan <- err
	}
	request.Header.Set("Content-Type", "application/json")

	client := http.Client{Timeout: 1 * time.Second}
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

func (s *payloadService) DoRequest(ctx context.Context, request items.Request) (*items.Response, rest_errors.RestErr) {
	if len(request.Items) > config.Cfg.RequestPayload {
		respErr := rest_errors.BadRequestError(fmt.Sprintf("max request qty is %d", config.Cfg.RequestPayload))
		return nil, respErr
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
			result := s.MakeRequest(ctx, req, errChan)
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

	return response, restErr
}
