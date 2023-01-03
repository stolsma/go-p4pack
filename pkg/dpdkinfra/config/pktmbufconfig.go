// SPDX-FileCopyrightText: 2022-present Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"

	"github.com/stolsma/go-p4pack/pkg/dpdkinfra"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/pktmbuf"
)

type PktmbufsConfig []*PktmbufConfig

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

// Create pktmbufs through the DpdkInfra API
func (c PktmbufsConfig) Apply() error {
	dpdki := dpdkinfra.Get()
	if dpdki == nil {
		return errors.New("dpdkinfra module is not initialized")
	}

	// Create PktMbuf memory pool
	for _, m := range c {
		name := m.GetName()
		_, err := dpdki.PktmbufCreate(name, m.GetBufferSize(), m.GetPoolSize(), m.GetCacheSize(), m.GetCPUID())
		if err != nil {
			return fmt.Errorf("pktmbuf %s create err: %d", name, err)
		}
		log.Infof("Pktmbuf Mempool %s ready!", name)
	}

	return nil
}
