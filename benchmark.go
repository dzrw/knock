package main

import (
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func RunBenchmark(conf *AppConfig, factory BehaviorFactory) {
	// Schedule all of the logical cores.
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Seed the RNG (doesn't happen automatically).
	rand.Seed(time.Now().UnixNano())

	// Start the benchmark.
	m := NewMaster(conf, factory)
	m.Start()

	// Wait for the benchmark to finish.
	await(conf, m)
}

// Blocks until SIGINT or SIGTERM.
func await(conf *AppConfig, m *master) {
	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case sig := <-ch:
			switch sig {
			case syscall.SIGQUIT:
				os.Exit(1)

			case syscall.SIGINT, syscall.SIGTERM:
				os.Exit(1)
			}

		case <-m.t.Dead():
			PrintReport(os.Stdout, m.Statistics(), conf)
			return

		case u, ok := <-m.SummaryEvents():
			if ok && conf.Verbose {
				printSummary(conf, u, m.t0)
			}
		}
	}
}
