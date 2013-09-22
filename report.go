package main

import (
	"fmt"
	_ "log"
	"os"
	"strconv"
	"strings"
)

type PrintFunc func(format string, args ...interface{})

func PrintReport(f *os.File, s Statistics, conf *BamConfig) {
	p := fmt.Fprintf
	//p := printer(f)

	p(f, "Setup\n")
	p(f, "-----\n")
	p(f, "\n")
	p(f, "clients=%d\n", conf.Clients)
	p(f, "duration=%d\n", conf.Duration)

	for k, v := range conf.Properties {
		p(f, "%s=%s\n", k, v)
	}
	p(f, "\n\n")

	p(f, "Overview\n")
	p(f, "--------\n")
	p(f, "\n")
	p(f, "Throughput (ops/sec):\t%f\n", s.Throughput())
	p(f, "Response Time (usec):\t%f\n", s.MeanResponseTimeUsec())
	p(f, "Load Efficiency (%%):\t%f\n", s.Efficiency())
	p(f, "\n\n")

	p(f, "Response Time Latencies Histogram\n")
	p(f, "---------------------------------\n")
	p(f, "\n")

	headers := []string{"usec", "total"}
	if s.IsClientTrackingEnabled() {
		for i := 0; i < conf.Clients; i++ {
			headers = append(headers, fmt.Sprintf("client-%d", i))
		}
	}

	p(f, strings.Join(headers, "\t"))

	spacers := []string{}
	for _, s := range headers {
		spacers = append(spacers, strings.Repeat("-", len(s)))
	}

	p(f, "\n")
	p(f, strings.Join(spacers, "\t"))

	iter := s.Histogram2().Iterator()
	for iter.Next() {
		usec := iter.Key().(int)
		vs := iter.Value().([]int)

		row := make([]string, len(vs)+1)
		row[0] = strconv.Itoa(usec)

		for i, v := range vs {
			row[i+1] = strconv.Itoa(v)
		}

		p(f, "\n")
		p(f, strings.Join(row, "\t"))
	}

	p(f, "\n")
}
