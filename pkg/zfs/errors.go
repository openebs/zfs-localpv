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
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Sentinel errors returned by classifiers. Callers should use errors.Is
// to test for these categories; the underlying *ZFSError carries the
// original stderr and operation context.
var (
	ErrPoolNotFound      = errors.New("zfs: pool not found")
	ErrDatasetNotFound   = errors.New("zfs: dataset not found")
	ErrDatasetExists     = errors.New("zfs: dataset already exists")
	ErrDatasetBusy       = errors.New("zfs: dataset is busy")
	ErrInsufficientSpace = errors.New("zfs: insufficient space")
	ErrPermissionDenied  = errors.New("zfs: permission denied")
	ErrInvalidArgument   = errors.New("zfs: invalid argument")
	ErrUnknown           = errors.New("zfs: unknown error")
)

// Reason constants identify the failure class on an operation. They are
// used both as the CamelCase `reason` on Kubernetes Events emitted by
// the node agent and as the key for mapping to gRPC status codes on the
// CSI controller. Keeping one enum for both sides guarantees that the
// value observed in `kubectl describe` matches what the controller
// logged.
const (
	// Operation-level reasons (terminal failures on a CR).
	ReasonProvisionFailed    = "ProvisionFailed"
	ReasonCloneFailed        = "CloneFailed"
	ReasonDestroyFailed      = "DestroyFailed"
	ReasonSetPropertyFailed  = "SetPropertyFailed"
	ReasonResizeFailed       = "ResizeFailed"
	ReasonSnapshotFailed     = "SnapshotFailed"
	ReasonSnapDestroyFailed  = "SnapshotDestroyFailed"
	ReasonBackupFailed       = "BackupFailed"
	ReasonRestoreFailed      = "RestoreFailed"

	// Fine-grained classified reasons (derived from stderr). These are
	// set on ZFSError.Reason and used by the controller to pick the
	// correct gRPC code.
	ReasonPoolNotFound       = "PoolNotFound"
	ReasonDatasetNotFound    = "DatasetNotFound"
	ReasonDatasetExists      = "DatasetExists"
	ReasonDatasetBusy        = "DatasetBusy"
	ReasonInsufficientSpace  = "InsufficientCapacity"
	ReasonPermissionDenied   = "PermissionDenied"
	ReasonInvalidArgument    = "InvalidArgument"
	ReasonUnknown            = "Unknown"

	// Success reasons for Normal events.
	ReasonProvisioned      = "Provisioned"
	ReasonCloned           = "Cloned"
	ReasonDestroyed        = "Destroyed"
	ReasonPropertySet      = "PropertySet"
	ReasonResized          = "Resized"
	ReasonSnapshotted      = "Snapshotted"
	ReasonSnapDestroyed    = "SnapshotDestroyed"
	ReasonBackupCompleted  = "BackupCompleted"
	ReasonRestoreCompleted = "RestoreCompleted"
)

// ZFSError is the typed error returned by every shell-out to the zfs /
// zpool binaries. It preserves the original exec error, the scrubbed
// stderr, the operation name, the dataset it was performed against, and
// a fine-grained classified reason. Callers should use errors.As to
// extract it or errors.Is against one of the sentinel errors above.
type ZFSError struct {
	Op      string // e.g. "zfs create", "zfs destroy"
	Dataset string // target dataset (may be empty)
	Reason  string // one of the Reason* constants above
	Stderr  string // trimmed & scrubbed combined output from the binary
	Err     error  // original error (exec.ExitError, etc.)
}

// Error implements the error interface. The message is deliberately
// compact and safe to emit verbatim as an Event message — callers
// should not re-wrap it with additional PII.
func (e *ZFSError) Error() string {
	if e == nil {
		return ""
	}
	switch {
	case e.Dataset != "" && e.Stderr != "":
		return fmt.Sprintf("%s on %q failed: %s: %s", e.Op, e.Dataset, e.Reason, e.Stderr)
	case e.Dataset != "":
		return fmt.Sprintf("%s on %q failed: %s: %v", e.Op, e.Dataset, e.Reason, e.Err)
	case e.Stderr != "":
		return fmt.Sprintf("%s failed: %s: %s", e.Op, e.Reason, e.Stderr)
	default:
		return fmt.Sprintf("%s failed: %s: %v", e.Op, e.Reason, e.Err)
	}
}

// Unwrap returns the chain root so errors.Is against sentinel errors
// works: callers can write errors.Is(err, ErrPoolNotFound).
func (e *ZFSError) Unwrap() error {
	if e == nil {
		return nil
	}
	if sentinel := sentinelFor(e.Reason); sentinel != nil {
		return sentinel
	}
	return e.Err
}

// sentinelFor maps a classified reason to its package-level sentinel.
// Operation-level reasons (ProvisionFailed etc.) do not have sentinels
// because they describe the operation, not the cause.
func sentinelFor(reason string) error {
	switch reason {
	case ReasonPoolNotFound:
		return ErrPoolNotFound
	case ReasonDatasetNotFound:
		return ErrDatasetNotFound
	case ReasonDatasetExists:
		return ErrDatasetExists
	case ReasonDatasetBusy:
		return ErrDatasetBusy
	case ReasonInsufficientSpace:
		return ErrInsufficientSpace
	case ReasonPermissionDenied:
		return ErrPermissionDenied
	case ReasonInvalidArgument:
		return ErrInvalidArgument
	case ReasonUnknown:
		return ErrUnknown
	}
	return nil
}

// stderrPattern pairs a regex with a fine-grained reason. Patterns are
// ordered most-specific first so that e.g. "no such pool" wins over
// "does not exist". All matching is case-insensitive.
type stderrPattern struct {
	re     *regexp.Regexp
	reason string
}

var stderrPatterns = []stderrPattern{
	{regexp.MustCompile(`(?i)no\s+such\s+pool`), ReasonPoolNotFound},
	{regexp.MustCompile(`(?i)pool\s+.*\s+does\s+not\s+exist`), ReasonPoolNotFound},
	{regexp.MustCompile(`(?i)cannot\s+open\s+'[^']+'[:\s]+no\s+such\s+pool`), ReasonPoolNotFound},
	{regexp.MustCompile(`(?i)dataset\s+does\s+not\s+exist`), ReasonDatasetNotFound},
	{regexp.MustCompile(`(?i)cannot\s+open\s+'[^']+'[:\s]+dataset\s+does\s+not\s+exist`), ReasonDatasetNotFound},
	{regexp.MustCompile(`(?i)filesystem\s+does\s+not\s+exist`), ReasonDatasetNotFound},
	{regexp.MustCompile(`(?i)snapshot\s+does\s+not\s+exist`), ReasonDatasetNotFound},
	{regexp.MustCompile(`(?i)dataset\s+already\s+exists`), ReasonDatasetExists},
	{regexp.MustCompile(`(?i)cannot\s+create\s+'[^']+'[:\s]+dataset\s+already\s+exists`), ReasonDatasetExists},
	{regexp.MustCompile(`(?i)snapshot\s+already\s+exists`), ReasonDatasetExists},
	{regexp.MustCompile(`(?i)dataset\s+is\s+busy`), ReasonDatasetBusy},
	{regexp.MustCompile(`(?i)pool\s+is\s+busy`), ReasonDatasetBusy},
	{regexp.MustCompile(`(?i)has\s+dependent\s+clones`), ReasonDatasetBusy},
	{regexp.MustCompile(`(?i)out\s+of\s+space`), ReasonInsufficientSpace},
	{regexp.MustCompile(`(?i)no\s+space\s+left`), ReasonInsufficientSpace},
	{regexp.MustCompile(`(?i)quota\s+exceeded`), ReasonInsufficientSpace},
	{regexp.MustCompile(`(?i)permission\s+denied`), ReasonPermissionDenied},
	{regexp.MustCompile(`(?i)operation\s+not\s+permitted`), ReasonPermissionDenied},
	{regexp.MustCompile(`(?i)invalid\s+argument`), ReasonInvalidArgument},
	{regexp.MustCompile(`(?i)invalid\s+option`), ReasonInvalidArgument},
	{regexp.MustCompile(`(?i)invalid\s+property`), ReasonInvalidArgument},
	{regexp.MustCompile(`(?i)bad\s+property`), ReasonInvalidArgument},
}

// classifyStderr maps a stderr blob from zfs/zpool to a fine-grained
// reason. Returns ReasonUnknown if no pattern matches.
func classifyStderr(stderr string) string {
	for _, p := range stderrPatterns {
		if p.re.MatchString(stderr) {
			return p.reason
		}
	}
	return ReasonUnknown
}

// NewZFSError constructs a classified error from the raw output of a
// zfs/zpool invocation. If err is nil, NewZFSError returns nil.
func NewZFSError(op, dataset string, err error, stderr []byte) error {
	if err == nil {
		return nil
	}
	msg := strings.TrimSpace(string(stderr))
	return &ZFSError{
		Op:      op,
		Dataset: dataset,
		Reason:  classifyStderr(msg),
		Stderr:  msg,
		Err:     err,
	}
}

// ReasonOf extracts the classified reason from err. If err is not a
// *ZFSError (or is nil), it returns ReasonUnknown. This is the canonical
// way for callers (event emitters, gRPC boundary) to obtain the reason
// without type-asserting everywhere.
func ReasonOf(err error) string {
	if err == nil {
		return ""
	}
	var zerr *ZFSError
	if errors.As(err, &zerr) {
		return zerr.Reason
	}
	return ReasonUnknown
}


