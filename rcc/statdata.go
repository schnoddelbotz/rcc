package rcc

import (
	"slices"
	"time"

	"github.com/hhatto/gocloc"
)

type statData struct {
	entries []statDataEntry
}

// type locData map[string]uint

type statDataEntry struct {
	sha      string
	date     time.Time
	loc      *gocloc.Result
	coverage float32
	duration time.Duration
}

func (sd *statData) sort() {
	slices.SortFunc(sd.entries, func(a, b statDataEntry) int {
		return a.date.Compare(b.date)
	})
}

func (sd *statData) languages() []string {
	langMap := map[string]bool{}
	for _, entry := range sd.entries {
		for lang := range entry.loc.Languages {
			langMap[lang] = true
		}
	}
	result := []string{}
	for lang := range langMap {
		result = append(result, lang)
	}
	slices.Sort(result)
	return result
}

func processCommits(parallel uint8, repoPath, shas []string, runLoc, runCov bool) (*statData, error) {
	s := statData{}
	// wg.wait
	return &s, nil
}

func processCommit(repoPath, sha string, runLoc, runCov bool) (statDataEntry, error) {
	s := statDataEntry{}
	return s, nil
}
