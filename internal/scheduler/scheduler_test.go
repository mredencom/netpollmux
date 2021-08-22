package scheduler

import (
	"sync"
	"testing"
	"time"
)

func TestSchedulerPreheat(t *testing.T) {
	var (
		total       = 10000000
		concurrency = 256
	)
	opts := DefaultOptions()
	opts.Threshold = 0
	s := NewScheduler(concurrency, opts)
	wg := &sync.WaitGroup{}
	for i := 0; i < total; i++ {
		wg.Add(1)
		job := func() {
			wg.Done()
		}
		s.Schedule(job)
	}
	wg.Wait()
	s.Close()
}

func TestScheduler(t *testing.T) {
	var (
		total       = 10000000
		concurrency = 1
	)
	opts := DefaultOptions()
	opts.Threshold = 0
	s := NewScheduler(concurrency, opts)
	wg := &sync.WaitGroup{}
	start := time.Now()
	for i := 0; i < total; i++ {
		wg.Add(1)
		job := func() {
			wg.Done()
		}
		s.Schedule(job)
	}
	wg.Wait()
	d := time.Now().Sub(start)
	t.Logf("Concurrency:%d, Batch %t, Time:%v, %vns/op", concurrency, DefaultOptions().Threshold > 1, d, int64(d)/int64(total))
	s.Close()
}

func TestSchedulerBatch(t *testing.T) {
	var (
		total       = 10000000
		concurrency = 1
	)
	opts := DefaultOptions()
	opts.Threshold = 2
	s := NewScheduler(concurrency, opts)
	wg := &sync.WaitGroup{}
	start := time.Now()
	for i := 0; i < total; i++ {
		wg.Add(1)
		job := func() {
			wg.Done()
		}
		s.Schedule(job)
	}
	wg.Wait()
	d := time.Now().Sub(start)
	t.Logf("Concurrency:%d, Batch %t, Time:%v, %vns/op", concurrency, opts.Threshold > 1, d, int64(d)/int64(total))
	s.Close()
}

func TestSchedulerConcurrency(t *testing.T) {
	var (
		total       = 10000000
		concurrency = 256
	)
	opts := DefaultOptions()
	opts.Threshold = 0
	s := NewScheduler(concurrency, opts)
	wg := &sync.WaitGroup{}
	start := time.Now()
	for i := 0; i < total; i++ {
		wg.Add(1)
		job := func() {
			wg.Done()
		}
		s.Schedule(job)
	}
	wg.Wait()
	d := time.Now().Sub(start)
	t.Logf("Concurrency:%d, Batch %t, Time:%v, %vns/op", concurrency, DefaultOptions().Threshold > 1, d, int64(d)/int64(total))
	s.Close()
}

func TestSchedulerBatchConcurrency(t *testing.T) {
	var (
		total       = 10000000
		concurrency = 256
	)
	opts := DefaultOptions()
	opts.Threshold = 2
	s := NewScheduler(concurrency, opts)
	wg := &sync.WaitGroup{}
	start := time.Now()
	for i := 0; i < total; i++ {
		wg.Add(1)
		job := func() {
			wg.Done()
		}
		s.Schedule(job)
	}
	wg.Wait()
	d := time.Now().Sub(start)
	t.Logf("Concurrency:%d, Batch %t, Time:%v, %vns/op", concurrency, opts.Threshold > 1, d, int64(d)/int64(total))
	s.Close()
}

func TestSchedulerOptions(t *testing.T) {
	{
		var (
			total       = 1000000
			concurrency = 0
		)
		s := NewScheduler(concurrency, nil)
		wg := &sync.WaitGroup{}
		for i := 0; i < total; i++ {
			wg.Add(1)
			job := func() {
				wg.Done()
			}
			s.Schedule(job)
		}
		wg.Wait()
		s.Close()
	}
	{
		var (
			total       = 1000000
			concurrency = 0
		)
		opts := &Options{}
		s := NewScheduler(concurrency, opts)
		wg := &sync.WaitGroup{}
		for i := 0; i < total; i++ {
			wg.Add(1)
			job := func() {
				wg.Done()
			}
			s.Schedule(job)
		}
		wg.Wait()
		s.Close()
	}
	{
		var (
			total       = 1000000
			concurrency = 1
		)
		opts := DefaultOptions()
		opts.Threshold = 2
		opts.IdleTime = time.Millisecond * 3
		opts.Interval = time.Millisecond * 1
		s := NewScheduler(concurrency, opts)
		wg := &sync.WaitGroup{}
		for i := 0; i < total; i++ {
			wg.Add(1)
			job := func() {
				wg.Done()
			}
			s.Schedule(job)
		}
		wg.Wait()
		time.Sleep(time.Millisecond * 10)
		s.Close()
	}
	{
		var (
			total       = 10000
			concurrency = 64
		)
		opts := DefaultOptions()
		opts.Threshold = 2
		opts.IdleTime = time.Millisecond * 30
		opts.Interval = time.Millisecond * 10
		s := NewScheduler(concurrency, opts)
		wg := &sync.WaitGroup{}
		for i := 0; i < total; i++ {
			wg.Add(1)
			job := func() {
				wg.Done()
				time.Sleep(time.Millisecond)
			}
			s.Schedule(job)
		}
		wg.Wait()
		s.Close()
	}
	func() {
		defer func() {
			if e := recover(); e == nil {
				t.Error("should panic")
			}
		}()
		var (
			total       = 1000000
			concurrency = 0
		)
		opts := &Options{}
		s := NewScheduler(concurrency, opts)
		wg := &sync.WaitGroup{}
		for i := 0; i < total; i++ {
			wg.Add(1)
			job := func() {
				wg.Done()
			}
			s.Schedule(job)
		}
		wg.Wait()
		s.Close()
		s.Schedule(func() {})
	}()
	{
		var (
			total       = 100000
			concurrency = 64
		)
		opts := &Options{}
		s := NewScheduler(concurrency, opts)
		wg := &sync.WaitGroup{}
		for i := 0; i < total; i++ {
			wg.Add(1)
			job := func() {
				wg.Done()
				time.Sleep(time.Millisecond)
			}
			s.Schedule(job)
		}
		numWorkers := s.NumWorkers()
		if numWorkers != concurrency && concurrency > 0 {
			t.Errorf("NumWorkers %d != Concurrency %d", numWorkers, concurrency)
		}
		wg.Wait()
		s.Close()
	}
}
