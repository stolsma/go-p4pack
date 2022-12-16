// SPDX-FileCopyrightText: 2022-present Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"strings"
	"time"

	"github.com/stolsma/go-p4pack/pkg/config"
	"github.com/stolsma/go-p4pack/pkg/dpdkinfra"
	dpdkiConfig "github.com/stolsma/go-p4pack/pkg/dpdkinfra/config"
	"github.com/stolsma/go-p4pack/pkg/flowtest"
	"github.com/stolsma/go-p4pack/pkg/logging"
	"github.com/stolsma/go-p4pack/pkg/signals"
	shell "github.com/stolsma/go-p4pack/pkg/sshshell"
)

var log logging.Logger

func init() {
	// keep the logger up to date, also after new log config
	logging.Register("main", func(logger logging.Logger) {
		log = logger
	})
}

type Config struct {
	*dpdkiConfig.Config `json:"chassis"`
	FlowTest            *flowtest.Config `json:"flowtest"`
	Logging             *logging.Config  `json:"logging"`
	SSHShell            *shell.Config    `json:"sshshell"`
}

func main() {
	// the context for the app with cancel function
	appCtx, cancelAppCtx := context.WithCancel(context.Background())

	// get application arguments
	cmd := CreateCmd()
	cmd.Execute()
	dpdkArgs, _ := cmd.Flags().GetString("dpdkargs")
	configFile, _ := cmd.Flags().GetString("config")

	// get configuration
	conf := &Config{Config: dpdkiConfig.Create()}
	err := config.LoadConfig(configFile, conf)
	if err != nil {
		log.Fatalf("Configuration load failed: ", err)
	}

	// configure logging with our app requirements
	if conf.Logging != nil {
		err = conf.Logging.Apply()
		if err != nil {
			log.Fatalf("Applying logging config failed:", err)
		}
	}

	// initialize the dpdkinfra singleton
	dpdki, err := dpdkinfra.CreateAndInit(strings.Split(dpdkArgs, " "))
	if err != nil {
		log.Fatalf("DPDKInfraInit failed:", err)
	}

	// Apply given dpdkinfra configuration
	if conf.Config != nil {
		err = conf.Config.Apply()
		if err != nil {
			log.Fatalf("Applying chassis config failed:", err)
		}
	}

	// initialize the flowtest singleton
	_, err = flowtest.CreateAndInit(appCtx)
	if err != nil {
		log.Fatalf("Flow tests initialization failed:", err)
	}

	// apply given flowtest configuration
	if conf.FlowTest != nil {
		err = conf.FlowTest.Apply()
		if err != nil {
			log.Fatalf("Applying predefined flowtests failed:", err)
		}
	}

	// start ssh shell cli server with our CLI
	if conf.SSHShell != nil {
		startSSHShell(appCtx, createShellRoot, conf.SSHShell)
	}

	// initialize wait for signals to react on during packet processing
	log.Info("p4vswitch pipeline and SSH CLI server running!")
	stopCh := signals.RegisterSignalHandlers()

	// wait for stop signal CTRL-C or forced termination
	<-stopCh

	// Cancel the App context to let all running SSH sessions and Flow tests close in a neat way
	cancelAppCtx()
	println()
	log.Info("p4vswitch pipeline requested to stop!")

	// cleanup DpdkInfra environment
	dpdki.Cleanup()

	// Wait a small time to let everything shutdown...
	time.Sleep(1000 * time.Millisecond)

	// All is handled...
	log.Info("p4vswitch stopped!")
}
