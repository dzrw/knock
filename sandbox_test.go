package main

import (
	_ "log"
	"sync"
	"testing"
	"time"
)

func TestSandbox(t *testing.T) {
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
			"mongodb.url": "mongodb://localhost:27017",
		},
		Duration:  15 * time.Second,
		StartTime: t0,
		Emitter:   tm,
		WaitGroup: wg,
	})

	sb.Start()

	countResponseTimes(tm)

	if time.Since(t0) < (15 * time.Second) {
		t.Errorf("This test should have taken at least %d seconds.", 15)
		return
	}
}
