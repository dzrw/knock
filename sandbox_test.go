package main

import (
	_ "log"
	"sync"
	"testing"
	"time"
)

func TestSandbox(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t0 := time.Now()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	tm := NewTaskMaster(&TaskMasterInfo{
		WaitGroup:  wg,
		Properties: make(map[string]string),
	})
	tm.Start()

	sb := NewSandbox(&SandboxInfo{
		Id: 1,
		Properties: map[string]string{
			"mongodb.run": "counters",
			"mongodb.url": "mongodb://localhost:27017",
		},
		Duration:  15 * time.Second,
		StartTime: t0,
		Emitter:   tm,
		WaitGroup: wg,
		Factory:   func() Behavior { return &mongodb_behavior{} },
	})

	sb.Start()

	countResponseTimes(tm)

	if time.Since(t0) < (15 * time.Second) {
		t.Errorf("This test should have taken at least %d seconds.", 15)
		return
	}
}
