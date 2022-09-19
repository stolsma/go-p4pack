// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkswx

/*
#cgo pkg-config: libdpdk

#include <stdlib.h>
#include <rte_launch.h>

#include "thread.h"

*/
import "C"

//
// Thread private functions
//

func threadPipelineEnable(threadid uint32, pipeline *Pipeline) error {
	res := C.thread_pipeline_enable(C.uint(threadid), pipeline.p, C.uint(pipeline.timerPeriodms))
	return err(res)
}

func threadPipelineDisable(pipeline *Pipeline) error {
	res := C.thread_pipeline_disable(C.uint32_t(pipeline.ThreadId()), pipeline.p)

	return err(res)
}

//
// Thread public functions
//

func ThreadInit() error {
	res := C.thread_init()
	return err(res)
}

func ThreadFree() error {
	res := C.thread_free()
	return err(res)
}

func ThreadIsRunning(threadId uint) bool {
	threadState := C.rte_eal_get_lcore_state((C.uint32_t)(threadId))

	if threadState == C.RUNNING {
		return true
	} else {
		return false
	}
}

func MainThreadInit() error {
	res := C.thread_start()
	return err(res)
}
