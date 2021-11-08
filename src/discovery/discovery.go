package discovery

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

	"github.com/btcsuite/btcd/btcec"
	"github.com/migalabs/armiarma/src/base"
	"github.com/migalabs/armiarma/src/db"
	"github.com/migalabs/armiarma/src/db/utils"
	"github.com/migalabs/armiarma/src/enode"
	"github.com/migalabs/armiarma/src/gossipsub/blockchaintopics"
	"github.com/migalabs/armiarma/src/info"

	geth_log "github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/discover"
	eth_enode "github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/protolambda/zrnt/eth2/beacon/common"

	"github.com/libp2p/go-libp2p-core/crypto"
	lib_peer "github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

const PKG_NAME string = "DV5"

type Discovery struct {
	*base.Base
	Node         *enode.LocalNode
	PeerStore    *db.PeerStore
	ListenPort   int
	BootNodeList []*eth_enode.Node
	info_data    *info.InfoData
	Dv5Listener  *discover.UDPv5
	// Filtering
	FilterDigest common.ForkDigest
}

func NewEmptyDiscovery() *Discovery {
	return &Discovery{}
}

// NewDiscovery
// * This method will create a Discovery object usign the given data
// @param input_opts the logging options object
// @return the modified logging options object

func NewDiscovery(ctx context.Context, input_node *enode.LocalNode, db *db.PeerStore, info_obj *info.InfoData, input_port int, stdOpts base.LogOpts) *Discovery {
	localLogger := dv5LoggerOpts(stdOpts)
	// instance base
	new_base, err := base.NewBase(
		base.WithContext(ctx),
		base.WithLogger(localLogger),
	)

	if err != nil {
		new_base.Log.Errorf(err.Error())
	}

	// return the Discovery object
	return &Discovery{
		Base:       new_base,
		Node:       input_node,
		PeerStore:  db,
		info_data:  info_obj,
		ListenPort: input_port,
	}
}

// Start_dv5
// * This method will initiate the discovery listener to receive new
// * peers connections. This will allow other peers to discover us.
func (d *Discovery) Start_dv5() {

	// udp address to listen
	udpAddr := &net.UDPAddr{
		//IP:   net.ParseIP(d.GetInfoObj().GetIPToString()),
		IP:   net.IPv4zero,
		Port: int(d.GetListenPort()),
	}

	// start listening and create a connection object
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		d.Log.Panicf(err.Error())
	}

	// Set custom logger for the discovery5 service (Debug)
	gethLogger := geth_log.New()
	gethLogger.SetHandler(geth_log.FuncHandler(func(r *geth_log.Record) error {

		// d.b.Log.Debugf("%+v\n", r)
		return nil
	}))

	d.ImportBootNodeList(d.GetInfoObj().GetBootNodeFile())

	// configuration of the discovery5
	cfg := discover.Config{
		PrivateKey:   (*ecdsa.PrivateKey)(d.GetInfoObj().GetPrivKey()),
		NetRestrict:  nil,
		Bootnodes:    d.GetBootNodeList(),
		Unhandled:    nil, // Not used in dv5
		Log:          gethLogger,
		ValidSchemes: eth_enode.ValidSchemes,
	}

	d.Log.Info("Starting dv5")

	// start the discovery5 service and listen using the given connection
	d.Dv5Listener, err = discover.ListenV5(conn, &d.Node.LocalNode, cfg)
	if err != nil {
		d.Log.Panicf(err.Error())
	}
}

// FindRandomNodes
// * This method will initiate the randomNodes method, which
// * will create an iterator over randomly generated peers.
// * For each peer, we will try to connect to it.
func (d *Discovery) FindRandomNodes() {
	iterator := d.Dv5Listener.RandomNodes()
	for iterator.Next() {
		d.Log.Debugf("new random node: %s\n", iterator.Node().ID().String())
		node := iterator.Node()
		err := d.HandleENR(node)
		if err != nil {
			// fo far printing a simple warp of the function and the received err
			// with the id of the enode
			d.Log.Debugf("unable to handle ENR from node %s, error: %s\n", node.ID().String(), err)
		}
	}
}

// HandleENR
// *
// @param res represents the enode of the newly discovered peer
func (d *Discovery) HandleENR(node *eth_enode.Node) error {
	eth2Dat, ok, err := enode.ParseNodeEth2Data(*node)
	if err != nil {
		return fmt.Errorf("enr parse error: %v", err)
	}
	// check if the peerexists
	if !ok {
		return fmt.Errorf("peer doesn't exist")
	}

	// check if the peer matches the given ForkDigest
	if eth2Dat.ForkDigest.String() != (blockchaintopics.ForkDigestPrefix + d.info_data.GetForkDigest()) {
		return fmt.Errorf("got ENR with other fork digest: %s", eth2Dat.ForkDigest.String())
	}

	// Get the public key and the peer.ID of the discovered peer
	pubkey := node.Pubkey()
	peerID, err := lib_peer.IDFromPublicKey(crypto.PubKey((*crypto.Secp256k1PublicKey)((*btcec.PublicKey)(pubkey))))
	if err != nil {
		return fmt.Errorf("error extracting peer.ID from node %s", node.ID())
	}
	// Gerearte the Multiaddres of the New Peer taht will be Updated or Stored
	peer := db.NewPeer(peerID.String())
	ipScheme := "ip4"
	if len(node.IP()) == net.IPv6len {
		ipScheme = "ip6"
	}

	multiAddrStr := fmt.Sprintf("/%s/%s/tcp/%d/p2p/%s", ipScheme, node.IP().String(), node.TCP(), peerID)
	multiAddr, err := ma.NewMultiaddr(multiAddrStr)
	if err != nil {
		return fmt.Errorf("error composing the maddrs from peer", err)
	}
	/* Unncesary here, peer.AddrInfo is only needed when connecting the peer
	newAddrInfo, err := lib_peer.AddrInfoFromP2pAddr(multiAddr)
	if err != nil {
		return fmt.Errorf(err)
	}
	*/
	// generate array of MAddr to fit the db.Peer struct
	mAddrs := make([]ma.Multiaddr, 0)
	mAddrs = append(mAddrs, multiAddr)

	// Fill db.Peer with given info
	pubBytes, _ := x509.MarshalPKIXPublicKey(pubkey) // get the []bytes of the pubkey
	peer.Pubkey = hex.EncodeToString(pubBytes)
	peer.NodeId = node.ID().String()
	peer.BlockchainNodeENR = (*node).String()
	peer.Ip = node.IP().String()
	peer.MAddrs = mAddrs

	country, city, countrycode, err := utils.GetLocationFromIp(node.IP().String())
	if err != nil {
		d.Log.Debugf("could not get location from ip: %s  error: %s", node.IP(), err)
	} else {
		peer.Country = country
		peer.CountryCode = countrycode
		peer.City = city
	}
	d.PeerStore.StoreOrUpdatePeer(peer)
	return nil
}

// ImportBootNodeList
// * This method will read the bootnodes list in string format and create an
// * enode array with the parsed ENRs of the bootnodes
// @param import_json_file represents the file where to read the bootnodes from.
// this file is configured in the config file
func (d *Discovery) ImportBootNodeList(import_json_file string) {

	// where we will store the result
	var bootNodeList []*eth_enode.Node

	// where we will unmarshal from file
	bootNodeListString := BootNodeListString{}

	// check if file exists
	if _, err := os.Stat(import_json_file); os.IsNotExist(err) {
		d.Log.Errorf("Bootnodes file does not exist")
	} else {
		// exists
		file, err := ioutil.ReadFile(import_json_file)
		if err == nil {
			err := json.Unmarshal([]byte(file), &bootNodeListString)
			if err != nil {
				d.Log.Errorf("Could not Unmarshal BootNodes file: %s", err)
			}
		} else {
			d.Log.Errorf("Could not read BootNodes file: %s", err)
		}
	}

	// parse bootnode strings into enodes
	for _, element := range bootNodeListString.BootNodes {
		bootNodeList = append(bootNodeList, eth_enode.MustParse(element))
	}

	//bootNodeList = append(bootNodeList, eth_enode.MustParse("enr:-Ku4QImhMc1z8yCiNJ1TyUxdcfNucje3BGwEHzodEZUan8PherEo4sF7pPHPSIB1NNuSg5fZy7qFsjmUKs2ea1Whi0EBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpD1pf1CAAAAAP__________gmlkgnY0gmlwhBLf22SJc2VjcDI1NmsxoQOVphkDqal4QzPMksc5wnpuC3gvSC8AfbFOnZY_On34wIN1ZHCCIyg"))
	d.SetBootNodeList(bootNodeList)
	d.Log.Debugf("Added %d bootNode/s", len(d.GetBootNodeList()))

}

// dv5LoggerOpts
// * This method will add logging options for the Discovery object
// @param input_opts: basic logging options
// @return the modified logging options object for the Discovery object
func dv5LoggerOpts(input_opts base.LogOpts) base.LogOpts {
	input_opts.ModName = PKG_NAME
	//input_opts.Level = "error" // HARDCODED

	return input_opts
}

// getters and setters

func (d Discovery) GetListenPort() int {
	return d.ListenPort
}

func (d Discovery) GetInfoObj() *info.InfoData {
	return d.info_data
}

func (d Discovery) GetDv5Listener() *discover.UDPv5 {
	return d.Dv5Listener
}

func (d *Discovery) SetBootNodeList(input_list []*eth_enode.Node) {
	d.BootNodeList = input_list
}

func (d Discovery) GetBootNodeList() []*eth_enode.Node {
	return d.BootNodeList
}
