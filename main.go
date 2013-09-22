package main

import (
	"os"
)

func main() {
	// Parse the command line.
	conf, err := parseArgs(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}

	// Run the benchmark with our own custom behavior.
	RunBenchmark(conf, func() Behavior {
		return &mgo_incr_client{}
	})
}
