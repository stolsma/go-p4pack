// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package flowtest

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

type Config struct {
	Start      bool              `json:"start"`
	Interfaces []InterfaceConfig `json:"interfaces"`
	FlowSets   []FlowSetConfig   `json:"flowsets"`
}

// Process everything in this config structure
func (c *Config) Apply() error {
	// get flowtest singleton and check if it is initialized
	t := Get()
	if t == nil {
		return errors.New("flowtest module is not initialized")
	}

	// add the interface references
	for _, intf := range c.Interfaces {
		err := t.AddInterface(intf.GetName(), intf.GetMAC(), intf.GetIP())
		if err != nil {
			return err
		}
	}

	// add the flowsets
	for _, test := range c.FlowSets {
		err := t.AddFlowSet(test)
		if err != nil {
			return err
		}
	}

	// start the flowsets right away when requested
	if c.GetStart() {
		err := t.StartAll()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) GetStart() bool {
	return c.Start
}

type InterfaceConfig struct {
	Name string   `json:"name"`
	MAC  HexArray `json:"mac"`
	IP   HexArray `json:"ip"`
}

func (i *InterfaceConfig) GetName() string {
	return i.Name
}

func (i *InterfaceConfig) GetMAC() HexArray {
	return i.MAC
}

func (i *InterfaceConfig) GetIP() HexArray {
	return i.IP
}

type FlowSetConfig struct {
	Name  string       `json:"name"`
	Flows []FlowConfig `json:"flows"`
}

func (t *FlowSetConfig) GetName() string {
	return t.Name
}

type FlowConfig struct {
	Source      EndpointConfig `json:"source"`
	Destination EndpointConfig `json:"destination"`
	Send        Packet         `json:"send"`
	Receive     Packet         `json:"receive"`
	Interval    int            `json:"interval"`
}

type EndpointConfig struct {
	Interface string `json:"interface"`
}

type Packet struct {
	Layout []string            `json:"layout"`
	Fields map[string]HexArray `json:"fields"`
}

// parses the given fields to an array of bytes by using the given fields and extra parameters
func (p Packet) ToByteArray(param map[string]HexArray) ([]byte, error) {
	var result []byte

	for _, fname := range p.Layout {
		field, ok := p.Fields[fname]
		if !ok {
			field, ok = param[fname]
			if !ok {
				return nil, fmt.Errorf("field '%s' not found when creating packet", fname)
			}
		}
		result = append(result, []byte(field)...)
	}

	return result, nil
}

// Array of value strings (integer or hex 0x00 type) in json on JSON unmarshal decoded to []byte
type HexArray []byte

func (ha *HexArray) UnmarshalJSON(data []byte) (err error) {
	var hexString []string

	if err = json.Unmarshal(data, &hexString); err != nil {
		return
	}

	for _, val := range hexString {
		j, _ := decode(val)
		*ha = append(*ha, j...)
	}

	return
}

// decodes a hex string with 0x prefix or plain number if no 0x prefix, into a byte array
func decode(input string) ([]byte, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("empty hex string")
	}

	if !has0xPrefix(input) {
		// convert text number to integer
		i, err := strconv.Atoi(input)
		// return nil, fmt.Errorf("hex string without 0x prefix")
		return []byte{byte(i)}, err
	}

	// create a string with odd len
	cleanInput := input[2:]
	if len(cleanInput)%2 != 0 {
		cleanInput = "0" + cleanInput
	}

	b, err := hex.DecodeString(cleanInput)
	if err != nil {
		return nil, mapError(err)
	}
	return b, err
}

// check on 0x as string prefix
func has0xPrefix(input string) bool {
	return len(input) >= 2 && input[0] == '0' && (input[1] == 'x' || input[1] == 'X')
}

// convert given error to readable error
func mapError(err error) error {
	if err, ok := err.(*strconv.NumError); ok {
		switch err.Err {
		case strconv.ErrRange:
			return fmt.Errorf("hex number > 64 bits")
		case strconv.ErrSyntax:
			return fmt.Errorf("invalid hex string")
		}
	}
	if _, ok := err.(hex.InvalidByteError); ok {
		return fmt.Errorf("invalid hex string")
	}
	if err == hex.ErrLength {
		return fmt.Errorf("hex string of odd length")
	}
	return err
}
