package main

import (
	"log"
	"os"
	"path/filepath"
)

type Language struct {
	GoclocName          string
	Description         string
	TestfilesRegex      string
	TestExecutable      string
	UnitTestArgs        []string
	IntegrationTestArgs []string

	UnitTestCmd        string
	IntegrationTestCmd string
}

const LanguageGeneric = "generic"

var languagesMap = map[string]*Language{
	"Go": {
		GoclocName:  "Go",
		Description: "golang",

		// TODO: Make overrideable via cli:
		TestfilesRegex: ".*_test.go",
		UnitTestCmd:    "go test -coverprofile cover.out ./... >/dev/null && go tool cover -func cover.out | tail -1 | awk '{ print $3 }' | cut -d% -f1",
		// IntegrationTestCmd: "go test -coverprofile cover.out -tags=integration ./... ",
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

func autoDetect(repoPath string) *Language {
	if _, err := os.Stat(filepath.Join(repoPath, "go.mod")); err == nil {
		return languagesMap["Go"]
	}
	return nil
}
