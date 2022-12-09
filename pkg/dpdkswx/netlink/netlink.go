// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package netlink

import (
	"github.com/stolsma/go-p4pack/pkg/logging"
	"github.com/vishvananda/netlink"
)

var log logging.Logger

func init() {
	// keep the logger up to date, also after new log config
	logging.Register("dpdkswx/netlink", func(logger logging.Logger) {
		log = logger
	})
}

// set interface up if not already up
func InterfaceUp(name string) error {
	localInterface, err := netlink.LinkByName(name)
	if err != nil {
		return err
	}

	if err = netlink.LinkSetUp(localInterface); err != nil {
		return err
	}

	return nil
}

// Remove all addresses from interface
func RemoveAllAddr(name string) error {
	localInterface, err := netlink.LinkByName(name)
	if err != nil {
		return err
	}

	list, _ := netlink.AddrList(localInterface, netlink.FAMILY_ALL)
	for _, addr := range list {
		a := addr
		err := netlink.AddrDel(localInterface, &a)
		if err != nil {
			log.Infof("Couldnt remove address %s from interface %s", addr.String(), name)
		}
	}

	return nil
}
