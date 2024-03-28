package ethereum

import (
	"encoding/hex"
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
	DenebKey     string = "Deneb"
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
		DenebKey:     "0x6a95a1a9",
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
)

var (
	MainnetGenesis time.Time     = time.Unix(1606824023, 0)
	GoerliGenesis  time.Time     = time.Unix(1616508000, 0)
	GnosisGenesis  time.Time     = time.Unix(1638968400, 0) // Dec 08, 2021, 13:00 UTC
	SecondsPerSlot time.Duration = 12 * time.Second
)

// CheckValidForkDigest checks if Fork Digest exists in the corresponding map (ForkDigests).
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

// utils
