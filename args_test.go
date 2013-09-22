package main

import (
	"testing"
)

func TestParseArguments(t *testing.T) {
	args := []string{
		"-c", "5",
		"-d", "30",
		"-v",
		"-p", "a:1",
		"-p", "b:2",
		"--client-stats",
		"--version",
	}

	opts, err := parseArgs(args)
	if err != nil {
		t.Error(err)
		return
	}

	if !expectInt(t, 5, opts.Clients) {
		return
	}

	if !expectInt(t, 30, opts.Duration) {
		return
	}

	if !expectBool(t, true, opts.Verbose) {
		return
	}

	if !expectBool(t, true, opts.PerClientStats) {
		return
	}

	if !expectBool(t, true, opts.Version) {
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

	if !expectInt(t, MIN_LOAD, opts.Clients) {
		return
	}

	if !expectInt(t, MIN_RUN_TIME, opts.Duration) {
		return
	}

	if !expectBool(t, false, opts.Verbose) {
		return
	}

	if !expectBool(t, false, opts.PerClientStats) {
		return
	}

	if !expectBool(t, false, opts.Version) {
		return
	}

	if !expectInt(t, 0, len(opts.Properties)) {
		return
	}
}

func TestLowerBoundsOfArguments(t *testing.T) {
	args := []string{"-d", "0", "-c", "0"}

	opts, err := parseArgs(args)
	if err != nil {
		t.Error(err)
		return
	}

	if !expectInt(t, MIN_LOAD, opts.Clients) {
		return
	}

	if !expectInt(t, MIN_RUN_TIME, opts.Duration) {
		return
	}
}
