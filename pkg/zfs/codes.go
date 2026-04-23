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

import "google.golang.org/grpc/codes"

// ReasonToGRPCCode maps a Kubernetes Event reason emitted by the node
// agent (see errors.go for the Reason* constants) to the gRPC status
// code the CSI controller should return. Unknown or unclassified
// reasons fall through to codes.Internal, which matches the existing
// pre-refactor behaviour.
//
// The mapping is deliberately small and conservative: external CSI
// provisioners (external-provisioner, external-snapshotter) treat most
// non-Internal codes as terminal, so we only use a non-Internal code
// when we are confident the condition will not clear on retry.
func ReasonToGRPCCode(reason string) codes.Code {
	switch reason {
	case ReasonInsufficientSpace:
		return codes.ResourceExhausted
	case ReasonPoolNotFound, ReasonInvalidArgument:
		return codes.FailedPrecondition
	case ReasonPermissionDenied:
		return codes.PermissionDenied
	case ReasonDatasetExists:
		return codes.AlreadyExists
	case ReasonDatasetNotFound:
		return codes.NotFound
	default:
		return codes.Internal
	}
}
