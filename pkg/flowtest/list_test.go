// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package flowtest

import "testing"

func TestList(t *testing.T) {
	var list = &ExpectedList{}
	var result *TestPacket

	var packet1 = &TestPacket{key: "packet1"}
	var packet2 = &TestPacket{key: "packet2"}
	var packet3 = &TestPacket{key: "packet3"}
	var packet4 = &TestPacket{key: "packet4"}

	list.Insert(packet1)

	result = list.Find(packet1.key)
	if result == nil {
		t.Fatalf("Did not find %s in list", packet1.key)
	}

	idwrong := "wrong1"
	result = list.Find(idwrong)
	if result != nil {
		t.Fatalf("Did find '%s' in list when nothing should be found", result.key)
	}

	list.Insert(packet2)
	list.Insert(packet3)
	list.Insert(packet4)

	result = list.Find(packet1.key)
	if result == nil {
		t.Fatalf("Did not find %s in list", packet1.key)
	}

	result = list.Find(idwrong)
	if result != nil {
		t.Fatalf("Did find '%s' in list when nothing should be found", result.key)
	}

	list.Remove(packet1)
	result = list.Find(packet1.key)
	if result != nil {
		t.Fatalf("Did find '%s' in list when nothing should be found", result.key)
	}

	result = list.Find(packet2.key)
	if result == nil {
		t.Fatalf("Did not find %s in list", packet2.key)
	}

	list.Remove(packet4)
	result = list.Find(packet4.key)
	if result != nil {
		t.Fatalf("Did find '%s' in list when nothing should be found", result.key)
	}

	result = list.Find(packet3.key)
	if result == nil {
		t.Fatalf("Did not find %s in list", packet3.key)
	}

	list.Remove(packet2)
	list.Remove(packet3)
	if list.head != nil && list.tail != nil {
		t.Fatalf("List wasn't empty!")
	}
}
