package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	err := run("testdata/orca.json")
	assert.NoError(t, err)
}

func TestRun2(t *testing.T) {
	err := run("testdata/lgov.json")
	assert.NoError(t, err)
}

func TestRunTag(t *testing.T) {
	err := run("testdata/lgov_tag.json")
	assert.NoError(t, err)
}
