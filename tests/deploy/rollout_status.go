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

package deploy

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

// rolloutStatus  is a typed function that
// abstracts status message formation logic
type rolloutStatus func(*Deploy) string

// rolloutStatuses contains a group of status message for
// each predicate checks. It uses predicateName as key.
var rolloutStatuses = map[PredicateName]rolloutStatus{
	// PredicateProgressDeadlineExceeded refer to rolloutStatus
	// for predicate IsProgressDeadlineExceeded.
	PredicateProgressDeadlineExceeded: func(d *Deploy) string {
		return "deployment exceeded its progress deadline"
	},
	// PredicateOlderReplicaActive refer to rolloutStatus for
	// predicate IsOlderReplicaActive.
	PredicateOlderReplicaActive: func(d *Deploy) string {
		if d.object.Spec.Replicas == nil {
			return "replica update in-progress: some older replicas were updated"
		}
		return fmt.Sprintf(
			"replica update in-progress: %d of %d new replicas were updated",
			d.object.Status.UpdatedReplicas, *d.object.Spec.Replicas)
	},
	// PredicateTerminationInProgress refer rolloutStatus
	// for predicate IsTerminationInProgress.
	PredicateTerminationInProgress: func(d *Deploy) string {
		return fmt.Sprintf(
			"replica termination in-progress: %d old replicas are pending termination",
			d.object.Status.Replicas-d.object.Status.UpdatedReplicas)
	},
	// PredicateUpdateInProgress refer to rolloutStatus for predicate IsUpdateInProgress.
	PredicateUpdateInProgress: func(d *Deploy) string {
		return fmt.Sprintf(
			"replica update in-progress: %d of %d updated replicas are available",
			d.object.Status.AvailableReplicas, d.object.Status.UpdatedReplicas)
	},
	// PredicateNotSpecSynced refer to status rolloutStatus for predicate IsNotSyncSpec.
	PredicateNotSpecSynced: func(d *Deploy) string {
		return "deployment rollout in-progress: waiting for deployment spec update"
	},
}

// rolloutChecks contains a group of predicate it uses predicateName as key.
var rolloutChecks = map[PredicateName]Predicate{
	// PredicateProgressDeadlineExceeded refer to predicate IsProgressDeadlineExceeded.
	PredicateProgressDeadlineExceeded: IsProgressDeadlineExceeded(),
	// PredicateOlderReplicaActive refer to predicate IsOlderReplicaActive.
	PredicateOlderReplicaActive: IsOlderReplicaActive(),
	// PredicateTerminationInProgress refer to predicate IsTerminationInProgress.
	PredicateTerminationInProgress: IsTerminationInProgress(),
	// PredicateUpdateInProgress refer to predicate IsUpdateInProgress.
	PredicateUpdateInProgress: IsUpdateInProgress(),
	// PredicateNotSpecSynced refer to predicate IsSyncSpec.
	PredicateNotSpecSynced: IsNotSyncSpec(),
}

// RolloutOutput struct contains message and boolean value to show rolloutstatus
type RolloutOutput struct {
	IsRolledout bool   `json:"isRolledout"`
	Message     string `json:"message"`
}

// rawFn is a typed function that abstracts
// conversion of rolloutOutput struct to raw byte
type rawFn func(r *RolloutOutput) ([]byte, error)

// Rollout enables getting various output format of rolloutOutput
type Rollout struct {
	output *RolloutOutput
	raw    rawFn
}

// rolloutBuildOption defines the
// abstraction to build a rollout instance
type rolloutBuildOption func(*Rollout)

// NewRollout returns new instance of rollout meant for
// rolloutOutput. caller can configure it with different
// rolloutOutputBuildOption
func NewRollout(opts ...rolloutBuildOption) *Rollout {
	r := &Rollout{}
	for _, o := range opts {
		o(r)
	}
	r.withDefaults()
	return r
}

// withOutputObject sets rolloutOutput in rollout instance
func withOutputObject(o *RolloutOutput) rolloutBuildOption {
	return func(r *Rollout) {
		r.output = o
	}
}

// withDefaults sets the default options of rolloutBuilder instance
func (r *Rollout) withDefaults() {
	if r.raw == nil {
		r.raw = func(o *RolloutOutput) ([]byte, error) {
			return json.Marshal(o)
		}
	}
}

// Raw returns raw bytes outpot of rollout
func (r *Rollout) Raw() ([]byte, error) {
	if r.output == nil {
		return nil, errors.New("unable to get rollout status output")
	}
	return r.raw(r.output)
}
