package enode

import (
	"context"
	"crypto/ecdsa"
	"log"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/migalabs/armiarma/src/base"
	"github.com/migalabs/armiarma/src/info"
)

const PKG_NAME = "ENODE"

type LocalNode struct {
	base      base.Base
	LocalNode enode.LocalNode
	info_data *info.InfoData
}

// This constructor will return an Enode object using the given parameters
func NewLocalNode(ctx context.Context, info_obj *info.InfoData, stdOpts base.LogOpts) *LocalNode {
	localOpts := nodeLoggerOpts(stdOpts)
	new_base, err := base.NewBase(
		base.WithContext(ctx),
		base.WithLogger(localOpts),
	)
	if err != nil {
		log.Panicf("Could not create base object %s", err)
	}

	new_db, err := enode.OpenDB("")
	if err != nil {
		log.Panicf("Could not create local DB %s", err)
	}
	new_base.Log.Infof("Creating Local Node")

	return &LocalNode{
		base:      *new_base,
		LocalNode: *enode.NewLocalNode(new_db, (*ecdsa.PrivateKey)(info_obj.GetPrivKey())),
		info_data: info_obj,
	}
}

// This function will fill the custom logging options to this struct
func nodeLoggerOpts(input_opts base.LogOpts) base.LogOpts {
	input_opts.ModName = PKG_NAME

	return input_opts
}

// Add needed Eth2 entries to the node
func (l *LocalNode) AddEntries() {
	l.LocalNode.Set(NewAttnetsENREntry("ffffffffffffffff"))
	l.LocalNode.Set(NewEth2DataEntry("b5303f2a"))
}

// getters and setters
