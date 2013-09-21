package main

import (
	_ "log"
	"log"
	"testing"
	"time"
)

func TestMaster(t *testing.T) {
	args := []string{
		"-c", "2",
		"-g", "2",
		"-d", "60",
		"-p", "mongodb.url:mongodb://localhost:27017",
	}

	conf, err := parseArgs(args)
	if err != nil {
		t.Error(err)
		return
	}

	m := NewMaster(conf)
	m.Start()

	reportProgress(m)

	dumpHistogram(m, 10)

	if time.Since(m.t0) < (10 * time.Second) {
		t.Errorf("This test should have taken at least %d seconds.", 10)
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

func dumpHistogram(m *master, limit int) {
	runtime := time.Since(m.t0).Seconds()
	hist := m.Histogram()

	log.Printf("Response Time Histogram (limit=%d)", limit)
	log.Print("usec,count")

	min := int64(1000000)
	max := int64(0)

	ops := int64(0)

	for usec, c := range hist {
		limit -= 1
		if limit >= 0 {
			log.Printf("%d,%d", usec, c)
		}

		if usec < min {
			min = usec
		}

		if usec > max {
			max = usec
		}

		ops += int64(c)
	}

	log.Printf("min (usec): %d, max (usec): %d, unique: %d", min, max, len(hist))
	log.Printf("ops/sec: %.3f", float64(ops)/runtime)
}
