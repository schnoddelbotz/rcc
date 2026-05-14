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
	StatData      *statData
	jobsDone      atomic.Uint32
	tmpDir        string
	wg            sync.WaitGroup
	done          chan struct{}
	sd            statData
}

type job struct {
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
		done:          make(chan struct{}),
	}
}

func (runner *Runner) Run(workers int, hashes []string) {
	var err error
	// create tmpDir root, shared by all workers
	runner.tmpDir, err = os.MkdirTemp(os.TempDir(), "rcc")
	if err != nil {
		panic(err)
	}
	defer func() {
		// log.Printf("removing %s", tmpDir)
		if err := os.RemoveAll(runner.tmpDir); err != nil {
			log.Printf("failed to remove %s", runner.tmpDir)
		}
		// log.Printf("removed  %s", tmpDir)
	}()

	// start workers consuming jobInputQueue
	runner.startBackgroundWorkers(workers, hashes)
	// feed jobs to jobInputQueue
	for _, h := range hashes {
		runner.jobInputQueue <- job{hash: h, runColoc: true}
	}

	// wait for all workers to complete
	close(runner.jobInputQueue)
	runner.wg.Wait()
	close(runner.resultQueue)
	<-runner.done

	runner.sd.sort()
	runner.StatData = &runner.sd
}

func (runner *Runner) startBackgroundWorkers(workers int, hashes []string) {
	for range workers {
		runner.wg.Go(func() {
			runner.startJobProcessor()
		})
	}

	go func() {
		for res := range runner.resultQueue {
			runner.sd.entries = append(runner.sd.entries, res)
			runner.jobsDone.Add(1)
		}
		close(runner.done)
	}()

	go func() {
		for runner.jobsDone.Load() != uint32(len(hashes)) {
			time.Sleep(200 * time.Millisecond)
			fmt.Printf("Processed %d of %d\r", runner.jobsDone.Load(), len(hashes))
		}
		log.Printf("Finished processing %d commits", len(hashes))
	}()
}

func (runner *Runner) startJobProcessor() {
	for job := range runner.jobInputQueue {
		// clone to clonePath with given hash
		clonePath := filepath.Join(runner.tmpDir, job.hash)
		commitTime, err := runner.repo.LocalClone(clonePath, job.hash)
		if err != nil {
			log.Printf("failed to clone: %s", err)
			continue
		}

		var loc *gocloc.Result
		if job.runColoc {
			languages := gocloc.NewDefinedLanguages() // HERE ?!?!?
			clocOpts := gocloc.NewClocOptions()       // HERE ?!?!
			loc, err = getLoc(languages, clocOpts, []string{clonePath})
			if err != nil {
				log.Printf("gocoloc failed: %s", err)
				continue
			}
			// log.Printf("LOC: %v", loc.Languages["Go"])
		}

		if err := os.RemoveAll(clonePath); err != nil {
			log.Printf("failed to remove clone %s", clonePath)
		}

		runner.resultQueue <- statDataEntry{ /* add loc, add cov*/
			date:     commitTime,
			coverage: 10.0,
			loc:      loc,
			sha:      job.hash,
		}
	}
}
