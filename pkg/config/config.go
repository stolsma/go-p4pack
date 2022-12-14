// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Type interface {
	SetBasePath(bp string)
	GetBasePath() string
}

type Base struct {
	basePath string
}

func (b *Base) SetBasePath(bp string) {
	b.basePath = bp
}

func (b *Base) GetBasePath() string {
	return b.basePath
}

func LoadConfig(filename string, c Type) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error when opening file: %s", err)
	}

	if err := json.Unmarshal(data, c); err != nil {
		return fmt.Errorf("JSON unmarshaling failed: %s", err)
	}

	// save the basepath of the config file read
	c.SetBasePath(filepath.Dir(filename))

	return nil
}
