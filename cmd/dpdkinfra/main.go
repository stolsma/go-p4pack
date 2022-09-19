package main

// CGO_LDFLAGS=`pkg-config --libs libdpdk` CGO_CFLAGS=`pkg-config --cflags libdpdk`

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/stolsma/go-p4dpdk-vswitch/pkg/dpdkinfra"
	"github.com/stolsma/go-p4dpdk-vswitch/pkg/signals"
)

func printStatistics(dpdki *dpdkinfra.DpdkInfra, interval int, ctx context.Context, stopCh <-chan struct{}) {
	go func(interval int, ctx context.Context) {
		var prevLines int = 0
		for {
			timeout := time.Duration(interval) * time.Second
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			select {
			case <-ctx.Done():
				// print statistics
				stats, err := dpdki.PipelineStats("PIPELINE0")
				if err != nil {
					log.Println("Pipeline Stats err:", err)
					return
				}
				for i := 0; i <= prevLines; i++ {
					fmt.Printf("\033[A") // move the cursor up
				}
				println(stats)
				prevLines = strings.Count(stats, "\n")

			case <-stopCh:
				// stop printing
				return
			}
		}
	}(interval, ctx)
}

//"-n", "4",

func main() {
	dpdkArgs := []string{"dummy", "-c", "3", "--log-level", ".*,8"}
	taps := [...]string{"sw0", "sw1", "sw2", "sw3"}
	ctx := context.Background()

	// os.Args
	dpdki, err := dpdkinfra.Init(dpdkArgs)
	if err != nil {
		log.Fatalln("DPDKInfraInit failed:", err)
	}

	// Create mempool
	// mempool MEMPOOL0 buffer 2304 pool 32K cache 256 cpu 0
	err = dpdki.MempoolCreate("MEMPOOL0", 2304, 32*1024, 256, 0)
	if err != nil {
		log.Fatalln("Pktmbuf Mempool create err:", err)
	}
	log.Println("Pktmbuf Mempool ready!")

	// Create TAP ports
	for _, t := range taps {
		err = dpdki.TapCreate(t)
		if err != nil {
			log.Fatalf("TAP %s create err: %d", t, err)
		}
		log.Printf("TAP %s created!", t)
	}

	// Create pipeline
	// pipeline PIPELINE0 create 0
	err = dpdki.PipelineCreate("PIPELINE0", 0)
	if err != nil {
		log.Fatalln("PIPELINE0 create err:", err)
	}
	log.Println("PIPELINE0 created!")

	// Add input ports to pipeline
	// pipeline PIPELINE0 port in <portindex> tap <tapname> mempool MEMPOOL0 mtu 1500 bsz 1
	for i, t := range taps {
		err = dpdki.PipelineAddInputPortTap("PIPELINE0", i, t, "MEMPOOL0", 1500, 1)
		if err != nil {
			log.Fatalf("AddInPortTap %s err: %d", t, err)
		}
		log.Printf("AddInPortTap %s ready!", t)
	}

	// Add output ports to pipeline
	// pipeline PIPELINE0 port out 0 tap sw0 bsz 1
	for i, t := range taps {
		err = dpdki.PipelineAddOutputPortTap("PIPELINE0", i, t, 1)
		if err != nil {
			log.Fatalf("AddOutPortTap %s err: %d", t, err)
		}
		log.Printf("AddOutPortTap %s ready!", t)
	}

	// Build the pipeline program
	// pipeline PIPELINE0 build ./examples/ipdk-simple_l3/simple_l3.spec
	err = dpdki.PipelineBuild("PIPELINE0", "../../examples/ipdk-simple_l3/simple_l3.spec")
	if err != nil {
		log.Fatalln("Pipelinebuild err:", err)
	}
	log.Println("Pipeline Build!")

	// Commit program to pipeline
	// pipeline PIPELINE0 commit
	err = dpdki.PipelineCommit("PIPELINE0")
	if err != nil {
		log.Fatalln("Pipelinecommit err:", err)
	}
	log.Println("Pipeline Commited!")

	// And run pipeline
	// thread 1 pipeline PIPELINE0 enable
	err = dpdki.PipelineEnable("PIPELINE0", 1)
	if err != nil {
		log.Fatalln("PipelineEnable err:", err)
	}
	log.Println("Pipeline Enabled!")

	// initialize wait for signals to react on during packet processing
	log.Println("p4vswitch pipeline running!")
	stopCh := signals.RegisterSignalHandlers()

	// start prining statistics every second
	printStatistics(dpdki, 1, ctx, stopCh)

	// wait for stop signal CTRL-C or forced termination
	<-stopCh
	println()
	log.Println("p4vswitch pipeline requested to stop!")

	// cleanup DpdkInfra environment
	dpdki.Cleanup()

	// All is handled...
	log.Println("p4vswitch stopped!")
}
