package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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
	debug               bool
}

type TestResult struct {
	Coverage float32
	Duration time.Duration
	Err      error
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
		// log.Printf("removed tmp dir %s", runner.tmpDir)
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
			result := runTestCmd(clonePath, runner.language.UnitTestCmd)
			stat.CoverageUnit = result.Coverage
			stat.UnitDuration = result.Duration
			if result.Err != nil || runner.jobOptions.debug {
				log.Printf("test-unit@%s '%s' took %.2fs => %3.2f | Err: %s",
					sha[0:8], runner.language.UnitTestCmd, result.Duration.Seconds(), result.Coverage, result.Err)
			}
		}

		if runner.jobOptions.runCoverIntegration {
			result := runTestCmd(clonePath, runner.language.IntegrationTestCmd)
			stat.CoverageIntegration = result.Coverage
			stat.IntegrationDuration = result.Duration
			if result.Err != nil || runner.jobOptions.debug {
				log.Printf("test-integration@%s '%s' took %.2fs => %3.2f | Err: %s",
					sha[0:8], runner.language.IntegrationTestCmd, result.Duration.Seconds(), result.Coverage, result.Err)
			}
		}

		if err := os.RemoveAll(clonePath); err != nil {
			log.Printf("failed to remove clone %s", clonePath)
		}

		runner.resultQueue <- stat
	}
}

func runTestCmd(clonePath string, command string) TestResult {
	start := time.Now()

	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = clonePath
	output, err := cmd.Output()
	if err != nil {
		log.Println(err)
		return TestResult{Err: fmt.Errorf("failed running test command '%s'; error: %s", command, err)}
	}
	cleanedOutput := strings.TrimRight(string(output), "\n")
	coverage, err := strconv.ParseFloat(cleanedOutput, 32)
	if err != nil {
		return TestResult{Err: fmt.Errorf("cannot parse coverage float from test cmd output: '%s'; error:%s", cleanedOutput, err)}
	}

	duration := time.Since(start)
	return TestResult{Duration: duration, Coverage: float32(coverage)}
}

func getLoc(languages *gocloc.DefinedLanguages, options *gocloc.ClocOptions, paths []string) (*gocloc.Result, error) {
	// https://github.com/hhatto/gocloc/blob/master/cmd/gocloc/main.go
	processor := gocloc.NewProcessor(languages, options)
	result, err := processor.Analyze(paths)
	if err != nil {
		return nil, fmt.Errorf("fail gocloc analyze. error: %w", err)
	}
	return result, nil
}
