package main

import (
	"netpollmux/internal/log"
	"netpollmux/internal/scheduler"
	"sync"
)

func main() {
	log.SetPrefix("😁")
	s := scheduler.NewScheduler(64, nil)
	wg := &sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		task := func() {
			log.Info("scheduler")
			wg.Done()
		}
		s.Schedule(task)
	}
	wg.Wait()
	s.Close()
}
