package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
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
	RunColoc            bool
	RunColocTests       bool
	runCoverUnit        bool
	runCoverIntegration bool
	includeDuration     bool
	IncludeLangs        []string
	TmpPath             string
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
		stat := StatDataEntry{Date: commitTime, sha: sha}

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
				// second run for tests - exclude target language's test files
				clocOpts.IncludeLangs[runner.language.GoclocName] = struct{}{}
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

			stat.Loc = loc
		}

		if runner.jobOptions.runCoverUnit {
			start := time.Now()

			cmd := exec.Command(runner.language.TestExecutable, runner.language.UnitTestArgs...)
			cmd.Dir = clonePath
			// out, err := cmd.Output()
			err := cmd.Run()
			if err != nil {
				log.Println(err)
				break
			}
			// log.Println(string(out))
			// Hmm... Go-specific ... must run 2nd tool to get overall coverage...
			cmd2 := exec.Command(runner.language.TestExecutable, "tool", "cover", "-func", "cover.out")
			cmd2.Dir = clonePath
			out, err := cmd2.Output()
			if err != nil {
				log.Println(err)
				break
			}
			pattern := regexp.MustCompile(`total:\s+\(statements\)\s+(\d+.\d+)%`)
			m := pattern.FindAllStringSubmatch(string(out), -1)
			// log.Printf("M: %+q", m[0][1]) // panic...
			ucov, _ := strconv.ParseFloat(m[0][1], 32) // panic...!
			stat.CoverageUnit = float32(ucov)

			duration := time.Since(start)
			log.Printf("run unit tests for %s [%s %s] took %s => %.2f ", sha, runner.language.TestExecutable, runner.language.UnitTestArgs, duration, ucov)
			// stat.CoverageUnit = 12
			stat.UnitDuration = duration
		}

		if err := os.RemoveAll(clonePath); err != nil {
			log.Printf("failed to remove clone %s", clonePath)
		}

		runner.resultQueue <- stat
	}
}
