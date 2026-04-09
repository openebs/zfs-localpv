/*
Copyright 2020 The OpenEBS Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package driver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoundOff(t *testing.T) {

	tests := map[string]struct {
		input    int64
		expected int64
	}{
		"Minimum allocatable is 1Mi": {input: 1, expected: Mi},
		"roundOff to same Mi size":   {input: Mi, expected: Mi},
		"roundOff to nearest Mi":     {input: Mi + 1, expected: Mi * 2},
		"roundOff to same Gi size":   {input: Gi, expected: Gi},
		"roundOff to nearest Gi":     {input: Gi + 1, expected: Gi * 2},
		"roundOff MB size":           {input: 5 * MB, expected: 5 * Mi},
		"roundOff GB size":           {input: 5 * GB, expected: 5 * Gi},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, getRoundedCapacity(test.input))
		})
	}
}

func TestIsSamePool(t *testing.T) {

	tests := map[string]struct {
		first    string
		second   string
		expected bool
	}{
		"root of same pool":                             {first: "data", second: "data", expected: true},
		"root and dataset of same pool":                 {first: "data", second: "data/sub1", expected: true},
		"same dataset of same pool":                     {first: "data/sub1", second: "data/sub1", expected: true},
		"different datasets of same pool":               {first: "data/sub1", second: "data/sub2", expected: true},
		"roots of different pools":                      {first: "data", second: "other-data", expected: false},
		"root and dataset of different pools":           {first: "data", second: "other-data/sub1", expected: false},
		"identically named datasets of different pools": {first: "data/sub1", second: "other-data/sub1", expected: false},
		"differently named datasets of different pools": {first: "data/sub1", second: "other-data/sub2", expected: false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, isSamePool(test.first, test.second))
		})
	}
}
