// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"github.com/spf13/cobra"
	"github.com/stolsma/go-p4pack/pkg/logging"
)

var log logging.Logger

func init() {
	// keep the logger up to date, also after new log config
	logging.Register("pcidevices/cli", func(logger logging.Logger) {
		log = logger
	})
}

// GetCommand returns the given parent (root) command with all pcidevices sub commands added
func GetCommand(parent *cobra.Command) *cobra.Command {
	log.Info("Adding pcidevices cli commands")

	// add all dpdkinfra cli commands
	PciCmd(parent)

	return parent
}
