package scheduler

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const (
	threshold = 2
	idleTime  = time.Second
	interval  = time.Second
)

// Scheduler represents a task scheduler.
type Scheduler interface {
	// Schedule schedules the task to an idle worker.
	Schedule(task func())
	// NumWorkers returns the number of workers.
	NumWorkers() int
	// Close closes the task scheduler.
	Close()
}

// Options represents options
type Options struct {
	// Threshold option represents the threshold at which batch task is enabled.
	Threshold int
	// IdleTime option represents the max idle time of the worker.
	IdleTime time.Duration
	// Interval option represents the check interval of the worker.
	Interval time.Duration
}

// DefaultOptions returns default options.
func DefaultOptions() *Options {
	return &Options{
		Threshold: threshold,
		IdleTime:  idleTime,
		Interval:  interval,
	}
}

type scheduler struct {
	lock       sync.Mutex
	cond       sync.Cond
	wg         sync.WaitGroup
	pending    []func()
	tasks      int64
	running    map[*worker]struct{}
	workers    int64
	maxWorkers int64
	done       chan struct{}
	closed     int32
	opts       Options
}

// NewScheduler returns a new task scheduler.
func NewScheduler(maxWorkers int, opts *Options) Scheduler {
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}
	if opts == nil {
		opts = DefaultOptions()
	} else {
		if opts.IdleTime <= 0 {
			opts.IdleTime = idleTime
		}
		if opts.Interval <= 0 {
			opts.Interval = interval
		}
	}
	s := &scheduler{
		running:    make(map[*worker]struct{}),
		maxWorkers: int64(maxWorkers),
		done:       make(chan struct{}),
		opts:       *opts,
	}
	s.cond.L = &s.lock
	s.wg.Add(1)
	go s.run()
	return s
}

// Schedule schedules the task to an idle worker.
func (s *scheduler) Schedule(task func()) {
	if atomic.LoadInt32(&s.closed) > 0 {
		panic("schedule tasks on a closed scheduler")
	}
	workers := atomic.LoadInt64(&s.workers)
	if atomic.AddInt64(&s.tasks, 1) > workers && workers < s.maxWorkers {
		if atomic.AddInt64(&s.workers, 1) <= s.maxWorkers {
			s.wg.Add(1)
			w := &worker{}
			s.lock.Lock()
			s.running[w] = struct{}{}
			s.lock.Unlock()
			go w.run(s, task)
			return
		}
	}
	s.lock.Lock()
	s.pending = append(s.pending, task)
	s.lock.Unlock()
	s.cond.Signal()
}

// NumWorkers returns the number of workers.
func (s *scheduler) NumWorkers() int {
	return int(atomic.LoadInt64(&s.workers))
}

// Close closes the task scheduler.
func (s *scheduler) Close() {
	if atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		close(s.done)
	}
	s.wg.Wait()
}

func (s *scheduler) run() {
	defer s.wg.Done()
	ticker := time.NewTicker(s.opts.Interval)
	var doned bool
	var idle bool
	var lastIdleTime time.Time
	for {
		if !doned {
			select {
			case <-ticker.C:
				if atomic.LoadInt64(&s.workers) > 0 && atomic.LoadInt64(&s.workers) > atomic.LoadInt64(&s.tasks) {
					if !idle {
						idle = true
						lastIdleTime = time.Now()
					} else if time.Now().Sub(lastIdleTime) > s.opts.IdleTime {
						s.lock.Lock()
						deletions := len(s.running) - len(s.pending)
						if deletions > 4 {
							deletions = deletions / 4
						} else if deletions > 0 {
							deletions = 1
						}
						if deletions > 0 {
							for w := range s.running {
								delete(s.running, w)
								w.close()
								deletions--
								if deletions == 0 {
									break
								}
							}
						}
						s.lock.Unlock()
						s.cond.Broadcast()
						idle = false
						lastIdleTime = time.Time{}
					}
				} else {
					idle = false
					lastIdleTime = time.Time{}
				}
			case <-s.done:
				ticker.Stop()
				doned = true
			}
		} else {
			if atomic.LoadInt64(&s.workers) == 0 {
				break
			}
			s.cond.Broadcast()
			time.Sleep(time.Millisecond)
		}
	}
}

type worker struct {
	closed int32
}

func (w *worker) run(s *scheduler, task func()) {
	var maxWorkers = int(s.maxWorkers)
	var batch []func()
	for {
		if len(batch) > 0 {
			for _, task = range batch {
				task()
			}
			atomic.AddInt64(&s.tasks, -int64(len(batch)))
			batch = batch[:0]
		} else {
			task()
			atomic.AddInt64(&s.tasks, -1)
		}
		s.lock.Lock()
		for {
			if s.opts.Threshold > 1 && len(s.pending) > maxWorkers*s.opts.Threshold {
				alloc := len(s.pending) / maxWorkers
				batch = s.pending[:alloc]
				s.pending = s.pending[alloc:]
				break
			} else if len(s.pending) > 0 {
				task = s.pending[0]
				s.pending = s.pending[1:]
				break
			}
			s.cond.Wait()
			if atomic.LoadInt32(&s.closed) > 0 || atomic.LoadInt32(&w.closed) > 0 {
				s.lock.Unlock()
				atomic.AddInt64(&s.workers, -1)
				s.wg.Done()
				return
			}
		}
		s.lock.Unlock()
	}
}

func (w *worker) close() {
	atomic.CompareAndSwapInt32(&w.closed, 0, 1)
}
