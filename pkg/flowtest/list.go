// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package flowtest

type TestPacket struct {
	prev *TestPacket
	next *TestPacket
	key  string
	// timeSend     time.Time
	// timeExpected time.Time
}

type ExpectedList struct {
	head *TestPacket
	tail *TestPacket
}

func (l *ExpectedList) Insert(record *TestPacket) {
	record.prev = nil
	record.next = nil

	// nothing in the list ?
	if l.head == nil {
		l.head = record
		l.tail = record
		return
	}

	// add to the back of the list
	tail := l.tail
	record.prev = tail
	tail.next = record
	l.tail = record
}

func (l *ExpectedList) Find(id string) *TestPacket {
	current := l.head
	for current != nil {
		// is this is the requested packet
		if current.key == id {
			return current
		}

		// look at next packet
		current = current.next
	}

	return nil
}

func (l *ExpectedList) Remove(record *TestPacket) {
	prev := record.prev
	next := record.next

	// first in list so make next the head
	if l.head == record {
		l.head = next
	}

	// last in list so make prev the tail
	if l.tail == record {
		l.tail = prev
	}

	// set previous record next to next if exist
	if prev != nil {
		prev.next = next
	}

	// set next record prev to prev if exist
	if next != nil {
		next.prev = prev
	}

	// clear pointers!
	record.prev = nil
	record.next = nil
}
