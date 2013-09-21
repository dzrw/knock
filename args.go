package main

import (
	goflags "github.com/jessevdk/go-flags"
	"time"
)

// -c, --clients CLIENTS, default=1
// -g, --goroutines GOROUTINES, default=1
// -d, --duration TIMEOUT_SECONDS, duration of test, default=notset
// -v, --verbose (boolean) whether or not to write summary information to os.Stderr
// -p, --property key:value (passed as a map[string]string to behaviors)

type BamConfig struct {
	Clients    int               `short:"c" long:"clients" value-name:"CLIENTS" description:"the number of individual load elements" default:"1" optional:"true"`
	Goroutines int               `short:"g" long:"goroutines" value-name:"GOROUTINES" description:"the number of goroutines used to host load elements" default:"1" optional:"true"`
	Duration   int               `short:"d" long:"duration" value-name:"SECONDS" description:"the number of seconds to run this benchmark" default:"10" optional:"true"`
	Verbose    bool              `short:"v" long:"verbose" default:"false" optional:"true"`
	Properties map[string]string `short:"p" description:"additional properties" optional:"true"`

	d time.Duration
}

// Parses the command-line arguments, and validates them.
func parseArgs(args []string) (opts *BamConfig, err error) {
	opts = &BamConfig{}

	_, err = goflags.ParseArgs(opts, args)
	if err != nil {
		return
	}

	// fix bad values...
	if opts.Clients <= 0 {
		opts.Clients = 1
	}

	if opts.Goroutines <= 0 {
		opts.Goroutines = 1
	}

	if opts.Goroutines > opts.Clients {
		opts.Goroutines = opts.Clients
	}

	if opts.Duration < 10 {
		opts.Duration = 10
	}

	opts.d = time.Duration(opts.Duration) * time.Second
	return
}
