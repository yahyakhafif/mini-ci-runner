package store

import (
	"errors"
	"sync"

	"mini-ci-runner-go/internal/job"
)

type MemoryStore struct {
	mu   sync.Mutex
	jobs map[string]*job.Job
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		jobs: make(map[string]*job.Job),
	}
}

func (s *MemoryStore) Save(j *job.Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[j.ID] = j
}

func (s *MemoryStore) Get(id string) (*job.Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, ok := s.jobs[id]
	if !ok {
		return nil, errors.New("job not found")
	}
	return j, nil
}

func (s *MemoryStore) Update(j *job.Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[j.ID] = j
}
