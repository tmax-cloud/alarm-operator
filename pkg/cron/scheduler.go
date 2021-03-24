package cron

import (
	"context"
	"fmt"
	"time"
)

var current string

type Scheduler struct {
	ctx  context.Context
	ch   chan Job
	jobs map[string]*Job
}

func NewScheduler(ctx context.Context, size int) *Scheduler {
	return &Scheduler{
		ctx:  ctx,
		ch:   make(chan Job, size),
		jobs: make(map[string]*Job),
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
	current = name
	if _, ok := r.jobs[current]; !ok {
		r.jobs[current] = NewJob(name)
	}

	return r
}

func (r *Scheduler) Every(interval interface{}) *Scheduler {
	switch interval := interval.(type) {
	case int:
		r.jobs[current].interval = time.Duration(interval)
	case time.Duration:
		r.jobs[current].interval = interval
	}
	return r
}

func (r *Scheduler) Second() *Scheduler {
	r.jobs[current].unit = seconds
	return r
}

func (r *Scheduler) Do(taskFn TaskFunc) *Scheduler {
	this := r.jobs[current]
	if this.isRunning {
		return r
	}

	go func() {
		this.isRunning = true
		this.taskFunc = taskFn

		calcDuration := func() time.Duration {
			var duration time.Duration
			switch this.unit {
			case seconds:
				duration = time.Second * this.interval
			}
			return duration
		}

		ticker := time.NewTicker(calcDuration())
		for {
			select {
			case <-ticker.C:
				r.ch <- *this
				ticker.Reset(calcDuration())
			case <-this.Done():
				fmt.Println(this.name, " Done")
				ticker.Stop()
				return
			}
		}
	}()

	return r
}

func (r *Scheduler) Cancel(name string) {
	if j, ok := r.jobs[name]; ok {
		j.Cancel()
	}
}
