package main

import (
	"log"
	"os"
	"path/filepath"
)

type Language struct {
	CoverageRegex      string
	Description        string
	GoclocName         string
	IntegrationTestCmd string
	TestfilesRegex     string
	UnitTestCmd        string
	AutoDetectFiles    []string
}

const LanguageGeneric = "generic"

var languagesMap = map[string]*Language{
	"Go": {
		AutoDetectFiles: []string{"go.mod"},
		CoverageRegex:   `total:\s+\(statements\)\s+\d+.\d+%`,
		Description:     "golang",
		GoclocName:      "Go",
		// IntegrationTestCmd: "go test -coverprofile cover.out -tags=integration ./... ", // YET UNUSED
		TestfilesRegex: ".*_test.go",
		UnitTestCmd:    "go test -coverprofile cover.out ./... && go tool cover -func cover.out",
	},
	"Python": {
		AutoDetectFiles: []string{"requirements.txt", "uv.lock", "pyproject.toml"},
		CoverageRegex:   `TOTAL.*? (100(?:\.0+)?\%|[1-9]?\d(?:\.\d+)?\%)$`,
		Description:     "python",
		GoclocName:      "Python",
		// IntegrationTestCmd: "go test -coverprofile cover.out -tags=integration ./... ", // YET UNUSED
		TestfilesRegex: "test_.*\\.py",
		UnitTestCmd:    "pytest --cov",
	},
	"Java": {
		AutoDetectFiles: []string{"gradlew"}, // + mvn pom...?
		CoverageRegex:   `Total.*?([0-9]{1,3})%`,
		Description:     "java",
		GoclocName:      "Java",
		// IntegrationTestCmd: "go test -coverprofile cover.out -tags=integration ./... ", // YET UNUSED
		TestfilesRegex: "test_.*\\.py",
		UnitTestCmd:    "./gradlew test jacocoTestReport",
	},
}

// GetLanguage returns language-specific configuration based on user-provided language name `userLang`.
// If userLang is empty and skipAutoDetect false, it will attempt auto-detection based on files in repo.
func GetLanguage(userLang string, skipAutoDetect bool, repoPath string) *Language {
	if userLang != "" {
		// user-specified language wins, if provided - even if not valid.
		log.Printf("Language: using user-specified '%s'", userLang)
		return languagesMap[userLang]
	}

	autoDetectStatus := "auto-detect disabled"
	if !skipAutoDetect {
		if result := autoDetect(repoPath); result != nil {
			log.Printf("Language: auto-detected '%s'", result.Description)
			return result
		}
		autoDetectStatus = "none auto-detected"
	}

	log.Printf("Language: no --language specified, %s, using '%s'", autoDetectStatus, LanguageGeneric)
	return &Language{Description: LanguageGeneric}
}

// autoDetect walks built-in languagesMap and tests for each language's AutoDetectFiles existence.
// Fix? As map order is random, may return random result if project contains matches from multiple languages.
func autoDetect(repoPath string) *Language {
	for lang, spec := range languagesMap {
		for _, file := range spec.AutoDetectFiles {
			if _, err := os.Stat(filepath.Join(repoPath, file)); err == nil {
				return languagesMap[lang]
			}
		}
	}
	return nil
}
