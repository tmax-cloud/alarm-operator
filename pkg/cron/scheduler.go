package cron

import (
	"context"
	"fmt"
	"sync"
	"time"
)

var current string

type Scheduler struct {
	ctx   context.Context
	ch    chan Job
	jobs  map[string]*Job
	mutex *sync.RWMutex
}

func NewScheduler(ctx context.Context, size int) *Scheduler {
	return &Scheduler{
		ctx:   ctx,
		ch:    make(chan Job, size),
		jobs:  make(map[string]*Job),
		mutex: new(sync.RWMutex),
	}
}

func (r *Scheduler) Start() {
	for {
		select {
		case job, ok := <-r.ch:
			if !ok {
				fmt.Println("closed job")
				continue
			}
			go job.Run()
		case <-r.ctx.Done():
			fmt.Println("scheduler done")
			return
		}
	}
}

func (r *Scheduler) Schedule(name string) *Scheduler {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	current = name
	if _, ok := r.jobs[current]; !ok {
		r.jobs[current] = NewJob(name)
	}

	return r
}

func (r *Scheduler) Every(interval interface{}) *Scheduler {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	switch interval := interval.(type) {
	case int:
		r.jobs[current].interval = time.Duration(interval)
	case time.Duration:
		r.jobs[current].interval = interval
	}
	return r
}

func (r *Scheduler) Second() *Scheduler {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	r.jobs[current].unit = seconds
	return r
}

func (r *Scheduler) Do(taskFn TaskFunc) *Scheduler {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.jobs[current].taskFunc = taskFn
	if r.jobs[current].isRunning {
		r.jobs[current].cancel()
		r.jobs[current] = r.jobs[current].Clone()
	}

	this := r.jobs[current]
	go func() {
		this.isRunning = true
		ticker := time.NewTicker(this.GetInterval())
		for {
			select {
			case <-ticker.C:
				r.ch <- *this
				ticker.Reset(this.GetInterval())
			case <-this.ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()

	return r
}

func (r *Scheduler) Cancel(name string) {
	if j, ok := r.jobs[name]; ok {
		j.cancel()
	}
}
