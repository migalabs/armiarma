package models

import (
	"github.com/migalabs/armiarma/src/utils"
	ma "github.com/multiformats/go-multiaddr"
)

type FilecoinPeer struct {
	PeerId          string
	UserAgent       string
	Ip              string
	Country         string
	CountryCode     string
	City            string
	MAddrs          []ma.Multiaddr
	Protocols       []string
	ProtocolVersion string
}

func NewFilecoinPeer(peerID string) FilecoinPeer {
	fpeer := FilecoinPeer{
		PeerId:    peerID,
		MAddrs:    make([]ma.Multiaddr, 0),
		Protocols: make([]string, 0),
	}
	return fpeer
}

func (p *FilecoinPeer) FetchInfoToFPeer(newPeer FilecoinPeer) {
	// Update PeerID ?
	// Somehow weird to update the peerID, since it is going to be the same one
	p.PeerId = getNonEmpty(p.PeerId, newPeer.PeerId)
	p.UserAgent = getNonEmpty(p.UserAgent, newPeer.UserAgent)

	p.Ip = getNonEmpty(p.Ip, newPeer.Ip)
	if p.City == "" || newPeer.City != "" {
		p.City = newPeer.City
		p.Country = newPeer.Country
	}
	p.MAddrs = getNonEmptyMAddrArray(p.MAddrs, newPeer.MAddrs)
	p.ProtocolVersion = getNonEmpty(p.ProtocolVersion, newPeer.ProtocolVersion)
	if len(newPeer.Protocols) != 0 {
		p.Protocols = newPeer.Protocols
	}
}

// AddAddr:
// This method adds a new multiaddress in string format to the MAddrs array.
// @return Any error. Otherwise nil.
func (p *FilecoinPeer) AddMAddr(input_addr string) error {
	new_ma, err := ma.NewMultiaddr(input_addr) // parse and format

	if err != nil {
		return err
	}
	p.MAddrs = append(p.MAddrs, new_ma)
	return nil
}

// ExtractPublicMAddr:
// This method loops over all multiaddress and extract the first one that has
// a public IP.
// @return the found multiaddress, nil if error.
func (p *FilecoinPeer) ExtractPublicAddr() ma.Multiaddr {

	// loop over all multiaddresses in the array
	for _, temp_addr := range p.MAddrs {
		temp_extracted_ip := utils.ExtractIPFromMAddr(temp_addr)

		// check if IP is public
		if utils.IsIPPublic(temp_extracted_ip) {
			// the IP is public
			return temp_addr
		}
	}
	return nil // ended loop without returning a public address

}
