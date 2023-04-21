package util

import (
	"fmt"
	"net/url"
)

const (
	WSProtocol         = "tunnel-protocol"
	RejectReasonHeader = "X-WebSocket-Reject-Reason"
)

func ParseURLDst(url *url.URL) (string, error) {
	dst := url.Query().Get("dst")
	if dst == "" {
		return "", fmt.Errorf("there is not dst")
	}

	// TODO: make sure dst is a valid destination (host and port)

	return dst, nil
}
