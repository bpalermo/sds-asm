package snapshot

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateSnapshot(t *testing.T) {
	snap, err := GenerateSnapshot()
	assert.NotNil(t, snap)
	assert.Nil(t, err)
}
