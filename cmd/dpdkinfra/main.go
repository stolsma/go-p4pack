// SPDX-FileCopyrightText: 2022-present Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"strings"
	"time"

	"github.com/stolsma/go-p4pack/pkg/config"
	"github.com/stolsma/go-p4pack/pkg/dpdkinfra"
	"github.com/stolsma/go-p4pack/pkg/flowtest"
	"github.com/stolsma/go-p4pack/pkg/logging"
	"github.com/stolsma/go-p4pack/pkg/signals"
)

var log logging.Logger

func init() {
	// keep the logger up to date, also after new log config
	logging.Register("main", func(logger logging.Logger) {
		log = logger
	})
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
	conf, err := config.CreateAndLoad(configFile)
	if err != nil {
		log.Fatalf("Configuration load failed: ", err)
	}

	// configure logging with our app requirements
	logging.Configure(conf.Logging)

	// initialize the dpdkinfra singleton
	dpdki, err := dpdkinfra.CreateAndInit(strings.Split(dpdkArgs, " "))
	if err != nil {
		log.Fatalf("DPDKInfraInit failed:", err)
	}

	// create Packet Mbuf mempools
	for _, m := range conf.PktMbufs {
		dpdki.CreatePktmbufWithConfig(m)
	}

	// create dpdkinfra interfaces through the API
	for _, i := range conf.Interfaces {
		dpdki.CreateInterfaceWithConfig(i)
	}

	// create and start dpdkinfra pipelines through the API
	for _, p := range conf.Pipelines {
		p.SetBasePath(conf.GetBasePath())
		dpdki.CreatePipelineWithConfig(p)
	}

	// define and start flowtests if requested
	tests, err := flowtest.Init(appCtx, conf.FlowTest)
	if err != nil {
		log.Fatalf("Tests initialization failed:", err)
	}
	if conf.FlowTest.GetStart() {
		err = tests.StartAll()
		if err != nil {
			log.Fatalf("Starting predefined flowtests failed:", err)
		}
	}

	// start ssh shell cli server with our CLI
	startSSHShell(appCtx, createShellRoot)

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
