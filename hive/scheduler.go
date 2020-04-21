package hive

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"
)

type scheduler struct {
	registered map[string]handler
	workers    map[string]worker
	sync.Mutex
}

type handler struct {
	runnable Runnable
	options  workerOpts
}

func newScheduler() *scheduler {
	s := &scheduler{
		registered: map[string]handler{},
		workers:    map[string]worker{},
		Mutex:      sync.Mutex{},
	}

	return s
}

func (s *scheduler) schedule(job Job) *Result {
	if s.workers == nil {
		s.workers = map[string]worker{}
	}

	s.Lock() // this is probably unneeded but just being safe
	w, isStarted := s.workers[job.jobType]
	s.Unlock()

	if !isStarted {
		handler := s.getHandler(job.jobType)
		if handler == nil {
			result := newResult()
			result.sendErr(fmt.Errorf("failed to getRunnable for jobType %q", job.jobType))
			return result
		}

		newWorker := newGoWorker(handler.runnable, handler.options)

		// "recursively" pass this function as the runFunc for the runnable
		if err := newWorker.start(s.schedule); err != nil {
			result := newResult()
			result.sendErr(errors.Wrapf(err, "failed start worker for jobType %q", job.jobType))
			return result
		}

		s.Lock()
		s.workers[job.jobType] = newWorker
		s.Unlock()

		w = newWorker
	}

	return w.schedule(job)
}

// handle adds a handler
func (s *scheduler) handle(jobType string, runnable Runnable, options ...Option) {
	s.Lock()
	defer s.Unlock()

	// apply the provided options
	opts := defaultOpts()
	for _, o := range options {
		opts = o(opts)
	}

	h := handler{runnable, opts}
	if s.registered == nil {
		s.registered = map[string]handler{jobType: h}
	} else {
		s.registered[jobType] = h
	}
}

func (s *scheduler) getHandler(jobType string) *handler {
	s.Lock()
	defer s.Unlock()

	if s.registered == nil {
		return nil
	}

	if r, ok := s.registered[jobType]; ok {
		return &r
	}

	return nil
}