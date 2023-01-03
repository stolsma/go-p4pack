// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"github.com/stolsma/go-p4pack/pkg/config"
	"github.com/stolsma/go-p4pack/pkg/logging"
)

var log logging.Logger

func init() {
	// keep the logger up to date, also after new log config
	logging.Register("dpdkinfra/config", func(logger logging.Logger) {
		log = logger
	})
}

type Config struct {
	*config.Base
	Pktmbufs   PktmbufsConfig   `json:"pktmbufs"`
	Devices    DevicesConfig    `json:"devices"`
	Interfaces InterfacesConfig `json:"interfaces"`
	Pipelines  PipelinesConfig  `json:"pipelines"`
}

// Process everything in this config structure
func (c *Config) Apply() error {
	// Order of processing is important because Pipeline needs Interface and Interface needs Pktmbuf!!
	if err := c.Pktmbufs.Apply(); err != nil {
		return err
	}

	if err := c.Devices.Apply(); err != nil {
		return err
	}

	if err := c.Interfaces.Apply(); err != nil {
		return err
	}

	if err := c.Pipelines.Apply(c.GetBasePath()); err != nil {
		return err
	}

	return nil
}

func Create() *Config {
	return &Config{Base: &config.Base{}}
}
