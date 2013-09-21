package main

// TIME=1 HOUR (spawn slaves/observer, compute stats (avgs, ops/sec, hist))
// master:
// master distributes CLIENTS over slaves ("gorotines"), hands slaves a waitgroup
// master waits on the observer to know when to proceed to reporting
// master streams <average latency(msec), ops/sec> every 2 seconds to caller
// master builds a histogram of latencies <msec, count>
// master terminates out after the run completes
//

import (
	"launchpad.net/tomb"
	"log"
	"sync"
	"time"
)

type SummaryEvent struct {
	Duration           time.Duration
	MeanResponseTimeMs float64
	OpsPerSecond       float64
}

type master struct {
	t     tomb.Tomb
	conf  *BamConfig
	t0    time.Time
	wg    *sync.WaitGroup
	tm    *taskmaster
	hosts []*sandbox

	// histogram
	hist map[int64]int

	// rolling stats
	curr_lag_sum int64
	last_lag_avg float64
	curr_ops     int64
	last_ops     int64

	// summary events
	statsChan chan *SummaryEvent
}

func NewMaster(conf *BamConfig) *master {
	t0 := time.Now().Add(2 * time.Second)
	wg := &sync.WaitGroup{}

	return &master{
		conf:      conf,
		t0:        t0,
		wg:        wg,
		tm:        NewTaskMaster(wg),
		hosts:     make([]*sandbox, conf.Goroutines),
		hist:      make(map[int64]int),
		statsChan: make(chan *SummaryEvent),
	}
}

func (this *master) Start() {
	go this.loop()
}

func (this *master) Stop() (err error) {
	this.t.Kill(nil)
	return this.t.Wait()
}

func (this *master) SummaryEvents() <-chan *SummaryEvent {
	return this.statsChan
}

func (this *master) Histogram() map[int64]int {
	return this.hist
}

func (this *master) loop() {
	defer this.t.Done()

	this.setup()

	rtChan := this.tm.ResponseTimes()
	prChan := time.After(2 * time.Second)

	for {
		select {
		case <-this.t.Dying():
			log.Print("master dying")
			this.shutdown()
			return

		case <-this.tm.t.Dead():
			log.Print("master observes taskmaster death")
			this.t.Kill(nil)

		case <-prChan:
			this.statsChan <- this.summarize()
			prChan = time.After(2 * time.Second)

		default:
			// Attempt to read from the channel a bunch of times
			// between each death check.
			this.capture(rtChan, 64)
		}
	}
}

func (this *master) setup() {
	// Describe sandboxes
	infos := make([]*SandboxInfo, this.conf.Goroutines)
	for i := 0; i < this.conf.Goroutines; i += 1 {
		infos[i] = &SandboxInfo{
			Id:          i,
			Properties:  this.conf.Properties,
			Duration:    this.conf.d,
			StartTime:   this.t0,
			Emitter:     this.tm,
			WaitGroup:   this.wg,
			ClientCount: 0,
		}
	}

	// Distribute clients evenly over sandboxes
	j := 0
	for i := 0; i < this.conf.Clients; i += 1 {
		infos[j].ClientCount += 1
		j += 1
	}

	// Initialize sandboxes
	this.hosts = make([]*sandbox, this.conf.Goroutines)
	for i, info := range infos {
		this.hosts[i] = NewSandbox(info)
		this.wg.Add(1)
	}

	// Spawn the taskmaster
	this.tm.Start()

	// Spawn each sandbox
	for _, host := range this.hosts {
		host.Start()
	}
}

func (this *master) shutdown() {
	close(this.statsChan)
}

func (this *master) capture(ch LatencyEventsChannel, reads int) {
	hist := this.hist
	chunk_lag := int64(0)
	chunk_ops := int64(0)

	for i := 0; i < reads; i += 1 {
		usec, ok := <-ch
		if !ok {
			break
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
	this.curr_ops += chunk_ops
}

func (this *master) summarize() *SummaryEvent {
	run_time := time.Since(this.t0)
	next_ops := this.last_ops + this.curr_ops

	// Compute the weighted average response time
	w0 := float64(this.last_ops) / float64(next_ops)
	w1 := float64(this.curr_ops) / float64(next_ops)
	curr_lag_avg := float64(this.curr_lag_sum) / float64(this.curr_ops)
	next_lag_avg := (w0 * this.last_lag_avg) + (w1 * curr_lag_avg)

	log.Printf("w: <%f, %f>", w0, w1)

	// Compute the current throughput ops/sec
	next_ops_sec := float64(next_ops) / run_time.Seconds()

	// Update
	this.last_lag_avg = next_lag_avg
	this.last_ops = next_ops

	// Reset counters
	this.curr_lag_sum = 0
	this.curr_ops = 0

	return &SummaryEvent{run_time, next_lag_avg, next_ops_sec}
}
