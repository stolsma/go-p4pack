// Copyright 2022 Google LLC
// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package kernel

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

// SetInterfaceHWAddr sets the MAC address of a network interface.
func SetInterfaceHWAddr(name string, addr string) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("failed to get interface: %w", err)
	}
	addrBytes, err := net.ParseMAC(addr)
	if err != nil {
		return fmt.Errorf("failed to set parse addres: %v", err)
	}
	if err := netlink.LinkSetHardwareAddr(link, addrBytes); err != nil {
		return fmt.Errorf("failed to get hwaddr of link: %w", err)
	}
	return nil
}

// SetInterfaceIP sets the IP addresses of a network interface.
func SetInterfaceIP(name string, ip string, prefixLen int) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("failed to get interface: %w", err)
	}
	ipNet := &net.IPNet{}
	ipNet.IP = net.ParseIP(ip)
	if ipNet.IP == nil {
		return fmt.Errorf("failed to parse ip")
	}
	ipNet.Mask = net.CIDRMask(prefixLen, 128)
	if ipNet.IP.To4() != nil { // If ip is IPv4.
		ipNet.Mask = net.CIDRMask(prefixLen, 32)
	}
	if err := netlink.AddrReplace(link, &netlink.Addr{IPNet: ipNet}); err != nil {
		return fmt.Errorf("failed to add ip to link: %w", err)
	}
	return nil
}

// DeleteInterfaceIP delete an IP addresses from a network interface.
func DeleteInterfaceIP(name string, ip *net.IPNet) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("failed to get interface: %w", err)
	}
	if err := netlink.AddrDel(link, &netlink.Addr{IPNet: ip}); err != nil {
		return fmt.Errorf("failed to add ip to link: %w", err)
	}
	return nil
}

// SetInterfaceState sets a links up or down.
func SetInterfaceState(name string, up bool) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("failed to get interface: %w", err)
	}
	if up {
		return netlink.LinkSetUp(link)
	}
	return netlink.LinkSetDown(link)
}

// CreateTAP creates kernel TAP interface.
func CreateTAP(name string) (int, error) {
	fd, err := unix.Open("/dev/net/tun", unix.O_RDWR, 0)
	if err != nil {
		return -1, fmt.Errorf("failed to open tun file: %w", err)
	}
	req, err := unix.NewIfreq(name)
	if err != nil {
		// TODO Close fd
		return -1, fmt.Errorf("failed to create interface req: %w", err)
	}
	req.SetUint16(unix.IFF_TAP | unix.IFF_NO_PI)
	if err := unix.IoctlIfreq(fd, unix.TUNSETIFF, req); err != nil {
		// TODO Close fd
		return -1, fmt.Errorf("failed to do ioctl: %v", err)
	}
	return fd, nil
}
