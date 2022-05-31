package main

// CGO_LDFLAGS=`pkg-config --libs libdpdk` CGO_CFLAGS=`pkg-config --cflags libdpdk`

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/stolsma/go-p4dpdk-vswitch/pkg/dpdkinfra"
)

func waitForSignal() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)

	go func() {
		<-sigs
		done <- true
	}()

	<-done
}

//"-n", "4",

func main() {
	dpdkArgs := []string{"dummy", "-c", "3", "--log-level", ".*,8"}

	// os.Args
	if err := dpdkinfra.DpdkInfraInit(dpdkArgs); err != nil {
		log.Fatalln("DPDKInfraInit failed:", err)
	}

	// Create mempool
	// mempool MEMPOOL0 buffer 2304 pool 32K cache 256 cpu 0
	var params dpdkinfra.MempoolParams
	params.Set(2304, 32*1024, 256, 0)
	num, err := dpdkinfra.MemPoolCreate("MEMPOOL0", &params)
	if num != 0 {
		log.Fatalln("Mempool num:", num)
	}
	if err != nil {
		log.Fatalln("Mempool create err:", err)
	}
	log.Println("Mempool ready!")

	// Create TAP ports
	// tap sw0
	_, err = dpdkinfra.TapCreate("sw0")
	if err != nil {
		log.Fatalln("TAP sw0 create err:", err)
	}
	log.Println("TAP sw0 created!")
	// tap sw1
	_, err = dpdkinfra.TapCreate("sw1")
	if err != nil {
		log.Fatalln("TAP sw1 create err:", err)
	}
	log.Println("TAP sw1 created!")
	// tap sw2
	_, err = dpdkinfra.TapCreate("sw2")
	if err != nil {
		log.Fatalln("TAP sw2 create err:", err)
	}
	log.Println("TAP sw2 created!")
	// tap sw3
	_, err = dpdkinfra.TapCreate("sw3")
	if err != nil {
		log.Fatalln("TAP sw3 create err:", err)
	}
	log.Println("TAP sw3 created!")

	// Create pipeline
	// pipeline PIPELINE0 create 0
	dpdkinfra.PipelineCreate("PIPELINE0", 0)

	// Add input ports to pipeline
	// pipeline PIPELINE0 port in 0 tap sw0 mempool MEMPOOL0 mtu 1500 bsz 1
	_, err = dpdkinfra.PipelineAddInputPortTap("PIPELINE0", 0, "sw0", "MEMPOOL0", 1500, 1)
	if num != 0 {
		log.Fatalln("AddInPort sw0 num:", num)
	}
	if err != nil {
		log.Fatalln("AddInPort sw0 err:", err)
	}
	log.Println("AddInPort sw0 ready!")
	// pipeline PIPELINE0 port in 1 tap sw1 mempool MEMPOOL0 mtu 1500 bsz 1
	_, err = dpdkinfra.PipelineAddInputPortTap("PIPELINE0", 1, "sw1", "MEMPOOL0", 1500, 1)
	if num != 0 {
		log.Fatalln("AddInPort sw1 num:", num)
	}
	if err != nil {
		log.Fatalln("AddInPort sw1 err:", err)
	}
	log.Println("AddInPort sw1 ready!")
	// pipeline PIPELINE0 port in 2 tap sw2 mempool MEMPOOL0 mtu 1500 bsz 1
	_, err = dpdkinfra.PipelineAddInputPortTap("PIPELINE0", 2, "sw2", "MEMPOOL0", 1500, 1)
	if num != 0 {
		log.Fatalln("AddInPort sw2 num:", num)
	}
	if err != nil {
		log.Fatalln("AddInPort sw2 err:", err)
	}
	log.Println("AddInPort sw2 ready!")
	// pipeline PIPELINE0 port in 3 tap sw3 mempool MEMPOOL0 mtu 1500 bsz 1
	_, err = dpdkinfra.PipelineAddInputPortTap("PIPELINE0", 3, "sw3", "MEMPOOL0", 1500, 1)
	if num != 0 {
		log.Fatalln("AddInPort sw3 num:", num)
	}
	if err != nil {
		log.Fatalln("AddInPort sw3 err:", err)
	}
	log.Println("AddInPort sw3 ready!")

	// Add output ports to pipeline
	// pipeline PIPELINE0 port out 0 tap sw0 bsz 1
	_, err = dpdkinfra.PipelineAddOutputPortTap("PIPELINE0", 0, "sw0", 1)
	if num != 0 {
		log.Fatalln("AddOutPort sw0 num:", num)
	}
	if err != nil {
		log.Fatalln("AddOutPort sw0 err:", err)
	}
	log.Println("AddOutPort sw0 ready!")
	// pipeline PIPELINE0 port out 1 tap sw1 bsz 1
	_, err = dpdkinfra.PipelineAddOutputPortTap("PIPELINE0", 1, "sw1", 1)
	if num != 0 {
		log.Fatalln("AddOutPort sw1 num:", num)
	}
	if err != nil {
		log.Fatalln("AddOutPort sw1 err:", err)
	}
	log.Println("AddOutPort sw1 ready!")
	// pipeline PIPELINE0 port out 2 tap sw2 bsz 1
	_, err = dpdkinfra.PipelineAddOutputPortTap("PIPELINE0", 2, "sw2", 1)
	if num != 0 {
		log.Fatalln("AddOutPort sw2 num:", num)
	}
	if err != nil {
		log.Fatalln("AddOutPort sw2 err:", err)
	}
	log.Println("AddOutPort sw2 ready!")
	// pipeline PIPELINE0 port out 3 tap sw3 bsz 1
	_, err = dpdkinfra.PipelineAddOutputPortTap("PIPELINE0", 3, "sw3", 1)
	if num != 0 {
		log.Fatalln("AddOutPort sw3 num:", num)
	}
	if err != nil {
		log.Fatalln("AddOutPort sw3 err:", err)
	}
	log.Println("AddOutPort sw3 ready!")

	// Build the pipeline program
	// pipeline PIPELINE0 build ./examples/ipdk-simple_l3/simple_l3.spec
	num, err = dpdkinfra.PipelineBuild("PIPELINE0", "../../examples/ipdk-simple_l3/simple_l3.spec")
	if num != 0 {
		log.Fatalln("Pipelinebuild num:", num)
	}
	if err != nil {
		log.Fatalln("Pipelinebuild err:", err)
	}
	log.Println("Pipeline Build!")

	// Commit program to pipeline
	// pipeline PIPELINE0 commit
	num, err = dpdkinfra.PipelineCommit("PIPELINE0")
	if num != 0 {
		log.Fatalln("Pipelinecommit num:", num)
	}
	if err != nil {
		log.Fatalln("Pipelinecommit err:", err)
	}
	log.Println("Pipeline Commited!")

	// And run pipeline
	// thread 1 pipeline PIPELINE0 enable
	num, err = dpdkinfra.PipelineEnable(1, "PIPELINE0")
	if num != 0 {
		log.Fatalln("PipelineEnable num:", num)
	}
	if err != nil {
		log.Fatalln("PipelineEnable err:", err)
	}
	log.Println("Pipeline Enabled!")

	// wait for signals to react on during packet processing
	waitForSignal()
	log.Println("p4vswitch pipeline requested to stop!")

	// thread 1 pipeline PIPELINE0 disable
	num, err = dpdkinfra.PipelineDisable(1, "PIPELINE0")
	if num != 0 {
		log.Fatalln("PipelineDisable num:", num)
	}
	if err != nil {
		log.Fatalln("PipelineDisable err:", err)
	}
	log.Println("Pipeline Disabled!")

	// TODO: cleanup EAL memory etc...
	//err = dpdkinfra.EalCleanup()
	//if err != nil {
	//	log.Fatalln("EAL cleanup err:", err)
	//}
	//log.Println("EAL cleanup ready!")

	// All is handled...
	log.Println("p4vswitch stopped!")
}
