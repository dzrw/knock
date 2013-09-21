package main

import (
	"fmt"
	"github.com/ryszard/goskiplist/skiplist"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func main() {
	// Schedule all of the logical cores.
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Seed the RNG (doesn't happen automatically).
	rand.Seed(time.Now().UnixNano())

	// Parse the command line.
	conf, err := parseArgs(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}

	run(conf)
}

func run(conf *BamConfig) {
	// Start the benchmark.
	m := NewMaster(conf)
	m.Start()

	// Wait for the benchmark to finish.
	await(conf, m)
}

// Blocks until SIGINT or SIGTERM.
func await(conf *BamConfig, m *master) {
	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	if conf.Verbose {
		fmt.Fprintln(os.Stderr, "starting benchmark...\015")
	}

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
			printHistogram(conf, m)
			return

		case u, ok := <-m.SummaryEvents():
			if ok && conf.Verbose {
				printSummary(conf, u, m.t0)
			}
		}
	}
}

func printSummary(conf *BamConfig, evt *SummaryEvent, t0 time.Time) {
	const msg = "[RUN %4ds] Throughput (ops/sec): %.3f, Response Time (usec): %.3f, Efficiency (%%): %.3f\015"
	const gmsg = "%4d\t%.3f\t%.3f\t%.3f\n"

	running := int(time.Since(t0).Seconds())

	planned := conf.Clients
	active := evt.MeanResponseTimeMs * (evt.OpsPerSecond / 1e6)
	efficiency := active / float64(planned)

	fmt.Fprintf(os.Stderr, gmsg, running, evt.OpsPerSecond, evt.MeanResponseTimeMs, efficiency)
}

func printHistogram(conf *BamConfig, m *master) {
	hist := m.Histogram()

	min := int64(1e9)
	max := int64(0)

	l := skiplist.NewIntMap()

	for usec, count := range hist {
		if usec < min {
			min = usec
		}

		if usec > max {
			max = usec
		}

		l.Set(int(usec), int(count))
	}

	fmt.Fprintf(os.Stdout, "Response Time Histogram\n\n")
	fmt.Fprintf(os.Stdout, "usec\tcount\n")
	fmt.Fprintf(os.Stdout, "----\t-----\n\n")

	iter := l.Iterator()
	for iter.Next() {
		k := iter.Key().(int)
		v := iter.Value().(int)

		fmt.Fprintf(os.Stdout, "%d\t%d\n", k, v)
	}

	fmt.Fprintf(os.Stdout, "min: %d, max: %d, unique: %d", min, max, len(hist))
}
