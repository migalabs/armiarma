package endpoint

import (
	"fmt"
	"strings"
)

func ReplaceEndpointWithRequest(endpoint string, toBeReplaced string, item string) string {
	replaceable := fmt.Sprintf("{%s}", toBeReplaced)
	return strings.Replace(endpoint, replaceable, item, -1)
}
