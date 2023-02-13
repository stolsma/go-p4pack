// Copyright(c) 2020 Intel Corporation
// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

#ifndef _INCLUDE_THREAD_H_
#define _INCLUDE_THREAD_H_

#include <stdint.h>
#include <rte_swx_pipeline.h>

/**
 * Control plane (CP) thread.
 */

int thread_init(void);

// pipeline

int pipeline_enable(struct rte_swx_pipeline *p, uint32_t thread_id);
void pipeline_disable(struct rte_swx_pipeline *p);

// block

typedef void (*block_run_f)(void *block);
int block_enable(block_run_f block_func, void *block, uint32_t thread_id);
void block_disable(void *block);

/**
 * Data plane (DP) threads.
 */

int thread_main(void *arg);

/*
 * Check that all configured WORKER lcores are in WAIT state, then run the thread_main function on all of them.
 *
 * Returns:
 * - 0: Success. Execution of thread_main function is started on all WORKER lcores.
 * - (-EBUSY): At least one WORKER lcore is not in a WAIT state. In this case, thread_main is not started on any of
 * the WORKER lcores.
 */
int thread_start(void);

#endif /* _INCLUDE_THREAD_H_ */
