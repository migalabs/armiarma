package blockchaintopics

import (
	"encoding/hex"
	"strings"

	"github.com/migalabs/armiarma/src/utils"
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
	AltairKey   string = "Altair"
	ForkDigests        = map[string]string{
		MainnetKey: "b5303f2a",
		AltairKey:  "afcaaba0",
	}
	DefaultForkDigest string = ForkDigests[AltairKey]

	/*BeaconBlockKey          string = "BeaconBlock"
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
	}*/

	MessageTypes = []string{
		"beacon_block",
		"beacon_aggregate_and_proof",
		"voluntary_exit",
		"proposer_slashing",
		"attester_slashing",
	}

	Encoding string = "ssz_snappy"
)

// GenerateEth2Topics
// * This method returns the built topic out of the given arguments
// * You may check the commented examples above
// @param forkDigest: the forDigest key in the map. You may use the Key constants.
// @param topic: the message type we want to use in the topic. You may use the Key constants.
// @param encoding: TODO: to be removed as it is always the same
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

// ReturnAllTopics
// * This method will iterate over the mesagetype map and return any possible topic for the
// * given fork digest
// @return the array of topics
func ReturnAllTopics(inputForkDigest string) []string {
	result_array := make([]string, 0)
	for _, messageValue := range MessageTypes {
		result_array = append(result_array, GenerateEth2Topics(inputForkDigest, messageValue))
	}
	return result_array
}

func ReturnTopics(forkDigest string, messageTypeName []string) []string {
	result_array := make([]string, 0)

	for _, messageTypeTmp := range messageTypeName {
		result_array = append(result_array, GenerateEth2Topics(forkDigest, messageTypeTmp))
	}
	return result_array
}

// CheckValidForkDigest
// * This method will check if Fork Digest exists in the corresponding map (ForkDigests)
// * @return the fork digest of the given network
// * @return a boolean (true for valid, false for not valid)
func CheckValidForkDigest(input_string string) (string, bool) {
	for forkDigestKey, _ := range ForkDigests {
		if strings.ToLower(forkDigestKey) == input_string {
			return ForkDigests[strings.ToLower(forkDigestKey)], true
		} else {
			newForkDigest := strings.TrimPrefix(input_string, "0x")
			forkDigestBytes, err := hex.DecodeString(newForkDigest)
			if err != nil {
				return "", false
			} else if len(forkDigestBytes) != 4 {
				return "", false
			} else {
				return newForkDigest, true
			}
		}
	}
	return "", false
}
