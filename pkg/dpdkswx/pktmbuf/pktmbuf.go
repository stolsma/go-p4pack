// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package pktmbuf

/*
#include <rte_mempool.h>
#include <rte_mbuf.h>

*/
import "C"

const RteMbufDefaultBufSize uint = C.RTE_MBUF_DEFAULT_BUF_SIZE
const RteMbufDefaultDataroom uint = C.RTE_MBUF_DEFAULT_DATAROOM
const RtePktmbufHeadroom uint = C.RTE_PKTMBUF_HEADROOM
const SizeofRteMbuf uint = C.sizeof_struct_rte_mbuf
