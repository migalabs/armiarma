package models

import (
	"time"
)

type ClientDiversity struct {
	Timestamp  time.Time `json:"snapshot_timestamp"`
	Prysm      int64     `json:"prysm"`
	Lighthouse int64     `json:"lighthouse"`
	Teku       int64     `json:"teku"`
	Nimbus     int64     `json:"nimbus"`
	Grandine   int64     `json:"grandine"`
	Lodestar   int64     `json:"lodestar"`
	Others     int64     `json:"others"`
}

func NewClientDiversity() ClientDiversity {
	return ClientDiversity{
		Timestamp: time.Now(),
	}
}
