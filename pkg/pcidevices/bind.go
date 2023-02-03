// Copyright 2018 Intel Corporation.
// Copyright 2023 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

// Package pcidevices helps to query DPDK compatibles devices and to bind/unbind drivers
package pcidevices

import (
	"errors"
	"strings"
)

// New returns a corresponding device by given input name
func New(id string) (*PciDevice, error) {
	switch {
	case IsPciID.Match([]byte(id)):
		return NewDeviceByPciID(id)
	default:
		return NewDeviceByNicName(id)
	}
}

// NewDeviceByPciID returns a PCI device by given PCI ID
func NewDeviceByPciID(pciID string) (*PciDevice, error) {
	device, err := GetPciDeviceByPciID(pciID)
	if err != nil {
		return nil, err
	}

	return device, nil
}

// NewDeviceByNicName returns a device by given NIC name, e.g. eth0.
func NewDeviceByNicName(nicName string) (*PciDevice, error) {
	devID, err := getDeviceID(nicName)
	if err != nil {
		return nil, err
	}

	if !IsPciID.Match([]byte(devID)) {
		return nil, ErrNoValidPciID
	}

	return GetPciDeviceByPciID(devID)
}

// IsModuleLoaded checks if the kernel has already loaded the driver or not.
func IsModuleLoaded(driver string) bool {
	output, err := cmdOutputWithTimeout(defaultTimeoutLimitation, "lsmod")
	if err != nil {
		// Can't run lsmod, return false
		return false
	}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if checkModuleLine(line, driver) {
			return true
		}
	}
	return false
}

// get the list of current PCI devices on this system. The list can be filteren by adding a classesfilter list
func GetPciDevices(classFilter ClassesFilter) ([]*PciDevice, error) {
	var devices = []*PciDevice{}
	var device *PciDevice

	out, err := cmdOutputWithTimeout(defaultTimeoutLimitation, "lspci", "-Dvmmnk")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(out), "\n")
	busy := false
	for _, line := range lines {
		if line != "" {
			if !busy {
				// get device ID
				sLine := strings.Split(line, "\t")
				if sLine[0] != "Slot:" {
					return nil, errors.New("error parsing device list")
				}

				// get device info
				device, err = GetPciDeviceByPciID(sLine[1])
				if err != nil {
					return nil, err
				}

				// only return requested types
				if deviceInFilter(device, classFilter) {
					devices = append(devices, device)
				}

				// forget all other lines
				busy = true
			}
			// go to next line
			continue
		} else {
			// next device
			busy = false
		}
	}

	return devices, nil
}

func deviceInFilter(device *PciDevice, filter ClassesFilter) bool {
	for _, class := range filter {
		if stringInList(device.class[0:2], class.Class) &&
			stringInList(device.vendor, class.Vendor) &&
			stringInList(device.device, class.Device) {
			return true
		}
	}

	return false
}

func stringInList(item string, list string) bool {
	if list == "" {
		return true
	}
	return strings.Contains(list, item)
}
