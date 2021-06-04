package utils

import "fmt"

// Information regarding the messages received on the beacon_lock topic
type MessageMetrics struct {
	Cnt              int64
	FirstMessageTime int64
	LastMessageTime  int64
}

func NewMessageMetrics() MessageMetrics {
	mm := MessageMetrics{
		Cnt:              0,
		FirstMessageTime: 0,
		LastMessageTime:  0,
	}
	return mm
}

// Increments the counter of the topic
func (c *MessageMetrics) IncrementCnt() int64 {
	c.Cnt++
	return c.Cnt
}

// Stamps linux_time(millis) on the FirstMessageTime/LastMessageTime from given args: time (int64), flag string("first"/"last")
func (c *MessageMetrics) StampTime(flag string) {
	unixMillis := GetTimeMiliseconds()

	switch flag {
	case "first":
		c.FirstMessageTime = unixMillis
	case "last":
		c.LastMessageTime = unixMillis
	default:
		fmt.Println("Metrics Package -> StampTime.flag wrongly parsed")
	}
}
