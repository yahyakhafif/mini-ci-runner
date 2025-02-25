package runner

import (
	"context"
	"sync"

	"mini-ci-runner-go/internal/executor"
	"mini-ci-runner-go/internal/job"
	"mini-ci-runner-go/internal/logger"
	"mini-ci-runner-go/internal/store"
)

var executorRunJob = func(ctx context.Context, j *job.Job) (string, error) {
	return executor.RunJob(ctx, j)
}

type WorkerPool struct {
	numWorkers int
	jobsCh     chan *job.Job
	store      *store.MemoryStore
	cancelFunc sync.Map
}

func NewWorkerPool(workers int, store *store.MemoryStore) *WorkerPool {
	return &WorkerPool{
		numWorkers: workers,
		jobsCh:     make(chan *job.Job),
		store:      store,
	}
}

func (p *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < p.numWorkers; i++ {
		go p.worker(ctx)
	}
}

func (p *WorkerPool) Submit(j *job.Job) {
	p.jobsCh <- j
}

func (p *WorkerPool) Cancel(jobID string) bool {
	if v, ok := p.cancelFunc.Load(jobID); ok {
		v.(context.CancelFunc)()
		return true
	}
	return false
}

func (p *WorkerPool) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case j := <-p.jobsCh:
			jobCtx, cancel := context.WithCancel(ctx)
			p.cancelFunc.Store(j.ID, cancel)

			j.Status = job.StatusRunning
			p.store.Update(j)
			logger.Info.Printf("Job %s started - Repo: %s, Commit: %s", j.ID, j.RepoURL, j.Commit)

			logs, err := executorRunJob(jobCtx, j)

			p.cancelFunc.Delete(j.ID)

			j.Logs = logs

			if err != nil {
				if jobCtx.Err() == context.Canceled {
					j.Status = job.StatusCanceled
					j.Error = "job canceled"
					logger.Info.Printf("Job %s cancelled successfully", j.ID)
				} else {
					j.Status = job.StatusFailed
					j.Error = err.Error()
					logger.Error.Printf("Job %s failed: %s", j.ID, err.Error())
				}
			} else {
				j.Status = job.StatusCompleted
				logger.Info.Printf("Job %s completed successfully", j.ID)
			}

			p.store.Update(j)
		}
	}
}
