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
)

type Config struct {
	Interfaces []dpdkinfra.InterfaceConfig
	Pipelines  []dpdkinfra.PipelineConfig
	FlowTest   flowtest.Config
}

func (c *Config) LoadConfig(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error when opening file: %s", err)
	}

	if err := json.Unmarshal(data, c); err != nil {
		return fmt.Errorf("JSON unmarshaling failed: %s", err)
	}

	path := filepath.Dir(filename)
	dpdkinfra.PathConfig(path, &c.Pipelines)

	return nil
}

func CreateAndLoad(filepath string) (*Config, error) {
	config := Create()
	return config, config.LoadConfig(filepath)
}

func Create() *Config {
	return &Config{}
}
