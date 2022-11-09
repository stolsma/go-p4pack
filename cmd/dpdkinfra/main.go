// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"log"
	"strings"

	"github.com/stolsma/go-p4pack/pkg/config"
	"github.com/stolsma/go-p4pack/pkg/dpdkinfra"
	"github.com/stolsma/go-p4pack/pkg/flowtest"
	"github.com/stolsma/go-p4pack/pkg/signals"
)

func main() {
	// the context for the app with cancel function
	appCtx, cancelAppCtx := context.WithCancel(context.Background())

	// get application arguments
	cmd := CreateCmd()
	cmd.Execute()
	dpdkArgs, _ := cmd.Flags().GetString("dpdkargs")
	configFile, _ := cmd.Flags().GetString("config")

	// get configuration
	config, err := config.CreateAndLoad(configFile)
	if err != nil {
		log.Fatalln("Configuration load failed: ", err)
	}

	// os.Args
	dpdki, err := dpdkinfra.CreateAndInit(strings.Split(dpdkArgs, " "))
	if err != nil {
		log.Fatalln("DPDKInfraInit failed:", err)
	}

	// create interfaces through the Go API
	for _, i := range config.Interfaces {
		dpdki.InterfaceWithConfig(i)
	}

	// create and start pipelines through the Go API
	for _, p := range config.Pipelines {
		dpdki.PipelineWithConfig(p)
	}

	// Define and start flowtests if requested
	tests, err := flowtest.Init(appCtx, config.FlowTest)
	if err != nil {
		log.Fatalln("Tests initialization failed:", err)
	}
	err = tests.StartAll()
	if err != nil {
		log.Fatalln("Starting predefined flowtests failed:", err)
	}

	// start ssh cli server
	startSSH(appCtx, dpdki)

	// initialize wait for signals to react on during packet processing
	log.Println("p4vswitch pipeline and SSH CLI server running!")
	stopCh := signals.RegisterSignalHandlers()

	// wait for stop signal CTRL-C or forced termination
	<-stopCh

	// Cancel the App context to let all running SSH sessions and Flow tests close in a neat way
	cancelAppCtx()
	println()
	log.Println("p4vswitch pipeline requested to stop!")

	// cleanup DpdkInfra environment
	dpdki.Cleanup()

	// All is handled...
	log.Println("p4vswitch stopped!")
}
