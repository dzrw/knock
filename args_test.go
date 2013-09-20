package main

import (
	"testing"
)

// -c, --clients CLIENTS, default=1
// -g, --goroutines GOROUTINES, default=1
// -d, --duration TIMEOUT_SECONDS, duration of test, default=notset
// -v, --verbose (boolean) whether or not to write summary information to os.Stderr
// -p, --property key:value (passed as a map[string]string to behaviors)

func TestParseArguments(t *testing.T) {
	args := []string{
		"-c", "5",
		"-g", "4",
		"-d", "30",
		"-v",
		"-p", "a:1",
		"-p", "b:2",
	}

	opts, err := parseArgs(args)
	if err != nil {
		t.Error(err)
		return
	}

	if !expectInt(t, 5, opts.Clients) {
		return
	}

	if !expectInt(t, 4, opts.Goroutines) {
		return
	}

	if !expectInt(t, 30, opts.Duration) {
		return
	}

	if !expectBool(t, true, opts.Verbose) {
		return
	}

	if opts.Properties == nil {
		t.Error("property map not initialized")
		return
	}

	if !expectInt(t, 2, len(opts.Properties)) {
		return
	}

	if !expectKeyValue(t, opts.Properties, "a", "1") {
		return
	}

	if !expectKeyValue(t, opts.Properties, "b", "2") {
		return
	}
}

func TestEmptyArguments(t *testing.T) {
	args := []string{}

	_, err := parseArgs(args)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestDefaultValuesOfArguments(t *testing.T) {
	args := []string{}

	opts, err := parseArgs(args)
	if err != nil {
		t.Error(err)
		return
	}

	if !expectInt(t, 1, opts.Clients) {
		return
	}

	if !expectInt(t, 1, opts.Goroutines) {
		return
	}

	if !expectInt(t, 10, opts.Duration) {
		return
	}

	if !expectBool(t, false, opts.Verbose) {
		return
	}

	if !expectInt(t, 0, len(opts.Properties)) {
		return
	}
}

func TestLowerBoundsOfArguments(t *testing.T) {
	args := []string{"-d", "0", "-c", "0", "-g", "0"}

	opts, err := parseArgs(args)
	if err != nil {
		t.Error(err)
		return
	}

	if !expectInt(t, 1, opts.Clients) {
		return
	}

	if !expectInt(t, 1, opts.Goroutines) {
		return
	}

	if !expectInt(t, 10, opts.Duration) {
		return
	}
}

func TestUpperBoundsOfArguments(t *testing.T) {
	args := []string{"-c", "1", "-g", "2"}

	opts, err := parseArgs(args)
	if err != nil {
		t.Error(err)
		return
	}

	if !expectInt(t, 1, opts.Clients) {
		return
	}

	if !expectInt(t, 1, opts.Goroutines) {
		return
	}
}
