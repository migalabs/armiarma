package enode

import (
	"github.com/ethereum/go-ethereum/p2p/enode"
)

type Node struct {
	base
	LocalNode enode.LocalNode
}
