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

func EnablePipeline(pl unsafe.Pointer, threadID uint) error {
	res := C.pipeline_enable((*C.struct_rte_swx_pipeline)(pl), C.uint(threadID))
	return common.Err(res)
}

func DisablePipeline(pl unsafe.Pointer) {
	C.pipeline_disable((*C.struct_rte_swx_pipeline)(pl))
}

//
// Block functions
//

func EnableBlock(fn unsafe.Pointer, block unsafe.Pointer, threadID uint) error {
	res := C.block_enable((C.block_run_f)(fn), block, C.uint(threadID))
	return common.Err(res)
}

func DisableBlock(block unsafe.Pointer) {
	C.block_disable(block)
}

//
// Thread functions
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
