package rcc

import (
	"log"
	"sync"
	"time"
)

type Runner struct {
	jobInputQueue chan job
	resultQueue   chan statDataEntry
}

type job struct {
	repo      *SourceRepository
	clonePath string
	hash      string
	runColoc  bool
	runCover  bool
}

func NewRunner() *Runner {
	return &Runner{
		jobInputQueue: make(chan job),
		resultQueue:   make(chan statDataEntry),
	}
}

func (r *Runner) Run(workers int, hashes []string) {
	var wg sync.WaitGroup
	for w := 1; w <= workers; w++ {
		wg.Go(func() {
			worker(w, r.jobInputQueue, r.resultQueue)
		})
	}

	sd := statData{}
	go func() {
		for res := range r.resultQueue {
			sd.entries = append(sd.entries, res)
		}
	}()

	for _, h := range hashes {
		r.jobInputQueue <- job{hash: h}
	}
	close(r.jobInputQueue)

	wg.Wait()
	log.Printf("Results: %+v", sd)
}

func worker(id int, jobs <-chan job, results chan<- statDataEntry) {
	for job := range jobs {
		// clone to clonePath
		// checkout hash
		log.Printf("worker %d, hash %s", id, job.hash)
		time.Sleep(1 * time.Second)
		if job.runColoc {
			loc, err := getLoc(nil, nil, []string{job.clonePath})
			if err != nil {
				log.Printf("gocoloc failed: %s", err)
			}
			log.Printf("LOC: %v", loc)
		}
		results <- statDataEntry{ /* add loc, add cov*/ coverage: 10.0}
	}
}
