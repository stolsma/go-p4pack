// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkinfra

/*
#cgo pkg-config: libdpdk

#include <rte_eal.h>

*/
import "C"

import (
	"github.com/yerden/go-dpdk/common"
)

//
// Infra Init functions
//

// call rte_eal_init and report its return value and rte_errno as an
// error. Should be run in main lcore thread only
func EalInit(args []string) (int, error) {
	mem := common.NewAllocatorSession(&common.StdAlloc{})
	defer mem.Flush()

	argc := C.int(len(args))
	argv := make([]*C.char, argc+1)
	for i := range args {
		argv[i] = (*C.char)(common.CString(mem, args[i]))
	}

	// initialize EAL
	n := int(C.rte_eal_init(argc, &argv[0]))
	if n < 0 {
		return n, err()
	}
	return n, nil
}

// EalCleanup releases EAL-allocated resources, ensuring that no hugepage
// memory is leaked. It is expected that all DPDK applications call
// rte_eal_cleanup() before exiting. Not calling this function could
// result in leaking hugepages, leading to failure during
// initialization of secondary processes.
func EalCleanup() error {
	return err(C.rte_eal_cleanup())
}
