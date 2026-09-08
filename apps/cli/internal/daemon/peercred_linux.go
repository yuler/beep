//go:build linux

package daemon

import (
	"fmt"
	"net"
	"os"
	"syscall"
)

func verifyPeerCredentials(conn net.Conn, claimedPID int) error {
	raw, ok := conn.(syscall.Conn)
	if !ok {
		return nil
	}
	sysConn, err := raw.SyscallConn()
	if err != nil {
		return nil
	}
	var ucred *syscall.Ucred
	var credErr error
	err = sysConn.Control(func(fd uintptr) {
		ucred, credErr = syscall.GetsockoptUcred(int(fd), syscall.SOL_SOCKET, syscall.SO_PEERCRED)
	})
	if err != nil || credErr != nil || ucred == nil {
		return nil
	}
	uid := os.Getuid()
	euid := os.Geteuid()
	if int(ucred.Uid) != uid && int(ucred.Uid) != euid {
		return fmt.Errorf("socket peer UID %d does not match current user (UID %d)", ucred.Uid, uid)
	}
	if claimedPID > 0 && int(ucred.Pid) != claimedPID {
		return fmt.Errorf("socket peer PID %d does not match claimed PID %d", ucred.Pid, claimedPID)
	}
	return nil
}
