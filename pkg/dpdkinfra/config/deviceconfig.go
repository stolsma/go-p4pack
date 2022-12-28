// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"

	"github.com/stolsma/go-p4pack/pkg/dpdkinfra"
)

type DevicesConfig []string

// Create hotplug ethdev interfaces with given DPDK device argument string
func (c DevicesConfig) Apply() error {
	dpdki := dpdkinfra.Get()
	if dpdki == nil {
		return errors.New("dpdkinfra module is not initialized")
	}

	// hotplug devices
	for _, devArgString := range c {
		devArgs, err := dpdki.HotplugAdd(devArgString)
		if err != nil {
			log.Infof("Hotplug (devargs: %s) error: %v", devArgString, err)
			return fmt.Errorf("error creating hotplug device: %v", err)
		}
		log.Infof("Device (%s) created via hotplug on bus %s!", devArgs.Name(), devArgs.Bus())
	}

	return nil
}
