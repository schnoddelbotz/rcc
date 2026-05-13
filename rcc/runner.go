package rcc

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hhatto/gocloc"
)

type Runner struct {
	jobInputQueue chan job
	resultQueue   chan statDataEntry
	repo          *SourceRepository
	statData      *statData
	jobsDone      atomic.Uint32
}

type job struct {
	repo                *SourceRepository
	clonePath           string
	hash                string
	runColoc            bool
	runColocNoTests     bool
	runCoverUnit        bool
	runCoverIntegration bool
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
			// log.Printf("Worker %d quit", w)
		})
	}

	sd := statData{}
	done := make(chan struct{})
	go func() {
		for res := range r.resultQueue {
			sd.entries = append(sd.entries, res)
			r.jobsDone.Add(1)
		}
		close(done)
	}()
	go func() {
		for r.jobsDone.Load() != uint32(len(hashes)) {
			time.Sleep(200 * time.Millisecond)
			fmt.Printf("Processed %d of %d\r", r.jobsDone.Load(), len(hashes))
		}
		log.Printf("Finished processing %d commits", len(hashes))
	}()

	tmpDir, err := os.MkdirTemp(os.TempDir(), "rcc")
	if err != nil {
		panic(err)
	}
	defer func() {
		// log.Printf("removing %s", tmpDir)
		if err := os.RemoveAll(tmpDir); err != nil {
			log.Printf("failed to remove %s", tmpDir)
		}
		// log.Printf("removed  %s", tmpDir)
	}()

	for _, h := range hashes {
		clonePath := filepath.Join(tmpDir, h)
		r.jobInputQueue <- job{hash: h, runColoc: true, repo: r.repo, clonePath: clonePath}
	}

	close(r.jobInputQueue)
	wg.Wait()
	close(r.resultQueue)
	<-done

	sd.sort()
	// for _, s := range sd.entries {
	// 	log.Printf("RES %+v", s)
	// }
	for _, l := range sd.languages() {
		log.Printf("LANG %+s", l)
	}
	r.statData = &sd
}

func worker(id int, jobs <-chan job, results chan<- statDataEntry) {
	for job := range jobs {
		// log.Printf("worker %02d, hash %s, wd %s", id, job.hash, job.clonePath)
		// clone to clonePath with given hash
		commitTime, err := job.repo.LocalClone(job.clonePath, job.hash)
		if err != nil {
			log.Printf("failed to clone: %s", err)
			continue
		}

		var loc *gocloc.Result
		if job.runColoc {
			languages := gocloc.NewDefinedLanguages() // HERE ?!?!?
			clocOpts := gocloc.NewClocOptions()       // HERE ?!?!
			loc, err = getLoc(languages, clocOpts, []string{job.clonePath})
			if err != nil {
				log.Printf("gocoloc failed: %s", err)
				continue
			}
			// log.Printf("LOC: %v", loc.Languages["Go"])
		}

		if err := os.RemoveAll(job.clonePath); err != nil {
			log.Printf("failed to remove clone %s", job.clonePath)
		}

		results <- statDataEntry{ /* add loc, add cov*/
			date:     commitTime,
			coverage: 10.0,
			loc:      loc,
			sha:      job.hash,
		}
	}
}
