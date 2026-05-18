package runner //nolint:testpackage // white-box tests for unexported run and newLogger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// immediateTask returns nil immediately, causing the task manager to cancel its context.
type immediateTask struct{}

func (immediateTask) Name() string              { return "immediate" }
func (immediateTask) Run(context.Context) error { return nil }

// failTask returns an error immediately.
type failTask struct{ err error }

func (failTask) Name() string                { return "fail" }
func (f failTask) Run(context.Context) error { return f.err }

func TestRun_runnableCalled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := slog.New(slog.DiscardHandler)
	called := false

	err := run(ctx, logger, func(tm Runner, _ *slog.Logger) error {
		called = true
		cancel()
		return nil
	})

	assert.NoError(t, err)
	assert.True(t, called)
}

func TestRun_runnableReceivesLogger(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := slog.New(slog.DiscardHandler)
	var received *slog.Logger

	err := run(ctx, logger, func(tm Runner, l *slog.Logger) error {
		received = l
		cancel()
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, logger, received)
}

func TestRun_runnableReceivesRunner(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := slog.New(slog.DiscardHandler)
	var received Runner

	err := run(ctx, logger, func(tm Runner, _ *slog.Logger) error {
		received = tm
		cancel()
		return nil
	})

	assert.NoError(t, err)
	assert.NotNil(t, received)
}

func TestRun_success_returnsNil(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := slog.New(slog.DiscardHandler)

	err := run(ctx, logger, func(_ Runner, _ *slog.Logger) error {
		cancel()
		return nil
	})

	assert.NoError(t, err)
}

func TestRun_runnableError_returnsError(t *testing.T) {
	// context.Background() is intentional: m.Stop() handles cancellation internally.
	t.Parallel()
	logger := slog.New(slog.DiscardHandler)
	want := errors.New("something failed")

	err := run(context.Background(), logger, func(_ Runner, _ *slog.Logger) error {
		return want
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, want)
}

func TestRun_taskError_returnsError(t *testing.T) {
	// context.Background() is intentional: the failing task cancels the manager internally.
	t.Parallel()
	logger := slog.New(slog.DiscardHandler)
	want := errors.New("task exploded")

	// tm.Run returns nil (task is started); the task error surfaces via m.Wait().
	err := run(context.Background(), logger, func(tm Runner, _ *slog.Logger) error {
		return tm.Run(failTask{err: want})
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, want)
}

func TestRun_taskStartedByRunner_completesCleanly(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.DiscardHandler)

	// immediateTask returns nil, which cancels the manager context and unblocks m.Wait().
	err := run(context.Background(), logger, func(tm Runner, _ *slog.Logger) error {
		return tm.Run(immediateTask{})
	})

	assert.NoError(t, err)
}

func TestRun_signalHandlerStartupFailure_returnsError(t *testing.T) {
	t.Parallel()
	// A pre-cancelled context makes the task manager refuse to start the ossignal task.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	logger := slog.New(slog.DiscardHandler)
	runnableCalled := false

	err := run(ctx, logger, func(_ Runner, _ *slog.Logger) error {
		runnableCalled = true
		return nil
	})

	assert.Error(t, err)
	assert.False(t, runnableCalled)
}

func TestRunner_runEphemeral(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.DiscardHandler)

	err := run(context.Background(), logger, func(tm Runner, _ *slog.Logger) error {
		if err := tm.RunEphemeral(immediateTask{}); err != nil {
			return err
		}
		// immediateTask via Run cancels the manager context, unblocking m.Wait().
		return tm.Run(immediateTask{})
	})

	assert.NoError(t, err)
}

func TestRunner_cleanup(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := slog.New(slog.DiscardHandler)
	cleanupRan := false

	err := run(ctx, logger, func(tm Runner, _ *slog.Logger) error {
		tm.Cleanup(func() error { cleanupRan = true; return nil })
		cancel()
		return nil
	})

	assert.NoError(t, err)
	assert.True(t, cleanupRan)
}

func TestRunner_context(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := slog.New(slog.DiscardHandler)
	var got context.Context

	err := run(ctx, logger, func(tm Runner, _ *slog.Logger) error {
		got = tm.Context()
		cancel()
		return nil
	})

	assert.NoError(t, err)
	assert.NotNil(t, got)
}

func TestNewLogger_serviceAttribute(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := newLogger("my-service", &buf)
	logger.Info("hello")

	var entry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
	assert.Equal(t, "my-service", entry["service"])
}

func TestNewLogger_defaultLevel_filtersDebug(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := newLogger("svc", &buf)
	logger.Debug("debug message")
	assert.Empty(t, buf.String())
}

func TestNewLogger_logLevelEnv_enablesDebug(t *testing.T) {
	// Not parallel: modifies environment.
	t.Setenv("LOG_LEVEL", "debug")
	var buf bytes.Buffer
	logger := newLogger("svc", &buf)
	logger.Debug("debug message")
	assert.NotEmpty(t, buf.String())
}
