package ethereum

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

var (
	ForkDigestPrefix string = "0x"
	ForkDigestSize   int    = 8 // without the ForkDigestPrefix
	BlockchainName   string = "eth2"

	// default fork_digests
	DefaultForkDigest string = ForkDigests[CapellaKey]
	AllForkDigest     string = "All"

	// Mainnet
	Phase0Key    string = "Mainnet"
	AltairKey    string = "Altair"
	BellatrixKey string = "Bellatrix"
	CapellaKey   string = "Capella"
	// Gnosis
	GnosisPhase0Key    string = "GnosisPhase0"
	GnosisAltairKey    string = "GnosisAltair"
	GnosisBellatrixKey string = "Gnosisbellatrix"
	// Goerli / Prater
	PraterPhase0Key    string = "PraterPhase0"
	PraterBellatrixKey string = "PraterBellatrix"
	PraterCapellaKey   string = "PraterCapella"
	// Sepolia
	SepoliaCapellaKey string = "SepoliaCapella"
	// Holesky
	HoleskyCapellaKey string = "HoleskyCapella"
	// Deneb
	DenebCancunKey string = "DenebCancun"

	ForkDigests = map[string]string{
		AllForkDigest: "all",
		// Mainnet
		Phase0Key:    "0xb5303f2a",
		AltairKey:    "0xafcaaba0",
		BellatrixKey: "0x4a26c58b",
		CapellaKey:   "0xbba4da96",
		// Gnosis
		GnosisPhase0Key:    "0xf925ddc5",
		GnosisBellatrixKey: "0x56fdb5e0",
		// Goerli-Prater
		PraterPhase0Key:    "0x79df0428",
		PraterBellatrixKey: "0xc2ce3aa8",
		PraterCapellaKey:   "0x628941ef",
		// Sepolia
		SepoliaCapellaKey: "0x47eb72b3",
		// Holesky
		HoleskyCapellaKey: "0x17e2dad3",
		// Deneb
		DenebCancunKey: "0xee7b3a32",
	}

	MessageTypes = []string{
		BeaconBlockTopicBase,
		BeaconAggregateAndProofTopicBase,
		VoluntaryExitTopicBase,
		ProposerSlashingTopicBase,
		AttesterSlashingTopicBase,
	}

	BeaconBlockTopicBase             string = "beacon_block"
	BeaconAggregateAndProofTopicBase string = "beacon_aggregate_and_proof"
	VoluntaryExitTopicBase           string = "voluntary_exit"
	ProposerSlashingTopicBase        string = "proposer_slashing"
	AttesterSlashingTopicBase        string = "attester_slashing"
	AttestationTopicBase             string = "beacon_attestation_{__subnet_id__}"
	SubnetLimit                             = 64

	Encoding string = "ssz_snappy"
)

var (
	MainnetGenesis time.Time     = time.Unix(1606824023, 0)
	GoerliGenesis  time.Time     = time.Unix(1616508000, 0)
	GnosisGenesis  time.Time     = time.Unix(1638968400, 0) // Dec 08, 2021, 13:00 UTC
	SecondsPerSlot time.Duration = 12 * time.Second
)

// GenerateEth2Topic returns the built topic out of the given arguments.
// You may check the commented examples above.nstants.
func ComposeTopic(forkDigest string, messageTypeName string) string {
	forkDigest = strings.Trim(forkDigest, ForkDigestPrefix)
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
		result_array = append(result_array, ComposeTopic(inputForkDigest, messageValue))
	}
	return result_array
}

// ReturnTopics:
// Returns topics for the given parameters.
// @param forkDigest: the forkDigest to use in the topic.
// @param messageTypeName: the type of topic.
// @return the list of generated topics with the given parameters (several messageTypes).
func ComposeTopics(forkDigest string, messageTypeName []string) []string {
	result_array := make([]string, 0)

	for _, messageTypeTmp := range messageTypeName {
		result_array = append(result_array, ComposeTopic(forkDigest, messageTypeTmp))
	}
	return result_array
}

// CheckValidForkDigest:
// This method will check if Fork Digest exists in the corresponding map (ForkDigests).
// @return the fork digest of the given network.
// @return a boolean (true for valid, false for not valid).
func CheckValidForkDigest(inStr string) (string, bool) {
	for forkDigestKey, forkDigest := range ForkDigests {
		if strings.ToLower(forkDigestKey) == inStr {
			return ForkDigests[strings.ToLower(forkDigestKey)], true
		}
		if forkDigest == inStr {
			return forkDigest, true
		}
	}
	forkDigestBytes, err := hex.DecodeString(inStr)
	if err != nil {
		return "", false
	}
	if len(forkDigestBytes) != 4 {
		return "", false
	}
	return inStr, true
}
