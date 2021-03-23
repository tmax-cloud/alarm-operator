package scheduler

import (
	"context"
	"fmt"
	"time"
)

type JobRunner struct {
	ctx  context.Context
	ch   chan IntervalJob
	jobs map[string]*IntervalJob
}

func NewJobRunner(ctx context.Context, size int) *JobRunner {
	return &JobRunner{
		ctx:  ctx,
		ch:   make(chan IntervalJob, size),
		jobs: make(map[string]*IntervalJob),
	}
}

func (r *JobRunner) Start() {
	for {
		select {
		case job, ok := <-r.ch:
			if !ok {
				fmt.Println("closed job channel")
				continue
			}
			go job.Run()
		case <-r.ctx.Done():
			fmt.Println("context done")
			return
		}
	}
}

func (r *JobRunner) Schedule(job *IntervalJob) {
	if j, ok := r.jobs[job.Name()]; ok {
		j.Cancel()
	}

	r.jobs[job.Name()] = job

	go func() {
		t := time.NewTimer(job.Interval())
		for {
			select {
			case <-t.C:
				r.ch <- *job
				t.Reset(job.Interval())
			case <-job.Done():
				return
			}
		}
	}()
}

func (r *JobRunner) CancelJob(id string) {
	if j, ok := r.jobs[id]; ok {
		j.Cancel()
	}
}
