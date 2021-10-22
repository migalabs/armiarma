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

// NewLocalNode
// * This method will create a LocalNode object using the given arguments
// @param ctx the context, usually inherited from the base
// @param info_obj the InfoData object where to get the configuration data from the user
// @param stdOpts the logging options object
// @return the LocalNode object
func NewLocalNode(ctx context.Context, info_obj *info.InfoData, stdOpts base.LogOpts) *LocalNode {
	localOpts := nodeLoggerOpts(stdOpts)
	new_base, err := base.NewBase(
		base.WithContext(ctx),
		base.WithLogger(localOpts),
	)
	if err != nil {
		log.Panicf("Could not create base object %s", err)
	}
	// db where to store the ENRs
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

// nodeLoggerOpts
// * This method will fill the custom logging options to this struct
// @param input_opts the logging options object
// @return the modified logging options object
func nodeLoggerOpts(input_opts base.LogOpts) base.LogOpts {
	input_opts.ModName = PKG_NAME
	return input_opts
}

// AddEntries
// * This method will add specific Eth2 Key Value entries to the created Node
// TODO: confirm which data to add and structure appropiately
func (l *LocalNode) AddEntries() {
	l.LocalNode.Set(NewAttnetsENREntry("ffffffffffffffff"))
	l.LocalNode.Set(NewEth2DataEntry("b5303f2a"))
}

// getters and setters
