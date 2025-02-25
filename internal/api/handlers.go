package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"mini-ci-runner-go/internal/job"
	"mini-ci-runner-go/internal/logger"
	"mini-ci-runner-go/internal/runner"
	"mini-ci-runner-go/internal/store"
)

type Handler struct {
	store *store.MemoryStore
	pool  *runner.WorkerPool
}

func NewHandler(store *store.MemoryStore, pool *runner.WorkerPool) *Handler {
	return &Handler{store: store, pool: pool}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/jobs", h.handleJobs)
	mux.HandleFunc("/jobs/", h.handleJobByID)
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return mux
}

func (h *Handler) handleJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RepoURL string   `json:"repo_url"`
		Commit  string   `json:"commit"`
		Steps   []string `json:"steps"`
	}

	json.NewDecoder(r.Body).Decode(&req)

	j := &job.Job{
		ID:        uuid.New().String(),
		RepoURL:   req.RepoURL,
		Commit:    req.Commit,
		Steps:     req.Steps,
		Status:    job.StatusQueued,
		CreatedAt: time.Now(),
	}

	h.store.Save(j)
	h.pool.Submit(j)
	logger.Info.Printf("Job %s submitted - Repo: %s", j.ID, j.RepoURL)

	json.NewEncoder(w).Encode(map[string]string{
		"job_id": j.ID,
		"status": string(j.Status),
	})
}

func (h *Handler) handleJobByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/jobs/")

	if strings.HasSuffix(id, "/cancel") {
		id = strings.TrimSuffix(id, "/cancel")
		if h.pool.Cancel(id) {
			logger.Info.Printf("Cancel request received for job %s", id)
			w.WriteHeader(http.StatusOK)
		} else {
			logger.Error.Printf("Cancel request failed - job %s not found or not running", id)
			w.WriteHeader(http.StatusNotFound)
		}
		return
	}

	j, err := h.store.Get(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(j)
}
