// Copyright 2018 Intel Corporation.
// Copyright 2023 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

// Package pcidevices helps to query DPDK compatibles devices and to bind/unbind drivers
package pcidevices

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

var defaultTimeoutLimitation = 5 * time.Second

// GetDeviceID returns the device ID of given NIC name.
func getDeviceID(nicName string) (string, error) {
	// DEV_ID=$(basename $(readlink /sys/class/net/<nicName>))
	raw, err := readlinkCmd(pathSysClassNetDevice.With(nicName))
	if err != nil {
		return "", err
	}

	// raw should be like /sys/devices/pci0002:00/0000:00:08.0/virtio2/net/ens8
	// or /sys/devices/pci0000:00/0000:00:01.0/0000:03:00.2/net/ens4f2
	raws := strings.Split(raw, "/")
	if len(raws) < 6 {
		return "", fmt.Errorf("path not correct")
	}

	switch {
	case IsPciID.Match([]byte(raws[5])):
		return raws[5], nil
	case IsPciID.Match([]byte(raws[4])):
		return raws[4], nil
	case len(raws) >= 11 && IsPciID.Match([]byte(raws[10])):
		return raws[10], nil
	default:
		return "", ErrNoDeviceID
	}
}

func writeToTargetWithData(sysfs string, flag int, mode os.FileMode, data string) error {
	writer, err := os.OpenFile(sysfs, flag, mode)
	if err != nil {
		return fmt.Errorf("OpenFile failed: %s", err.Error())
	}
	defer writer.Close()

	_, err = writer.Write([]byte(data))
	if err != nil {
		return fmt.Errorf("WriteFile failed: %s", err.Error())
	}

	return nil
}

func readlinkCmd(path string) (string, error) {
	output, err := cmdOutputWithTimeout(defaultTimeoutLimitation, "readlink", "-f", path)
	if err != nil {
		return "", fmt.Errorf("cmd execute readlink failed: %s", err.Error())
	}
	outputStr := strings.Trim(string(output), "\n")
	return outputStr, nil
}

func checkModuleLine(line, driver string) bool {
	if !strings.Contains(line, driver) {
		return false
	}
	elements := strings.Split(line, " ")
	return elements[0] == driver
}

func cmdOutputWithTimeout(duration time.Duration, cmd string, parameters ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	output, err := exec.CommandContext(ctx, cmd, parameters...).Output()
	if err != nil {
		return nil, err
	}
	return output, nil
}
