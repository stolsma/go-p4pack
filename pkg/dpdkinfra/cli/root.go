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

// GetCommand returns the root command after adding the dpdkinfra service commands
func GetCommand(root *cobra.Command) *cobra.Command {
	log.Info("Adding dpdkinfra cli")

	// add all supported root commands
	initPipeline(root)
	initPktmbuf(root)
	initInterface(root)
	return root
}
