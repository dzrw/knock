package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Parse the command line.
	opts, err := parseArgs(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}

	run(opts)

	println("\ngoodbye")
}

func run(opts *BamConfig) {

	// Wait for OS signals.
	await(opts)
}

// Blocks until SIGINT or SIGTERM.
func await(opts *BamConfig) {
	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	println("CTRL-C to exit...")

	// Block until we receive a signal.
	sig := <-ch

	if opts.Verbose {
		log.Println("Got signal: ", sig.String())
	}
}
