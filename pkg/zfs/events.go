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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
)

// EmitFailureEvent records a Warning event on obj describing a failed
// operation. The event reason is the fine-grained cause (e.g.
// "DatasetExists") when err classifies to a known ZFS failure, and
// falls back to opReason (e.g. "ProvisionFailed") otherwise — this
// follows the Kubernetes convention of "reason = cause, not operation"
// while still yielding a meaningful reason for unclassified errors.
//
// The controller side can map the event reason directly to a gRPC
// status code (see pkg/driver for the table).
//
// recorder may be nil (e.g. in unit tests) in which case this is a
// no-op.
func EmitFailureEvent(recorder record.EventRecorder, obj runtime.Object, opReason string, err error) {
	if recorder == nil || err == nil || obj == nil {
		return
	}
	reason := ReasonOf(err)
	if reason == "" || reason == ReasonUnknown {
		reason = opReason
	}
	recorder.Event(obj, corev1.EventTypeWarning, reason, err.Error())
}

// EmitSuccessEvent records a Normal event on obj. It is intentionally
// only called for terminal, user-visible transitions (provisioned,
// destroyed, ...) so the event stream stays signal-dense.
func EmitSuccessEvent(recorder record.EventRecorder, obj runtime.Object, reason, message string) {
	if recorder == nil || obj == nil {
		return
	}
	recorder.Event(obj, corev1.EventTypeNormal, reason, message)
}
