package main

import (
	"errors"
	"launchpad.net/tomb"
	"log"
	"strconv"
	"sync"
)

const (
	DEFAULT_LATENCY_EVENT_CHANNEL_SIZE = 0 // Unbuffered
)

type LatencyEmitter interface {
	PublishResponseTime(clientId int, latency int64, res WorkResult)
}

type LatencyEvent struct {
	id   int
	usec int64
}

type LatencyEventsChannel <-chan *LatencyEvent

type TaskMasterInfo struct {
	WaitGroup  *sync.WaitGroup
	Properties map[string]string
}

type taskmaster struct {
	t      tomb.Tomb
	ch     chan *LatencyEvent
	wg     *sync.WaitGroup
	props  map[string]string
	chSize int
}

func NewTaskMaster(info *TaskMasterInfo) *taskmaster {
	return &taskmaster{
		wg:     info.WaitGroup,
		props:  info.Properties,
		ch:     nil,
		chSize: DEFAULT_LATENCY_EVENT_CHANNEL_SIZE,
	}
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

func (this *taskmaster) PublishResponseTime(clientId int, latency int64, res WorkResult) {
	this.ch <- &LatencyEvent{clientId, latency}
}

func (this *taskmaster) loop() {
	defer this.t.Done()

	this.setup()

	for {
		select {
		case <-this.t.Dying():
			this.shutdown()
			return
		default:
			this.wg.Wait()
			this.t.Kill(nil)
		}
	}
}

func (this *taskmaster) setup() {
	err := this.parseProperties(this.props)
	if err != nil {
		log.Fatal(err)
	}

	if this.chSize == 0 {
		this.ch = make(chan *LatencyEvent)
	} else {
		this.ch = make(chan *LatencyEvent, this.chSize)
	}
}

func (this *taskmaster) shutdown() {
	close(this.ch)
}

func (this *taskmaster) parseProperties(props map[string]string) (err error) {
	if v, ok := props["internals.LatencyEventChannelSize"]; ok {
		w, err := strconv.Atoi(v)
		if err != nil || w < 0 {
			return errors.New("internals.LatencyEventChannelSize must be >= 0")
		}

		this.chSize = w
	} else {
		this.chSize = DEFAULT_LATENCY_EVENT_CHANNEL_SIZE
	}

	return
}
