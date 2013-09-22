package main

import (
	"github.com/ryszard/goskiplist/skiplist"
	"time"
)

type Statistics interface {
	StartTime() time.Time

	Histogram() (hist map[int64]int)
	Throughput() float64
	MeanResponseTimeUsec() float64
	Efficiency() float64
	Histogram2() *skiplist.SkipList

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

	bucket
}

func NewCalculator(conf *BamConfig, ch LatencyEventsChannel, emitter SummaryEmitter, t0 time.Time) *calculator {
	this := &calculator{
		t0:          t0,
		ch:          ch,
		emitter:     emitter,
		clients:     nil,
		clientCount: conf.Clients,
		clientStats: conf.PerClientStats,
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

func (this *calculator) Histogram() map[int64]int {
	return this.hist
}

func (this *calculator) Throughput() float64 {
	return float64(this.prev_ops_sum) / time.Since(this.t0).Seconds()
}

func (this *calculator) MeanResponseTimeUsec() float64 {
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

func (this *calculator) Histogram2() *skiplist.SkipList {
	l := skiplist.NewIntMap()

	min := int64(1e9)
	max := int64(0)

	for usec, total := range this.hist {
		if usec < min {
			min = usec
		}

		if usec > max {
			max = usec
		}

		v := []int{total}
		if this.IsClientTrackingEnabled() {
			v = make([]int, this.clientCount+1)
			v[0] = total
		}

		l.Set(int(usec), v)
	}

	if this.IsClientTrackingEnabled() {
		for id, bucket := range this.clients {
			for usec, count := range bucket.hist {
				v, ok := l.Get(int(usec))
				if ok {
					v.([]int)[id+1] = count
				}
			}
		}
	}

	return l
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

	// Compute the current throughput ops/sec
	next_ops_per_sec := float64(next_ops_sum) / d.Seconds()

	// Compute the active load and load efficiency
	active_load := next_lag_avg * (next_ops_per_sec / 1e6)
	planned_load := float64(this.clientCount)
	efficiency := active_load / planned_load

	// Update
	this.prev_lag_avg = next_lag_avg
	this.prev_ops_sum = next_ops_sum

	// Reset counters
	this.curr_lag_sum = 0
	this.curr_ops_sum = 0

	this.emitter.PublishSummaryEvent(
		d, next_ops_per_sec, next_lag_avg, efficiency)
}
