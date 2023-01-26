package enode

import (
	"context"
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"

	eth "github.com/migalabs/armiarma/pkg/networks/ethereum"
	"github.com/sirupsen/logrus"
)

var (
	ModuleName = "ENODE"
	Log        = logrus.WithField(
		"module", ModuleName,
	)
)

type LocalNode struct {
	ctx       context.Context
	LocalNode enode.LocalNode
}

// NewLocalNode will create a LocalNode object using the given arguments.
func NewLocalNode(ctx context.Context, privKey *ecdsa.PrivateKey) *LocalNode {
	// db where to store the ENRs
	ethDB, err := enode.OpenDB("")
	if err != nil {
		Log.Panicf("Could not create local DB %s", err)
	}
	Log.Infof("Creating Local Node")

	return &LocalNode{
		ctx:       ctx,
		LocalNode: *enode.NewLocalNode(ethDB, privKey),
	}
}

// SetForkDigest adds any given ForkDigest into the local node's enr
func (l *LocalNode) SetForkDigest(forkDigest string) {
	// TODO: parse to see if it's a valid ForkDigets (len, blabla)
	l.addEntries(eth.NewEth2DataEntry("b5303f2a"))
}

// SetAttNetworks adds any given set of Attnets into the local node's enr
func (l *LocalNode) SetAttNetworks(networks string) {
	// TODO: parse to see if it's a valid ForkDigets (len, blabla)
	l.addEntries(eth.NewAttnetsENREntry("ffffffffffffffff"))
}

// AddEntries modifies the local Ethereum Node's ENR adding a new entry to the Key-Value
func (l *LocalNode) addEntries(entry enr.Entry) {
	l.LocalNode.Set(entry)
}
