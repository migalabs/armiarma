package types

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

type Genesis struct {
	GenesisTime           time.Time
	GenesisValidatorsRoot Root
	GenesisForkVersion    Version
}

type genesisJSON struct {
	GenesisTime           string `json:"genesis_time"`
	GenesisValidatorsRoot string `json:"genesis_validators_root"`
	GenesisForkVersion    string `json:"genesis_fork_version"`
}

func (c *Genesis) MarshalJSON() ([]byte, error) {
	return json.Marshal(&genesisJSON{
		GenesisTime:           fmt.Sprintf("%d", c.GenesisTime.Unix()),
		GenesisValidatorsRoot: string(hex.EncodeToString(c.GenesisValidatorsRoot[:])),
		GenesisForkVersion:    string(hex.EncodeToString(c.GenesisForkVersion[:])),
	})
}

func (c *Genesis) UnmarshalJSON(raw []byte) error {
	var err error

	var vJson genesisJSON
	err = json.Unmarshal(raw, &vJson)
	if err != nil {
		return errors.Wrap(err, "invalid JSON")
	}
	// Check if the fields are emtpy or not
	// Genesis Time
	if vJson.GenesisTime == "" {
		return errors.New("missing genesis time")
	}
	genTime, err := strconv.ParseInt(vJson.GenesisTime, 10, 64)
	if err != nil {
		return errors.Wrap(err, "invalid genesis time")
	}
	c.GenesisTime = time.Unix(genTime, 0)
	// Genesis Validators Root
	if vJson.GenesisValidatorsRoot == "" {
		return errors.New("missing genesis validators root")
	}
	var genValRoot Root
	err = genValRoot.UnmarshalText([]byte(vJson.GenesisValidatorsRoot))
	if err != nil {
		return errors.Wrap(err, "invalid value for validators root")
	}
	c.GenesisValidatorsRoot = genValRoot
	// Fork Digest Version
	var forkVersion Version
	err = forkVersion.UnmarshalText([]byte(vJson.GenesisForkVersion))
	if err != nil {
		return errors.Wrap(err, "invalid fork version value")
	}
	c.GenesisForkVersion = forkVersion
	return nil
}
