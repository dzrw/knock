package main

import (
	_ "log"
	"log"
	"strconv"
	"testing"
	"time"
)

func TestMaster(t *testing.T) {
	const (
		RunTime = 15
	)

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	args := []string{
		"-c", "8",
		"-d", strconv.Itoa(RunTime),
		"-v",
		"-p", "mongodb.url:mongodb://localhost:27017",
		"-p", "mongodb.run:counters",
	}

	conf, err := parseArgs(args)
	if err != nil {
		t.Error(err)
		return
	}

	log.Printf("Running a %d second test...", conf.Duration)

	m := NewMaster(conf, func() Behavior { return &mongodb_behavior{} })
	m.Start()

	reportProgress(m)

	if time.Since(m.t0) < (RunTime * time.Second) {
		t.Errorf("This test should have taken at least %d seconds.", RunTime)
		return
	}
}

func reportProgress(m *master) {
	reads_on_closed_chan := 0

	for {
		select {
		case <-m.t.Dead():
			log.Printf("awaiter observes master death (%d reads on closed chan)", reads_on_closed_chan)
			return
		case u, ok := <-m.SummaryEvents():
			if !ok {
				reads_on_closed_chan++
				break
			}

			efficiency := u.MeanResponseTimeMs * (u.OpsPerSecond / 1e6) / float64(m.conf.Clients)

			log.Printf("Response Time (usec): %.3f, Throughput (ops/sec): %.3f, Efficiency (%%): %.3f",
				u.MeanResponseTimeMs, u.OpsPerSecond, efficiency)
		}
	}
}
