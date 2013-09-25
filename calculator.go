package main

import (
	"github.com/ryszard/goskiplist/skiplist"
	_ "log"
	"math"
	"time"
)

type Statistics interface {
	StartTime() time.Time

	Throughput() float64
	MeanResponseTimeUsec() float64
	Efficiency() float64
	Histogram2() (res *HistogramResult)
	Errors() map[WorkResult]int

	IsClientTrackingEnabled() (ok bool)
	HistogramByClientId(clientId int) (hist map[int64]int, ok bool)
}

type bucket struct {
	id int

	hist map[int64]int

	curr_lag_sum int64
	prev_lag_avg float64
	curr_ops_sum int64
	prev_ops_sum int64
}

func (this *bucket) observe(usec int64) {
	if count, ok := this.hist[usec]; ok {
		this.hist[usec] = count + 1
	} else {
		this.hist[usec] = 1
	}

	this.curr_lag_sum += usec
	this.curr_ops_sum += 1
}

type calculator struct {
	t0          time.Time
	ch          LatencyEventsChannel
	emitter     SummaryEmitter
	clients     map[int]*bucket
	clientStats bool
	clientCount int
	errors      map[WorkResult]int

	bucket
}

func NewCalculator(conf *AppConfig, ch LatencyEventsChannel, emitter SummaryEmitter, t0 time.Time) *calculator {
	this := &calculator{
		t0:          t0,
		ch:          ch,
		emitter:     emitter,
		clients:     nil,
		clientCount: conf.Clients,
		clientStats: conf.PerClientStats,
		errors:      make(map[WorkResult]int),
		bucket: bucket{
			id:   -1,
			hist: make(map[int64]int),
		},
	}

	if this.clientStats {
		this.clients = make(map[int]*bucket)
		for i := 0; i < this.clientCount; i++ {
			this.clients[i] = &bucket{id: i, hist: make(map[int64]int)}
		}
	}

	return this
}

func (this *calculator) StartTime() time.Time {
	return this.t0
}

func (this *calculator) Errors() map[WorkResult]int {
	return this.errors
}

func (this *calculator) Throughput() float64 {
	return float64(this.prev_ops_sum) / time.Since(this.t0).Seconds()
}

func (this *calculator) MeanResponseTimeUsec() float64 {
	const EPSILON = float64(1*time.Microsecond) / 10

	if this.prev_lag_avg < EPSILON {
		return EPSILON
	}

	return this.prev_lag_avg
}

func (this *calculator) Efficiency() float64 {
	throughput := this.Throughput()
	responseTime := this.MeanResponseTimeUsec()
	active_load := responseTime * (throughput / 1e6)
	planned_load := float64(this.clientCount)
	efficiency := active_load / planned_load
	return efficiency
}

func (this *calculator) IsClientTrackingEnabled() bool {
	return this.clientStats
}

func (this *calculator) HistogramByClientId(clientId int) (hist map[int64]int, ok bool) {
	bucket, ok := this.clients[clientId]
	if !ok {
		return
	}

	return bucket.hist, true
}

type HistogramResult struct {
	dist         *skiplist.SkipList
	cdf          *skiplist.SkipList
	p5, p95, p99 int
	min, max     int64
	errors       map[WorkResult]int
}

func (this *calculator) Histogram2() (res *HistogramResult) {
	res = &HistogramResult{
		dist: skiplist.NewIntMap(),
		cdf:  skiplist.NewIntMap(),
	}

	min := int64(1e9)
	max := int64(0)

	// Copy our histogram map into an ordered skiplist.
	for usec, freq := range this.hist {
		if usec < min {
			min = usec
		}

		if usec > max {
			max = usec
		}

		v := []int{freq}
		if this.IsClientTrackingEnabled() {
			v = make([]int, this.clientCount+1)
			v[0] = freq
		}

		res.dist.Set(int(usec), v)
	}

	res.min = min
	res.max = max

	// Build the CDF.
	sum := int64(0)

	iter := res.dist.Iterator()
	for iter.Next() {
		usec := iter.Key().(int)
		vs := iter.Value().([]int)
		freq := int64(vs[0])

		sum += freq

		v := float64(sum) / float64(this.prev_ops_sum)

		switch {
		case res.p5 == 0 && v >= 0.05:
			res.p5 = usec
		case res.p95 == 0 && v >= 0.95:
			res.p95 = usec
		case res.p99 == 0 && v >= 0.99:
			res.p99 = usec
		}

		res.cdf.Set(usec, v)
	}

	// Extend the skiplist with per-client stats, if enabled.
	if this.IsClientTrackingEnabled() {
		for id, bucket := range this.clients {
			for usec, count := range bucket.hist {
				v, ok := res.dist.Get(int(usec))
				if ok {
					v.([]int)[id+1] = count
				}
			}
		}
	}

	return
}

// func concat(old1, old2 []int) []int {
// 	newslice := make([]int, len(old1)+len(old2))
// 	copy(newslice, old1)
// 	copy(newslice[len(old1):], old2)
// 	return newslice
// }

func (this *calculator) capture(count int) {
	ch := this.ch
	hist := this.hist
	chunk_lag := int64(0)
	chunk_ops := int64(0)

	clients := this.clients
	clientStats := this.clientStats

	for i := 0; i < count; i += 1 {
		evt, ok := <-ch
		if !ok {
			break
		}

		// Count errors, but don't pollute the ops counter.
		if evt.result != WRK_OK {
			if v, ok := this.errors[evt.result]; ok {
				this.errors[evt.result] = v + 1
			} else {
				this.errors[evt.result] = 1
			}

			continue
		}

		usec := evt.usec

		// Update the per-client stats, if necessary.
		if clientStats {
			clients[evt.id].observe(evt.usec)
		}

		// Update the histogram
		if count, ok := hist[usec]; ok {
			hist[usec] = count + 1
		} else {
			hist[usec] = 1
		}

		chunk_lag += usec
		chunk_ops += 1
	}

	// Update the intermediate sums
	this.curr_lag_sum += chunk_lag
	this.curr_ops_sum += chunk_ops
}

func (this *calculator) summarize() {
	// Run Time
	d := time.Since(this.t0)

	// Total Ops
	next_ops_sum := this.prev_ops_sum + this.curr_ops_sum

	// Compute the weighted average response time
	w0 := float64(this.prev_ops_sum) / float64(next_ops_sum)
	w1 := float64(this.curr_ops_sum) / float64(next_ops_sum)
	curr_lag_avg := float64(this.curr_lag_sum) / float64(this.curr_ops_sum)
	next_lag_avg := (w0 * this.prev_lag_avg) + (w1 * curr_lag_avg)

	// TODO: I think w1 is underflowing, but I don't have time to
	// check right now.  Instead, since this only seems to happen
	// towards the end of the run with really fast work units,
	// we'll just stop updating these variables (and emit the same
	// summaries).
	if math.IsNaN(next_lag_avg) {
		next_ops_sum = this.prev_ops_sum
		next_lag_avg = this.prev_lag_avg
	}

	// Compute the current throughput ops/sec
	next_ops_per_sec := float64(next_ops_sum) / d.Seconds()

	// Compute the active load and load efficiency
	eff := efficiency(this.clientCount, next_ops_per_sec, next_lag_avg)

	// Update
	this.prev_lag_avg = next_lag_avg
	this.prev_ops_sum = next_ops_sum

	// Reset counters
	this.curr_lag_sum = 0
	this.curr_ops_sum = 0

	this.emitter.PublishSummaryEvent(d, next_ops_per_sec, next_lag_avg, eff)
}

func efficiency(load int, throughput, responseTimeUs float64) float64 {
	active_load := responseTimeUs * (throughput / 1e6)
	planned_load := float64(load)
	efficiency := active_load / planned_load
	return efficiency
}
