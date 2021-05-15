package custom

import (
	"fmt"
	"strings"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/protolambda/rumor/p2p/track"
)

// ---- Ports and Peers in Peerstore ----

// Read the address for each of the peers in the peerstore 
// counting the number of peers (total), peers with ports 13000, 9000 or others
func TotalPeers(h host.Host) int {
	p := h.Peerstore()
	peerList := p.Peers()
	return len(peerList)
}  

// returns number of peers with ports exposed in address
// (0) -> 13000 | (1) -> 9000 | (2) -> Others
func GetPeersWithPorts(h host.Host, ep track.ExtendedPeerstore) (int, int, int){
	x := 0 // port 13000
	y := 0 // port 9000
	z := 0 // other ports
	peerList := h.Peerstore().Peers()
	for _, peerId := range peerList {
		peerData := ep.GetAllData(peerId)
		for _, address := range peerData.Addrs {
			if strings.Contains(address, "/13000/") {
				x += 1
			} else if strings.Contains(address, "/9000/") {
				y += 1
			} else {
				z += 1
			}
		}
	}
	w := x + y + z
	if w != len(peerList) {
		fmt.Println("Peerstore at host and ExtendedPeerstore don't match", w, len(peerList))
	}
	return x, y, z
}

