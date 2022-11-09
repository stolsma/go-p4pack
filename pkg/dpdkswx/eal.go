// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkswx

/*
#cgo pkg-config: libdpdk

#include <stdlib.h>
#include <rte_eal.h>

*/
import "C"
import "unsafe"

// Call rte_eal_init and report its return value and rte_errno as an error.
func EalInit(args []string) (int, error) {
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
		return n, err()
	}
	return n, nil
}

// EalCleanup releases DPDK EAL-allocated resources, ensuring that no hugepage memory is leaked. It is expected that all
// DPDK SWX applications call EalCleanup() before exiting. Not calling this function could result in leaking hugepages,
// leading to failure during initialization of secondary processes.
func EalCleanup() error {
	return err(C.rte_eal_cleanup())
}
