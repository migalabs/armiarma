package events

type Topic string

// Ethereum events
const (
	// TopicEthereumAttestation is the topic for Ethereum Attestation events
	TopicEthereumAttestation string = "ethereum_attestation"
	// TopicTimedEthereumAttestation is the topic for Timed Ethereum Attestation events
	TopicTimedEthereumAttestation string = "timed_ethereum_attestation"
)
