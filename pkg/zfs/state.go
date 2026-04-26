/*
Copyright 2024 The OpenEBS Authors

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

package zfs

import (
	"os/exec"
	"strings"
)

// State probes used by classifyByState. Defined as package vars so
// tests can swap them out without exec'ing real zfs binaries.
var (
	datasetExistsFn = datasetExistsCmd
	poolExistsFn    = poolExistsCmd
)

// datasetExistsCmd reports whether a dataset (filesystem, volume,
// snapshot or bookmark) is visible to `zfs list`. Exit status is the
// signal: `zfs list <name>` returns 0 iff the name resolves. We do not
// parse stdout or stderr.
func datasetExistsCmd(name string) bool {
	if name == "" {
		return false
	}
	return exec.Command(ZFSVolCmd, "list", "-t", "all", "-H", "-o", "name", name).Run() == nil
}

// poolExistsCmd reports whether a pool's root dataset is visible. The
// pool root is itself a dataset, so this uses the same `zfs list`
// exit-status contract as datasetExistsCmd. Avoids depending on a
// separate `zpool` binary that this codebase does not otherwise call.
func poolExistsCmd(pool string) bool {
	if pool == "" {
		return false
	}
	return exec.Command(ZFSVolCmd, "list", "-H", "-o", "name", pool).Run() == nil
}

// poolFromDataset returns the pool component of a ZFS dataset name.
// For "tank/foo/bar@snap" it returns "tank". Returns the input
// unchanged if no separator is present (i.e. a bare pool name).
func poolFromDataset(name string) string {
	if name == "" {
		return ""
	}
	if i := strings.IndexAny(name, "/@"); i > 0 {
		return name[:i]
	}
	return name
}

// classifyByState derives a fine-grained reason for a failed zfs
// operation by inspecting current ZFS state. It is the sole source of
// fine-grained reasons in this package; the stderr from the failing
// command is preserved verbatim on the ZFSError for human consumption
// but is intentionally not parsed.
//
// The set of reasons returned here is deliberately small: only those
// the controller maps to a non-Internal gRPC code (see codes.go) and
// that can be derived unambiguously from state. Everything else falls
// through to ReasonUnknown, where the raw stderr is what users see in
// `kubectl describe`.
func classifyByState(op, dataset string) string {
	if dataset == "" {
		return ReasonUnknown
	}
	pool := poolFromDataset(dataset)
	if pool != "" && pool != dataset && !poolExistsFn(pool) {
		return ReasonPoolNotFound
	}
	exists := datasetExistsFn(dataset)
	creating := isCreateOp(op)
	switch {
	case creating && exists:
		return ReasonDatasetExists
	case !creating && !exists:
		return ReasonDatasetNotFound
	}
	return ReasonUnknown
}

// isCreateOp reports whether op is a write that introduces a new
// dataset. Destroys are explicitly excluded so that a name like
// "zfs destroy snapshot" is correctly classified as a destroy.
func isCreateOp(op string) bool {
	if strings.Contains(op, "destroy") {
		return false
	}
	return strings.Contains(op, "create") ||
		strings.Contains(op, "clone") ||
		strings.Contains(op, "snapshot") ||
		strings.Contains(op, "recv")
}
