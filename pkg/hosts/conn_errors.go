package hosts

import (
	"strings"

	swarm "github.com/libp2p/go-libp2p-swarm"
)

// Much easier/prettier way of filtering the Error returned by the libp2p.Host.Connect
// extracted from `@dennis-tra`'s nebula-crawler repo:
// https://github.com/dennis-tra/nebula-crawler/blob/f2b3ba376d221fed886dad204acfc0dfe8e492ea/pkg/db/errors.go#L28
const (
	// list errors
	NoConnError                                         = "none"
	ErrorRequestingMetadta                              = "error requesting metadata"
	ResourceLimitError                                  = "resource_limit"
	DialErrorRoutingNotFound                            = "routing_not_found"
	DialErrorNoRecentNetworkActivity                    = "no_recent_network_activity"
	DialErrorStreamErrorCode0                           = "canceled_stream_err_code_0"
	DialErrorMsgSenderInvalidated                       = "msg_sender_invalidated"
	DialBlacklistedPeer                                 = "hydra_booster_peer"
	DialErrorIoTimeout                                  = "io_timeout"
	DialErrorSelfAttempt                                = "dial to self attempted"
	DialErrorConnectionRefused                          = "connection_refused"
	DialErrorConnectionResetByPeer                      = "connection_reset_by_peer"
	DialErrorBackOff                                    = "backoff"
	DialErrorProtocolNotSupported                       = "protocol_not_supported"
	DialErrorPeerIDMismatch                             = "peer_id_mismatch"
	DialErrorNoRouteToHost                              = "no_route_to_host"
	DialErrorNetworkUnreachable                         = "network_unreachable"
	DialErrorNoGoodAddresses                            = "no_good_addresses"
	DialErrorNoAddress                                  = "no_addresses"
	DialErrorMaddrReset                                 = "maddr_reset"
	DialErrorContextDeadlineExceeded                    = "context_deadline_exceeded"
	DialErrorNoPublicIP                                 = "no_public_ip"
	DialErrorSecurityProtocolNegotiation                = "security_protocol_negotiation"
	DialErrorMaxDialAttemptsExceeded                    = "max_dial_attempts_exceeded"
	DialErrorUnknown                                    = "unknown"
	DialErrorStreamReset                                = "stream_reset"
	DialErrorHostIsDown                                 = "host_is_down"
	DialErrorTooManyOpenFiles                           = "too_many_open_files"
	DialErrorNegotiateSecurityProtocolNoTrailingNewLine = "negotiate_security_protocol_no_trailing_new_line"
)

var KnownErrors = map[string]string{
	ResourceLimitError:                                  "resource limit exceeded",
	DialErrorRoutingNotFound:                            "routing: not found",
	DialErrorNoRecentNetworkActivity:                    "no recent network activity",
	DialErrorStreamErrorCode0:                           "canceled with error code 0",
	DialErrorMsgSenderInvalidated:                       "message sender has been invalidated",
	DialErrorIoTimeout:                                  "i/o timeout",
	DialErrorSelfAttempt:                                "dial to self attempted",
	DialErrorConnectionRefused:                          "connection refused",
	DialErrorConnectionResetByPeer:                      "connection reset by peer",
	DialErrorBackOff:                                    "backoff",
	DialErrorProtocolNotSupported:                       "protocol not supported",
	DialErrorPeerIDMismatch:                             "peer id mismatch",
	DialErrorNoRouteToHost:                              "no route to host",
	DialErrorNetworkUnreachable:                         "network is unreachable",
	DialErrorNoGoodAddresses:                            "no good addresses",
	DialErrorNoAddress:                                  "no addresses",
	DialErrorContextDeadlineExceeded:                    "context deadline exceeded",
	DialErrorNoPublicIP:                                 "no public IP address",
	DialErrorSecurityProtocolNegotiation:                "failed to negotiate security protocol: EOF",
	DialErrorMaxDialAttemptsExceeded:                    "max dial attempts exceeded",
	DialErrorHostIsDown:                                 "host is down",
	DialErrorStreamReset:                                "stream reset",
	DialErrorTooManyOpenFiles:                           "too many open files",
	DialErrorNegotiateSecurityProtocolNoTrailingNewLine: "failed to negotiate security protocol: message did not have trailing newline",
}

func ParseConError(err error) string {
	// nested error casting to the swarm.DialError to process all the errors that might come
	connErr, ok := err.(*swarm.DialError)
	if !ok && connErr != nil {
		if connErr.Cause != nil {
			return ParseConError(connErr.Cause)
		}
	}

	// check if the connError is one of the ones that we have identified
	for key, errStr := range KnownErrors {
		if strings.Contains(err.Error(), errStr) {
			return key
		}
	}
	return DialErrorUnknown
}
