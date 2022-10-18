package snapshot

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateSnapshot(t *testing.T) {
	assert.NotNil(t, GenerateSnapshot())
}
