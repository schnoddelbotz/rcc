package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
)

type SourceRepository struct {
	path string
	repo *git.Repository
}

func NewSourceRepository(path string) (*SourceRepository, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}
	return &SourceRepository{path: path, repo: repo}, nil
}

func (repo *SourceRepository) GetCommits(from, to time.Time, branch string) ([]string, error) {
	// based on: https://github.com/go-git/go-git/blob/main/_examples/log/main.go
	// get list of commits to process - limit using time range.
	// list commits within range from any branch by default, or limit to one (or more?).
	hashes := []string{}
	ref, err := repo.repo.Head()
	if err != nil {
		return hashes, err
	}
	// ... retrieves the commit history
	cIter, err := repo.repo.Log(&git.LogOptions{From: ref.Hash(), Since: &from, Until: &to})
	if err != nil {
		return hashes, err
	}
	// ... just iterates over the commits, printing it
	err = cIter.ForEach(func(c *object.Commit) error {
		hashes = append(hashes, c.Hash.String())
		return nil
	})
	if err != nil {
		return hashes, err
	}
	return hashes, nil
}

func (repo *SourceRepository) LocalClone(to, sha string) (time.Time, error) {
	var commitTime time.Time

	if err := os.CopyFS(filepath.Join(to, ".git"), os.DirFS(filepath.Join(repo.path, ".git"))); err != nil {
		return commitTime, fmt.Errorf("copy failed: %w", err)
	}
	r, err := git.PlainOpen(to)

	if err != nil {
		return commitTime, fmt.Errorf("git open %s failed: %w", to, err)
	}
	defer func() { _ = r.Close() }()

	w, err := r.Worktree()
	if err != nil {
		return commitTime, err
	}
	err = w.Checkout(&git.CheckoutOptions{
		Hash:  plumbing.NewHash(sha),
		Force: true, // override unstaged changes - we copied .git - with empty worktree
	})
	if err != nil {
		return commitTime, err
	}

	head, err := r.Head()
	if err != nil {
		return commitTime, err
	}
	logi, err := r.Log(&git.LogOptions{From: head.Hash()}) // To: head.Hash()
	if err != nil {
		return commitTime, err
	}
	loge, err := logi.Next()
	if err != nil {
		return commitTime, err
	}

	return loge.Committer.When, nil
}
