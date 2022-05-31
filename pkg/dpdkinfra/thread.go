// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkinfra

/*
#cgo pkg-config: libdpdk

#include <stdlib.h>
#include <string.h>
#include <netinet/in.h>
#ifdef RTE_EXEC_ENV_LINUX
#include <linux/if.h>
#include <linux/if_tun.h>
#endif
#include <sys/ioctl.h>
#include <fcntl.h>
#include <unistd.h>
#include <stdint.h>
#include <sys/queue.h>
#include <endian.h>

#include <rte_common.h>
#include <rte_byteorder.h>
#include <rte_cycles.h>
#include <rte_lcore.h>
#include <rte_ring.h>

#include <rte_table_acl.h>
#include <rte_table_array.h>
#include <rte_table_hash.h>
#include <rte_table_lpm.h>
#include <rte_table_lpm_ipv6.h>

#include <rte_mempool.h>
#include <rte_mbuf.h>
#include <rte_ethdev.h>
#include <rte_swx_port_fd.h>
#include <rte_swx_pipeline.h>
#include <rte_swx_ctl.h>

#ifndef NAME_SIZE
#define NAME_SIZE 64
#endif

//
// obj
//
#define ntoh64(x) rte_be_to_cpu_64(x)
#define hton64(x) rte_cpu_to_be_64(x)

#if __BYTE_ORDER == __LITTLE_ENDIAN
#define dpdk_field_ntoh(val, n_bits) (ntoh64((val) << (64 - n_bits)))
#define dpdk_field_hton(val, n_bits) (hton64((val) << (64 - n_bits)))
#else
#define dpdk_field_ntoh(val, n_bits) (val)
#define dpdk_field_hton(val, n_bits) (val)
#endif

#define PNA_DIR_REG_NAME "direction"

//
// mempool
//
struct mempool_params {
	uint32_t buffer_size;
	uint32_t pool_size;
	uint32_t cache_size;
	uint32_t cpu_id;
};

struct mempool {
	TAILQ_ENTRY(mempool) node;
	char name[NAME_SIZE];
	struct rte_mempool *m;
	uint32_t buffer_size;
};

//
// link
//
#ifndef LINK_RXQ_RSS_MAX
#define LINK_RXQ_RSS_MAX                                   16
#endif

struct link_params_rss {
	uint32_t queue_id[LINK_RXQ_RSS_MAX];
	uint32_t n_queues;
};

struct link_params {
	const char *dev_name;
	const char *dev_args;
	uint16_t port_id; // Valid only when *dev_name* is NULL.
	uint16_t dev_hotplug_enabled;

	struct {
		uint32_t n_queues;
		uint32_t queue_size;
		const char *mempool_name;
		struct link_params_rss *rss;
	} rx;

	struct {
		uint32_t n_queues;
		uint32_t queue_size;
	} tx;

	int promiscuous;
};

struct link {
	TAILQ_ENTRY(link) node;
	char name[NAME_SIZE];
	char dev_name[NAME_SIZE];
	uint16_t port_id;
	uint32_t n_rxq;
	uint32_t n_txq;
};

//
// ring
//
struct ring_params {
	uint32_t size;
	uint32_t numa_node;
};

struct ring {
	TAILQ_ENTRY(ring) node;
	char name[NAME_SIZE];
};

//
// tap
//
struct tap {
	TAILQ_ENTRY(tap) node;
	char name[NAME_SIZE];
	int fd;
};

//
// pipeline
//
struct pipeline {
	TAILQ_ENTRY(pipeline) node;
	char name[NAME_SIZE];

	struct rte_swx_pipeline *p;
	struct rte_swx_ctl_pipeline *ctl;

	uint32_t timer_period_ms;
	int enabled;
	uint32_t thread_id;
	uint32_t cpu_id;
	uint64_t net_port_mask[4];
};

//
// mempool
//
TAILQ_HEAD(mempool_list, mempool);

//
// link
//
TAILQ_HEAD(link_list, link);

//
// ring
//
TAILQ_HEAD(ring_list, ring);

//
// tap
//
TAILQ_HEAD(tap_list, tap);

//
// pipeline
//
TAILQ_HEAD(pipeline_list, pipeline);

//
// obj
//
struct obj {
	struct mempool_list mempool_list;
	struct link_list link_list;
	struct ring_list ring_list;
	struct pipeline_list pipeline_list;
	struct tap_list tap_list;
};

struct obj *obj;

//
// mempool
//
#define BUFFER_SIZE_MIN (sizeof(struct rte_mbuf) + RTE_PKTMBUF_HEADROOM)

struct mempool * mempool_find(const char *name) {
	struct mempool *mempool;

	if (!name)
		return NULL;

	TAILQ_FOREACH(mempool, &obj->mempool_list, node)
		if (strcmp(mempool->name, name) == 0)
			return mempool;

	return NULL;
}

int mempool_create(const char *name, struct mempool_params *params) {
	struct mempool *mempool;
	struct rte_mempool *m;

	// Check input params
	if ((name == NULL) ||
		mempool_find(name) ||
		(params == NULL) ||
		(params->buffer_size < BUFFER_SIZE_MIN) ||
		(params->pool_size == 0))
		return 1;

	// Resource create
	m = rte_pktmbuf_pool_create(
		name,
		params->pool_size,
		params->cache_size,
		0,
		params->buffer_size - sizeof(struct rte_mbuf),
		params->cpu_id);
	if (m == NULL)
		return 2;

	// Node allocation
	mempool = calloc(1, sizeof(struct mempool));
	if (mempool == NULL) {
		rte_mempool_free(m);
		return 3;
	}

	// Node fill in
	strlcpy(mempool->name, name, sizeof(mempool->name));
	mempool->m = m;
	mempool->buffer_size = params->buffer_size;

	// Node add to list
	TAILQ_INSERT_TAIL(&obj->mempool_list, mempool, node);

	return 0;
}

//
// link
//
static struct rte_eth_conf port_conf_default = {
	.link_speeds = 0,
	.rxmode = {
		.mq_mode = RTE_ETH_MQ_RX_NONE,
		.mtu = 9000 - (RTE_ETHER_HDR_LEN + RTE_ETHER_CRC_LEN), // Jumbo frame MTU
		.split_hdr_size = 0, // Header split buffer size
	},
	.rx_adv_conf = {
		.rss_conf = {
			.rss_key = NULL,
			.rss_key_len = 40,
			.rss_hf = 0,
		},
	},
	.txmode = {
		.mq_mode = RTE_ETH_MQ_TX_NONE,
	},
	.lpbk_mode = 0,
};

#define RETA_CONF_SIZE     (RTE_ETH_RSS_RETA_SIZE_512 / RTE_ETH_RETA_GROUP_SIZE)

static int rss_setup(uint16_t port_id, uint16_t reta_size, struct link_params_rss *rss) {
	struct rte_eth_rss_reta_entry64 reta_conf[RETA_CONF_SIZE];
	uint32_t i;
	int status;

	// RETA setting
	memset(reta_conf, 0, sizeof(reta_conf));

	for (i = 0; i < reta_size; i++)
		reta_conf[i / RTE_ETH_RETA_GROUP_SIZE].mask = UINT64_MAX;

	for (i = 0; i < reta_size; i++) {
		uint32_t reta_id = i / RTE_ETH_RETA_GROUP_SIZE;
		uint32_t reta_pos = i % RTE_ETH_RETA_GROUP_SIZE;
		uint32_t rss_qs_pos = i % rss->n_queues;

		reta_conf[reta_id].reta[reta_pos] =
			(uint16_t) rss->queue_id[rss_qs_pos];
	}

	// RETA update
	status = rte_eth_dev_rss_reta_update(port_id,
		reta_conf,
		reta_size);

	return status;
}

struct link * link_find(const char *name) {
	struct link *link;

	if (!name)
		return NULL;

	TAILQ_FOREACH(link, &obj->link_list, node)
		if (strcmp(link->name, name) == 0)
			return link;

	return NULL;
}

struct link * link_create(const char *name, struct link_params *params) {
	struct rte_eth_dev_info port_info;
	struct rte_eth_conf port_conf;
	struct link *link;
	struct link_params_rss *rss;
	struct mempool *mempool;
	uint32_t cpu_id, i;
	int status;
	uint16_t port_id;

	// Check input params
	if ((name == NULL) ||
		link_find(name) ||
		(params == NULL) ||
		(params->rx.n_queues == 0) ||
		(params->rx.queue_size == 0) ||
		(params->tx.n_queues == 0) ||
		(params->tx.queue_size == 0))
		return NULL;

	printf("LINK CREATE: Dev:%s Args:%s dev_hotplug_enabled: %d\n",
		params->dev_name, params->dev_args, params->dev_hotplug_enabled);

	// Performing Device Hotplug and valid for only VDEVs
	if (params->dev_hotplug_enabled) {
		if (rte_eal_hotplug_add("vdev", params->dev_name,
					params->dev_args)) {
			printf("LINK CREATE: Dev:%s probing failed\n",
				params->dev_name);
			return NULL;
		}
		printf("LINK CREATE: Dev:%s probing successful\n",
			params->dev_name);
	}

	port_id = params->port_id;
	if (params->dev_name) {
		status = rte_eth_dev_get_port_by_name(params->dev_name,
			&port_id);

		if (status)
			return NULL;
	} else
		if (!rte_eth_dev_is_valid_port(port_id))
			return NULL;

	if (rte_eth_dev_info_get(port_id, &port_info) != 0)
		return NULL;

	mempool = mempool_find(params->rx.mempool_name);
	if (mempool == NULL)
		return NULL;

	rss = params->rx.rss;
	if (rss) {
		if ((port_info.reta_size == 0) ||
			(port_info.reta_size > RTE_ETH_RSS_RETA_SIZE_512))
			return NULL;

		if ((rss->n_queues == 0) ||
			(rss->n_queues >= LINK_RXQ_RSS_MAX))
			return NULL;

		for (i = 0; i < rss->n_queues; i++)
			if (rss->queue_id[i] >= port_info.max_rx_queues)
				return NULL;
	}

	//
	// Resource create
	//
	// Port
	memcpy(&port_conf, &port_conf_default, sizeof(port_conf));
	if (rss) {
		port_conf.rxmode.mq_mode = RTE_ETH_MQ_RX_RSS;
		port_conf.rx_adv_conf.rss_conf.rss_hf =
			(RTE_ETH_RSS_IP | RTE_ETH_RSS_TCP | RTE_ETH_RSS_UDP) &
			port_info.flow_type_rss_offloads;
	}

	cpu_id = (uint32_t) rte_eth_dev_socket_id(port_id);
	if (cpu_id == (uint32_t) SOCKET_ID_ANY)
		cpu_id = 0;

	status = rte_eth_dev_configure(
		port_id,
		params->rx.n_queues,
		params->tx.n_queues,
		&port_conf);

	if (status < 0)
		return NULL;

	if (params->promiscuous) {
		status = rte_eth_promiscuous_enable(port_id);
		if (status != 0)
			return NULL;
	}

	// Port RX
	for (i = 0; i < params->rx.n_queues; i++) {
		status = rte_eth_rx_queue_setup(
			port_id,
			i,
			params->rx.queue_size,
			cpu_id,
			NULL,
			mempool->m);

		if (status < 0)
			return NULL;
	}

	// Port TX
	for (i = 0; i < params->tx.n_queues; i++) {
		status = rte_eth_tx_queue_setup(
			port_id,
			i,
			params->tx.queue_size,
			cpu_id,
			NULL);

		if (status < 0)
			return NULL;
	}

	// Port start
	status = rte_eth_dev_start(port_id);
	if (status < 0)
		return NULL;

	if (rss) {
		status = rss_setup(port_id, port_info.reta_size, rss);

		if (status) {
			rte_eth_dev_stop(port_id);
			return NULL;
		}
	}

	// Port link up
	status = rte_eth_dev_set_link_up(port_id);
	if ((status < 0) && (status != -ENOTSUP)) {
		rte_eth_dev_stop(port_id);
		return NULL;
	}

	// Node allocation
	link = calloc(1, sizeof(struct link));
	if (link == NULL) {
		rte_eth_dev_stop(port_id);
		return NULL;
	}

	// Node fill in
	strlcpy(link->name, name, sizeof(link->name));
	link->port_id = port_id;
	rte_eth_dev_get_name_by_port(port_id, link->dev_name);
	link->n_rxq = params->rx.n_queues;
	link->n_txq = params->tx.n_queues;

	// Node add to list
	TAILQ_INSERT_TAIL(&obj->link_list, link, node);

	return link;
}

int link_is_up(const char *name) {
	struct rte_eth_link link_params;
	struct link *link;

	// Check input params
	if (!name)
		return 0;

	link = link_find(name);
	if (link == NULL)
		return 0;

	// Resource
	if (rte_eth_link_get(link->port_id, &link_params) < 0)
		return 0;

	return (link_params.link_status == RTE_ETH_LINK_DOWN) ? 0 : 1;
}

struct link * link_next(struct link *link) {
	return (link == NULL) ?
		TAILQ_FIRST(&obj->link_list) : TAILQ_NEXT(link, node);
}

//
// ring
//
struct ring * ring_find(const char *name) {
	struct ring *ring;

	if (!name)
		return NULL;

	TAILQ_FOREACH(ring, &obj->ring_list, node)
		if (strcmp(ring->name, name) == 0)
			return ring;

	return NULL;
}

struct ring * ring_create(const char *name, struct ring_params *params) {
	struct ring *ring;
	struct rte_ring *r;
	unsigned int flags = RING_F_SP_ENQ | RING_F_SC_DEQ;

	// Check input params
	if (!name || ring_find(name) || !params || !params->size)
		return NULL;

	//
	// Resource create
	//
	r = rte_ring_create(
		name,
		params->size,
		params->numa_node,
		flags);
	if (!r)
		return NULL;

	// Node allocation
	ring = calloc(1, sizeof(struct ring));
	if (!ring) {
		rte_ring_free(r);
		return NULL;
	}

	// Node fill in
	strlcpy(ring->name, name, sizeof(ring->name));

	// Node add to list
	TAILQ_INSERT_TAIL(&obj->ring_list, ring, node);

	return ring;
}

//
// tap
//
#define TAP_DEV		"/dev/net/tun"

struct tap * tap_find(const char *name) {
	struct tap *tap;

	if (!name)
		return NULL;

	TAILQ_FOREACH(tap, &obj->tap_list, node)
		if (strcmp(tap->name, name) == 0)
			return tap;

	return NULL;
}

struct tap * tap_next(struct tap *tap) {
	return (tap == NULL) ?
		TAILQ_FIRST(&obj->tap_list) : TAILQ_NEXT(tap, node);
}

#ifndef RTE_EXEC_ENV_LINUX

int tap_create(struct obj *obj __rte_unused, const char *name __rte_unused) {
	return 0;
}

#else

int tap_create(const char *name) {
	struct tap *tap;
	struct ifreq ifr;
	int fd, status;

	// Check input params
	if ((name == NULL) ||
		tap_find(name))
		return 0; //NULL

	// Resource create
	fd = open(TAP_DEV, O_RDWR | O_NONBLOCK);
	if (fd < 0)
		return fd;

	memset(&ifr, 0, sizeof(ifr));
	ifr.ifr_flags = IFF_TAP | IFF_NO_PI; // No packet information
	strlcpy(ifr.ifr_name, name, IFNAMSIZ);

	status = ioctl(fd, TUNSETIFF, (void *) &ifr);
	if (status < 0) {
		close(fd);
		return status; //NULL
	}

	// Node allocation
	tap = calloc(1, sizeof(struct tap));
	if (tap == NULL) {
		close(fd);
		return 0; //NULL
	}
	// Node fill in
	strlcpy(tap->name, name, sizeof(tap->name));
	tap->fd = fd;

	// Node add to list
	TAILQ_INSERT_TAIL(&obj->tap_list, tap, node);

	return fd;
}

#endif

//
// pipeline
//
#ifndef PIPELINE_MSGQ_SIZE
#define PIPELINE_MSGQ_SIZE                                 64
#endif

struct pipeline * pipeline_find(const char *name) {
	struct pipeline *pipeline;

	if (!name)
		return NULL;

	TAILQ_FOREACH(pipeline, &obj->pipeline_list, node)
		if (strcmp(name, pipeline->name) == 0)
			return pipeline;

	return NULL;
}

void pipeline_create(const char *name, int numa_node) {
	struct pipeline *pipeline;
	struct rte_swx_pipeline *p = NULL;
	int status;

	// Check input params
	if ((name == NULL) ||	pipeline_find(name))
		return;

	// Resource create
	status = rte_swx_pipeline_config(&p, numa_node);
	if (status)
		goto error;

	// Node allocation
	pipeline = calloc(1, sizeof(struct pipeline));
	if (pipeline == NULL)
		goto error;

	// Node fill in
	strlcpy(pipeline->name, name, sizeof(pipeline->name));
	pipeline->p = p;
	pipeline->timer_period_ms = 10;

	// Node add to list
	TAILQ_INSERT_TAIL(&obj->pipeline_list, pipeline, node);

	return;

error:
	rte_swx_pipeline_free(p);
	return;
}

//
// Validate the number of ports added to the
// pipeline in input and output directions
//
int pipeline_port_is_valid(struct pipeline *pipe) {
	struct rte_swx_ctl_pipeline_info pipe_info = {0};

	if (rte_swx_ctl_pipeline_info_get(pipe->p, &pipe_info) < 0) {
		printf("%s failed at %d for pipeinfo \n",__func__, __LINE__);
		return 0;
	}

	if (!pipe_info.n_ports_in || !(rte_is_power_of_2(pipe_info.n_ports_in)))
		return 0;

	if (!pipe_info.n_ports_out)
		return 0;

	return 1;
}

//
// obj
//
int obj_init(void) {
	obj = calloc(1, sizeof(struct obj));
	if (!obj)
		return -1;

	TAILQ_INIT(&obj->mempool_list);
	TAILQ_INIT(&obj->link_list);
	TAILQ_INIT(&obj->ring_list);
	TAILQ_INIT(&obj->pipeline_list);
	TAILQ_INIT(&obj->tap_list);

	return 0;
}

void table_entry_free(struct rte_swx_table_entry *entry) {
	if (!entry)
		return;

	free(entry->key);
	free(entry->key_mask);
	free(entry->action_data);
	free(entry);
}

uint64_t get_action_id(struct pipeline *pipe, const char *action_name) {
	uint64_t i;
	int ret;
	struct rte_swx_ctl_action_info action;
	struct rte_swx_ctl_pipeline_info pipe_info = {0};

	if (action_name == NULL || pipe == NULL || pipe->p == NULL) {
		printf("%s failed at %d\n",__func__, __LINE__);
		goto action_error;
	}
	ret = rte_swx_ctl_pipeline_info_get(pipe->p, &pipe_info);
	if (ret < 0) {
		printf("%s failed at %d for pipeinfo \n",__func__, __LINE__);
		goto action_error;
	}
	for (i = 0; i < pipe_info.n_actions; i++) {
		memset(&action, 0, sizeof(action));
		ret = rte_swx_ctl_action_info_get (pipe->p, i, &action);
		if (ret < 0) {
			printf("%s failed at %d for actioninfo\n",
				__func__, __LINE__);
			break;
		}
		if (!strncmp(action_name, action.name, RTE_SWX_CTL_NAME_SIZE))
			return i;
	}
action_error:
	printf("%s failed at %d end\n",__func__, __LINE__);
	return UINT64_MAX;
}

uint32_t get_table_id(struct pipeline *pipe, const char *table_name) {
	uint32_t i;
	int ret;
	struct rte_swx_ctl_table_info table;
	struct rte_swx_ctl_pipeline_info pipe_info = {0};

	if (table_name == NULL || pipe == NULL || pipe->p == NULL) {
		printf("%s failed at %d\n",__func__, __LINE__);
		goto table_error;
	}

	ret = rte_swx_ctl_pipeline_info_get(pipe->p, &pipe_info);
	if (ret < 0) {
		printf("%s failed at %d for pipeinfo\n",__func__, __LINE__);
		goto table_error;
	}
	for (i = 0; i < pipe_info.n_tables; i++) {
		memset(&table, 0, sizeof(table));
		ret = rte_swx_ctl_table_info_get (pipe->p, i, &table);
		if (ret < 0) {
			printf("%s failed at %d for tableinfo\n",
				__func__, __LINE__);
			break;
		}
		if (!strncmp(table_name, table.name, RTE_SWX_CTL_NAME_SIZE))
			return i;
	}
table_error:
	printf("%s failed at %d end\n",__func__, __LINE__);
	return UINT32_MAX;
}


#ifndef THREAD_PIPELINES_MAX
#define THREAD_PIPELINES_MAX                               256
#endif

#ifndef THREAD_MSGQ_SIZE
#define THREAD_MSGQ_SIZE                                   64
#endif

#ifndef THREAD_TIMER_PERIOD_MS
#define THREAD_TIMER_PERIOD_MS                             100
#endif

// Pipeline instruction quanta: Needs to be big enough to do some meaningful
// work, but not too big to avoid starving any other pipelines mapped to the
// same thread. For a pipeline that executes 10 instructions per packet, a
// quanta of 1000 instructions equates to processing 100 packets.
//
#ifndef PIPELINE_INSTR_QUANTA
#define PIPELINE_INSTR_QUANTA                              1000
#endif

//
// Control thread: data plane thread context
//
struct thread {
	struct rte_ring *msgq_req;
	struct rte_ring *msgq_rsp;

	uint32_t enabled;
};

static struct thread thread[RTE_MAX_LCORE];

//
// Data plane threads: context
//
struct pipeline_data {
	struct rte_swx_pipeline *p;
	uint64_t timer_period; // Measured in CPU cycles.
	uint64_t time_next;
};

struct thread_data {
	struct rte_swx_pipeline *p[THREAD_PIPELINES_MAX];
	uint32_t n_pipelines;

	struct pipeline_data pipeline_data[THREAD_PIPELINES_MAX];
	struct rte_ring *msgq_req;
	struct rte_ring *msgq_rsp;
	uint64_t timer_period; // Measured in CPU cycles.
	uint64_t time_next;
	uint64_t time_next_min;
} __rte_cache_aligned;

static struct thread_data thread_data[RTE_MAX_LCORE];

//
// Control thread: data plane thread init
//
static void thread_free(void) {
	uint32_t i;

	for (i = 0; i < RTE_MAX_LCORE; i++) {
		struct thread *t = &thread[i];

		if (!rte_lcore_is_enabled(i))
			continue;

		// MSGQs
		if (t->msgq_req)
			rte_ring_free(t->msgq_req);

		if (t->msgq_rsp)
			rte_ring_free(t->msgq_rsp);
	}
}

int thread_init(void) {
	uint32_t i;

	RTE_LCORE_FOREACH_WORKER(i) {
		char name[NAME_MAX];
		struct rte_ring *msgq_req, *msgq_rsp;
		struct thread *t = &thread[i];
		struct thread_data *t_data = &thread_data[i];
		uint32_t cpu_id = rte_lcore_to_socket_id(i);

		// MSGQs
		snprintf(name, sizeof(name), "THREAD-%04x-MSGQ-REQ", i);

		msgq_req = rte_ring_create(name,
			THREAD_MSGQ_SIZE,
			cpu_id,
			RING_F_SP_ENQ | RING_F_SC_DEQ);

		if (msgq_req == NULL) {
			thread_free();
			return -1;
		}

		snprintf(name, sizeof(name), "THREAD-%04x-MSGQ-RSP", i);

		msgq_rsp = rte_ring_create(name,
			THREAD_MSGQ_SIZE,
			cpu_id,
			RING_F_SP_ENQ | RING_F_SC_DEQ);

		if (msgq_rsp == NULL) {
			thread_free();
			return -1;
		}

		// Control thread records
		t->msgq_req = msgq_req;
		t->msgq_rsp = msgq_rsp;
		t->enabled = 1;

		// Data plane thread records
		t_data->n_pipelines = 0;
		t_data->msgq_req = msgq_req;
		t_data->msgq_rsp = msgq_rsp;
		t_data->timer_period =
			(rte_get_tsc_hz() * THREAD_TIMER_PERIOD_MS) / 1000;
		t_data->time_next = rte_get_tsc_cycles() + t_data->timer_period;
		t_data->time_next_min = t_data->time_next;
	}

	return 0;
}

static inline int thread_is_running(uint32_t thread_id) {
	enum rte_lcore_state_t thread_state;

	thread_state = rte_eal_get_lcore_state(thread_id);
	return (thread_state == RUNNING) ? 1 : 0;
}

//
// Control thread & data plane threads: message passing
//
enum thread_req_type {
	THREAD_REQ_PIPELINE_ENABLE = 0,
	THREAD_REQ_PIPELINE_DISABLE,
	THREAD_REQ_MAX
};

struct thread_msg_req {
	enum thread_req_type type;

	union {
		struct {
			struct rte_swx_pipeline *p;
			uint32_t timer_period_ms;
		} pipeline_enable;

		struct {
			struct rte_swx_pipeline *p;
		} pipeline_disable;
	};
};

struct thread_msg_rsp {
	int status;
};

//
// Control thread
//
static struct thread_msg_req * thread_msg_alloc(void) {
	size_t size = RTE_MAX(sizeof(struct thread_msg_req),
		sizeof(struct thread_msg_rsp));

	return calloc(1, size);
}

static void thread_msg_free(struct thread_msg_rsp *rsp) {
	free(rsp);
}

static struct thread_msg_rsp * thread_msg_send_recv(uint32_t thread_id, struct thread_msg_req *req) {
	struct thread *t = &thread[thread_id];
	struct rte_ring *msgq_req = t->msgq_req;
	struct rte_ring *msgq_rsp = t->msgq_rsp;
	struct thread_msg_rsp *rsp;
	int status;

	// send
	do {
		status = rte_ring_sp_enqueue(msgq_req, req);
	} while (status == -ENOBUFS);

	// recv
	do {
		status = rte_ring_sc_dequeue(msgq_rsp, (void **) &rsp);
	} while (status != 0);

	return rsp;
}

int thread_pipeline_enable(uint32_t thread_id, const char *pipeline_name) {
	struct pipeline *p = pipeline_find(pipeline_name);
	struct thread *t;
	struct thread_msg_req *req;
	struct thread_msg_rsp *rsp;
	int status;

	// Check input params
	if ((thread_id >= RTE_MAX_LCORE) ||
		(p == NULL))
		return -1;

	t = &thread[thread_id];
	if (t->enabled == 0)
		return -1;

	if (!thread_is_running(thread_id)) {
		struct thread_data *td = &thread_data[thread_id];
		struct pipeline_data *tdp = &td->pipeline_data[td->n_pipelines];

		if (td->n_pipelines >= THREAD_PIPELINES_MAX)
			return -1;

		// Data plane thread
		td->p[td->n_pipelines] = p->p;

		tdp->p = p->p;
		tdp->timer_period =
			(rte_get_tsc_hz() * p->timer_period_ms) / 1000;
		tdp->time_next = rte_get_tsc_cycles() + tdp->timer_period;

		td->n_pipelines++;

		// Pipeline
		p->thread_id = thread_id;
		p->enabled = 1;

		return 0;
	}

	// Allocate request
	req = thread_msg_alloc();
	if (req == NULL)
		return -1;

	// Write request
	req->type = THREAD_REQ_PIPELINE_ENABLE;
	req->pipeline_enable.p = p->p;
	req->pipeline_enable.timer_period_ms = p->timer_period_ms;

	// Send request and wait for response
	rsp = thread_msg_send_recv(thread_id, req);

	// Read response
	status = rsp->status;

	// Free response
	thread_msg_free(rsp);

	// Request completion
	if (status)
		return status;

	p->thread_id = thread_id;
	p->enabled = 1;

	return 0;
}

int thread_pipeline_disable(uint32_t thread_id, const char *pipeline_name) {
	struct pipeline *p = pipeline_find(pipeline_name);
	struct thread *t;
	struct thread_msg_req *req;
	struct thread_msg_rsp *rsp;
	int status;

	// Check input params
	if ((thread_id >= RTE_MAX_LCORE) ||
		(p == NULL))
		return -1;

	t = &thread[thread_id];
	if (t->enabled == 0)
		return -1;

	if (p->enabled == 0)
		return 0;

	if (p->thread_id != thread_id)
		return -1;

	if (!thread_is_running(thread_id)) {
		struct thread_data *td = &thread_data[thread_id];
		uint32_t i;

		for (i = 0; i < td->n_pipelines; i++) {
			struct pipeline_data *tdp = &td->pipeline_data[i];

			if (tdp->p != p->p)
				continue;

			// Data plane thread
			if (i < td->n_pipelines - 1) {
				struct rte_swx_pipeline *pipeline_last =
					td->p[td->n_pipelines - 1];
				struct pipeline_data *tdp_last =
					&td->pipeline_data[td->n_pipelines - 1];

				td->p[i] = pipeline_last;
				memcpy(tdp, tdp_last, sizeof(*tdp));
			}

			td->n_pipelines--;

			// Pipeline
			p->enabled = 0;

			break;
		}

		return 0;
	}

	// Allocate request
	req = thread_msg_alloc();
	if (req == NULL)
		return -1;

	// Write request
	req->type = THREAD_REQ_PIPELINE_DISABLE;
	req->pipeline_disable.p = p->p;

	// Send request and wait for response
	rsp = thread_msg_send_recv(thread_id, req);

	// Read response
	status = rsp->status;

	// Free response
	thread_msg_free(rsp);

	// Request completion
	if (status)
		return status;

	p->enabled = 0;

	return 0;
}

//
// Data plane threads: message handling
//
static inline struct thread_msg_req * thread_msg_recv(struct rte_ring *msgq_req) {
	struct thread_msg_req *req;

	int status = rte_ring_sc_dequeue(msgq_req, (void **) &req);
	if (status != 0)
		return NULL;

	return req;
}

static inline void thread_msg_send(struct rte_ring *msgq_rsp,	struct thread_msg_rsp *rsp) {
	int status;

	do {
		status = rte_ring_sp_enqueue(msgq_rsp, rsp);
	} while (status == -ENOBUFS);
}

static struct thread_msg_rsp *
thread_msg_handle_pipeline_enable(struct thread_data *t,
	struct thread_msg_req *req)
{
	struct thread_msg_rsp *rsp = (struct thread_msg_rsp *) req;
	struct pipeline_data *p = &t->pipeline_data[t->n_pipelines];

	// Request
	if (t->n_pipelines >= THREAD_PIPELINES_MAX) {
		rsp->status = -1;
		return rsp;
	}

	t->p[t->n_pipelines] = req->pipeline_enable.p;

	p->p = req->pipeline_enable.p;
	p->timer_period = (rte_get_tsc_hz() *
		req->pipeline_enable.timer_period_ms) / 1000;
	p->time_next = rte_get_tsc_cycles() + p->timer_period;

	t->n_pipelines++;

	// Response
	rsp->status = 0;
	return rsp;
}

static struct thread_msg_rsp * thread_msg_handle_pipeline_disable(struct thread_data *t, struct thread_msg_req *req) {
	struct thread_msg_rsp *rsp = (struct thread_msg_rsp *) req;
	uint32_t n_pipelines = t->n_pipelines;
	struct rte_swx_pipeline *pipeline = req->pipeline_disable.p;
	uint32_t i;

	// find pipeline
	for (i = 0; i < n_pipelines; i++) {
		struct pipeline_data *p = &t->pipeline_data[i];

		if (p->p != pipeline)
			continue;

		if (i < n_pipelines - 1) {
			struct rte_swx_pipeline *pipeline_last =
				t->p[n_pipelines - 1];
			struct pipeline_data *p_last =
				&t->pipeline_data[n_pipelines - 1];

			t->p[i] = pipeline_last;
			memcpy(p, p_last, sizeof(*p));
		}

		t->n_pipelines--;

		rsp->status = 0;
		return rsp;
	}

	// should not get here
	rsp->status = 0;
	return rsp;
}

static void thread_msg_handle(struct thread_data *t) {
	for ( ; ; ) {
		struct thread_msg_req *req;
		struct thread_msg_rsp *rsp;

		req = thread_msg_recv(t->msgq_req);
		if (req == NULL)
			break;

		printf("Request: %d\n", req->type);

		switch (req->type) {
		case THREAD_REQ_PIPELINE_ENABLE:
			rsp = thread_msg_handle_pipeline_enable(t, req);
			printf("Pipeline Enable: %d\n", 1);
			break;

		case THREAD_REQ_PIPELINE_DISABLE:
			rsp = thread_msg_handle_pipeline_disable(t, req);
			break;

		default:
			rsp = (struct thread_msg_rsp *) req;
			rsp->status = -1;
		}

		thread_msg_send(t->msgq_rsp, rsp);
	}
}

//
// Data plane threads: main
//
int thread_main(void *arg __rte_unused) {
	struct thread_data *t;
	uint32_t thread_id, i;

	thread_id = rte_lcore_id();
	t = &thread_data[thread_id];

	// Dispatch loop
	for (i = 0; ; i++) {
		uint32_t j;

		// Data Plane
		for (j = 0; j < t->n_pipelines; j++)
			rte_swx_pipeline_run(t->p[j], PIPELINE_INSTR_QUANTA);

		// Control Plane
		if ((i & 0xF) == 0) {
			uint64_t time = rte_get_tsc_cycles();
			uint64_t time_next_min = UINT64_MAX;

			if (time < t->time_next_min)
				continue;

			// Thread message queues
			{
				uint64_t time_next = t->time_next;

				if (time_next <= time) {
					thread_msg_handle(t);
					time_next = time + t->timer_period;
					t->time_next = time_next;
				}

				if (time_next < time_next_min)
					time_next_min = time_next;
			}

			t->time_next_min = time_next_min;
		}
	}

	return 0;
}

int main_thread_init() {
	int status = 0;

	status = rte_eal_mp_remote_launch(
		thread_main,
		NULL,
		SKIP_MAIN);

	return status;
}

//
// Runtime supporting functions
//
//

int pipeline_build(char *pipename, char *specfname) {
	struct pipeline *p = NULL;
	FILE *spec = NULL;
	uint32_t err_line;
	const char *err_msg;
	int status;

	p = pipeline_find(pipename);
	if (!p || p->ctl) {
		return 1;
	}

	spec = fopen(specfname, "r");
	if (!spec) {
		return 2;
	}

	status = rte_swx_pipeline_build_from_spec(p->p,
		spec,
		&err_line,
		&err_msg);
	fclose(spec);
	if (status) {
		printf("Err build from spec:%s line: %d\n", err_msg, err_line);
		return status;
	}

	p->ctl = rte_swx_ctl_pipeline_create(p->p);
	if (!p->ctl) {
		rte_swx_pipeline_free(p->p);
		return 4;
	}

	return 0;
}

int pipeline_commit(char *pipename) {
	struct pipeline *p;
	char *pipeline_name;
	int status;

	p = pipeline_find(pipename);
	if (!p || !p->ctl) {
		return 1;
	}

	status = rte_swx_ctl_pipeline_commit(p->ctl, 1);
	if (status)
		return 2;
}

int pipeline_in_addtap(char *pipelinename, uint32_t port_id, char *tapname, char *mempoolname, uint32_t mtu, uint32_t bsz) {
	struct pipeline *p;
	int status;
	struct rte_swx_port_fd_reader_params params;
	struct tap *tap;
	struct mempool *mp;

	p = pipeline_find(pipelinename);
	if (!p || p->ctl) {
		return 1;
	}

	tap = tap_find(tapname);
	if (!tap) {
		return 2;
	}

	mp = mempool_find(mempoolname);
	if (!mp) {
		return 3;
	}

	params.fd = tap->fd;
	params.mempool = mp->m;
	params.mtu = mtu;
	params.burst_size = bsz;

	status = rte_swx_pipeline_port_in_config(p->p, port_id, "fd", &params);

	return status;
}

int pipeline_out_addtap(char *pipelinename, uint32_t port_id, char *tapname, uint32_t bsz) {
	struct pipeline *p;
	int status;
	struct rte_swx_port_fd_writer_params params;
	struct tap *tap;

	p = pipeline_find(pipelinename);
	if (!p || p->ctl) {
		return 1;
	}

	tap = tap_find(tapname);
	if (!tap) {
		return 2;
	}

	params.fd = tap->fd;
	params.burst_size = bsz;

	status = rte_swx_pipeline_port_out_config(p->p, port_id, "fd", &params);

	return status;
}

*/
import "C"

import (
	"unsafe"

	"github.com/yerden/go-dpdk/common"
)

func err(n ...interface{}) error {
	if len(n) == 0 {
		return common.RteErrno()
	}

	return common.IntToErr(n[0])
}

//
// Infra Init functions
//

func ObjInit() error {
	return err(C.obj_init())
}

func ThreadInit() error {
	return err(C.thread_init())
}

func MainThreadInit() error {
	return err(C.main_thread_init())
}

//
// Runtime functions
//

// Mempool functions
type MempoolParams C.struct_mempool_params

// set the values of the MempoolParams struct
func (mpp *MempoolParams) Set(bufferSize uint32, poolSize uint32, cacheSize uint32, cpuId uint32) {
	mpp.buffer_size = C.uint32_t(bufferSize)
	mpp.pool_size = C.uint32_t(poolSize)
	mpp.cache_size = C.uint32_t(cacheSize)
	mpp.cpu_id = C.uint32_t(cpuId)
}

// create the MempoolParams structure
func MemPoolCreate(name string, params *MempoolParams) (int, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	res := C.mempool_create(cname, (*C.struct_mempool_params)(params))
	return common.IntOrErr(res)
}

// Pipeline functions
func PipelineCreate(name string, numaNode int) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	C.pipeline_create(cname, (C.int)(numaNode))
}

func PipelineBuild(name string, specfile string) (int, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	cspecfile := C.CString(specfile)
	defer C.free(unsafe.Pointer(cspecfile))

	res := C.pipeline_build(cname, cspecfile)
	return common.IntOrErr(res)
}

func PipelineCommit(name string) (int, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	res := C.pipeline_commit(cname)
	return common.IntOrErr(res)
}

func PipelineEnable(threadid int, pipename string) (int, error) {
	pname := C.CString(pipename)
	defer C.free(unsafe.Pointer(pname))

	res := C.thread_pipeline_enable(C.uint32_t(threadid), pname)
	return common.IntOrErr(res)
}

func PipelineDisable(threadid int, pipename string) (int, error) {
	pname := C.CString(pipename)
	defer C.free(unsafe.Pointer(pname))

	res := C.thread_pipeline_disable(C.uint32_t(threadid), pname)
	return common.IntOrErr(res)
}

// pipeline PIPELINE0 port in 0 tap sw0 mempool MEMPOOL0 mtu 1500 bsz 1
func PipelineAddInputPortTap(pipelineName string, portId int, tapName string, mempool string, mtu int, bsz int) (int, error) {
	pname := C.CString(pipelineName)
	defer C.free(unsafe.Pointer(pname))
	tname := C.CString(tapName)
	defer C.free(unsafe.Pointer(tname))
	mname := C.CString(mempool)
	defer C.free(unsafe.Pointer(mname))

	res := C.pipeline_in_addtap(pname, C.uint32_t(portId), tname, mname, C.uint32_t(mtu), C.uint32_t(bsz))
	return common.IntOrErr(res)
}

// pipeline PIPELINE0 port out 0 tap sw0 bsz 1
func PipelineAddOutputPortTap(pipelineName string, portId int, tapName string, bsz int) (int, error) {
	pname := C.CString(pipelineName)
	defer C.free(unsafe.Pointer(pname))
	tname := C.CString(tapName)
	defer C.free(unsafe.Pointer(tname))

	res := C.pipeline_out_addtap(pname, C.uint32_t(portId), tname, C.uint32_t(bsz))
	return common.IntOrErr(res)
}

// TAP functions
func TapCreate(name string) (int, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	res := C.tap_create(cname)
	return common.IntOrErr(res)
}
