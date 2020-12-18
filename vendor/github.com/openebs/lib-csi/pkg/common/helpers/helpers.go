/*
Copyright 2020 The OpenEBS Authors

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

package helpers

import (
	"strings"
)

// GetCaseInsensitiveMap coercs the map's keys to lower case, which only works
// when unicode char is in ASCII subset. May overwrite key-value pairs on
// different permutations of key case as in Key and key. DON'T force values to the
// lower case unconditionally, because values for keys such as mountpoint or
// keylocation are case-sensitive.
// Note that although keys such as 'comPREssion' are accepted and processed,
// even if they are technically invalid, updates to rectify such typing will be
// prohibited as a forbidden update.
func GetCaseInsensitiveMap(dict *map[string]string) map[string]string {
	insensitiveDict := map[string]string{}

	for k, v := range *dict {
		insensitiveDict[strings.ToLower(k)] = v
	}
	return insensitiveDict
}

// GetInsensitiveParameter handles special case ofGetCaseInsensitiveMap looking up one
// key-value pair only
func GetInsensitiveParameter(dict *map[string]string, key string) string {
	insensitiveDict := GetCaseInsensitiveMap(dict)
	return insensitiveDict[strings.ToLower(key)]
}
