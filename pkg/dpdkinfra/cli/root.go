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
	logging.Register("dpdkinfra/cli", func(logger logging.Logger) {
		log = logger
	})
}

// GetCommand returns the given parent (root) command with all dpdkinfra sub commands added
func GetCommand(parent *cobra.Command) *cobra.Command {
	log.Info("Adding dpdkinfra cli commands")

	// add all dpdkinfra cli commands
	pktmbufCmd(parent)
	interfaceCmd(parent)
	pipelineCmd(parent)

	return parent
}
