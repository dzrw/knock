package main

import (
	"log"
	_ "math"
	"strconv"
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

func TestSandboxWithMinWorkTime(t *testing.T) {
	const (
		TestDuration  = 5 * time.Second
		OpsPerStall   = 10000
		SleepDuration = 5 * time.Microsecond
		StallDuration = 1 * time.Second
	)

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	log.Printf("Running a %.2f second test...", TestDuration.Seconds())

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
			"mongodb.run":           "counters",
			"mongodb.url":           "mongodb://localhost:27017",
			"internals.OpsPerStall": strconv.Itoa(OpsPerStall),
		},
		Duration:  TestDuration,
		StartTime: t0,
		Emitter:   tm,
		WaitGroup: wg,
		Factory: func() Behavior {
			return &dummy_behavior{
				sleep: SleepDuration,
			}
		},
	})

	sb.Start()

	ops := countResponseTimes(tm)

	if time.Since(t0) < TestDuration {
		t.Errorf("This test should have taken at least %.1f seconds.", TestDuration.Seconds())
		return
	}

	// When the response times are very small (on the order of nanoseconds),
	// then other channels aren't serviced as fairly by the Go scheduler
	// (I think -- I'm certainly not seeing code associated with some channels
	// run (e.g. IO)).  At those response times, the millisecond stall interval
	// should dominate the ops/sec.  That is, performing OpsPerStall ops should
	// take far far less time that performing the stall.  If we through out that
	// term, then we can estimate the # of stalls, and put an upper bound on the
	// number of ops.
	expected_stalls := float64(TestDuration.Seconds()) / float64(StallDuration.Seconds())
	if float64(sb.stalls) < expected_stalls {
		t.Errorf("expected at least: %.4f stalls, observed: %d stalls",
			expected_stalls, sb.stalls)
		return
	}

	// Add a 1% fudge factor to the upper bound because of error
	// in the measurement of the actual duration of the test.
	expected_max_ops := float64(OpsPerStall) * expected_stalls * 1.01
	if float64(ops) > expected_max_ops {
		t.Errorf("expected at most: %.4f ops, observed: %d ops",
			expected_max_ops, ops)
		return
	}
}

type dummy_behavior struct {
	sleep time.Duration
}

func (*dummy_behavior) Init(props map[string]string) (err error) {
	return
}

func (*dummy_behavior) Close() {}

func (this *dummy_behavior) Work(t0 time.Time) (res WorkResult) {
	<-time.After(this.sleep)
	return WRK_OK
}
