package main

import (
	_ "launchpad.net/tomb"
	"log"
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
	id       int
	props    map[string]string
	d        time.Duration
	start    time.Time
	emitter  LatencyEmitter
	wg       *sync.WaitGroup
	behavior Behavior
	factory  BehaviorFactory
}

func NewSandbox(info *SandboxInfo) *sandbox {
	return &sandbox{
		info.Id,
		info.Properties,
		info.Duration,
		info.StartTime,
		info.Emitter,
		info.WaitGroup,
		nil,
		info.Factory,
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
	usec := int64(time.Since(t0) / time.Microsecond)

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
