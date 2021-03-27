// Copyright (c) 2021 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin,redo ios,redo

package ipnlocal

import (
	"fmt"
	"log"
	"net"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
	"inet.af/netaddr"
	"tailscale.com/net/interfaces"
)

func init() {
	initListenConfig = initListenConfigNetworkExtension
}

// initListenConfigNetworkExtension configures nc for listening on IP
// through the iOS/macOS Network/System Extension (Packet Tunnel
// Provider) sandbox.
func initListenConfigNetworkExtension(nc *net.ListenConfig, ip netaddr.IP, st *interfaces.State, tunIfName string) error {
	tunIf, ok := st.Interface[tunIfName]
	if !ok {
		return fmt.Errorf("no interface with name %q", tunIfName)
	}
	nc.Control = func(network, address string, c syscall.RawConn) error {
		var sockErr error
		err := c.Control(func(fd uintptr) {

			v6 := strings.Contains(address, "]:") || strings.HasSuffix(network, "6") // hacky test for v6
			proto := unix.IPPROTO_IP
			opt := unix.IP_BOUND_IF
			if v6 {
				proto = unix.IPPROTO_IPV6
				opt = unix.IPV6_BOUND_IF
			}

			sockErr = unix.SetsockoptInt(int(fd), proto, opt, tunIf.Index)
			log.Printf("peerapi: bind(%q, %q) on index %v = %v", network, address, tunIf.Index, sockErr)
		})
		if err != nil {
			return err
		}
		return sockErr
	}
	return nil
}