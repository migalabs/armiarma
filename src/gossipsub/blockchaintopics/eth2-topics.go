package blockchaintopics

import (
	"strings"
)

var (
	// Eth2 Mainnet topics
	/*
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

	MainnetKey  string = "Mainnet"
	ForkDigests        = map[string]string{
		MainnetKey: "b5303f2a",
	}

	BeaconBlockKey          string = "BeaconBlock"
	BeaconAggregateProofKey string = "BeaconAggregateProof"
	VoluntaryExitKey        string = "VoluntaryExit"
	ProposerSlashingKey     string = "ProposerSlashing"
	AttesterSlashingKey     string = "AttesterSlashing"

	MessageTypes = map[string]string{
		BeaconBlockKey:          "beacon_block",
		BeaconAggregateProofKey: "beacon_aggregate_and_proof",
		VoluntaryExitKey:        "voluntary_exit",
		ProposerSlashingKey:     "proposer_slashing",
		AttesterSlashingKey:     "attester_slashing",
	}

	Encoding string = "ssz_snappy"
)

// GenerateEth2Topics
// * This method returns the built topic out of the given arguments
// * You may check the commented examples above
// @param forkDigest: the forDigest key in the map. You may use the Key constants.
// @param topic: the message type we want to use in the topic. You may use the Key constants.
// @param encoding: TODO: to be removed as it is always the same
func GenerateEth2Topics(forkDigestKey string, messageTypeKey string) string {

	// check if inputs exists in corresponding maps
	// if any does not exist, return ""
	forkDigest, ok := ForkDigests[forkDigestKey]

	if !ok {
		return ""
	}

	messageType, ok := MessageTypes[messageTypeKey]

	if !ok {
		return ""
	}

	// if we reach here, inputs were okay
	return "/" + BlockchainName +
		"/" + forkDigest +
		"/" + messageType +
		"/" + Encoding

}

// ReturnAllTopics
// * This method will iterate over the mesagetype map and return any possible topic for the
// * given fork digest
// @return the array of topics
func ReturnAllTopics(inputForkDigestKey string) []string {
	result_array := make([]string, 0)

	for _, messageValue := range MessageTypes {
		result_array = append(result_array, GenerateEth2Topics(inputForkDigestKey, messageValue))
	}
	return result_array
}

func ReturnTopics(forkDigestKey string, messageTypesKeys []string) []string {
	result_array := make([]string, 0)

	for _, messageTypeKeyTmp := range messageTypesKeys {
		result_array = append(result_array, GenerateEth2Topics(forkDigestKey, messageTypeKeyTmp))
	}
	return result_array
}

// CheckValidForkDigest
// * This method will check if Fork Digest exists in the corresponding map (ForkDigests)
// * @return the key string
// * @return a boolean (true for valid, false for not valid)
func CheckValidForkDigest(input_string string) (string, bool) {

	for forkDigestKey, _ := range ForkDigests {
		if strings.ToLower(forkDigestKey) == strings.ToLower(forkDigestKey) {
			return forkDigestKey, true
		}
	}
	return "", false
}
