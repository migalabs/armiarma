package utils

import (
	"time"
)

// Connection event model
type ConnectionEvents struct {
	ConnectionType string
	TimeMili       int64
}

func GetTimeMiliseconds() int64 {
	now := time.Now()
	//secs := now.Unix()
	nanos := now.UnixNano()
	millis := nanos / 1000000

	return millis
}
