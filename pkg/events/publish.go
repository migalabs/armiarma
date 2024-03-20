package events

import (
	"encoding/json"

	"github.com/r3labs/sse/v2"
)

// publishEthereumAttestation publishes an EthereumAttestation event
func (f *Forwarder) publishEthereumAttestation(event *EthereumAttestation) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	f.server.Publish(string(TopicEthereumAttestation), &sse.Event{
		Data: data,
	})

	return nil
}

// publishTimedEthereumAttestation publishes a TimedEthereumAttestation event
func (f *Forwarder) publishTimedEthereumAttestation(event *TimedEthereumAttestation) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	f.server.Publish(string(TopicTimedEthereumAttestation), &sse.Event{
		Data: data,
	})

	return nil
}
