package main

import (
	"fmt"
	"time"
)

type Job struct {
	ID   int
	Data string
}

type Result struct {
	Job    Job
	Result string
}

func worker(id int, jobs <-chan Job, results chan<- Result) {
	for job := range jobs {
		fmt.Printf("Worker %d 处理任务 %d\n", id, job.ID)
		time.Sleep(100 * time.Millisecond)

		results <- Result{
			Job:    job,
			Result: fmt.Sprintf("处理完成:%s", job.Data),
		}
	}
}

func main() {
	jobs := make(chan Job, 10)
	results := make(chan Result, 10)

	for w := 1; w <= 3; w++ {
		go worker(w, jobs, results)
	}

	for j := 1; j <= 9; j++ {
		jobs <- Job{
			ID:   j,
			Data: fmt.Sprintf("交易 #%d", j),
		}
	}
	close(jobs)

	for a := 1; a <= 9; a++ {
		result := <-results
		fmt.Println(result.Result)
	}
}
