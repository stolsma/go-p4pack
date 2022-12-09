// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var test1 = "1\n2\n3\n4\n"
var test1Result = "\t1\n\t2\n\t3\n\t4\n"
var test2 = "1\n2\n3\n4"
var test2Result = "\t1\n\t2\n\t3\n\t4"
var test3 = "1\n2\n3\n4\n"
var test3Result = "  1\n  2\n  3\n  4\n"
var test4 = "1\n2\n3\n4"
var test4Result = "  1\n  2\n  3\n  4"

func TestIndent(t *testing.T) {
	assert.Equal(t, test1Result, indent("\t", test1))
	assert.Equal(t, test2Result, indent("\t", test2))
	assert.Equal(t, test3Result, indent("  ", test3))
	assert.Equal(t, test4Result, indent("  ", test4))
}
