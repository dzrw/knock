package main

import (
	_ "launchpad.net/tomb"
	"log"
	"sync"
	"time"
)

type SandboxInfo struct {
	Id          int
	Properties  map[string]string
	Duration    time.Duration
	StartTime   time.Time
	Emitter     LatencyEmitter
	WaitGroup   *sync.WaitGroup
	ClientCount int
}

type sandbox struct {
	id      int
	props   map[string]string
	d       time.Duration
	start   time.Time
	emitter LatencyEmitter
	wg      *sync.WaitGroup
	count   int

	clients []Client
}

func NewSandbox(info *SandboxInfo) *sandbox {
	return &sandbox{
		info.Id,
		info.Properties,
		info.Duration,
		info.StartTime,
		info.Emitter,
		info.WaitGroup,
		info.ClientCount,
		make([]Client, info.ClientCount),
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
	for i := 0; i < this.count; i += 1 {
		res, err := this.init(i)
		if err != nil {
			log.Fatal(err)
		}

		this.clients[i] = res
	}
}

func (this *sandbox) init(clientId int) (res Client, err error) {
	defer func() {
		e := recover()
		if e != nil {
			if u, ok := e.(error); ok {
				err = u
			}
		}

		return
	}()

	res = &mgo_incr_client{}
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
	for i, client := range this.clients {
		if client != nil {
			_ = this.work(i, client)
		}
	}
}

func (this *sandbox) work(clientId int, client Client) (err error) {
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
	_ = client.Work()
	rt := time.Since(t0) / time.Microsecond

	this.emitter.WriteResponseTime(this.id, clientId, int64(rt))
	return
}

func (this *sandbox) teardown() {
	for i, client := range this.clients {
		if client != nil {
			this.close(client)
			this.clients[i] = nil
		}
	}

	this.wg.Done()
}

func (this *sandbox) close(client Client) (err error) {
	defer func() {
		e := recover()
		if e != nil {
			if u, ok := e.(error); ok {
				err = u
			}
		}

		return
	}()

	client.Close()
	return
}
