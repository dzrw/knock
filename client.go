package main

type WorkResult int

const (
	WRK_OK WorkResult = iota
	WRK_WTF
	WRK_TIMEOUT
	WRK_ERROR
)

type Client interface {
	// Initialize any state for this client.
	Init(props map[string]string) (err error)

	// Cleanup any state for this client.
	Close()

	// Perform one unit of work.
	Work() (res WorkResult)
}
