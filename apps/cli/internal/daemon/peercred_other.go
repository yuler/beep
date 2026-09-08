//go:build !linux

package daemon

import (
	"net"
)

func verifyPeerCredentials(conn net.Conn, claimedPID int) error {
	return nil
}
