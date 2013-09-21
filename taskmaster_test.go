package main

import (
	"log"
	"sync"
	"testing"
	"time"
)

func spawnTask(emitter LatencyEmitter, wg *sync.WaitGroup, count int) {
	go func() {
		log.Print("test task starting")

		for i := 0; i < count; i += 1 {
			emitter.PublishResponseTime(-1, int64(i), WRK_OK)
			<-time.After(25 * time.Millisecond)
		}

		log.Print("test task completed")

		wg.Done()
	}()
}

func countResponseTimes(tm *taskmaster) (count int) {
	c := 0
	for {
		select {
		case <-tm.t.Dead():
			return c
		case _, ok := <-tm.ResponseTimes():
			// If we don't check whether the channel is closed,
			// then the count will sometimes be off (high).
			if !ok {
				break
			}
			c += 1
		}
	}
}

func TestTaskMaster(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	tm := NewTaskMaster(wg)
	tm.Start()

	spawnTask(tm, wg, 50)

	c := countResponseTimes(tm)

	if !expectInt(t, 50, c) {
		return
	}
}
