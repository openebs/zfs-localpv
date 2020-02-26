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

package usage

import "testing"

func TestToGigaUnits(t *testing.T) {
	tests := map[string]struct {
		stringSize    string
		expectedGsize int64
		positiveTest  bool
	}{
		"One Hundred Twenty Three thousand Four Hundred Fifty Six Teribytes": {
			"123456 TiB",
			123456000,
			true,
		},
		"One Gibibyte": {
			"1 GiB",
			1,
			true,
		},
		"One Megabyte": {
			"1 MB",
			0, // One cannot express <1GB in integer
			true,
		},
		"One Megabyte negative-case": {
			"1 MB",
			1,
			false,
			// 1 MB isn't 1 GB
		},
		"One hundred four point five gigabyte": {
			"104.5 GB",
			104,
			true,
		},
	}

	for testKey, testSuite := range tests {
		gotValue, err := toGigaUnits(testSuite.stringSize)
		if (gotValue != testSuite.expectedGsize || err != nil) && testSuite.positiveTest {
			t.Fatalf("Tests failed for %s, expected=%d, got=%d", testKey, testSuite.expectedGsize, gotValue)
		}
	}
}
