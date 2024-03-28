package ethereum

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type EthereumGossipTopic uint8

const (
	// gossip topic types
	BeaconBlockTopic EthereumGossipTopic = iota
	BeaconAggregateAndProofTopic
	BeaconSubnetAttestationTopic
	BeaconVoluntaryExitTopic
	BeaconProposerSlashingTopic
	BeaconAttesterSlashingTopic
	BeaconSyncCommitteeAggregationTopic
	BeaconSubnetSyncCommitteeVoteTopic
	BeaconBLSExectionChangesTopic
	BeaconSubnetBlobsTopic

	// gossip topic bases
	BeaconBlockTopicBase               string = "beacon_block"
	BeaconAggregateAndProofTopicBase   string = "beacon_aggregate_and_proof"
	VoluntaryExitTopicBase             string = "voluntary_exit"
	ProposerSlashingTopicBase          string = "proposer_slashing"
	AttesterSlashingTopicBase          string = "attester_slashing"
	AttestationSubnetsTopicBase        string = "beacon_attestation_{__subnet_id__}"
	SubnetLimit                               = 64
	SyncCommitteeAggregationsTopicBase string = "sync_committee_contribution_and_proof"
	SyncCommitteeSubnetsTopicBase      string = "sync_committee_{__subnet_id__}"
	SyncCommitteeLimit                        = 4
	BLStoExectionChangeTopicBase       string = "bls_to_execution_change"
	BlobsSubnetsTopicBase              string = "blob_sidecar_{__subnet_id__}"

	// encoding-compression
	Encoding string = "ssz_snappy"
)

// EthTopicPretty returns the topic based on its message type in a pretty version of it.
// It would return "beacon_block" out of the given "/eth2/b5303f2a/beacon_block/ssz_snappy" topic
func EthTopicPretty(eth2topic string) string {
	return strings.Split(eth2topic, "/")[3]
}

// GenerateEth2Topic returns the built topic out of the given arguments
func ComposeTopic(forkDigest string, messageTypeName string) string {
	forkDigest = strings.Trim(forkDigest, ForkDigestPrefix)
	// if we reach here, inputs were okay
	return "/" + BlockchainName +
		"/" + forkDigest +
		"/" + messageTypeName +
		"/" + Encoding
}

// ComposeSubnetTopic generates the GossipSub topic for the given ForkDigest, base, and subnet
func ComposeSubnetTopic(base, forkDigest string, subnet int) string {
	if subnet > SubnetLimit || subnet <= 0 {
		return ""
	}

	// trim "0x" if exists
	forkDigest = strings.Trim(forkDigest, "0x")
	name := strings.Replace(base, "{__subnet_id__}", fmt.Sprintf("%d", subnet), -1)
	return "/" + BlockchainName +
		"/" + forkDigest +
		"/" + name +
		"/" + Encoding
}

func GetSubnetFromTopic(base, topic string) (int, error) {
	re := regexp.MustCompile(base + `_([0-9]+)`)
	match := re.FindAllString(topic, -1)
	if len(match) < 1 {
		return -1, ErrorNoSubnet
	}

	re2 := regexp.MustCompile("([0-9]+)")
	match = re2.FindAllString(match[0], -1)
	if len(match) < 1 {
		return -1, ErrorNotParsableSubnet
	}
	subnet, err := strconv.Atoi(match[0])
	if err != nil {
		return -1, errors.Wrap(err, "unable to conver subnet to int")
	}
	return subnet, nil
}
