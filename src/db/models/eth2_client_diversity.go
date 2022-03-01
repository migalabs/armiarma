package models

import (
	"time"
)

type ClientDiversity struct {
	Timestamp  time.Time `json:"snapshot_timestamp"`
	Prysm      int       `json:"prysm"`
	Lighthouse int       `json:"lighthouse"`
	Teku       int       `json:"teku"`
	Nimbus     int       `json:"nimbus"`
	Grandine   int       `json:"grandine"`
	Lodestar   int       `json:"lodestar"`
	Others     int       `json:"others"`
}

func NewClientDiversity() ClientDiversity {
	return ClientDiversity{
		Timestamp: time.Now(),
	}
}
