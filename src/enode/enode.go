package enode

import (
	"context"
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/migalabs/armiarma/src/info"
	all_utils "github.com/migalabs/armiarma/src/utils"
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
	info_data *info.Eth2InfoData
}

// NewLocalNode:
// This method will create a LocalNode object using the given arguments.
// @param ctx the context, usually inherited from the base.
// @param info_obj the InfoData object where to get the configuration data from the user.
// @param stdOpts the logging options object.
// @return the LocalNode object.
func NewLocalNode(ctx context.Context, infObj *info.Eth2InfoData) *LocalNode {
	// db where to store the ENRs
	new_db, err := enode.OpenDB("")
	if err != nil {
		Log.Panicf("Could not create local DB %s", err)
	}
	Log.Infof("Creating Local Node")

	return &LocalNode{
		ctx:       ctx,
		LocalNode: *enode.NewLocalNode(new_db, (*ecdsa.PrivateKey)(infObj.PrivateKey)),
		info_data: infObj,
	}
}

// AddEntries:
// This method will add specific Eth2 Key Value entries to the created Node.
// TODO: confirm which data to add and structure appropiately
func (l *LocalNode) AddEntries() {
	l.LocalNode.Set(all_utils.NewAttnetsENREntry("ffffffffffffffff"))
	l.LocalNode.Set(all_utils.NewEth2DataEntry("b5303f2a"))
}

// getters and setters
