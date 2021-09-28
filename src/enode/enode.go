package enode

import (
	"context"
	"crypto/ecdsa"
	"log"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/migalabs/armiarma/src/base"
	"github.com/migalabs/armiarma/src/info"
)

type LocalNode struct {
	base      base.Base
	LocalNode enode.LocalNode
	info_data *info.InfoData
}

func NewLocalNode(ctx context.Context, info_obj *info.InfoData, opts base.LogOpts) *LocalNode {

	new_base, err := base.NewBase(
		base.WithContext(ctx),
		base.WithLogger(base.LogOpts{
			ModName:   opts.ModName,
			Output:    opts.Output,
			Formatter: opts.Formatter,
			Level:     opts.Level,
		}),
	)
	if err != nil {
		log.Panicf("Could not create base object %s", err)
	}

	new_db, err := enode.OpenDB("")
	if err != nil {
		log.Panicf("Could not create local DB %s", err)
	}
	new_base.Log.Debugf("Creating Local Node")
	return &LocalNode{
		base:      *new_base,
		LocalNode: *enode.NewLocalNode(new_db, (*ecdsa.PrivateKey)(info_obj.GetPrivKey())),
		info_data: info_obj,
	}
}

// getters and setters
