package peerstore

import (
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/migalabs/armiarma/pkg/utils"

	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

func NewAddInfo(pid peer.ID, addrss []ma.Multiaddr, network utils.NetworkType) *PersistablePeer {
	persistable := &PersistablePeer{
		ID:      pid,
		Addrs:   make([]ma.Multiaddr, 0),
		Network: network,
	}
	persistable.Addrs = addrss
	return persistable
}

// PersistablePeer is the main unit of PeerInfo that will be stored locally in the place of running the Crawler
// with the main intetion of having a local AddrssBook of info about the Peers previously found/contacted
type PersistablePeer struct {
	ID      peer.ID
	Addrs   []ma.Multiaddr
	Network utils.NetworkType
}

// PersistablePeerJSON is the JSON formateable struct that simplifies the variable types to the basic ones
type PersistablePeerJSON struct {
	ID      string   `json:"id"`
	Addrs   []string `json:"addrs"`
	Network string   `json:"network"`
}

func (p *PersistablePeer) MarshalJSON() ([]byte, error) {

	var maddresses []string

	for _, maddr := range p.Addrs {
		maddresses = append(maddresses, maddr.String())
	}

	return json.Marshal(&PersistablePeerJSON{
		p.ID.String(),
		maddresses,
		string(p.Network),
	})
}

func (p *PersistablePeer) UnmarshalJSON(input []byte) error {
	// marshal into JSON struct
	var jsonStr PersistablePeerJSON

	var err error
	err = json.Unmarshal(input, &jsonStr)
	if err != nil {
		return errors.Wrap(err, "invalid JSON")
	}

	if jsonStr.ID != "" {
		id, err := peer.Decode(jsonStr.ID)
		if err != nil {
			return errors.Wrap(err, "invalid peer.ID")
		}
		p.ID = id
	}

	if len(jsonStr.Addrs) > 0 {
		addresses := make([]ma.Multiaddr, 0)
		for _, value := range jsonStr.Addrs {
			maddr, err := ma.NewMultiaddr(value)
			if err != nil {
				return errors.Wrap(err, "invalid Maddrs")
			}
			addresses = append(addresses, maddr)
		}
		p.Addrs = addresses
	}

	if jsonStr.Network != "" {
		network := utils.NetworkType(jsonStr.Network)
		p.Network = network
	}
	return nil
}
