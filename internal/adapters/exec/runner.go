package execadapter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"sync"

	"github.com/shellcell/cnvrt/internal/ports"
)

type lookPathResult struct {
	path string
	err  error
}

type Runner struct {
	mu      sync.Mutex
	lookups map[string]lookPathResult
}

func NewRunner() *Runner {
	return &Runner{lookups: map[string]lookPathResult{}}
}

// LookPath caches results: availability checks run for every backend and
// format choice, and PATH scans are comparatively expensive.
func (r *Runner) LookPath(name string) (string, error) {
	r.mu.Lock()
	cached, ok := r.lookups[name]
	r.mu.Unlock()
	if ok {
		return cached.path, cached.err
	}

	path, err := exec.LookPath(name)
	r.mu.Lock()
	r.lookups[name] = lookPathResult{path: path, err: err}
	r.mu.Unlock()
	return path, err
}

func (r *Runner) Run(ctx context.Context, command ports.Command) (ports.CommandResult, error) {
	cmd := exec.CommandContext(ctx, command.Name, command.Args...)
	cmd.Dir = command.Dir

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := ports.CommandResult{Stdout: stdout.String(), Stderr: stderr.String()}
	if err == nil {
		return result, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		result.ExitCode = exitErr.ExitCode()
		return result, fmt.Errorf("exit code %d", result.ExitCode)
	}

	return result, err
}
