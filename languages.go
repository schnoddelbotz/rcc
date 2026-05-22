package main

type Language struct {
	GoclocName          string
	Description         string
	TestfilesRegex      string
	TestExecutable      string
	UnitTestArgs        []string
	IntegrationTestArgs []string
}

const LanguageGeneric = "generic"

func GetLanguage(l string) *Language {
	if l == "" {
		return &Language{Description: LanguageGeneric}
	}
	return languagesMap[l]
}

var languagesMap = map[string]*Language{
	"Go": {
		GoclocName:          "Go",
		Description:         "golang",
		TestfilesRegex:      ".*_test.go",
		TestExecutable:      "go",
		UnitTestArgs:        []string{"test", "-coverprofile", "cover.out", "./..."},
		IntegrationTestArgs: []string{"test", "-coverprofile", "cover.out", "-tags=integration", "./..."},
	},
}
