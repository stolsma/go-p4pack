// Copyright 2023 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

// Package pcidevices helps to query DPDK compatibles devices and to bind/unbind drivers
package pcidevices

import (
	"testing"
)

func TestGetPciDevices(t *testing.T) {
	devices, err := GetPciDevices(AllDevices)
	if err != nil {
		t.Error(err)
	}

	if devices == nil {
		t.Fail()
	}

	devices, err = GetPciDevices(NetworkDevices)
	if err != nil {
		t.Error(err)
	}

	if devices == nil {
		t.Fail()
	}
}

func TestDeviceInFilter(t *testing.T) {
	device1 := &PciDevice{
		class:  "02",
		vendor: "8086",
		device: "1521",
	}

	device2 := &PciDevice{
		class:  "08",
		vendor: "8087",
		device: "6f23",
	}

	device3 := &PciDevice{
		class:  "08",
		vendor: "8086",
		device: "6f23",
	}

	device4 := &PciDevice{
		class:  "08",
		vendor: "8086",
		device: "6f50",
	}

	if !deviceInFilter(device1, NetworkDevices) {
		t.Fail()
	}

	if deviceInFilter(device2, DmaDevices) {
		t.Fail()
	}

	if !deviceInFilter(device3, DmaDevices) {
		t.Fail()
	}

	if deviceInFilter(device4, DmaDevices) {
		t.Fail()
	}
}

func TestNewDeviceByNicName(t *testing.T) {
	res, err := NewDeviceByNicName("eth0")
	if err != nil && err != ErrNoDeviceID {
		t.Error(err)
	}
	if res == nil {
		return
	}
}
