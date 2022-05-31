/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright(c) 2020 Intel Corporation
 */

#include <stdio.h>
#include <string.h>
#include <fcntl.h>
#include <unistd.h>
#include <getopt.h>
#include <signal.h>

#include "dpdk_infra.h"

static char *dpdk_args[] = {"dummy", "-n", "4", "-c", "3"};

/*
/* Simple signal handler for now
 */
void signal_handler(int signum) {
	if (signum == SIGUSR1) {
			printf("Received SIGUSR1!\n");
	}
}

/*
/* Simple main program to run the dpdk swx pipeline threads and cli
 */
int main(int argc, char **argv) {
	/* show pid */
	int pid;
	pid = getpid();
	printf("PID: %d\n", pid);

	/* start pipeline threads + cli */
	size_t arr_size = sizeof(dpdk_args) / sizeof(*dpdk_args);
	dpdk_infra_init(arr_size, dpdk_args, true);

	/* handle signals and go to pause */
	signal(SIGUSR1, signal_handler);
	pause();

	/* TODO: clean up the EAL */
	// rte_eal_cleanup();
}