package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCliArgs(t *testing.T) {
	os.Args = []string{"foo", "bar", "-v"}

	args := getCliArgs() // ugly, improve: may only be called once due to global pflags :/

	assert.Equal(t, []string{"bar"}, args.argv)
	assert.True(t, args.printVersionOnly)
}

func TestGetVersionOnly(t *testing.T) {
	err := runRCC(CliArgs{printVersionOnly: true})
	assert.Nil(t, err)
}
