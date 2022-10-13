// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package flowtest

import (
	"bytes"
	"testing"
)

func TestDecode(t *testing.T) {
	// check odd number of hex characters
	value, err := decode("0xA00")
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if !bytes.Equal(value, []byte{0xA, 0x0}) {
		t.Fatalf("Decoded to wrong result (0xA00 <> [10,0]): %v", value)
	}

	// check correct hex number
	value, err = decode("0x0800")
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if !bytes.Equal(value, []byte{0x8, 0x0}) {
		t.Fatalf("Decoded to wrong result (0x0800 <> [8,0]): %v", value)
	}

	// check odd number of hex characters
	_, err = decode("0xJ00")
	if err == nil {
		t.Fatalf("Decoded with wrong hex number (0xJ00)")
	}
}
