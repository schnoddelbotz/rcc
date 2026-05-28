package main

import "log"

// JobOptions ... is partially redundant with CliArgs. fix.
type JobOptions struct {
	RunColoc            bool
	RunColocTests       bool
	runCoverUnit        bool
	runCoverIntegration bool
	includeDuration     bool
	IncludeLangs        []string
	TmpPath             string
	debug               bool
	titleParts          string
}

func getJobOptions(opts CliArgs, langdata *Language) JobOptions {
	jobOpts := JobOptions{
		RunColoc:      true,
		RunColocTests: true,
		IncludeLangs:  opts.includeLanguages,
		TmpPath:       opts.tmpPath,
		debug:         opts.printDebug,
	}

	titleParts := ""
	if opts.coverUnitCmd != "" {
		langdata.UnitTestCmd = opts.coverUnitCmd
	}
	if opts.coverIntegrationCmd != "" {
		langdata.IntegrationTestCmd = opts.coverIntegrationCmd
	}
	if opts.customCoverageRegex != "" {
		langdata.CoverageRegex = opts.customCoverageRegex
	}
	if langdata.UnitTestCmd != "" && !opts.noCoverU {
		log.Printf("Coverage (unit tests): enabled, command: %s", langdata.UnitTestCmd)
		jobOpts.runCoverUnit = true
		titleParts = " + Coverage (Unit-Tests)"
	}
	if langdata.IntegrationTestCmd != "" && opts.CoverI {
		log.Printf("Coverage (integration tests): enabled, command: %s", langdata.IntegrationTestCmd)
		jobOpts.runCoverIntegration = true
		titleParts = " + Coverage (Integration-Tests)"
	}
	if (jobOpts.runCoverUnit || jobOpts.runCoverIntegration) && !opts.noCoverD {
		log.Println("Graphing of coverage test duration: enabled")
		jobOpts.includeDuration = true
		titleParts += " + Test Duration"
	}
	if jobOpts.runCoverUnit || jobOpts.runCoverIntegration {
		log.Printf("Coverage output regex: %s", langdata.CoverageRegex)
	}
	log.Printf("LoC test exclusion regex: %s", langdata.TestfilesRegex)

	jobOpts.titleParts = titleParts
	return jobOpts
}
