// SPDX-FileCopyrightText: 2022-present Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkinfra

import "github.com/stolsma/go-p4pack/pkg/dpdkswx/pktmbuf"

type PktmbufConfig struct {
	Name       string `json:"name"`
	BufferSize uint   `json:"buffersize"`
	PoolSize   uint32 `json:"poolsize"`
	CacheSize  uint32 `json:"cachesize"`
	CPUID      int    `json:"cpuid"`
}

func (mpc *PktmbufConfig) GetName() string {
	return mpc.Name
}

func (mpc *PktmbufConfig) GetBufferSize() uint {
	if mpc.BufferSize == 0 {
		mpc.BufferSize = pktmbuf.RteMbufDefaultBufSize
	}
	return mpc.BufferSize
}

func (mpc *PktmbufConfig) GetPoolSize() uint32 {
	return mpc.PoolSize
}

func (mpc *PktmbufConfig) GetCacheSize() uint32 {
	return mpc.CacheSize
}

func (mpc *PktmbufConfig) GetCPUID() int {
	return mpc.CPUID
}

// Create pipelines through the DpdkInfra API
func (dpdki *DpdkInfra) CreatePktmbufWithConfig(m *PktmbufConfig) {
	// Create PktMbuf memory pool
	name := m.GetName()
	_, err := dpdki.PktmbufCreate(name, m.GetBufferSize(), m.GetPoolSize(), m.GetCacheSize(), m.GetCPUID())
	if err != nil {
		log.Fatalf("Pktmbuf Mempool %s create err: %d", name, err)
	}
	log.Infof("Pktmbuf Mempool %s ready!", name)
}
