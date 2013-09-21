package main

import (
	"launchpad.net/tomb"
	"sync"
)

type LatencyEmitter interface {
	PublishResponseTime(clientId int, latency int64)
}

type LatencyEvent struct {
	id   int
	usec int64
}

type LatencyEventsChannel <-chan *LatencyEvent

type taskmaster struct {
	ch chan *LatencyEvent
	t  tomb.Tomb
	wg *sync.WaitGroup
}

func NewTaskMaster(wg *sync.WaitGroup) *taskmaster {
	return &taskmaster{ch: make(chan *LatencyEvent), wg: wg}
}

func (this *taskmaster) Start() {
	go this.loop()
}

func (this *taskmaster) Stop() (err error) {
	this.t.Kill(nil)
	return this.t.Wait()
}

func (this *taskmaster) ResponseTimes() LatencyEventsChannel {
	return this.ch
}

func (this *taskmaster) PublishResponseTime(clientId int, latency int64) {
	this.ch <- &LatencyEvent{clientId, latency}
}

func (this *taskmaster) loop() {
	defer this.t.Done()

	for {
		select {
		case <-this.t.Dying():
			this.dispose()
			return
		default:
			this.wg.Wait()
			this.t.Kill(nil)
		}
	}
}

func (this *taskmaster) dispose() {
	close(this.ch)
}
