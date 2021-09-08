package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cirocosta/gyr"
)

func run(ctx context.Context) error {
	scm, err := gyr.NewSCMResolver()
	if err != nil {
		return fmt.Errorf("new scm resolver: %w", err)
	}

	docker, err := gyr.NewDockerResolver()
	if err != nil {
		return fmt.Errorf("new docker resolver: %w", err)
	}

	nodes, err := gyr.NodesFromFiles(filenames())
	if err != nil {
		return fmt.Errorf("nodes from files: %w", err)
	}

	err = gyr.NewResolver(scm, docker).Resolve(ctx, nodes...)
	if err != nil {
		return fmt.Errorf("resolve: %w", err)
	}

	err = gyr.WriteYAML(os.Stdout, nodes)
	if err != nil {
		return fmt.Errorf("write as yaml: %w", err)
	}

	return nil
}

func filenames() []string {
	fnames := os.Args[1:]
	if len(fnames) == 0 {
		return []string{"/dev/stdin"}
	}

	for idx, fname := range fnames {
		if fname == "-" {
			fnames[idx] = "/dev/stdin"
		}
	}

	return fnames
}

func cancelOnSignalContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	sigC := make(chan os.Signal, 2)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigC
		cancel()
		<-sigC
		os.Exit(1)
	}()

	return ctx
}

func main() {
	if err := run(cancelOnSignalContext()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
