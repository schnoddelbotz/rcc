package rcc

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
	jobInputQueue chan job
	resultQueue   chan StatDataEntry
	repo          *SourceRepository
	StatData      *StatData
	jobsDone      atomic.Uint32
	tmpDir        string
	wg            sync.WaitGroup
	done          chan struct{}
	sd            StatData
	language      string
}

type job struct {
	hash                string
	runColoc            bool
	runColocNoTests     bool
	runCoverUnit        bool
	runCoverIntegration bool
}

func NewRunner(repo *SourceRepository, language string) *Runner {
	return &Runner{
		repo:          repo,
		jobInputQueue: make(chan job),
		resultQueue:   make(chan StatDataEntry),
		done:          make(chan struct{}),
		language:      language,
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
		if err := os.RemoveAll(runner.tmpDir); err != nil {
			log.Printf("failed to remove %s", runner.tmpDir)
		}
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

			// second run for tests - only target language's test files
			clocOpts.IncludeLangs[runner.language] = struct{}{}
			clocOpts.ReMatch = regexp.MustCompile(languageTestfilesReMap[runner.language])
			loc2, err := getLoc(languages, clocOpts, []string{clonePath})
			if err != nil {
				log.Printf("gocoloc for tests failed: %s", err)
				continue
			}
			if testLoc, exists := loc2.Languages[runner.language]; exists {
				// log.Printf("LOC2: %v", testLoc.Code)
				loc.Languages[runner.language+"Tests"] = testLoc
			}

			// log.Printf("LOC: %v", loc.Languages)
		}

		if err := os.RemoveAll(clonePath); err != nil {
			log.Printf("failed to remove clone %s", clonePath)
		}

		runner.resultQueue <- StatDataEntry{ /* add loc, add cov*/
			Date:     commitTime,
			Coverage: 10.0,
			Loc:      loc,
			sha:      job.hash,
		}
	}
}

var languageTestfilesReMap = map[string]string{
	"Go": ".*_test.go",
}
