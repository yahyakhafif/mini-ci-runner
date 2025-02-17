package job

import "time"

type Status string

const (
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusCanceled  Status = "canceled"
)

type Job struct {
	ID        string
	RepoURL   string
	Commit    string
	Steps     []string
	Status    Status
	Logs      string
	Error     string
	CreatedAt time.Time
}
