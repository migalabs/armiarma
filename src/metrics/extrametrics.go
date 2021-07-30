package metrics

import (
	"fmt"
	"strings"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
)

// AddNewAttempts adds the resuts of a new attempt over an existing peer
// increasing the attempt counter and the respective fields
func (gm *GossipMetrics) AddNewConnectionAttempt(id peer.ID, succeed bool, err string) error {
	v, ok := gm.GossipMetrics.Load(id)
	if !ok { // the peer was already in the sync.Map return true
		return fmt.Errorf("Not peer found with that ID %s", id.String())
	}
	// Update the counter and connection status
	p := v.(PeerMetrics)

	if !p.Attempted {
		p.Attempted = true
		//fmt.Println("Original ", err)
		// MIGHT be nice to try if we can change the uncertain errors for the dial backoff
		if err != "" || err != "dial backoff" {
			p.Error = FilterError(err)
		}
	}
	if succeed {
		p.Succeed = succeed
		p.Error = "None"
	}
	p.Attempts += 1

	// Store the new struct in the sync.Map
	gm.GossipMetrics.Store(id, p)
	return nil
}

// AddNewAttempts adds the resuts of a new attempt over an existing peer
// increasing the attempt counter and the respective fields
func (gm *GossipMetrics) AddNewConnection(id peer.ID) error {
	v, ok := gm.GossipMetrics.Load(id)
	if !ok { // the peer was already in the sync.Map return true
		return fmt.Errorf("Not peer found with that ID %s", id.String())
	}
	// Update the counter and connection status
	p := v.(PeerMetrics)

	p.Connected = true

	// Store the new struct in the sync.Map
	gm.GossipMetrics.Store(id, p)
	return nil
}

// CheckIdConnected check if the given peer was already connected
// returning true if it was connected before or false if wasn't
func (gm *GossipMetrics) CheckIfConnected(id peer.ID) bool {
	v, ok := gm.GossipMetrics.Load(id)
	if !ok { // the peer was already in the sync.Map we didn't connect the peer -> false
		return false
	}
	// Check if the peer was connected
	p := v.(PeerMetrics)
	if p.Succeed {
		return true
	} else {
		return false
	}
}

// GetConnectionsMetrics returns the analysis over the peers found in the
// ExtraMetrics. Return Values = (0)->succeed | (1)->failed | (2)->notattempted
func (gm *GossipMetrics) GetConnectionMetrics(h host.Host) (int, int, int) {
	totalrecorded := 0
	succeed := 0
	failed := 0
	notattempted := 0
	// Read from the recorded ExtraMetrics the status of each peer connections
	gm.GossipMetrics.Range(func(key interface{}, value interface{}) bool {
		p := value.(PeerMetrics)
		totalrecorded += 1
		// Catalog each of the peers for the experienced status
		if p.Attempted {
			if p.Succeed {
				succeed += 1
			} else {
				failed += 1
			}
		} else {
			notattempted += 1
		}
		return true
	})
	// get the len of the peerstore to complete the number of notattempted peers
	peerList := h.Peerstore().Peers()
	peerstoreLen := len(peerList)
	notattempted = notattempted + (peerstoreLen - totalrecorded)
	// MAYBE -> include here the error reader?
	return succeed, failed, notattempted
}

// GetConnectionsMetrics returns the analysis over the peers found in the ExtraMetrics.
// Return Values = (0)->resetbypeer | (1)->timeout | (2)->dialtoself | (3)->dialbackoff | (4)->uncertain
func (gm *GossipMetrics) GetErrorCounter(h host.Host) (int, int, int, int, int) {
	totalfailed := 0
	dialbackoff := 0
	timeout := 0
	resetbypeer := 0
	dialtoself := 0
	uncertain := 0
	// Read from the recorded ExtraMetrics the status of each peer connections
	gm.GossipMetrics.Range(func(key interface{}, value interface{}) bool {
		p := value.(PeerMetrics)
		// Catalog each of the peers for the experienced status
		if p.Attempted && !p.Succeed { // atempted and failed should have generated an error
			erro := p.Error
			totalfailed += 1
			switch erro {
			case "Connection reset by peer":
				resetbypeer += 1
			case "i/o timeout":
				timeout += 1
			case "dial to self attempted":
				dialtoself += 1
			case "dial backoff":
				dialbackoff += 1
			case "Uncertain":
				uncertain += 1
			default:
				fmt.Println("The recorded error type doesn't match any of the error on the list", erro)
			}
		}
		return true
	})
	return resetbypeer, timeout, dialtoself, dialbackoff, uncertain
}

// funtion that formats the error into a Pretty understandable (standard) way
// also important to cohesionate the extra-metrics output csv
func FilterError(err string) string {
	errorPretty := "Uncertain"
	// filter the error type
	if strings.Contains(err, "connection reset by peer") {
		errorPretty = "Connection reset by peer"
	} else if strings.Contains(err, "i/o timeout") {
		errorPretty = "i/o timeout"
	} else if strings.Contains(err, "dial to self attempted") {
		errorPretty = "dial to self attempted"
	} else if strings.Contains(err, "dial backoff") {
		errorPretty = "dial backoff"
	}

	return errorPretty
}
