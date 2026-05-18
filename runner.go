// Package runner provides boilerplate-free bootstrapping for server processes:
// logger creation, task manager setup, OS signal handling, and graceful shutdown.
package runner

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/wood-jp/task"
	"github.com/wood-jp/task/ossignal"
	"github.com/wood-jp/xerrors"
)

const exitError = 1

// Runner limits the task.Manager surface exposed to a Runnable.
type Runner interface {
	Run(tasks ...task.Task) error
	RunEphemeral(tasks ...task.Task) error
	Cleanup(f func() error)
	Context() context.Context
}

var _ Runner = (*task.Manager)(nil)

// Runnable is the server logic function signature. It sets up tasks and returns;
// the runner waits for all tasks to complete.
type Runnable func(tm Runner, logger *slog.Logger) error

// Run bootstraps a server process: creates a JSON logger tagged with serviceName,
// wires up an OS signal handler, calls runnable, then waits for all tasks to
// complete. Calls os.Exit on error; main() can be a single call.
func Run(serviceName string, runnable Runnable) {
	logger := newLogger(serviceName, os.Stdout)
	if err := run(context.Background(), logger, runnable); err != nil {
		os.Exit(exitError)
	}
}

func run(ctx context.Context, logger *slog.Logger, runnable Runnable) error {
	m := task.NewManager(task.WithLogger(logger), task.WithContext(ctx))

	if err := m.Run(ossignal.NewTask(ossignal.WithLogger(logger))); err != nil {
		logger.Error("failed to start signal handler", xerrors.Log(err))
		return err
	}

	if err := runnable(m, logger); err != nil {
		logger.Error("service failed", xerrors.Log(err))
		if stopErr := m.Stop(); stopErr != nil {
			logger.Error("shutdown failed", xerrors.Log(stopErr))
		}
		return err
	}

	if err := m.Wait(); err != nil {
		logger.Error("service exited with error", xerrors.Log(err))
		return err
	}

	return nil
}

// newLogger creates a JSON slog.Logger tagged with the given service name.
// The log level defaults to INFO and can be overridden via the LOG_LEVEL env var.
func newLogger(serviceName string, w io.Writer) *slog.Logger {
	level := slog.LevelInfo
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		var l slog.Level
		if err := l.UnmarshalText([]byte(v)); err == nil {
			level = l
		}
	}
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level})
	return slog.New(h).With(slog.String("service", serviceName))
}
