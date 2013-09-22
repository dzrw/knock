package main

import (
	"fmt"
	_ "log"
	"os"
	"strconv"
	"strings"
	"time"
)

type PrintFunc func(format string, args ...interface{})

func PrintReport(f *os.File, s Statistics, conf *AppConfig) {
	p := fmt.Fprintf
	//p := printer(f)

	res := s.Histogram2()
	dist := res.dist
	cdf := res.cdf

	if conf.Verbose {
		printSummaryTrailer(os.Stderr, s, res)
	}

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
	p(f, "Run Time (s):\t%8.4f\n", time.Since(s.StartTime()).Seconds())
	p(f, "Throughput (ops/sec):\t%f\n", s.Throughput())
	p(f, "Mean Response Time (μs):\t%8.4f\n", s.MeanResponseTimeUsec())
	p(f, "Load Efficiency (%%):\t%f\n", s.Efficiency())
	p(f, "\n")

	p(f, "Response Time Details:\n")
	p(f, "  Min: %dμs\n", res.min)
	p(f, "  Max: %dμs\n", res.max)
	p(f, "  Mean: %8.4fμs\n", s.MeanResponseTimeUsec())
	p(f, "  5th Percentile: %dμs\n", res.p5)
	p(f, "  95th Percentile: %dμs\n", res.p95)
	p(f, "  99th Percentile: %dμs\n", res.p99)
	p(f, "\n\n")

	p(f, "Response Time CDF and Frequency Histogram\n")
	p(f, "-----------------------------------------\n")
	p(f, "(cut and paste the tab-delimited table below into Google Spreadsheets)")
	p(f, "\n\n")

	headers := []string{"usec", "CDF", "total"}
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

	iter := dist.Iterator()
	for iter.Next() {
		usec := iter.Key().(int)
		vs := iter.Value().([]int)

		cdfi, _ := cdf.Get(usec)

		row := make([]string, len(vs)+2)
		row[0] = strconv.Itoa(usec)
		row[1] = fmt.Sprintf("%2.6f", cdfi.(float64))

		for i, v := range vs {
			row[i+2] = strconv.Itoa(v)
		}

		p(f, "\n")
		p(f, strings.Join(row, "\t"))
	}

	p(f, "\n")
}

func printSummary(conf *AppConfig, evt *SummaryEvent, t0 time.Time) {
	const format = "\015Runtime: %4.fs, Throughput (ops/sec): %8.3f, Response Time (μs): %8.3f, Efficiency (%%): %2.3f"
	//const format2 = "%4.2f\t%.3f\t%.3f\t%.3f\n"

	running := time.Since(t0).Seconds()

	fmt.Fprintf(os.Stderr, format,
		running, evt.OpsPerSecond, evt.MeanResponseTimeMs, evt.Efficiency)
}

func printSummaryTrailer(f *os.File, s Statistics, res *HistogramResult) {
	p := fmt.Fprintf

	p(f, "\n")
	p(f, "Time's up! Fastest: %s, Percentiles: [5th: %s, 95th: %s, 99th: %s], Slowest: %s",
		wash(int(res.min)), wash(res.p5), wash(res.p95), wash(res.p99), wash(int(res.max)))
	p(f, "\n")
}

func wash(usec int) string {
	switch {
	case usec < 1000:
		return fmt.Sprintf("%dμs", usec)
	case usec < 1e6:
		return fmt.Sprintf("%.3fms", float64(usec)/1000)
	default:
		return fmt.Sprintf("%.3fs", float64(usec)/1e6)
	}
}
