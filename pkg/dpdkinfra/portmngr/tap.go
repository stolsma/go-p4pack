// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package portmngr

/*
#include <stdlib.h>
#include <string.h>
#include <bsd/string.h>
#include <netinet/in.h>
#include <linux/if.h>
#include <linux/if_tun.h>
#include <sys/ioctl.h>
#include <fcntl.h>
#include <unistd.h>

#define TAP_DEV		"/dev/net/tun"

int tap_create(const char *name) {
	struct ifreq ifr;
	int fd, status;

	// Resource create
	fd = open(TAP_DEV, O_RDWR | O_NONBLOCK);
	if (fd < 0)
		return fd;

	memset(&ifr, 0, sizeof(ifr));
	ifr.ifr_flags = IFF_TAP | IFF_NO_PI; // No packet information
	strlcpy(ifr.ifr_name, name, IFNAMSIZ);

	status = ioctl(fd, TUNSETIFF, (void *) &ifr);
	if (status < 0) {
		close(fd);
		return status;
	}

  return fd;
}

*/
import "C"

import (
	"unsafe"
)

// Tap represents a Tap record stored in a tap store
type Tap struct {
	name  string
	fd    C.int
	clean func()
}

// Create Tap interface. Returns a pointer to a Tap structure or nil with error.
func (tap *Tap) Init(name string, clean func()) error {
	// create fd of tap interface
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	fd, err := C.tap_create(cname)
	if err != nil {
		return err
	}

	// Node fill in
	tap.name = name
	tap.fd = fd
	tap.clean = clean

	return nil
}

// Name returns the name of the Tap interface
func (tap *Tap) Name() string {
	return tap.name
}

// Fd returns the File descripter of the Tap interface
func (tap *Tap) Fd() C.int {
	return tap.fd
}

// Free deletes the current Tap record and calls the clean callback function given at init
func (tap *Tap) Free() {
	// TODO remove TAP interface from the system
	// call given clean callback function if given during init
	if tap.clean != nil {
		tap.clean()
	}
}
