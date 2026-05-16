package main

import (
	"fmt"

	"github.com/hhatto/gocloc"
)

func getLoc(languages *gocloc.DefinedLanguages, options *gocloc.ClocOptions, paths []string) (*gocloc.Result, error) {
	// https://github.com/hhatto/gocloc/blob/master/cmd/gocloc/main.go
	processor := gocloc.NewProcessor(languages, options)
	result, err := processor.Analyze(paths)
	if err != nil {
		return nil, fmt.Errorf("fail gocloc analyze. error: %w", err)
	}
	return result, nil
}
