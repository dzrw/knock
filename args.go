package main

import (
	goflags "github.com/jessevdk/go-flags"
	"time"
)

const (
	MIN_RUN_TIME = 5
	MIN_LOAD     = 1
)

type AppConfig struct {
	Clients        int               `short:"c" long:"clients" value-name:"CLIENTS" description:"the number of individual load elements" default:"0" optional:"true"`
	Duration       int               `short:"d" long:"duration" value-name:"SECONDS" description:"the number of seconds to run this benchmark" default:"0" optional:"true"`
	Verbose        bool              `short:"v" long:"verbose" default:"false" optional:"true"`
	PerClientStats bool              `long:"client-stats" default:"false" optional:"true" description:"whether or not to track individual client statistics"`
	Properties     map[string]string `short:"p" description:"additional properties" optional:"true"`
	Version        bool              `long:"version" optional:"true" default:"false" description:"display version information"`
	Profiles       map[string]string `short:"r" long:"runtime-profile" optional:"true" description:"Go runtime profiles (e.g. cpu, memory, block, threadcount, or behavior-specifc)"`

	d time.Duration
}

// Parses the command-line arguments, and validates them.
func parseArgs(args []string) (opts *AppConfig, err error) {
	opts = &AppConfig{}

	_, err = goflags.ParseArgs(opts, args)
	if err != nil {
		return
	}

	// fix bad values...
	if opts.Clients < MIN_LOAD {
		opts.Clients = MIN_LOAD
	}

	if opts.Duration < MIN_RUN_TIME {
		opts.Duration = MIN_RUN_TIME
	}

	opts.d = time.Duration(opts.Duration) * time.Second
	return
}
