// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	"github.com/stolsma/go-p4pack/pkg/flowtest"
	"github.com/stolsma/go-p4pack/pkg/logging"
	"github.com/stretchr/testify/assert"
)

type empty struct{}

var filename = "../../examples/default/config.json"

type Config struct {
	*Base
	FlowTest *flowtest.Config `json:"flowtest"`
	Logging  *logging.Config  `json:"logging"`
}

func TestLoad(t *testing.T) {
	var c = Config{Base: &Base{}}
	err := LoadConfig(filename, &c)
	if err != nil {
		return
	}
	assert.Equal(t, "../../examples/default", c.GetBasePath())
	assert.NotEqual(t, empty{}, c)
}
