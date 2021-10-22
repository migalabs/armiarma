package gossipsub

import (
	"strings"
)

var (
	// Eth2 Mainnet topics
	MainnetForkDigest string = "b5303f2a"

	BeaconBlock          string = "/eth2/b5303f2a/beacon_block/ssz_snappy"
	BeaconAggregateProof string = "/eth2/b5303f2a/beacon_aggregate_and_proof/ssz_snappy"
	VoluntaryExit        string = "/eth2/b5303f2a/voluntary_exit/ssz_snappy"
	ProposerSlashing     string = "/eth2/b5303f2a/proposer_slashing/ssz_snappy"
	AttesterSlashing     string = "/eth2/b5303f2a/attester_slashing/ssz_snappy"
)

func GenerateEth2Topics(forkDigest string, topic string, encoding string) string {
	if strings.Contains(forkDigest, "0x") { // given forkDigest, check if it has the proper format for gossip topics
		forkDigest = strings.Replace(forkDigest, "0x", "", 1)
	}
	topicComposedName := "/eth2/" + forkDigest + "/" + topic + "/" + encoding
	return topicComposedName
}
