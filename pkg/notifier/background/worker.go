package background

import (
	"context"

	"go.uber.org/zap"
)

type Worker struct {
	queue  chan Job
	stopCh <-chan struct{}
	logger *zap.SugaredLogger
}

func NewWorker(ctx context.Context, logger *zap.SugaredLogger) chan<- Job {
	worker := &Worker{
		queue:  make(chan Job, 1),
		stopCh: ctx.Done(),
		logger: logger,
	}

	defer worker.start()

	return worker.queue
}

func (w *Worker) start() {
	go func() {
		for {
			select {
			case job, isOpened := <-w.queue:
				if !isOpened {
					return
				}

				if err := job.Execute(job); err != nil {
					w.logger.Error(err)
					return
				}

			case <-w.stopCh:
				return
			}
		}
	}()
}
