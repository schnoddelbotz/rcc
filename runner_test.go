package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var goTestOutput = `Leading output ...
github.com/schnoddelbotz/rcc/statdata.go:26:		sort				0.0%
github.com/schnoddelbotz/rcc/statdata.go:32:		languages			100.0%
total:							(statements)			30.1%
... Trailing output.
`
var pytestOutput = `--------------------------------------------
TOTAL                      166      20   100%
=========================== short test summary info ============================`

func TestExtractGo(t *testing.T) {
	pattern := languagesMap["Go"].CoverageRegex

	val := extractCoverage(goTestOutput, pattern)

	assert.Equal(t, float32(30.1), val)
}

func TestExtractPytest(t *testing.T) {
	pattern := languagesMap["Python"].CoverageRegex

	val := extractCoverage(pytestOutput, pattern)

	assert.Equal(t, float32(100), val)
}
