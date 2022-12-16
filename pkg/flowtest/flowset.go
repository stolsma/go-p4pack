// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package flowtest

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/mdlayher/ethernet"
	"github.com/mdlayher/packet"
	"golang.org/x/sys/unix"
)

type ExpectedFlows ExpectedList

type Flow struct {
	sourceName       string
	sourceIP         HexArray
	sourceMAC        HexArray
	destName         string
	destIP           HexArray
	destMAC          HexArray
	send             Packet
	receive          Packet
	interval         int
	destExpectedList *ExpectedFlows
}

type Flows []Flow

func (f *Flows) AddFlow(flow Flow) {
	*f = append(*f, flow)
}

type Intf struct {
	name string
	intf *net.Interface
	mtu  int
	conn *packet.Conn
}

// Open a raw socket if not already open on this interface
func (i *Intf) Open() {
	if i.conn != nil {
		return
	}

	c, err := packet.Listen(i.intf, packet.Raw, unix.ETH_P_ALL, nil)
	if err != nil {
		log.Errorf("failed to open RAW socket: %v", err)
		return
	}
	c.SetPromiscuous(true)

	i.conn = c
	log.Infof("RAW socket opened on interface: %s", i.name)
}

// close the raw socket if open on this interface
func (i *Intf) Close() {
	if i.conn == nil {
		return
	}
	i.conn.Close()
	i.conn = nil
	log.Infof("RAW socket closed on interface: %s", i.name)
}

type FlowSet struct {
	name         string
	interfaces   map[string]*Intf
	flowsSend    map[string]*Flows
	flowsReceive map[string]*ExpectedFlows
	ctx          context.Context
	cancelFn     context.CancelFunc
	running      bool
}

func FlowSetCreate(name string) *FlowSet {
	return &FlowSet{
		name:         name,
		interfaces:   make(map[string]*Intf),
		flowsSend:    make(map[string]*Flows),
		flowsReceive: make(map[string]*ExpectedFlows),
	}
}

func (fs *FlowSet) Init(config []FlowConfig, ifaces IfaceMap) error {
	// configure the requested flows
	for _, flow := range config {
		err := fs.AddFlow(flow, ifaces)
		if err != nil {
			return err
		}
	}

	return nil
}

func (fs *FlowSet) Start(ctx context.Context) error {
	if fs.running {
		return fmt.Errorf("flowset already running")
	}

	ctx, cancelFn := context.WithCancel(ctx)
	fs.ctx = ctx
	fs.cancelFn = cancelFn

	// Start receiveFlows GoThreads process
	for name := range fs.flowsReceive { // expectedFlows
		iface := fs.interfaces[name]
		iface.Open()
		go receiveFlows(fs.ctx, iface) // expectedFlows
	}

	// startup sendFlows GoThreads
	for name, flows := range fs.flowsSend {
		iface := fs.interfaces[name]
		iface.Open()
		go sendFlows(fs.ctx, iface, flows)
	}

	fs.running = true

	return nil
}

func receiveFlows(ctx context.Context, intf *Intf) {
	// Accept frames up to interface's MTU in size.
	b := make([]byte, intf.mtu)
	var f ethernet.Frame

	// Keep reading frames.
	for {
		select {
		case <-ctx.Done(): // Quit receive goroutine
			log.Infof("Stopped checking interface %s", intf.name)
			return
		default:
			n, addr, err := intf.conn.ReadFrom(b)
			if err != nil {
				log.Errorf("failed to receive message: %v", err)
				break
			}

			// Unpack Ethernet frame into Go representation.
			if err := (&f).UnmarshalBinary(b[:n]); err != nil {
				log.Errorf("failed to unmarshal ethernet frame: %v", err)
				break
			}

			// TODO: Handle timecheck on received frame
			// Display source of message and message itself.
			// s := net.HardwareAddr(f.Payload[:6])
			log.Infof("Packet received [%s %s] [%s %s]", intf.name, addr.String(), f.Source.String(), f.Destination.String())
		}
	}
}

func sendFlows(ctx context.Context, intf *Intf, flows *Flows) {
	for {
		select {
		case <-ctx.Done(): // Quit send goroutine
			log.Infof("Stopped sending flows on interface %s", intf.name)
			return
		default:
			for _, f := range *flows {
				// TODO: add generic paramaters to packet byte creation from Interface definition
				var param = map[string]HexArray{}
				frame, err := f.send.ToByteArray(param)
				if err != nil {
					log.Fatalf("failed to create frame: %v", err)
					return
				}

				// Write the frame to the source network interface
				addr := &packet.Addr{HardwareAddr: net.HardwareAddr(f.destMAC)}
				if _, err := intf.conn.WriteTo(frame, addr); err != nil {
					log.Errorf("failed to write frame: %v", err)
					return
				}

				log.Infof("Packet send     [%s %s]", intf.name, addr.String())
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func (fs *FlowSet) Stop() error {
	if !fs.running {
		return fmt.Errorf("flowset already stopped")
	}
	// stop SendFlow GoThreads
	fs.cancelFn()

	// stop ReceiveCheck GoThreads
	// TODO: Implement

	fs.ctx = nil
	fs.cancelFn = nil
	fs.running = false

	return nil
}

func (fs *FlowSet) Delete() error {
	// Kill all Sendflow & ReceiveCheck GoThreads
	// TODO: Implement

	return nil
}

func (fs *FlowSet) AddFlow(config FlowConfig, ifaces IfaceMap) error {
	newFlow := Flow{
		sourceName: config.Source.Interface,
		destName:   config.Destination.Interface,
		send:       config.Send,
		receive:    config.Receive,
		interval:   config.Interval,
	}

	// get source and destination interface configs
	if _, ok := fs.interfaces[newFlow.sourceName]; !ok {
		ifs, err := fs.getInterface(newFlow.sourceName)
		if err != nil {
			return err
		}
		fs.interfaces[newFlow.sourceName] = &Intf{
			name: newFlow.sourceName,
			intf: ifs,
			mtu:  ifs.MTU,
		}
	}

	if _, ok := fs.interfaces[newFlow.destName]; !ok {
		ifs, err := fs.getInterface(newFlow.destName)
		if err != nil {
			return err
		}
		fs.interfaces[newFlow.destName] = &Intf{
			name: newFlow.destName,
			intf: ifs,
			mtu:  ifs.MTU,
		}
	}

	// get source variables
	ifConfig := ifaces[newFlow.sourceName]
	if ifConfig != nil {
		newFlow.sourceIP = ifConfig.GetIP()
		newFlow.sourceMAC = ifConfig.GetMAC()
	}
	if _, ok := fs.flowsSend[newFlow.sourceName]; !ok {
		fs.flowsSend[newFlow.sourceName] = &Flows{}
	}

	// get destination variables
	ifConfig = ifaces[newFlow.destName]
	if ifConfig != nil {
		newFlow.destIP = ifConfig.GetIP()
		newFlow.destMAC = ifConfig.GetMAC()
	}
	if fs.flowsReceive[newFlow.destName] == nil {
		fs.flowsReceive[newFlow.destName] = &ExpectedFlows{}
	}
	newFlow.destExpectedList = fs.flowsReceive[newFlow.destName]

	// add to source flow send list
	fs.flowsSend[newFlow.sourceName].AddFlow(newFlow)

	return nil
}

func (fs *FlowSet) getInterface(name string) (*net.Interface, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return nil, fmt.Errorf("failed to open interface: %v", err)
	}
	return iface, nil
}
