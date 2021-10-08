package discovery

/**
This file implements the discovery5 service using the go-ethereum library
With this implementation, you can create a Discovery5 Service and inititate the service itself.

*/

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"github.com/btcsuite/btcd/btcec"
	"github.com/migalabs/armiarma/src/base"
	"github.com/migalabs/armiarma/src/enode"
	"github.com/migalabs/armiarma/src/hosts"
	"github.com/migalabs/armiarma/src/info"

	eth_enode "github.com/ethereum/go-ethereum/p2p/enode"

	geth_log "github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

const PKG_NAME string = "DV5"

type Discovery struct {
	b            *base.Base
	Node         *enode.LocalNode
	ListenPort   int
	BootNodeList []*eth_enode.Node
	info_data    *info.InfoData
	Dv5Listener  *discover.UDPv5
}

func NewEmptyDiscovery() *Discovery {
	return &Discovery{}
}

// constructor
func NewDiscovery(ctx context.Context, input_node *enode.LocalNode, info_obj *info.InfoData, input_port int, stdOpts base.LogOpts) *Discovery {

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
		b:          new_base,
		Node:       input_node,
		info_data:  info_obj,
		ListenPort: input_port,
	}
}

// start dv5 service and listening in given port
func (d *Discovery) Start_dv5() {

	// udp address to listen
	udpAddr := &net.UDPAddr{
		IP:   net.ParseIP(d.GetInfoObj().GetIPToString()),
		Port: int(d.GetListenPort()),
	}

	// start listening and create a connection object
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		d.b.Log.Panicf(err.Error())
	}

	// logger for the discovery5 service
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

	d.b.Log.Infof("dv5 starting to listen: ")

	// start the discovery5 service and listen using the given connection
	d.Dv5Listener, err = discover.ListenV5(conn, &d.Node.LocalNode, cfg)
	if err != nil {
		d.b.Log.Panicf(err.Error())
	}
}

func (d *Discovery) FindRandomNodes(h hosts.BasicLibp2pHost) {
	iterator := d.Dv5Listener.RandomNodes()

	for iterator.Next() {
		d.b.Log.Infof("new randon node:  %s\n", iterator.Node().ID().String())
		node := iterator.Node()

		ipScheme := "ip4"
		if len(node.IP()) == net.IPv6len {
			ipScheme = "ip6"
		}
		pubkey := node.Pubkey()

		peerID, _ := peer.IDFromPublicKey(crypto.PubKey((*crypto.Secp256k1PublicKey)((*btcec.PublicKey)(pubkey))))

		multiAddrStr := fmt.Sprintf("/%s/%s/tcp/%d/p2p/%s", ipScheme, node.IP().String(), node.TCP(), peerID)
		multiAddr, err := ma.NewMultiaddr(multiAddrStr)
		if err != nil {
			fmt.Println(err)
			continue
		}

		new_addr_info, err := peer.AddrInfoFromP2pAddr(multiAddr)

		if err != nil {
			fmt.Println(err)
			continue
		}

		h.Host().Connect(h.Ctx(), *new_addr_info)
		// d.b.Log.Infof("%+v\n", h.Host().Network().Peerstore().Peers())

	}
}

// function which will return the boot node array to initialize our discovery5 listener
// Overrides the bootNodeList attribute inside the Discovery struct
func (d *Discovery) ImportBootNodeList(import_json_file string) {

	var bootNodeList []*eth_enode.Node

	bootNodeListString := BootNodeListString{}

	if _, err := os.Stat(import_json_file); os.IsNotExist(err) {
		d.b.Log.Errorf("Bootnodes file does not exist")
	} else {

		file, err := ioutil.ReadFile(import_json_file)
		if err == nil {
			err := json.Unmarshal([]byte(file), &bootNodeListString)
			if err != nil {
				d.b.Log.Errorf("Could not Unmarshal BootNodes file: %s", err)
			}
		} else {
			d.b.Log.Errorf("Could not read BootNodes file: %s", err)
		}
	}
	// parse bootnode strings into enodes
	for _, element := range bootNodeListString.BootNodes {

		bootNodeList = append(bootNodeList, eth_enode.MustParse(element))

	}

	//bootNodeList = append(bootNodeList, eth_enode.MustParse("enr:-Ku4QImhMc1z8yCiNJ1TyUxdcfNucje3BGwEHzodEZUan8PherEo4sF7pPHPSIB1NNuSg5fZy7qFsjmUKs2ea1Whi0EBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpD1pf1CAAAAAP__________gmlkgnY0gmlwhBLf22SJc2VjcDI1NmsxoQOVphkDqal4QzPMksc5wnpuC3gvSC8AfbFOnZY_On34wIN1ZHCCIyg"))
	d.SetBootNodeList(bootNodeList)
	d.b.Log.Debugf("Added %d bootNode/s", len(d.GetBootNodeList()))

}

func dv5LoggerOpts(input_opts base.LogOpts) base.LogOpts {
	input_opts.ModName = PKG_NAME

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
