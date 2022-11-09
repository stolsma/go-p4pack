// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

#ifndef _INCLUDE_THREAD_H_
#define _INCLUDE_THREAD_H_

#include <stdint.h>
#include <rte_swx_pipeline.h>

#ifndef NAME_MAX
#define NAME_MAX 64
#endif

void thread_free(void);
int thread_pipeline_enable(uint32_t thread_id, struct rte_swx_pipeline *pl, uint32_t timer_period_ms);
int thread_pipeline_disable(uint32_t thread_id, struct rte_swx_pipeline *pl);
int thread_init(void);

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
