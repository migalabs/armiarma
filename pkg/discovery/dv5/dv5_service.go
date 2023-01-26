package dv5

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
	"sync"

	"github.com/pkg/errors"

	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/migalabs/armiarma/pkg/discovery"
	"github.com/migalabs/armiarma/pkg/enode"
	eth "github.com/migalabs/armiarma/pkg/networks/ethereum"
	"github.com/migalabs/armiarma/pkg/utils"
	log "github.com/sirupsen/logrus"

	gethlog "github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/discover"
	ethenode "github.com/ethereum/go-ethereum/p2p/enode"

	"github.com/libp2p/go-libp2p-core/peer"
)

var (
	ModuleName           = "DV5"
	NoNewPeerError error = errors.New("no new peer to read")
)

type Discovery5 struct {
	// Service control variables
	ctx context.Context

	Node        *enode.LocalNode
	Dv5Listener *discover.UDPv5
	Iterator    ethenode.Iterator

	// node notifier
	nodeNotC chan *models.HostInfo
	wg       sync.WaitGroup
	doneF    bool

	// Filtering
	FilterDigest string
}

// NewDiscovery
func NewDiscovery5(
	ctx context.Context,
	node *enode.LocalNode,
	privkey *ecdsa.PrivateKey,
	bootnodes []*ethenode.Node,
	fdigest string,
	port int) (*Discovery5, error) {

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
	return &Discovery5{
		ctx:          ctx,
		Node:         node,
		Dv5Listener:  dv5Listener,
		FilterDigest: fdigest,
		nodeNotC:     make(chan *models.HostInfo),
		doneF:        false,
	}, nil
}

// Start
func (d *Discovery5) Start() chan *models.HostInfo {
	// Generate the iterator over the foud peers
	d.Iterator = d.Dv5Listener.RandomNodes()

	d.wg.Add(1)
	go d.nodeIterator()

	return d.nodeNotC
}

func (d *Discovery5) nodeIterator() {
	defer d.wg.Done()

	for {
		if d.doneF || d.ctx.Err() != nil {
			log.Info("shutdown detected, closing routine")
			return
		}

		if d.Iterator.Next() {
			log.Debug("new ENR discovered")
			// fill the given DiscoveredPeer interface with the next found peer
			node := d.Iterator.Node()

			hInfo, err := d.handleENR(node)
			if err != nil {
				log.Error(errors.Wrap(err, "error handling new ENR"))
			}
			log.Debug("notifying of new ENR")
			d.nodeNotC <- hInfo
			log.Debug("notification done!")
		}
	}

}

// Stop
func (d *Discovery5) Stop() {
	d.doneF = true
	d.wg.Wait()

	d.Iterator.Close()
	d.Dv5Listener.Close()
	close(d.nodeNotC)
}

// handleENR parses and identifies all the advertised fields of a newly discovered peer
func (d *Discovery5) handleENR(node *ethenode.Node) (*models.HostInfo, error) {

	// Parse ENR
	enr, err := eth.ParseEnr(node)
	// TODO: Add fork digest filter when needed
	if err != nil {
		return &models.HostInfo{}, errors.Wrap(err, "unable to parse new discovered ENR")
	}

	// Get the public key and the peer.ID of the discovered peer
	pubkey := node.Pubkey()
	libp2pKey, err := utils.ConvertECDSAPubkeyToSecp2561k(pubkey)
	if err != nil {
		return &models.HostInfo{}, errors.Wrap(err, "unable to convert Geth pubkey to Libp2p")
	}
	peerID, err := peer.IDFromPublicKey(libp2pKey)
	if err != nil {
		return &models.HostInfo{}, errors.Wrap(err, fmt.Sprintf("error extracting peer.ID from node %s", node.ID()))
	}

	hInfo := models.NewHostInfo(
		peerID,
		utils.EthereumNetwork,
		models.WithIPAndPorts(
			enr.IP.String(),
			enr.TCP,
		),
	)

	hInfo.AddAtt(eth.EnrHostInfoAttribute, enr)

	return hInfo, nil
}

// ImportBootNodeList reads the Eth2 bootnodes from a given file
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