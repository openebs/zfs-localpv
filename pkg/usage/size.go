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

import units "github.com/docker/go-units"

// toGigaUnits converts a size from xB to bytes where x={k,m,g,t,p...}
// and return the number of Gigabytes as an integer
// 1 gigabyte=1000 megabyte
func toGigaUnits(size string) (int64, error) {
	sizeInBytes, err := units.FromHumanSize(size)
	return sizeInBytes / units.GB, err
}
