// Copyright 2016 DeepFabric, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package task

import (
	"errors"
	"fmt"
	"sync"
	"time"
	"context"
)

var (
	errUnavailable = errors.New("Stopper is unavailable")
)

var (
	defaultWaitStoppedTimeout       = time.Minute
	defaultQueueName                = "__default__"
	batch                     int64 = 64

	jobPool sync.Pool
)

type state int

const (
	running  = state(0)
	stopping = state(1)
	stopped  = state(2)
)

var (
	// ErrJobCancelled error job cancelled
	ErrJobCancelled = errors.New("Job cancelled")
)

// JobState is the job state
type JobState int

const (
	// Pending job is wait to running
	Pending = JobState(0)
	// Running job is running
	Running = JobState(1)
	// Cancelling job is cancelling
	Cancelling = JobState(2)
	// Cancelled job is cancelled
	Cancelled = JobState(3)
	// Finished job is complete
	Finished = JobState(4)
	// Failed job is failed when execute
	Failed = JobState(5)
)

func acquireJob() *Job {
	v := jobPool.Get()
	if v == nil {
		return &Job{}
	}
	return v.(*Job)
}

func releaseJob(job *Job) {
	job.reset()
	jobPool.Put(job)
}

// Job is do for something with state
type Job struct {
	sync.RWMutex

	desc        string
	fun         func() error
	state       JobState
	result      interface{}
	needRelease bool
}

func newJob(desc string, fun func() error) *Job {
	job := acquireJob()
	job.fun = fun
	job.state = Pending
	job.desc = desc
	job.needRelease = true

	return job
}

func (job *Job) reset() {
	job.Lock()
	job.fun = nil
	job.state = Pending
	job.desc = ""
	job.result = nil
	job.needRelease = true
	job.Unlock()
}

// IsComplete return true means the job is complete.
func (job *Job) IsComplete() bool {
	return !job.IsNotComplete()
}

// IsNotComplete return true means the job is not complete.
func (job *Job) IsNotComplete() bool {
	job.RLock()
	yes := job.state == Pending || job.state == Running || job.state == Cancelling
	job.RUnlock()

	return yes
}

// SetResult set result
func (job *Job) SetResult(result interface{}) {
	job.Lock()
	job.result = result
	job.Unlock()
}

// GetResult returns job result
func (job *Job) GetResult() interface{} {
	job.RLock()
	r := job.result
	job.RUnlock()
	return r
}

// Cancel cancel the job
func (job *Job) Cancel() {
	job.Lock()
	if job.state == Pending {
		job.state = Cancelling
	}
	job.Unlock()
}

// IsRunning returns true if job state is Running
func (job *Job) IsRunning() bool {
	return job.isSpecState(Running)
}

// IsPending returns true if job state is Pending
func (job *Job) IsPending() bool {
	return job.isSpecState(Pending)
}

// IsFinished returns true if job state is Finished
func (job *Job) IsFinished() bool {
	return job.isSpecState(Finished)
}

// IsCancelling returns true if job state is Cancelling
func (job *Job) IsCancelling() bool {
	return job.isSpecState(Cancelling)
}

// IsCancelled returns true if job state is Cancelled
func (job *Job) IsCancelled() bool {
	return job.isSpecState(Cancelled)
}

// IsFailed returns true if job state is Failed
func (job *Job) IsFailed() bool {
	return job.isSpecState(Failed)
}

func (job *Job) isSpecState(spec JobState) bool {
	job.RLock()
	yes := job.state == spec
	job.RUnlock()

	return yes
}

func (job *Job) setState(state JobState) {
	job.state = state
}

func (job *Job) getState() JobState {
	job.RLock()
	s := job.state
	job.RUnlock()

	return s
}

// Runner TODO
type Runner struct {
	sync.RWMutex

	stop       sync.WaitGroup
	stopC      chan struct{}
	lastID     uint64
	cancels    map[uint64]context.CancelFunc
	state      state
	namedQueue map[string]*Queue
}

// NewRunner returns a task runner
func NewRunner() *Runner {
	t := &Runner{
		stopC:      make(chan struct{}),
		state:      running,
		namedQueue: make(map[string]*Queue),
		cancels:    make(map[uint64]context.CancelFunc),
	}

	t.AddNamedWorker(defaultQueueName)
	return t
}

// AddNamedWorker add a named worker, the named worker has uniq queue, so jobs are linear execution
func (s *Runner) AddNamedWorker(name string) (uint64, error) {
	s.Lock()
	q, ok := s.namedQueue[name]
	if !ok {
		q = &Queue{}
		s.namedQueue[name] = q
	}
	s.Unlock()

	return s.startWorker(name, q)
}

// IsNamedWorkerBusy returns true if named queue is not empty
func (s *Runner) IsNamedWorkerBusy(worker string) bool {
	s.RLock()
	defer s.RUnlock()

	return s.getNamedQueue(worker).Len() > 0
}

func (s *Runner) startWorker(name string, q *Queue) (uint64, error) {
	return s.RunCancelableTask(func(ctx context.Context) {
		jobs := make([]interface{}, batch, batch)

		for {
			n, err := q.Get(batch, jobs)
			if err != nil {
				return
			}

			for i := int64(0); i < n; i++ {
				job := jobs[i].(*Job)

				job.Lock()

				switch job.state {
				case Pending:
					job.setState(Running)
					job.Unlock()
					err := job.fun()
					job.Lock()
					if err != nil {
						if err == ErrJobCancelled {
							job.setState(Cancelled)
						} else {
							job.setState(Failed)
						}
					} else {
						if job.state == Cancelling {
							job.setState(Cancelled)
						} else {
							job.setState(Finished)
						}
					}
				case Cancelling:
					job.setState(Cancelled)
				}

				job.Unlock()

				if job.needRelease {
					releaseJob(job)
				}
			}
		}
	})
}

// RunJob run a job
func (s *Runner) RunJob(desc string, task func() error) error {
	return s.RunJobWithNamedWorker(desc, defaultQueueName, task)
}

// RunJobWithNamedWorker run a job in a named worker
func (s *Runner) RunJobWithNamedWorker(desc, worker string, task func() error) error {
	return s.RunJobWithNamedWorkerWithCB(desc, worker, task, nil)
}

// RunJobWithNamedWorkerWithCB run a job in a named worker
func (s *Runner) RunJobWithNamedWorkerWithCB(desc, worker string, task func() error, cb func(*Job)) error {
	s.Lock()

	if s.state != running {
		s.Unlock()
		return errUnavailable
	}

	job := newJob(desc, task)
	q := s.getNamedQueue(worker)
	if q == nil {
		s.Unlock()
		return fmt.Errorf("named worker %s is not exists", worker)
	}

	if cb != nil {
		job.needRelease = false
		cb(job)
	}

	q.Put(job)

	s.Unlock()
	return nil
}

// RunCancelableTask run a task that can be cancelled
// Example:
// err := s.RunCancelableTask(func(ctx context.Context) {
// 	select {
// 	case <-ctx.Done():
// 	// cancelled
// 	case <-time.After(time.Second):
// 		// do something
// 	}
// })
// if err != nil {
// 	// hanle error
// 	return
// }
func (s *Runner) RunCancelableTask(task func(context.Context)) (uint64, error) {
	s.Lock()
	defer s.Unlock()

	if s.state != running {
		return 0, errUnavailable
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.lastID++
	id := s.lastID
	s.cancels[id] = cancel
	s.stop.Add(1)

	go func() {
		if err := recover(); err != nil {
			panic(err)
		}
		defer s.stop.Done()
		task(ctx)
	}()

	return id, nil
}

// RunTask runs a task in new goroutine
func (s *Runner) RunTask(task func()) error {
	s.Lock()
	defer s.Unlock()

	if s.state != running {
		return errUnavailable
	}

	s.stop.Add(1)

	go func() {
		defer s.stop.Done()
		task()
	}()

	return nil
}

// StopCancelableTask stop cancelable spec task
func (s *Runner) StopCancelableTask(id uint64) error {
	s.Lock()
	defer s.Unlock()

	if s.state == stopping ||
		s.state == stopped {
		return errors.New("stopper is already stoppped")
	}

	cancel, ok := s.cancels[id]
	if !ok {
		return fmt.Errorf("target task<%d> not found", id)
	}

	delete(s.cancels, id)
	cancel()
	return nil
}

// Stop stop all task
// RunTask will failure with an error
// Wait complete for the tasks that already in execute
// Cancel the tasks that is not start
func (s *Runner) Stop() error {
	s.Lock()
	defer s.Unlock()

	if s.state == stopping ||
		s.state == stopped {
		return errors.New("stopper is already stoppped")
	}
	s.state = stopping

	for _, cancel := range s.cancels {
		cancel()
	}

	for _, q := range s.namedQueue {
		q.Dispose()
	}

	go func() {
		s.stop.Wait()
		s.stopC <- struct{}{}
	}()

	select {
	case <-time.After(defaultWaitStoppedTimeout):
		return errors.New("waitting for task complete timeout")
	case <-s.stopC:
	}

	s.state = stopped
	return nil
}

func (s *Runner) getDefaultQueue() *Queue {
	return s.getNamedQueue(defaultQueueName)
}

func (s *Runner) getNamedQueue(name string) *Queue {
	return s.namedQueue[name]
}
