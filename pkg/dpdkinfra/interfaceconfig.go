// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkinfra

import (
	"log"

	"github.com/vishvananda/netlink"
)

type InterfaceConfig struct {
	Name string
	Type string
}

func (i *InterfaceConfig) GetName() string {
	return i.Name
}

func (i *InterfaceConfig) GetType() string {
	return i.Type
}

// Create interfaces through the DpdkInfra API
func (dpdki *DpdkInfra) InterfaceWithConfig(interfaceConfig InterfaceConfig) {
	// Create (TAP) interface ports
	// TODO: Implement other interface types!!
	name := interfaceConfig.GetName()
	err := dpdki.TapCreate(name)
	if err != nil {
		log.Fatalf("TAP %s create err: %d", name, err)
	}

	// TODO Temporaraly set interface up here but refactor interfaces into seperate dpdki module!
	dpdki.InterfaceUp(name)
	log.Printf("TAP %s created!", name)
}

// Create interfaces through the DpdkInfra API
func (dpdki *DpdkInfra) InterfaceUp(name string) error {
	// set interface up if not already up
	localInterface, err := netlink.LinkByName(name)
	if err != nil {
		return err
	}
	if err = netlink.LinkSetUp(localInterface); err != nil {
		return err
	}

	list, _ := netlink.AddrList(localInterface, netlink.FAMILY_ALL)
	for _, addr := range list {
		a := addr
		err := netlink.AddrDel(localInterface, &a)
		if err != nil {
			log.Printf("Couldnt remove address %s from interface %s", addr.String(), name)
		}
	}

	return nil
}
