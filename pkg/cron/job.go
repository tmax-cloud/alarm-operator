package cron

import (
	"context"
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

func (j *Job) Clone() *Job {
	ctx, cancel := context.WithCancel(context.Background())
	return &Job{
		name:      j.name,
		interval:  j.interval,
		unit:      j.unit,
		taskFunc:  j.taskFunc,
		ctx:       ctx,
		cancel:    cancel,
		isRunning: j.isRunning,
	}
}

func (j *Job) Run() {
	j.taskFunc(j.ctx)
}

func (j *Job) GetInterval() time.Duration {
	// TODO:
	switch j.unit {
	case seconds:
		return time.Second * j.interval
	}

	return time.Second * j.interval
}
