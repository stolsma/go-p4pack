// Copyright 2018 Intel Corporation.
// Copyright 2023 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

// Package pcidevices helps to query DPDK compatibles devices and to bind/unbind drivers
package pcidevices

import (
	"errors"
	"fmt"
	"regexp"
)

// Driver names
const (
	DriverUioPciGeneric = "uio_pci_generic"
	DriverIgbUio        = "igb_uio"
	DriverVfioPci       = "vfio-pci"
)

// Path to PCI
const (
	PathSysPciDevices     = "/sys/bus/pci/devices"
	PathSysPciDrivers     = "/sys/bus/pci/drivers"
	PathSysPciDriverProbe = "/sys/bus/pci/drivers_probe"
)

// Path to net
const (
	PathSysClassNet = "/sys/class/net"
)

// Regular expressions for PCI-ID
var (
	IsPciID *regexp.Regexp
)

// DPDK related drivers
var (
	DefaultDpdkDriver = DriverVfioPci
	DpdkDrivers       = [...]string{DriverUioPciGeneric, DriverIgbUio, DriverVfioPci}
	DpdkPciDrivers    = [...]string{DriverUioPciGeneric, DriverIgbUio, DriverVfioPci}
)

type stringBuilder string

func (s stringBuilder) With(args ...interface{}) string {
	return fmt.Sprintf(string(s), args...)
}

var (
	// pathSysPciDevicesBind           stringBuilder = PathSysPciDevices + "/%s/driver/bind"
	pathSysPciDevicesUnbind         stringBuilder = PathSysPciDevices + "/%s/driver/unbind"
	pathSysPciDevicesOverrideDriver stringBuilder = PathSysPciDevices + "/%s/driver_override"

	pathSysPciDriversBind   stringBuilder = PathSysPciDrivers + "/%s/bind"
	pathSysPciDriversUnbind stringBuilder = PathSysPciDrivers + "/%s/unbind"
	pathSysPciDriversNewID  stringBuilder = PathSysPciDrivers + "/%s/new_id"
)

var (
	pathSysClassNetDevice stringBuilder = PathSysClassNet + "/%s"
)

// lspci like print
var (
	pciDeviceStringer stringBuilder = "Slot:\t%s\nClass:\t%s\nVendor:\t%s\nDevice:\t%s\nDriver:\t%s"
)

var (
	// for lspci output
	rPciClassExt  *regexp.Regexp
	rPciVendorExt *regexp.Regexp
	rPciDeviceExt *regexp.Regexp
	rPciDriver    *regexp.Regexp
)

func init() {
	rPciClassExt = regexp.MustCompile(`[Cc]lass:\s(.*)\[(.*)\]`)
	rPciVendorExt = regexp.MustCompile(`[Vv]endor:\s(.*)\[(.*)\]`)
	rPciDeviceExt = regexp.MustCompile(`[Dd]evice:\s(.*)\[(.*)\]`)
	rPciDriver = regexp.MustCompile(`[Dd]river:\s(\S+)`)

	// domains are numbered from 0 to ffff), bus (0 to ff), slot (0 to 1f) and function (0 to 7)
	IsPciID = regexp.MustCompile("^[[:xdigit:]]{4}:[[:xdigit:]]{2}:[0-1][[:xdigit:]].[0-7]$")
}

type ClassFilter struct {
	Class  string
	Vendor string
	Device string
}

// The PCI base classes for all devices
var (
	allClasses = ClassFilter{Class: "", Vendor: "", Device: ""}

	networkClass        = ClassFilter{Class: "02", Vendor: "", Device: ""}
	accelerationClass   = ClassFilter{Class: "12", Vendor: "", Device: ""}
	ifpgaClass          = ClassFilter{Class: "12", Vendor: "8086", Device: "0b30"}
	encryptionClass     = ClassFilter{Class: "10", Vendor: "", Device: ""}
	intelProcessorClass = ClassFilter{Class: "0b", Vendor: "8086", Device: ""}
	caviumSso           = ClassFilter{Class: "08", Vendor: "177d", Device: "a04b,a04d"}
	caviumFpa           = ClassFilter{Class: "08", Vendor: "177d", Device: "a053"}
	caviumPkx           = ClassFilter{Class: "08", Vendor: "177d", Device: "a0dd,a049"}
	caviumTim           = ClassFilter{Class: "08", Vendor: "177d", Device: "a051"}
	caviumZip           = ClassFilter{Class: "12", Vendor: "177d", Device: "a037"}
	avpVnic             = ClassFilter{Class: "05", Vendor: "1af4", Device: "1110"}

	cnxkBphy    = ClassFilter{Class: "08", Vendor: "177d", Device: "a089"}
	cnxkBphyCgx = ClassFilter{Class: "08", Vendor: "177d", Device: "a059,a060"}
	cnxkDma     = ClassFilter{Class: "08", Vendor: "177d", Device: "a081"}
	cnxkInlDev  = ClassFilter{Class: "08", Vendor: "177d", Device: "a0f0,a0f1"}

	hisiliconDma = ClassFilter{Class: "08", Vendor: "19e5", Device: "a122"}

	intelDlb     = ClassFilter{Class: "0b", Vendor: "8086", Device: "270b,2710,2714"}
	intelIoatBdw = ClassFilter{Class: "08", Vendor: "8086", Device: "6f20,6f21,6f22,6f23,6f24,6f25,6f26,6f27,6f2e,6f2f"}
	intelIoatSkx = ClassFilter{Class: "08", Vendor: "8086", Device: "2021"}
	intelIoatIcx = ClassFilter{Class: "08", Vendor: "8086", Device: "0b00"}
	intelIdxdSpr = ClassFilter{Class: "08", Vendor: "8086", Device: "0b25"}
	intelNtbSkx  = ClassFilter{Class: "06", Vendor: "8086", Device: "201c"}
	intelNtbIcx  = ClassFilter{Class: "06", Vendor: "8086", Device: "347e"}

	cnxkSso = ClassFilter{Class: "08", Vendor: "177d", Device: "a0f9,a0fa"}
	cnxkNpa = ClassFilter{Class: "08", Vendor: "177d", Device: "a0fb,a0fc"}
	cn9kRee = ClassFilter{Class: "08", Vendor: "177d", Device: "a0f4"}

	virtioBlk = ClassFilter{Class: "01", Vendor: "1af4", Device: "1001,1042"}
)

type ClassesFilter []ClassFilter

var (
	AllDevices      = ClassesFilter{allClasses}
	NetworkDevices  = ClassesFilter{networkClass, caviumPkx, avpVnic, ifpgaClass}
	BasebandDevices = ClassesFilter{accelerationClass}
	CryptoDevices   = ClassesFilter{encryptionClass, intelProcessorClass}
	DmaDevices      = ClassesFilter{cnxkDma, hisiliconDma, intelIdxdSpr, intelIoatBdw, intelIoatIcx, intelIoatSkx}
	EventdevDevices = ClassesFilter{caviumSso, caviumTim, intelDlb, cnxkSso}
	MempoolDevices  = ClassesFilter{caviumFpa, cnxkNpa}
	CompressDevices = ClassesFilter{caviumZip}
	RegexDevices    = ClassesFilter{cn9kRee}
	MiscDevices     = ClassesFilter{cnxkBphy, cnxkBphyCgx, cnxkInlDev, intelNtbSkx, intelNtbIcx, virtioBlk}
)

// Errors of devices package
var (
	ErrNoBoundDriver         = errors.New("no driver is bound to the device")
	ErrAlreadyBoundDriver    = errors.New("device has already bound the selected driver")
	ErrBind                  = errors.New("fail to bind the driver")
	ErrUnbind                = errors.New("fail to unbind the driver")
	ErrUnsupportedDriver     = errors.New("unsupported DPDK driver")
	ErrNotProbe              = errors.New("device doesn't support 'drive_probe'")
	ErrKernelModuleNotLoaded = errors.New("kernel module is not loaded")
	ErrNoValidPciID          = errors.New("no valid pci identifier")
	ErrNoDeviceID            = errors.New("can't get device ID from NIC")
)
