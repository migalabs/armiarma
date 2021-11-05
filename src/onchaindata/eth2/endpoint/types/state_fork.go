package types

import (
	"encoding/hex"
	"encoding/json"
	"strconv"

	"github.com/pkg/errors"
)

type StateFork struct {
	PreviousVersion Version
	CurrentVersion  Version
	Epoch           Epoch
}

type StateForkJSON struct {
	PreviousVersion string `json:"previous_version"`
	CurrentVersion  string `json:"current_version"`
	Epoch           string `json:"epoch"`
}

func (c *StateFork) MarshalJSON() ([]byte, error) {
	return json.Marshal(StateForkJSON{
		PreviousVersion: string(hex.EncodeToString(c.PreviousVersion[:])),
		CurrentVersion:  string(hex.EncodeToString(c.CurrentVersion[:])),
		Epoch:           c.Epoch.String(),
	})
}

func (c *StateFork) UnmarshalJSON(raw []byte) error {
	var err error

	var vJson StateForkJSON
	err = json.Unmarshal(raw, &vJson)
	if err != nil {
		return errors.Wrap(err, "unable to unmarshal bytes into state fork")
	}
	// Check if the received data is okey
	if vJson.PreviousVersion == "" {
		return errors.New("missing previous version")
	}
	var prevVers Version
	err = prevVers.UnmarshalText([]byte(vJson.PreviousVersion))
	if err != nil {
		return errors.Wrap(err, "invalid previous version value")
	}
	var currentVers Version
	err = currentVers.UnmarshalText([]byte(vJson.CurrentVersion))
	if err != nil {
		return errors.Wrap(err, "invalid current version value")
	}
	e, err := strconv.ParseInt(vJson.Epoch, 10, 64)
	epoch := Epoch(e)
	// Fill data
	c.PreviousVersion = prevVers
	c.CurrentVersion = currentVers
	c.Epoch = epoch
	return nil
}
