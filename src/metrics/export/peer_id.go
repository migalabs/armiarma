package export

import (
	"github.com/libp2p/go-libp2p-core/peer"
)

type PeerIdList []peer.ID

// add new item to the list
func (pl *PeerIdList) AddItem(newItem peer.ID) {
	*pl = append(*pl, newItem)
}

// get item form the list by index
func (pl *PeerIdList) GetByIndex(idx int) peer.ID {
	return (*pl)[idx]
}

// get the array sorted by list of indexes
func (pl PeerIdList) GetArrayByIndexes(idxs []int) []peer.ID {
	var sortedArray []peer.ID
	for _, i := range idxs {
		sortedArray = append(sortedArray, pl[i])
	}
	return sortedArray
}

// NOTE: There is no need to sort the peerIds
