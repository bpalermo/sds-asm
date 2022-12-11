package helper

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

const (
	testEnvVarKey = "AWS_ENDPOINT"
)

func TestGetEnv(t *testing.T) {
	expected := "test"
	_ = os.Setenv(testEnvVarKey, expected)
	defer func() {
		_ = os.Unsetenv(testEnvVarKey)
	}()

	actual := GetEnv(testEnvVarKey, "fallback")
	assert.Equal(t, expected, actual)

	actual = GetEnv("doesnt_matter", "fallback")
	assert.Equal(t, "fallback", actual)
}
