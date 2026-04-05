//go:build linux
// +build linux

package preparer

import (
	"log"

	"golang.org/x/sys/unix"
)

func (c *ConnConfig) Cork() error {
	log.Println("Using TCP_NODELAY WITH TCP_CORK..")
	if err := c.RawTcpSyscall.Control(func(fd uintptr) {
		_ = unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_CORK, 1)
	}); err != nil {
		return err
	}
	return nil
}

func (c *ConnConfig) UnCork() error {
	if err := c.RawTcpSyscall.Control(func(fd uintptr) {
		_ = unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_CORK, 0)
	}); err != nil {
		return err
	}
	return nil
}
