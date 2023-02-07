package ethereum

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/migalabs/armiarma/pkg/utils"
)

var (
	// Eth2 Mainnet topics
	/* Deprecated for dynamic Eth2 topic construction see bellow
	MainnetForkDigest string = "b5303f2a"
	BeaconBlock          string = "/eth2/b5303f2a/beacon_block/ssz_snappy"
	BeaconAggregateProof string = "/eth2/b5303f2a/beacon_aggregate_and_proof/ssz_snappy"
	VoluntaryExit        string = "/eth2/b5303f2a/voluntary_exit/ssz_snappy"
	ProposerSlashing     string = "/eth2/b5303f2a/proposer_slashing/ssz_snappy"
	AttesterSlashing     string = "/eth2/b5303f2a/attester_slashing/ssz_snappy"
	*/

	// new
	ForkDigestPrefix string = "0x"
	ForkDigestSize   int    = 8 // without the ForkDigestPrefix
	BlockchainName   string = "eth2"

	// mainnet
	Phase0Key    string = "Mainnet"
	AltairKey    string = "Altair"
	BellatrixKey string = "Bellatrix"

	AllForkDigest string = "All"

	// Gnosis
	GnosisKey          string = "Gnosis"
	GnosisAltairKey    string = "GnosisAltair"
	GnosisBellatrixKey string = "Genosisbellatrix"

	ForkDigests = map[string]string{
		AllForkDigest:      "all",
		Phase0Key:          "0xb5303f2a",
		AltairKey:          "0xafcaaba0",
		BellatrixKey:       "0x4a26c58b",
		GnosisKey:          "0xf925ddc5",
		GnosisAltairKey:    "0x56fdb5e0",
		GnosisBellatrixKey: "",
	}

	MessageTypes = []string{
		"beacon_block",
		"beacon_aggregate_and_proof",
		"voluntary_exit",
		"proposer_slashing",
		"attester_slashing",
	}

	AttestationTopicBase = "beacon_attestation_{__subnet_id__}"
	SubnetLimit          = 64

	Encoding string = "ssz_snappy"
)

// GenerateEth2Topics returns the built topic out of the given arguments.
// You may check the commented examples above.nstants.
func GenerateEth2Topics(forkDigest string, messageTypeName string) string {
	// check valid messagetype
	if !utils.ExistsInArray(MessageTypes, messageTypeName) {
		return ""
	}
	// check valid fork digest
	if !utils.ExistsInMapValue(ForkDigests, forkDigest) {
		return ""
	}
	// if we reach here, inputs were okay
	return "/" + BlockchainName +
		"/" + forkDigest +
		"/" + messageTypeName +
		"/" + Encoding
}

// ComposeAttnetsTopic generates the GossipSub topic for the given ForkDigest and subnet
func ComposeAttnetsTopic(forkDigest string, subnet int) string {
	if subnet > SubnetLimit || subnet <= 0 {
		return ""
	}

	// trim "0x" if exists
	forkDigest = strings.Trim(forkDigest, "0x")

	name := strings.Replace(AttestationTopicBase, "{__subnet_id__}", fmt.Sprintf("%d", subnet), -1)
	return "/" + BlockchainName +
		"/" + forkDigest +
		"/" + name +
		"/" + Encoding
}

// Eth2TopicPretty:
// This method returns the topic based on it's message type
// in a pretty version of it.
// It would return "beacon_block" out of the given "/eth2/b5303f2a/beacon_block/ssz_snappy" topic
// @param eth2topic:the entire composed eth2 topic with fork digest and compression.
// @return topic pretty.
func Eth2TopicPretty(eth2topic string) string {
	return strings.Split(eth2topic, "/")[3]
}

// ReturnAllTopics:
// This method will iterate over the mesagetype map and return any possible topic for the
// given fork digest.
// @return the array of topics.
func ReturnAllTopics(inputForkDigest string) []string {
	result_array := make([]string, 0)
	for _, messageValue := range MessageTypes {
		result_array = append(result_array, GenerateEth2Topics(inputForkDigest, messageValue))
	}
	return result_array
}

// ReturnTopics:
// Returns topics for the given parameters.
// @param forkDigest: the forkDigest to use in the topic.
// @param messageTypeName: the type of topic.
// @return the list of generated topics with the given parameters (several messageTypes).
func ReturnTopics(forkDigest string, messageTypeName []string) []string {
	result_array := make([]string, 0)

	for _, messageTypeTmp := range messageTypeName {
		result_array = append(result_array, GenerateEth2Topics(forkDigest, messageTypeTmp))
	}
	return result_array
}

// CheckValidForkDigest:
// This method will check if Fork Digest exists in the corresponding map (ForkDigests).
// @return the fork digest of the given network.
// @return a boolean (true for valid, false for not valid).
func CheckValidForkDigest(input_string string) (string, bool) {
	for forkDigestKey, _ := range ForkDigests {
		if strings.ToLower(forkDigestKey) == input_string {
			return ForkDigests[strings.ToLower(forkDigestKey)], true
		} else {
			forkDigestBytes, err := hex.DecodeString(forkDigestKey)
			if err != nil {
				return "", false
			} else if len(forkDigestBytes) != 4 {
				return "", false
			} else {
				return forkDigestKey, true
			}
		}
	}
	return "", false
}
