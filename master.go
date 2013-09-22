package main

import (
	"launchpad.net/tomb"
	_ "log"
	"sync"
	"time"
)

type SummaryEmitter interface {
	PublishSummaryEvent(d time.Duration, throughput, responseTime, efficiency float64)
}

type SummaryEvent struct {
	Duration           time.Duration
	MeanResponseTimeMs float64
	OpsPerSecond       float64
	Efficiency         float64
}

type master struct {
	t         tomb.Tomb
	conf      *AppConfig
	t0        time.Time
	wg        *sync.WaitGroup
	tm        *taskmaster
	hosts     []*sandbox
	stats     *calculator
	statsChan chan *SummaryEvent
}

func NewMaster(conf *AppConfig) *master {
	wg := &sync.WaitGroup{}

	return &master{
		conf:      conf,
		wg:        wg,
		tm:        nil,
		hosts:     make([]*sandbox, conf.Clients),
		stats:     nil,
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

func (this *master) PublishSummaryEvent(d time.Duration, throughput, responseTime, activeLoad float64) {
	this.statsChan <- &SummaryEvent{d, responseTime, throughput, activeLoad}
}

// Only call this after the goroutine is dead.
func (this *master) Statistics() Statistics {
	return this.stats
}

func (this *master) loop() {
	const (
		ProgressInterval = 1 * time.Second
		MaxChannelReads  = 64
	)

	defer this.t.Done()

	this.setup()

	prChan := time.After(ProgressInterval)

	for {
		select {
		case <-this.t.Dying():
			this.shutdown()
			return

		case <-this.tm.t.Dead():
			this.t.Kill(nil)

		case <-prChan:
			this.stats.summarize()
			prChan = time.After(ProgressInterval)

		default:
			// Attempt to read from the channel a bunch of times
			// between each death check.
			this.stats.capture(MaxChannelReads)
		}
	}
}

func (this *master) setup() {
	// Record approximate start time
	this.t0 = time.Now()

	// Initialize the taskmaster
	this.tm = NewTaskMaster(&TaskMasterInfo{
		WaitGroup:  this.wg,
		Properties: this.conf.Properties,
	})

	// Initialize the stats recorder
	this.stats = NewCalculator(
		this.conf, this.tm.ResponseTimes(), this, this.t0)

	// Initialize client sandboxes
	count := this.conf.Clients
	for i := 0; i < count; i += 1 {
		info := &SandboxInfo{
			Id:         i,
			Properties: this.conf.Properties,
			Duration:   this.conf.d,
			StartTime:  this.t0,
			Emitter:    this.tm,
			WaitGroup:  this.wg,
		}

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
	this.stats.summarize()
	close(this.statsChan)
}
