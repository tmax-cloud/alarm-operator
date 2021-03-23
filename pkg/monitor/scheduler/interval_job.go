package scheduler

import (
	"context"
	"fmt"
	"time"
)

type IntervalJob struct {
	ctx      context.Context
	cancel   context.CancelFunc
	name     string
	interval time.Duration
	task     Task
}

func NewIntervalJob(name string, interval time.Duration, task Task) *IntervalJob {
	ctx, cancel := context.WithCancel(context.Background())
	return &IntervalJob{
		ctx:      ctx,
		cancel:   cancel,
		name:     name,
		interval: time.Millisecond * interval,
		task:     task,
	}
}

func (j *IntervalJob) Run() {
	fmt.Println(">>   ", j.name)
	if err := j.task.Run(); err != nil {
		fmt.Println(err)
	}

	fmt.Println("  << ", j.name)
}

func (j *IntervalJob) Name() string {
	return j.name
}

func (j *IntervalJob) Interval() time.Duration {
	return j.interval
}

func (j *IntervalJob) Cancel() {
	j.cancel()
}

func (j *IntervalJob) Done() <-chan struct{} {
	return j.ctx.Done()
}
