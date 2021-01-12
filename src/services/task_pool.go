package services

type Task func()

type Worker struct {
	Jobs chan Task
	Quit chan struct{}
}

func NewWorker(maxQueueSize int) *Worker {
	return &Worker{Jobs: make(chan Task, maxQueueSize)}
}

func (w *Worker) Run() {
	go func() {
		for job := range w.Jobs{
			job()
			select {
			case <- w.Quit:
				return
			default:

			}
		}
	}()
}

func (w *Worker) Stop() {
	go func() {
		w.Quit <- struct {}{}
	}()
}

