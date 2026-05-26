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

func TestXxx(t *testing.T) {
	pattern := `total:\s+\(statements\)\s+\d+.\d+%`

	val := extractCoverage(goTestOutput, pattern)

	assert.Equal(t, float32(30.1), val)
}
