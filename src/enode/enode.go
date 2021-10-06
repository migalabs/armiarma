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

func NewLocalNode(ctx context.Context, info_obj *info.InfoData, stdOpts base.LogOpts) *LocalNode {
	localOpts := nodeLoggerOpts(stdOpts, info_obj)
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

func nodeLoggerOpts(input_opts base.LogOpts, info_data *info.InfoData) base.LogOpts {
	input_opts.ModName = PKG_NAME
	input_opts.Level = info_data.GetLogLevel()

	return input_opts
}

func (l *LocalNode) AddEntries() {
	l.LocalNode.Set(NewAttnetsENREntry("ffffffffffffffff"))
	l.LocalNode.Set(NewEth2DataEntry("b5303f2a"))
}

// getters and setters
