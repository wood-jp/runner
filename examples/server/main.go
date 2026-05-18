// Package main demonstrates a server process using runner.
// It runs a poll task that logs a heartbeat every 5 seconds.
// Send SIGINT or SIGTERM (Ctrl+C) to trigger graceful shutdown.
package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/wood-jp/task/poll"

	"github.com/wood-jp/runner"
)

const serviceName = "example-server"

func main() {
	runner.Run(serviceName, run)
}

func run(tm runner.Runner, logger *slog.Logger) error {
	heartbeatTask := poll.NewTask(
		func(ctx context.Context) error {
			logger.Info("heartbeat")
			return nil
		},
		"heartbeat-task",
		5*time.Second,
		poll.WithRunAtStart(),
		poll.WithLogger(logger),
		poll.WithContinueOnError(),
	)

	if err := tm.Run(heartbeatTask); err != nil {
		return err
	}

	return nil
}
