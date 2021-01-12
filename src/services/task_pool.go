package services

// Task is a task to be done
type Task func()

// Worker is a wrapper for a job queue
type Worker struct {
	Jobs chan Task
	Quit chan struct{}
}

// NewWorker creates a new Pool
func NewWorker(maxQueueSize int) *Worker {
	return &Worker{Jobs: make(chan Task, maxQueueSize)}
}

// Run runs a worker pool
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
