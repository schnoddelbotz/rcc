package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutoDetectSelf(t *testing.T) {
	language := autoDetect(".")
	require.NotNil(t, language)
	assert.Equal(t, "Go", language.GoclocName)
}

func TestGetLanguageInvalid(t *testing.T) {
	lang := GetLanguage("no-such-lang", false, "/nix")
	assert.Nil(t, lang)
}

func TestGetLanguageSpecific(t *testing.T) {
	lang := GetLanguage("Go", false, "/nix")
	assert.Equal(t, "Go", lang.GoclocName)
}

func TestGetLanguageAuto(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := os.Create(filepath.Join(tmpDir, "go.mod"))
	require.Nil(t, err)

	lang := GetLanguage("", false, tmpDir)

	assert.Equal(t, "Go", lang.GoclocName)
}

func TestGetLanguageNone(t *testing.T) {
	lang := GetLanguage("", false, t.TempDir())

	assert.Equal(t, "generic", lang.Description)
}
