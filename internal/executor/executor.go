package executor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"mini-ci-runner-go/internal/job"
)

func RunJob(ctx context.Context, j *job.Job) (string, error) {
	var logs bytes.Buffer

	dir, err := os.MkdirTemp("", "job-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)

	run := func(cmd *exec.Cmd) error {
		cmd.Stdout = &logs
		cmd.Stderr = &logs
		return cmd.Run()
	}

	if err := run(exec.CommandContext(ctx, "git", "clone", j.RepoURL, dir)); err != nil {
		return logs.String(), err
	}

	if j.Commit != "" {
		cmd := exec.CommandContext(ctx, "git", "checkout", j.Commit)
		cmd.Dir = dir
		if err := run(cmd); err != nil {
			return logs.String(), err
		}
	}

	for _, step := range j.Steps {
		cmd := exec.CommandContext(ctx, "sh", "-c", step)
		cmd.Dir = filepath.Clean(dir)
		if err := run(cmd); err != nil {
			return logs.String(), fmt.Errorf("step failed: %s", step)
		}
	}

	return logs.String(), nil
}
