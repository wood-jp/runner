# runner

<!-- badges -->
[![Go Version](https://img.shields.io/github/go-mod/go-version/wood-jp/runner)](https://pkg.go.dev/github.com/wood-jp/runner)
[![CI](https://github.com/wood-jp/runner/actions/workflows/ci.yml/badge.svg)](https://github.com/wood-jp/runner/actions/workflows/ci.yml)
[![Coverage Status](https://coveralls.io/repos/github/wood-jp/runner/badge.svg?branch=main)](https://coveralls.io/github/wood-jp/runner?branch=main)
[![Release](https://img.shields.io/github/v/release/wood-jp/runner)](https://github.com/wood-jp/runner/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/wood-jp/runner)](https://goreportcard.com/report/github.com/wood-jp/runner)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/wood-jp/runner.svg)](https://pkg.go.dev/github.com/wood-jp/runner)
<!-- /badges -->

Bootstrap a server process in one call: JSON logger, OS signal handling, and graceful shutdown via [task](https://github.com/wood-jp/task).

## Stability

v1.x releases make no breaking changes to exported APIs. New functionality may be added in minor releases; patches are bug fixes, or administrative work only.

## Installation

Go 1.26.2 or later.

```bash
go get github.com/wood-jp/runner
```

## Usage

Implement a `Runnable` and pass it to `Run`:

```go
const serviceName = "my-service"

func main() {
    runner.Run(serviceName, run)
}

func run(tm runner.Runner, logger *slog.Logger) error {
    return tm.Run(myTask{})
}
```

`Run` creates a JSON logger tagged with `serviceName`, starts an OS signal handler (SIGINT, SIGTERM, SIGQUIT), calls the `Runnable`, then waits for all tasks to finish. On error it calls `os.Exit(1)`, so `main()` can be a single call.

The `LOG_LEVEL` environment variable overrides the log level (e.g. `LOG_LEVEL=debug`). It defaults to `INFO`.

The `Runner` interface exposes the subset of [task.Manager](https://github.com/wood-jp/task) that a `Runnable` needs:

```go
type Runner interface {
    Run(tasks ...task.Task) error
    RunEphemeral(tasks ...task.Task) error
    Cleanup(f func() error)
    Context() context.Context
}
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Security

See [SECURITY.md](SECURITY.md).

## Attribution

*This library is a simplified fork of one written by [wood-jp](https://github.com/wood-jp) at [Zircuit](https://www.zircuit.com/). The original code is available here: [zkr-go-common-public/runner](https://github.com/zircuit-labs/zkr-go-common-public/tree/main/runner)*
