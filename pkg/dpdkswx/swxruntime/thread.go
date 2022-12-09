// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package swxruntime

/*
#include <stdlib.h>
#include <rte_launch.h>

#include "thread.h"

*/
import "C"
import (
	"unsafe"

	"github.com/stolsma/go-p4pack/pkg/dpdkswx/common"
)

//
// Pipeline functions
//

func EnablePipeline(pl unsafe.Pointer, threadID uint, timerPeriodms uint) error {
	res := C.thread_pipeline_enable(
		C.uint(threadID), (*C.struct_rte_swx_pipeline)(pl), C.uint(timerPeriodms))
	if res != 0 {
		return common.Err(res)
	}

	return nil
}

func DisablePipeline(pl unsafe.Pointer, threadID uint) error {
	res := C.thread_pipeline_disable(C.uint(threadID), (*C.struct_rte_swx_pipeline)(pl))
	return common.Err(res)
}

//
// Thread public functions
//

func ThreadsInit() error {
	res := C.thread_init()
	return common.Err(res)
}

func ThreadsStart() error {
	res := C.thread_start()
	return common.Err(res)
}

func ThreadsStop() error {
	return nil
}

func ThreadsFree() error {
	res := C.thread_free()
	return common.Err(res)
}
