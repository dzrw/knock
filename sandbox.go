package main

import (
	"errors"
	_ "launchpad.net/tomb"
	"log"
	_ "math"
	"strconv"
	"sync"
	"time"
)

type SandboxInfo struct {
	Id         int
	Properties map[string]string
	Duration   time.Duration
	StartTime  time.Time
	Emitter    LatencyEmitter
	WaitGroup  *sync.WaitGroup
	Factory    BehaviorFactory
}

type sandbox struct {
	id            int
	props         map[string]string
	d             time.Duration
	start         time.Time
	emitter       LatencyEmitter
	wg            *sync.WaitGroup
	behavior      Behavior
	factory       BehaviorFactory
	stall         bool
	opsPerStall   int
	stall_counter int
	stalls        int
}

func NewSandbox(info *SandboxInfo) *sandbox {
	return &sandbox{
		id:      info.Id,
		props:   info.Properties,
		d:       info.Duration,
		start:   info.StartTime,
		emitter: info.Emitter,
		wg:      info.WaitGroup,
		factory: info.Factory,
	}
}

func (this *sandbox) Start() {
	go this.loop()
}

// func (this *sandbox) Stop() (err error) {
// 	this.t.Kill(nil)
// 	return this.t.Wait()
// }

func (this *sandbox) loop() {
	//defer this.t.Done()

	this.setup()

	defer this.teardown()

	for !this.expired() {
		this.update()
	}
}

func (this *sandbox) setup() {
	err := this.parseProperties()
	if err != nil {
		log.Fatal(err)
	}

	res, err := this.init()
	if err != nil {
		log.Fatal(err)
	}

	this.behavior = res
}

func (this *sandbox) init() (res Behavior, err error) {
	defer func() {
		e := recover()
		if e != nil {
			if u, ok := e.(error); ok {
				err = u
			}
		}

		return
	}()

	res = this.factory()
	err = res.Init(this.props)
	if err != nil {
		return
	}

	return
}

func (this *sandbox) expired() (ok bool) {
	return time.Since(this.start) > this.d
}

func (this *sandbox) update() {
	_ = this.work()
}

func (this *sandbox) work() (err error) {
	defer func() {
		e := recover()
		if e != nil {
			if u, ok := e.(error); ok {
				err = u
			}
		}

		return
	}()

	t0 := time.Now()
	res := this.behavior.Work(t0)
	d := time.Since(t0)
	usec := int64(d / time.Microsecond)

	// If stalling is enabled, then introduce a 1s pause
	// every N ops. Hopefully, this will fix the issue with
	// some Work operations completing far too fast for
	// the progress/stats reporting goroutines to work.
	// I'm not sure why the select{} construct isn't picking
	// the other channels...
	if this.stall {
		this.stall_counter += 1

		if this.stall_counter > this.opsPerStall {
			this.stalls += 1
			this.stall_counter = 0
			<-time.After(1 * time.Second)
		}
	}

	this.emitter.PublishResponseTime(this.id, usec, res)
	return
}

func (this *sandbox) teardown() {
	this.close()
	this.wg.Done()
}

func (this *sandbox) close() (err error) {
	defer func() {
		e := recover()
		if e != nil {
			if u, ok := e.(error); ok {
				err = u
			}
		}

		return
	}()

	this.behavior.Close()
	this.behavior = nil
	return
}

func (this *sandbox) parseProperties() (err error) {
	props := this.props

	const (
		MinOpsPerStall = 10000
	)

	if v, ok := props["internals.OpsPerStall"]; ok {
		u, err := strconv.Atoi(v)
		if err != nil || u < MinOpsPerStall {
			return errors.New("internals.OpsPerStall must be >= 10000 ops (e.g. don't use this option when response times are slow enough to measure)")
		}

		this.stall = true
		this.opsPerStall = u
	} else {
		this.stall = false
	}

	return
}
