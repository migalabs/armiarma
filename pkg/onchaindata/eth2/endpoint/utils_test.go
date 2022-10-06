package endpoint

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestReplaceEndpointWithRequest(t *testing.T) {
	endpoint := "/eth/v1/beacon/states/{state}/fork"
	tobereplaced := "state"
	itemtoreplace := "head"
	newItem := ReplaceEndpointWithRequest(endpoint, tobereplaced, itemtoreplace)
	require.Equal(t, "/eth/v1/beacon/states/head/fork", newItem)

}
