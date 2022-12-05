// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package kvstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	clean func()
	name  string
}

func (ts *TestStruct) Init(key string, clean func()) error {
	ts.clean = clean
	ts.name = key
	return nil
}

func (ts *TestStruct) Free() {

}

func TestKVStore(t *testing.T) {
	store := New[string, *TestStruct]()
	ts := &TestStruct{}
	store.Set("eerste", ts)
	tsGet, ok := store.Get("eerste")
	assert.Equal(t, ok, true)
	assert.Equal(t, ts, tsGet)
}
