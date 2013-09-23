package main

import (
	"time"
)

type WorkResult int

const (
	WRK_OK WorkResult = iota
	WRK_WTF
	WRK_TIMEOUT
	WRK_ERROR
)

type BehaviorFactory func() Behavior

type Behavior interface {
	// Initialize any state for this client.
	Init(props map[string]string) (err error)

	// Cleanup any state for this client.
	Close()

	// Perform one unit of work.
	Work(t0 time.Time) (res WorkResult)
}
