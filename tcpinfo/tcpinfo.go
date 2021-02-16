// +build linux

package tcpinfo

import (
	"errors"
	"fmt"
	"net"
	"syscall"
	"unsafe"
)

func Get(conn *net.TCPConn) (*syscall.TCPInfo, error) {
	if conn == nil {
		return nil, errors.New("nil conn")
	}

	// Fetch the underlying raw connection.
	rawConn, err := conn.SyscallConn()
	if err != nil {
		return nil, fmt.Errorf("rawConn err: %v", err)
	}

	tcpInfo := syscall.TCPInfo{}
	tcpInfoSize := unsafe.Sizeof(tcpInfo)
	var errno syscall.Errno

	// Instruct the kernel to deliver the TCP_INFO data into the data structure provided.
	if err := rawConn.Control(func(fd uintptr) {
		_, _, errno = syscall.Syscall6(syscall.SYS_GETSOCKOPT, fd, syscall.SOL_TCP, syscall.TCP_INFO,
			uintptr(unsafe.Pointer(&tcpInfo)), uintptr(unsafe.Pointer(&tcpInfoSize)), 0)
	}); err != nil {
		return nil, fmt.Errorf("rawConn control err: %v", err)
	}

	// Perhaps the syscall failed, if it did then wrap it so that the caller might do something with it.
	if errno != 0 {
		return nil, fmt.Errorf("syscall errno: %w", errno)
	}

	return &tcpInfo, nil
}
