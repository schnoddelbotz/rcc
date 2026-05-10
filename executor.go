package main

import "time"

type statDataEntry struct {
	sha      string
	date     time.Time
	loc      map[string]uint
	coverage float32
	duration time.Duration
}

func processCommits(parallel uint8, repoPath, shas []string, runLoc, runCov bool) ([]statDataEntry, error) {
	s := []statDataEntry{}
	// wg.wait
	return s, nil
}

func processCommit(repoPath, sha string, runLoc, runCov bool) (statDataEntry, error) {
	s := statDataEntry{}
	return s, nil
}
