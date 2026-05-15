package rcc

import (
	"slices"
	"time"

	"github.com/hhatto/gocloc"
)

type StatData struct {
	entries []StatDataEntry
}

// type locData map[string]uint

type StatDataEntry struct {
	sha      string
	Date     time.Time
	Loc      *gocloc.Result
	Coverage float32
	Duration time.Duration
}

func (sd *StatData) sort() {
	slices.SortFunc(sd.entries, func(a, b StatDataEntry) int {
		return a.Date.Compare(b.Date)
	})
}

func (sd *StatData) languages() []string {
	langMap := map[string]bool{}
	for _, entry := range sd.entries {
		for lang := range entry.Loc.Languages {
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
