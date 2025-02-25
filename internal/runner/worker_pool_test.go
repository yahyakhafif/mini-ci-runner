package runner

import (
	"context"
	"io"
	"log"
	"strconv"
	"sync"
	"testing"
	"time"

	"mini-ci-runner-go/internal/job"
	"mini-ci-runner-go/internal/logger"
	"mini-ci-runner-go/internal/store"
)

func init() {
	logger.Info = log.New(io.Discard, "", 0)
	logger.Error = log.New(io.Discard, "", 0)

	executorRunJob = func(ctx context.Context, j *job.Job) (string, error) {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(50 * time.Millisecond):
			return "done", nil
		}
	}
}

func TestWorkerPoolConcurrency(t *testing.T) {
	jobStore := store.NewMemoryStore()
	pool := NewWorkerPool(5, jobStore)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool.Start(ctx)

	numJobs := 20
	var wg sync.WaitGroup
	wg.Add(numJobs)

	for i := 0; i < numJobs; i++ {
		id := strconv.Itoa(i)
		j := &job.Job{
			ID:     id,
			Steps:  []string{"step1"},
			Status: job.StatusQueued,
		}
		jobStore.Save(j)
		pool.Submit(j)
		go func(jobID string) {
			defer wg.Done()
			// Poll job until completed
			for {
				j, _ := jobStore.Get(jobID)
				if j.Status == job.StatusCompleted {
					return
				}
				time.Sleep(10 * time.Millisecond)
			}
		}(id)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// all jobs completed
	case <-time.After(2 * time.Second):
		t.Fatal("jobs did not complete in time, concurrency issue")
	}
}

func TestJobCancellation(t *testing.T) {
	jobStore := store.NewMemoryStore()
	pool := NewWorkerPool(2, jobStore)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)

	j := &job.Job{
		ID:     "cancel-test",
		Steps:  []string{"step1"},
		Status: job.StatusQueued,
	}
	jobStore.Save(j)
	pool.Submit(j)

	time.Sleep(10 * time.Millisecond)

	canceled := pool.Cancel(j.ID)
	if !canceled {
		t.Fatal("failed to cancel job")
	}

	for {
		j, _ := jobStore.Get("cancel-test")
		if j.Status == job.StatusCanceled {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	j, _ = jobStore.Get("cancel-test")
	if j.Status != job.StatusCanceled {
		t.Fatalf("expected status canceled, got %s", j.Status)
	}
}

func TestConcurrentStoreAccess(t *testing.T) {
	jobStore := store.NewMemoryStore()
	numJobs := 50
	var wg sync.WaitGroup
	wg.Add(numJobs)

	for i := 0; i < numJobs; i++ {
		go func(i int) {
			defer wg.Done()
			id := strconv.Itoa(i)
			j := &job.Job{
				ID:     id,
				Steps:  []string{"step"},
				Status: job.StatusQueued,
			}
			jobStore.Save(j)
			j2, err := jobStore.Get(id)
			if err != nil || j2.ID != j.ID {
				t.Errorf("store access error: %v", err)
			}
		}(i)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("concurrent store access did not finish in time")
	}
}
