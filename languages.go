package main

type Language struct {
	GoclocName     string
	Description    string
	TestfilesRegex string
}

func GetLanguage(l string) *Language {
	if l == "" {
		return &Language{Description: "generic"}
	}
	return languagesMap[l]
}

var languagesMap = map[string]*Language{
	"Go": {
		GoclocName:     "Go",
		Description:    "golang",
		TestfilesRegex: ".*_test.go",
	},
}
