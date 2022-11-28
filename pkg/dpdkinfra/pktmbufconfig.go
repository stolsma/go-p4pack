// SPDX-FileCopyrightText: 2022-present Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkinfra

import "github.com/stolsma/go-p4pack/pkg/dpdkswx/pktmbuf"

type PktMbufConfig struct {
	Name       string `json:"name"`
	BufferSize uint   `json:"buffersize"`
	PoolSize   uint32 `json:"poolsize"`
	CacheSize  uint32 `json:"cachesize"`
	CPUID      int    `json:"cpuid"`
}

func (mpc *PktMbufConfig) GetName() string {
	return mpc.Name
}

func (mpc *PktMbufConfig) GetBufferSize() uint {
	if mpc.BufferSize == 0 {
		mpc.BufferSize = pktmbuf.RteMbufDefaultBufSize
	}
	return mpc.BufferSize
}

func (mpc *PktMbufConfig) GetPoolSize() uint32 {
	return mpc.PoolSize
}

func (mpc *PktMbufConfig) GetCacheSize() uint32 {
	return mpc.CacheSize
}

func (mpc *PktMbufConfig) GetCPUID() int {
	return mpc.CPUID
}

// Create pipelines through the DpdkInfra API
func (dpdki *DpdkInfra) PktMbufWithConfig(m *PktMbufConfig) {
	// Create PktMbuf memory pool
	name := m.GetName()
	err := dpdki.PktMbufCreate(name, m.GetBufferSize(), m.GetPoolSize(), m.GetCacheSize(), m.GetCPUID())
	if err != nil {
		log.Fatalf("Pktmbuf Mempool %s create err: %d", name, err)
	}
	log.Infof("Pktmbuf Mempool %s ready!", name)
}
