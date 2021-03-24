package cron

import (
	"context"
	"fmt"
	"time"
)

type TaskFunc func(context.Context) error

type timeUnit int

const (
	// default unit is seconds
	milliseconds timeUnit = iota
	seconds
	minutes
	hours
	days
	weeks
	months
	duration
)

type Job struct {
	name      string
	interval  time.Duration
	unit      timeUnit
	taskFunc  TaskFunc
	ctx       context.Context
	cancel    context.CancelFunc
	isRunning bool
}

func NewJob(name string) *Job {
	ctx, cancel := context.WithCancel(context.Background())
	return &Job{
		name:      name,
		unit:      seconds,
		ctx:       ctx,
		cancel:    cancel,
		isRunning: false,
	}
}

func (j *Job) Run() {
	fmt.Println(">>   ", j.name)
	j.taskFunc(j.ctx)
	fmt.Println("  << ", j.name)
}

func (j *Job) Name() string {
	return j.name
}

func (j *Job) Cancel() {
	j.cancel()
}

func (j *Job) Done() <-chan struct{} {
	return j.ctx.Done()
}
