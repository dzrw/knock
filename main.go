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
			printHistogram(m.Statistics())
			return

		case u, ok := <-m.SummaryEvents():
			if ok && conf.Verbose {
				printSummary(conf, u, m.t0)
			}
		}
	}
}

func printSummary(conf *BamConfig, evt *SummaryEvent, t0 time.Time) {
	const format = "\015Runtime: %4.fs, Throughput (ops/sec): %8.3f, Response Time (usec): %8.3f, Efficiency (%%): %2.3f"
	//const format2 = "%4.2f\t%.3f\t%.3f\t%.3f\n"

	running := time.Since(t0).Seconds()

	fmt.Fprintf(os.Stderr, format,
		running, evt.OpsPerSecond, evt.MeanResponseTimeMs, evt.Efficiency)
}

func printHistogram(stats Statistics) {
	hist := stats.Histogram()

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
