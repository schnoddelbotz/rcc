package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hhatto/gocloc"
)

type Runner struct {
	jobInputQueue chan string
	resultQueue   chan StatDataEntry
	repo          *SourceRepository
	StatData      *StatData
	jobsDone      atomic.Uint32
	tmpDir        string // within tmpPath, will be removed at exit
	wg            sync.WaitGroup
	done          chan struct{}
	sd            StatData
	language      *Language
	jobOptions    JobOptions
}

type JobOptions struct {
	RunColoc      bool
	RunColocTests bool
	// runCoverUnit        bool
	// runCoverIntegration bool
	IncludeLangs []string
	TmpPath      string
}

func NewRunner(repo *SourceRepository, language *Language, options JobOptions) *Runner {
	return &Runner{
		repo:          repo,
		jobInputQueue: make(chan string),
		resultQueue:   make(chan StatDataEntry),
		done:          make(chan struct{}),
		language:      language,
		jobOptions:    options,
	}
}

func (runner *Runner) Run(workers int, hashes []string) {
	var err error
	// create tmpDir root, shared by all workers
	runner.tmpDir, err = os.MkdirTemp(runner.jobOptions.TmpPath, "rcc")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := os.RemoveAll(runner.tmpDir); err != nil {
			log.Printf("failed to remove %s", runner.tmpDir)
			return
		}
		log.Printf("removed tmp dir %s", runner.tmpDir)
	}()
	log.Printf("Runner started, using clone tmp dir: %s", runner.tmpDir)

	// start workers consuming jobInputQueue
	runner.startBackgroundWorkers(workers, hashes)
	// feed jobs to jobInputQueue
	for _, h := range hashes {
		runner.jobInputQueue <- h
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
			runner.startJobInputQueueWorker()
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

func (runner *Runner) startJobInputQueueWorker() {
	languages := gocloc.NewDefinedLanguages()

	for sha := range runner.jobInputQueue {
		// clone to clonePath with given hash
		clonePath := filepath.Join(runner.tmpDir, sha)
		commitTime, err := runner.repo.LocalClone(clonePath, sha)
		if err != nil {
			log.Printf("failed to clone: %s", err)
			continue
		}

		var loc *gocloc.Result
		if runner.jobOptions.RunColoc {
			clocOpts := gocloc.NewClocOptions() // HERE ?!?!
			for _, x := range runner.jobOptions.IncludeLangs {
				clocOpts.IncludeLangs[x] = struct{}{}
			}
			loc, err = getLoc(languages, clocOpts, []string{clonePath})
			if err != nil {
				log.Printf("gocoloc failed: %s", err)
				continue
			}

			if runner.jobOptions.RunColocTests {
				// second run for tests - only target language's test files
				clocOpts.IncludeLangs[runner.language.GoclocName] = struct{}{}
				// clocOpts.ReMatch = regexp.MustCompile(runner.language.TestfilesRegex)
				clocOpts.ReNotMatch = regexp.MustCompile(runner.language.TestfilesRegex)
				loc2, err := getLoc(languages, clocOpts, []string{clonePath})
				if err != nil {
					log.Printf("gocoloc for tests failed: %s", err)
					continue
				}
				if testLoc, exists := loc2.Languages[runner.language.GoclocName]; exists {
					loc.Languages[runner.language.GoclocName+"ExcludingTests"] = testLoc
				}
			}
		}

		if err := os.RemoveAll(clonePath); err != nil {
			log.Printf("failed to remove clone %s", clonePath)
		}

		runner.resultQueue <- StatDataEntry{ /* add loc, add cov*/
			Date:     commitTime,
			Coverage: 10.0,
			Loc:      loc,
			sha:      sha,
		}
	}
}
