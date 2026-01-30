# 🚀 Mini CI Runner

A tiny but mighty CI runner written in Go. Think GitHub Actions, but fits in your pocket.

## What's this?

Ever wanted your own CI system without the corporate overhead? This little guy clones repos, runs your build steps, and tells you if things went 💥 or ✅.

## Features

- 🏃 **Concurrent job execution** — runs multiple jobs in parallel
- 🛑 **Job cancellation** — changed your mind? cancel anytime
- 📝 **Logging** — writes to console AND file (because why choose?)
- 🧹 **Auto cleanup** — no leftover files cluttering your disk
- 🤝 **Simple REST API** — curl-friendly, no PhD required

## Quick Start

After cloning the repo, cd to it
```bash
# Build it
go build -o mini-ci-runner .

# Run it
./mini-ci-runner
```

Server starts on `http://localhost:8080` 🎉

## API

### Submit a Job

the follozing is a test repo, keep in mind it runs for a long time. but you can run multiple jobs at the same time as long as your CPU can handle them.

```bash
curl -X POST http://localhost:8080/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "repo_url":"https://github.com/stretchr/testify",
    "commit":"master",
    "steps":[
      "go mod download",
      "go test ./... ./assert ./require"
    ]
  }'
```

Returns:
```json
{"job_id": "abc-123", "status": "queued"}
```

### Check Job Status

```bash
curl http://localhost:8080/jobs/abc-123
```

### Cancel a Job

```bash
curl -X POST http://localhost:8080/jobs/abc-123/cancel
```

### Health Check

```bash
curl http://localhost:8080/health
```

## Job Statuses

| Status | Meaning |
|--------|---------|
| `queued` | Waiting in line |
| `running` | Doing the thing |
| `completed` | Success! 🎊 |
| `failed` | Something broke 😢 |
| `canceled` | You changed your mind |

## Project Structure

```
├── main.go                 # Entry point
├── internal/
│   ├── api/               # HTTP handlers
│   ├── executor/          # Git clone + run steps
│   ├── job/               # Job data model
│   ├── logger/            # Logging utility
│   ├── runner/            # Worker pool magic
│   └── store/             # In-memory storage
```

## Logs

Logs go to both your terminal and `ci-runner.log`. Example:

```
[INFO] 2026/01/30 12:00:00 Server running on :8080
[INFO] 2026/01/30 12:00:05 Job abc-123 submitted - Repo: https://github.com/you/repo
[INFO] 2026/01/30 12:00:05 Job abc-123 started
[INFO] 2026/01/30 12:00:10 Job abc-123 completed successfully
```

## Requirements

- Go 1.21+
- Git (for cloning repos)

## License

Do whatever you want with it. Seriously.

---

*Built with ☕ and mass amounts of procrastination*
*README file written with the help of copilot*
