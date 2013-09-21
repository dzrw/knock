package main

import (
	"launchpad.net/tomb"
	"sync"
)

// TIME=1 HOUR (stats funnel, waitgroup)
// observer:
// spawn a goroutine with a tomb and a waitgroup to monitor slaves
// observer exposes a Receive() <-chan for master to listen for stats, and
//  a Report(slaveid, latency) method for slaves to write stats.
//  internally, this wraps around an unbuffered channel that can later be
//  expanded into a bufferred or elastic channel (or multiple channels)
//  if it becomes a bottleneck.

type LatencyEmitter interface {
	WriteResponseTime(hostId, clientId int, latency int64)
}

type LatencyEventsChannel <-chan int64

type taskmaster struct {
	ch chan int64
	t  tomb.Tomb
	wg *sync.WaitGroup
}

func NewTaskMaster(wg *sync.WaitGroup) *taskmaster {
	return &taskmaster{ch: make(chan int64), wg: wg}
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

func (this *taskmaster) WriteResponseTime(hostId, clientId int, latency int64) {
	this.ch <- latency
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
