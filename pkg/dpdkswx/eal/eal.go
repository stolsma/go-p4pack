// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package eal

/*
#include <stdlib.h>
#include <string.h>

#include <rte_eal.h>
#include <rte_lcore.h>
#include <rte_devargs.h>

*/
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/stolsma/go-p4pack/pkg/dpdkswx/common"
)

// Call rte_eal_init and report its return value and rte_errno as an error.
func RteEalInit(args []string) (int, error) {
	argc := C.int(len(args))
	argv := make([]*C.char, argc+1)
	for i := range args {
		cstring := C.CString(args[i])
		defer C.free(unsafe.Pointer(cstring))
		argv[i] = cstring
	}

	// initialize EAL
	n := int(C.rte_eal_init(argc, &argv[0]))
	if n < 0 {
		return n, common.Err()
	}
	return n, nil
}

// EalCleanup releases DPDK EAL-allocated resources, ensuring that no hugepage memory is leaked. It is expected that all
// DPDK SWX applications call EalCleanup() before exiting. Not calling this function could result in leaking hugepages,
// leading to failure during initialization of secondary processes.
func RteEalCleanup() error {
	return common.Err(C.rte_eal_cleanup())
}

type lcoresIter struct {
	i  C.uint
	sm C.int
}

func (iter *lcoresIter) next() bool {
	iter.i = C.rte_get_next_lcore(iter.i, iter.sm, 0)
	return iter.i < C.RTE_MAX_LCORE
}

// If skipMain is 0, main lcore will be included in the result.
// Otherwise, it will miss the output.
func getLcores(skipMain int) (out []uint) {
	c := &lcoresIter{i: ^C.uint(0), sm: C.int(skipMain)}
	for c.next() {
		out = append(out, uint(c.i))
	}
	return out
}

// Returns all lcores registered in EAL.
func GetLcores() []uint {
	return getLcores(0)
}

// Returns all worker lcores registered in EAL. Lcore is worker if it is not main.
func GetLcoresWorkers() []uint {
	return getLcores(1)
}

// Returns CPU logical core id (Lcore) where the main thread is executed.
func GetMainLcore() uint {
	return uint(C.rte_get_main_lcore())
}

// Returns number of CPU logical cores configured by EAL.
func GetLcoreCount() uint {
	return uint(C.rte_lcore_count())
}

func LcoreIsRunning(lcoreID uint) bool {
	threadState := C.rte_eal_get_lcore_state((C.uint32_t)(lcoreID))

	return threadState == C.RUNNING
}

// HasHugePages tells if huge pages are activated.
func HasHugePages() bool {
	return int(C.rte_eal_has_hugepages()) != 0
}

// HasPCI tells whether EAL is using PCI bus. Disabled by –no-pci option.
func HasPCI() bool {
	return int(C.rte_eal_has_pci()) != 0
}

// Returns the current process type.
func ProcessType() int {
	return int(C.rte_eal_process_type())
}

// Type of generic device
type RteDevtype uint32

const (
	RteDevtypeAllowed RteDevtype = iota
	RteDevtypeBlocked
	RteDevtypeVirtual
)

type DevArgs struct {
	dType   RteDevtype // the device type
	bus     string     // name of the bus
	name    string     // name of the device
	drvArgs string     // driver arguments of the device string
}

func (d *DevArgs) Bus() string {
	return d.bus
}

func (d *DevArgs) Name() string {
	return d.name
}

func (d *DevArgs) DrvArgs() string {
	return d.drvArgs
}

func (d *DevArgs) Type() RteDevtype {
	return d.dType
}

func (d *DevArgs) SetArgs(bus string, name string, drvArgs string, dType ...RteDevtype) {
	if len(dType) == 0 {
		d.dType = RteDevtypeAllowed
	} else {
		d.dType = dType[0]
	}
	d.bus = bus
	d.name = name
	d.drvArgs = drvArgs
}

// parses device arguments like "virtio_user4,path=/dev/vhost-net,queues=1,queue_size=32,iface=sw3" to devargs struct
func (d *DevArgs) Parse(id string) error {
	var da C.struct_rte_devargs

	cID := C.CString(id)
	defer C.free(unsafe.Pointer(cID))

	res := C.rte_devargs_parse(&da, cID) //nolint:gocritic
	if res != 0 {
		return common.Err(res)
	}
	defer C.rte_devargs_reset(&da) //nolint:gocritic

	// get all values and transfer to go struct
	d.dType = RteDevtype(da._type)
	d.name = C.GoString(&da.name[0])
	if da.bus != nil {
		d.bus = C.GoString(da.bus.name)
	}
	d.drvArgs = C.GoString(*(**C.char)(unsafe.Pointer(&da.anon0[0])))

	return nil
}

// Hotplug add (attach) a DPDK device. Returns error when something went wrong.
func HotplugAdd(d *DevArgs) error {
	cBus := C.CString(d.bus)
	defer C.free(unsafe.Pointer(cBus))
	cDevName := C.CString(d.name)
	defer C.free(unsafe.Pointer(cDevName))
	cDrvStr := C.CString(d.drvArgs)
	defer C.free(unsafe.Pointer(cDrvStr))

	status := C.rte_eal_hotplug_add(cBus, cDevName, cDrvStr)
	if status != 0 {
		return fmt.Errorf("hotplug add failed (%w)", common.Err(status))
	}

	return nil
}

// Hotplug remove (detach) a DPDK device. Returns error when something went wrong.
func HotplugRemove(d *DevArgs) error {
	cBus := C.CString(d.bus)
	defer C.free(unsafe.Pointer(cBus))
	cDevName := C.CString(d.name)
	defer C.free(unsafe.Pointer(cDevName))

	status := C.rte_eal_hotplug_remove(cBus, cDevName)
	if status != 0 {
		return fmt.Errorf("hotplug remove failed (%w)", common.Err(status))
	}

	return nil
}
