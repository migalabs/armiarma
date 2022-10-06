package dv5

/**
This file implements the discovery5 service using the go-ethereum library
With this implementation, you can create a Discovery5 Service and inititate the service itself.

*/

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"github.com/pkg/errors"

	"github.com/btcsuite/btcd/btcec"
	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/migalabs/armiarma/pkg/discovery"
	"github.com/migalabs/armiarma/pkg/enode"
	"github.com/migalabs/armiarma/pkg/gossipsub/blockchaintopics"
	"github.com/migalabs/armiarma/pkg/utils"
	"github.com/sirupsen/logrus"

	gethlog "github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/discover"
	ethenode "github.com/ethereum/go-ethereum/p2p/enode"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

var (
	ModuleName = "DV5"
	log        = logrus.WithField(
		"module", ModuleName,
	)
)

type Discovery struct {
	// Service control variables
	ctx context.Context

	Node        *enode.LocalNode
	Dv5Listener *discover.UDPv5
	Iterator    ethenode.Iterator
	// Filtering
	FilterDigest string
}

// NewDiscovery
func NewDiscovery(ctx context.Context, node *enode.LocalNode, privkey *ecdsa.PrivateKey, bootnodes []*ethenode.Node, fdigest string, port int) (*Discovery, error) {

	if len(bootnodes) == 0 {
		log.Panic("unable to start dv5 peer discovery, no bootnodes provided")
	}

	// udp address to listen
	udpAddr := &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: port,
	}

	// start listening and create a connection object
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Panic(err.Error())
	}

	// Set custom logger for the discovery5 service (Debug)
	gethLogger := gethlog.New()
	gethLogger.SetHandler(gethlog.FuncHandler(func(r *gethlog.Record) error {
		return nil
	}))

	// configuration of the discovery5
	cfg := discover.Config{
		PrivateKey:   privkey,
		NetRestrict:  nil,
		Bootnodes:    bootnodes,
		Unhandled:    nil, // Not used in dv5
		Log:          gethLogger,
		ValidSchemes: ethenode.ValidSchemes,
	}

	// start the discovery5 service and listen using the given connection
	dv5Listener, err := discover.ListenV5(conn, &node.LocalNode, cfg)
	if err != nil {
		log.Panic(err.Error())
	}

	// return the Discovery object
	return &Discovery{
		ctx:          ctx,
		Node:         node,
		Dv5Listener:  dv5Listener,
		FilterDigest: fdigest,
	}, nil
}

// Start
func (d *Discovery) Start() {
	// Generate the iterator over the foud peers
	d.Iterator = d.Dv5Listener.RandomNodes()
}

// Next
func (d *Discovery) Next() bool {
	return d.Iterator.Next()
}

// Peer
func (d *Discovery) Peer() (models.Peer, bool) {
	// check if there is a new peer to read
	if !d.Iterator.Next() {
		return models.Peer{}, false
	}
	// fill the given DiscoveredPeer interface with the next found peer
	node := d.Iterator.Node()
	if node == nil {
		return models.Peer{}, false
	}
	bp, err := d.handleENR(node)
	if err != nil {
		// fo far printing a simple wrap of the function and the received err
		// with the id of the enode
		log.Debugf("unable to handle ENR from node %s, error: %s\n", node.ID().String(), err)
		return models.Peer{}, false
	}

	return bp, true
}

// HandleENR
// Retrieve information from the ENR
// @param node: represents the enode of the newly discovered peer
func (d *Discovery) handleENR(node *ethenode.Node) (models.Peer, error) {
	bp := models.NewPeer("")
	eth2Dat, ok, err := utils.ParseNodeEth2Data(*node)
	if err != nil {
		return models.Peer{}, fmt.Errorf("enr parse error: %v", err)
	}
	// check if the peer exists
	if !ok {
		return models.Peer{}, fmt.Errorf("peer doesn't exist")
	}

	// check if the peer matches the given ForkDigest
	if eth2Dat.ForkDigest.String() != (blockchaintopics.ForkDigestPrefix + d.FilterDigest) {
		return models.Peer{}, fmt.Errorf("got ENR with other fork digest: %s", eth2Dat.ForkDigest.String())
	}

	// Get the public key and the peer.ID of the discovered peer
	pubkey := node.Pubkey()
	peerid, err := peer.IDFromPublicKey(crypto.PubKey((*crypto.Secp256k1PublicKey)((*btcec.PublicKey)(pubkey))))
	if err != nil {
		return models.Peer{}, fmt.Errorf("error extracting peer.ID from node %s", node.ID())
	}
	// Gerearte the Multiaddres of the New Peer taht will be Updated or Stored
	ipScheme := "ip4"
	if len(node.IP()) == net.IPv6len {
		ipScheme = "ip6"
	}

	multiAddrStr := fmt.Sprintf("/%s/%s/tcp/%d/p2p/%s", ipScheme, node.IP().String(), node.TCP(), peerid)
	multiAddr, err := ma.NewMultiaddr(multiAddrStr)
	if err != nil {
		return models.Peer{}, fmt.Errorf("error composing the maddrs from peer %s", err)
	}

	// generate array of MAddr to fit the db.Peer struct
	mAddrs := make([]ma.Multiaddr, 0)
	mAddrs = append(mAddrs, multiAddr)

	// Fill models.Peer with given info
	pubBytes, _ := x509.MarshalPKIXPublicKey(pubkey) // get the []bytes of the pubkey
	bp.SetAtt("pubkey", hex.EncodeToString(pubBytes))
	bp.SetAtt("nodeid", node.ID().String())
	bp.SetAtt("enr", (*node).String())
	bp.MAddrs = mAddrs
	bp.PeerId = peerid.String()

	return bp, nil
}

// ImportBootNodeList
// This method will read the bootnodes list in string format and create an
// enode array with the parsed ENRs of the bootnodes.
// @param import_json_file represents the file where to read the bootnodes from.
// This file is configured in the config file.
func ReadEth2BootnodeFile(jfile string) ([]*ethenode.Node, error) {

	// where we will store the result
	bootNodeList := make([]*ethenode.Node, 0)

	// where we will unmarshal from file
	bootNodeListString := discovery.BootNodeListString{}

	// check if file exists
	if _, err := os.Stat(jfile); os.IsNotExist(err) {
		return bootNodeList, errors.New("Bootnodes file does not exist")
	} else {
		// exists
		file, err := ioutil.ReadFile(jfile)
		if err == nil {
			err := json.Unmarshal([]byte(file), &bootNodeListString)
			if err != nil {
				return bootNodeList, errors.Wrap(err, "Could not Unmarshal BootNodes file: "+jfile)
			}
		} else {
			return bootNodeList, errors.Wrap(err, "Could not read BootNodes file: %s"+jfile)
		}
	}

	// parse bootnode strings into enodes
	for _, element := range bootNodeListString.BootNodes {
		bootNodeList = append(bootNodeList, ethenode.MustParse(element))
	}
	return bootNodeList, nil

}
