/*
Copyright 2019 The OpenEBS Authors

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

package env

import (
	"os"
	"strconv"
	"strings"
)

const (
	// KubeConfig is the ENV variable to fetch kubernetes kubeconfig
	KubeConfig = "OPENEBS_IO_KUBE_CONFIG"

	// KubeMaster is the ENV variable to fetch kubernetes master's address
	KubeMaster = "OPENEBS_IO_K8S_MASTER"
)

// EnvironmentSetter abstracts setting of environment variable
type EnvironmentSetter func(envKey string, value string) (err error)

// EnvironmentGetter abstracts fetching value from an environment variable
type EnvironmentGetter func(envKey string) (value string)

// EnvironmentLookup abstracts looking up an environment variable
type EnvironmentLookup func(envKey string) (value string, present bool)

// Set sets the provided environment variable
//
// NOTE:
//  This is an implementation of EnvironmentSetter
func Set(envKey string, value string) (err error) {
	return os.Setenv(string(envKey), value)
}

// Get fetches value from the provided environment variable
//
// NOTE:
//  This is an implementation of EnvironmentGetter
func Get(envKey string) (value string) {
	return getEnv(string(envKey))
}

// GetOrDefault fetches value from the provided environment variable
// which on empty returns the defaultValue
// NOTE: os.Getenv is used here instead of os.LookupEnv because it is
// not required to know if the environment variable is defined on the system
func GetOrDefault(e string, defaultValue string) (value string) {
	envValue := Get(e)
	if len(envValue) == 0 {
		// ENV not defined or set to ""
		return defaultValue
	}
	return envValue
}

// Lookup looks up an environment variable
//
// NOTE:
//  This is an implementation of EnvironmentLookup
func Lookup(envKey string) (value string, present bool) {
	return lookupEnv(string(envKey))
}

// Truthy returns boolean based on the environment variable's value
//
// The lookup value can be truthy (i.e. 1, t, TRUE, true) or falsy (0, false,
// etc) based on strconv.ParseBool logic
func Truthy(envKey string) (truth bool) {
	v, found := Lookup(envKey)
	if !found {
		return
	}
	truth, _ = strconv.ParseBool(v)
	return
}

// LookupOrFalse looks up an environment variable and returns a string "false"
// if environment variable is not present. It returns appropriate values for
// other cases.
func LookupOrFalse(envKey string) string {
	val, present := Lookup(envKey)
	if !present {
		return "false"
	}
	return strings.TrimSpace(val)
}

// getEnv fetches the provided environment variable's value
func getEnv(envKey string) (value string) {
	return strings.TrimSpace(os.Getenv(envKey))
}

// lookupEnv looks up the provided environment variable
func lookupEnv(envKey string) (value string, present bool) {
	value, present = os.LookupEnv(envKey)
	value = strings.TrimSpace(value)
	return
}
