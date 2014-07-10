package s3utils

import (
	"sync"
)

type S3Worker struct {
	Worker int
	Tasks  []S3Task
}

// Adds download tasks to the queue of things to download
func (s *S3Worker) Add(task S3Task) {
	s.Tasks = append(s.Tasks, task)
}

// Performs the actual downloads
func (s *S3Worker) Run() {
	tasks := make(chan S3Task, 64)
	// spawn goroutines workers
	var wg sync.WaitGroup
	for i := 0; i < s.Worker; i++ {
		wg.Add(1)
		go func() {
			for task := range tasks {
				task.Perform()
			}
			wg.Done()
		}()
	}

	for _, task := range s.Tasks {
		tasks <- task
	}
	close(tasks)

	// wait for the workers to finish
	wg.Wait()
}
