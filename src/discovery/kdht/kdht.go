/*
	Copyright © 2021 Miga Labs
*/
package kdht

import (
	"context"

	kdh "github.com/libp2p/go-libp2p-kad-dht"
)

type KDHTService struct {
	ctx    context.Context
	cancel context.CancelFunc

	IpfsDHT *kdh.IpfsDHT
}

func NewKDHTService() {
	//return
}
