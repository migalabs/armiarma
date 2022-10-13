package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func ParseTestTime(test *testing.T, tm string) time.Time {
	pt, err := time.Parse(time.RFC3339, tm)
	require.NoError(test, err)
	return pt
}
