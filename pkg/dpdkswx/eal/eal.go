// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package eal

/*
#include <stdlib.h>
#include <rte_eal.h>
#include <rte_lcore.h>
*/
import "C"
import (
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
