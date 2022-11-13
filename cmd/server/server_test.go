package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetup(t *testing.T) {
	srv := setup(true)
	assert.NotNil(t, srv)
}
