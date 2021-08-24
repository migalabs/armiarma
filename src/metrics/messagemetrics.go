package metrics

import (
	"fmt"
	"time"
)

// Information regarding the messages received on the beacon_lock topic
type MessageMetrics struct {
	Count            uint64
	FirstMessageTime time.Time
	LastMessageTime  time.Time
}

func NewMessageMetrics() MessageMetrics {
	mm := MessageMetrics{
		Count:            uint64(0),
		FirstMessageTime: time.Time{},
		LastMessageTime:  time.Time{},
	}
	return mm
}

// Increments the counter of the topic
func (c *MessageMetrics) IncrementCnt() uint64 {
	c.Count++
	return c.Count
}

// Stamps linux_time(millis) on the FirstMessageTime/LastMessageTime from given args: time (int64), flag string("first"/"last")
func (c *MessageMetrics) StampTime(flag string) {
	now := time.Now()
	switch flag {
	case "first":
		c.FirstMessageTime = now
	case "last":
		c.LastMessageTime = now
	default:
		fmt.Println("Metrics Package -> StampTime.flag wrongly parsed")
	}
}
