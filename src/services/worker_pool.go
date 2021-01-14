package services

import (
	"github.com/implicithash/simple_gateway/src/utils/config"
	"sync"
	"sync/atomic"
)

// Task is a task to be done
type Task func()

// Worker is a wrapper for a job queue
type Worker struct {
	Jobs chan Task
	Quit chan struct{}
}

// RateLimiter is a load balancer between incoming and outgoing requests
type RateLimiter struct {
	IncomingQueue   chan struct{}
	OutgoingQueue   chan struct{}
	IncomingCounter uint64
	OutgoingCounter uint64
	Quit            chan struct{}
	Cond            *sync.Cond
}

// NewWorker creates a new Pool
func NewWorker(maxQueueSize int) *Worker {
	return &Worker{Jobs: make(chan Task, maxQueueSize)}
}

// Run starts a worker pool
func (w *Worker) Run() {
	go func() {
		for job := range w.Jobs {
			job()
			select {
			case <-w.Quit:
				return
			default:

			}
		}
	}()
}

// Stop stops a worker pool
func (w *Worker) Stop() {
	go func() {
		w.Quit <- struct{}{}
	}()
}

// NewRateLimiter creates a rate limiter
func NewRateLimiter(incomingReqQty int, outgoingReqQty int) *RateLimiter {
	return &RateLimiter{
		IncomingQueue: make(chan struct{}, incomingReqQty),
		OutgoingQueue: make(chan struct{}, outgoingReqQty),
		Quit:          make(chan struct{}, 1),
		Cond:          sync.NewCond(&sync.Mutex{}),
	}
}

// Run starts a rate limiter
func (r *RateLimiter) Run() {
	go func() {
		for {
			select {
			case <-r.IncomingQueue:
				atomic.AddUint64(&r.IncomingCounter, 1)
				r.Cond.L.Lock()
				length := config.Cfg.OutgoingReqQty - len(r.OutgoingQueue)
				for i := 0; i < length; i++ {
					r.OutgoingQueue <- struct{}{}
					//log.Println(len(r.OutgoingQueue))
				}
				r.Cond.L.Unlock()
				r.Cond.Broadcast()
			case <-r.Quit:
				return
			default:
			}
		}
	}()
}

// Stop stops a rate limiter
func (r *RateLimiter) Stop() {
	go func() {
		r.Quit <- struct{}{}
	}()
}
