package rcc

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Runner struct {
	jobInputQueue chan job
	resultQueue   chan statDataEntry
	repo          *SourceRepository
}

type job struct {
	repo            *SourceRepository
	clonePath       string
	hash            string
	runColoc        bool
	runColocNoTests bool
	runCover        bool
}

func NewRunner(repo *SourceRepository) *Runner {
	return &Runner{
		repo:          repo,
		jobInputQueue: make(chan job),
		resultQueue:   make(chan statDataEntry),
	}
}

func (r *Runner) Run(workers int, hashes []string) {
	var wg sync.WaitGroup
	for w := 1; w <= workers; w++ {
		wg.Go(func() {
			worker(w, r.jobInputQueue, r.resultQueue)
			log.Printf("Worker %d quit", w)
		})
	}

	sd := statData{}
	go func() {
		for res := range r.resultQueue {
			sd.entries = append(sd.entries, res)
		}
	}()

	tmpDir, err := os.MkdirTemp(os.TempDir(), "rcc")
	if err != nil {
		panic(err)
	}
	for _, h := range hashes {
		clonePath := filepath.Join(tmpDir, h)
		r.jobInputQueue <- job{hash: h, runColoc: false, repo: r.repo, clonePath: clonePath}
	}
	close(r.jobInputQueue)
	wg.Wait()

	// sd.sort()
	for _, s := range sd.entries {
		log.Printf("RES %+v", s)
	}
	for _, l := range sd.languages() {
		log.Printf("LANG %+s", l)
	}

	if err := os.RemoveAll(tmpDir); err != nil {
		log.Printf("failed to remove %s", tmpDir)
	}
}

func worker(id int, jobs <-chan job, results chan<- statDataEntry) {
	for job := range jobs {
		log.Printf("worker %d, hash %s, wd %s", id, job.hash, job.clonePath)
		// clone to clonePath with given hash
		err := job.repo.LocalClone(job.clonePath, job.hash)
		if err != nil {
			log.Printf("failed to clone: %s", err)
		}

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
