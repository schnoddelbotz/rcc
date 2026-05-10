package main

import (
	"time"
)

func getCommits(from, to time.Time, branch string) []string {
	// get list of commits to process - limit using time range.
	// list commits within range from any branch by default, or limit to one (or more?).
	return []string{}
}

func localClone(from, to string) error {
	return nil
}
