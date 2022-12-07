// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/stolsma/go-p4pack/pkg/dpdkinfra"
	"github.com/stolsma/go-p4pack/pkg/flowtest"
	"github.com/stolsma/go-p4pack/pkg/logging"
)

type Config struct {
	BasePath   string
	PktMbufs   []*dpdkinfra.PktmbufConfig   `json:"pktmbufs"`
	Interfaces []*dpdkinfra.InterfaceConfig `json:"interfaces"`
	Pipelines  []*dpdkinfra.PipelineConfig  `json:"pipelines"`
	FlowTest   *flowtest.Config             `json:"flowtest"`
	Logging    *logging.Config              `json:"logging"`
}

func (c *Config) LoadConfig(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error when opening file: %s", err)
	}

	if err := json.Unmarshal(data, c); err != nil {
		return fmt.Errorf("JSON unmarshaling failed: %s", err)
	}

	// save the basepath of the config file read
	c.BasePath = filepath.Dir(filename)

	return nil
}

func (c *Config) GetBasePath() string {
	return c.BasePath
}

func CreateAndLoad(filepath string) (*Config, error) {
	config := Create()
	return config, config.LoadConfig(filepath)
}

func Create() *Config {
	return &Config{}
}
