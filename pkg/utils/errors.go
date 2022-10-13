package utils

import (
	"strings"
)

type ConnError string

const (
	NoneErr                    ConnError = "none"
	ConnectionResetErr         ConnError = "connection reset by peer"
	IOTimeoutErr               ConnError = "i/o timeout"
	SelfAttemptErr             ConnError = "dial to self attempted"
	DialBackoffErr             ConnError = "dial backoff"
	ConnRefusedErr             ConnError = "connection refused"
	ContextDeadlineExceededErr ConnError = "context deadline exceeded"
	NoRouteToHostErr           ConnError = "no route to host"
	UnReachableNetworkErr      ConnError = "unreachable network"
	PeerIDMissmatchErr         ConnError = "peer id mismatch"
	MetadataReqErr             ConnError = "error requesting metadata"
	NoGoodAddressErr           ConnError = "no good addresses"
	UncertainErr               ConnError = "uncertain"
)

// TODO: Refactor for a much beter and cleaner Error handling

// Funtion that formats the error into a Pretty understandable (standard) way.
// Also important to cohesionate the extra-metrics output csv.
func FilterError(err string) string {
	err = strings.ToLower(err)
	errorPretty := "Uncertain"
	// filter the error type
	if strings.Contains(err, "connection reset by peer") {
		// The peer that we tried to connect resets/drops the connection
		errorPretty = "connection reset by peer"
	} else if strings.Contains(err, "i/o timeout") {
		// When trying to connect a peer, the timeout waiting for stablishing the connection was triggered
		errorPretty = "i/o timeout"
	} else if strings.Contains(err, "dial to self attempted") {
		// When the host tries to connect to itself
		errorPretty = "dial to self attempted"
	} else if strings.Contains(err, "dial backoff") {
		errorPretty = "dial backoff"
	} else if strings.Contains(err, "connection refused") {
		// The peer that we tried to connect refuses/drops the connection
		errorPretty = "connection refused"
	} else if strings.Contains(err, "context deadline exceeded") {
		// When the crawler is able to stablish the connection, but we are unable to identify it
		// h.Connect() internally calls/identifies the peer which reports the error
		errorPretty = "context deadline exceeded"
	} else if strings.Contains(err, "no route to host") {
		// Unable to find a host in that IP
		errorPretty = "no route to host"
	} else if strings.Contains(err, "network is unreachable") || strings.Contains(err, "unreachable network") {
		errorPretty = "unreachable network"
	} else if strings.Contains(err, "peer id mismatch") {
		// Dialing a peer that does not longer exist
		// Hoever there is a new one with another peerID
		errorPretty = "peer id mismatch"
	} else if strings.Contains(err, "none") {
		errorPretty = "none"
	} else if strings.Contains(err, "error requesting metadata") {
		errorPretty = "metadata error"
	} else if strings.Contains(err, "no good addresses") {
		errorPretty = "no good addresses"
	} else {
		// Uncertain (not tracked one)
		log.Errorf("uncertain error: %s", err)
	}
	// TODO: Further encountered errors:
	// 		 - stream reset

	return errorPretty
}
