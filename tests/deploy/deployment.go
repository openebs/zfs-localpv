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
	"github.com/openebs/zfs-localpv/pkg/common/errors"
	templatespec "github.com/openebs/zfs-localpv/tests/pts"
	"github.com/openebs/zfs-localpv/tests/stringer"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Predicate abstracts conditional logic w.r.t the deployment instance
//
// NOTE:
// predicate is a functional approach versus traditional approach to mix
// conditions such as *if-else* within blocks of business logic
//
// NOTE:
// predicate approach enables clear separation of conditionals from
// imperatives i.e. actions that form the business logic
type Predicate func(*Deploy) bool

// Deploy is the wrapper over k8s deployment Object
type Deploy struct {
	// kubernetes deployment instance
	object *appsv1.Deployment
}

// Builder enables building an instance of
// deployment
type Builder struct {
	deployment *Deploy     // kubernetes deployment instance
	checks     []Predicate // predicate list for deploy
	errors     []error
}

// PredicateName type is wrapper over string.
// It is used to refer predicate and status msg.
type PredicateName string

const (
	// PredicateProgressDeadlineExceeded refer to
	// predicate IsProgressDeadlineExceeded.
	PredicateProgressDeadlineExceeded PredicateName = "ProgressDeadlineExceeded"
	// PredicateNotSpecSynced refer to predicate IsNotSpecSynced
	PredicateNotSpecSynced PredicateName = "NotSpecSynced"
	// PredicateOlderReplicaActive refer to predicate IsOlderReplicaActive
	PredicateOlderReplicaActive PredicateName = "OlderReplicaActive"
	// PredicateTerminationInProgress refer to predicate IsTerminationInProgress
	PredicateTerminationInProgress PredicateName = "TerminationInProgress"
	// PredicateUpdateInProgress refer to predicate IsUpdateInProgress.
	PredicateUpdateInProgress PredicateName = "UpdateInProgress"
)

// String implements the stringer interface
func (d *Deploy) String() string {
	return stringer.Yaml("deployment", d.object)
}

// GoString implements the goStringer interface
func (d *Deploy) GoString() string {
	return d.String()
}

// NewBuilder returns a new instance of builder meant for deployment
func NewBuilder() *Builder {
	return &Builder{
		deployment: &Deploy{
			object: &appsv1.Deployment{},
		},
	}
}

// WithName sets the Name field of deployment with provided value.
func (b *Builder) WithName(name string) *Builder {
	if len(name) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build deployment: missing name"),
		)
		return b
	}
	b.deployment.object.Name = name
	return b
}

// WithNamespace sets the Namespace field of deployment with provided value.
func (b *Builder) WithNamespace(namespace string) *Builder {
	if len(namespace) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build deployment: missing namespace"),
		)
		return b
	}
	b.deployment.object.Namespace = namespace
	return b
}

// WithAnnotations merges existing annotations if any
// with the ones that are provided here
func (b *Builder) WithAnnotations(annotations map[string]string) *Builder {
	if len(annotations) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build deployment object: missing annotations"),
		)
		return b
	}

	if b.deployment.object.Annotations == nil {
		return b.WithAnnotationsNew(annotations)
	}

	for key, value := range annotations {
		b.deployment.object.Annotations[key] = value
	}
	return b
}

// WithAnnotationsNew resets existing annotaions if any with
// ones that are provided here
func (b *Builder) WithAnnotationsNew(annotations map[string]string) *Builder {
	if len(annotations) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build deployment object: no new annotations"),
		)
		return b
	}

	// copy of original map
	newannotations := map[string]string{}
	for key, value := range annotations {
		newannotations[key] = value
	}

	// override
	b.deployment.object.Annotations = newannotations
	return b
}

// WithNodeSelector Sets the node selector with the provided argument.
func (b *Builder) WithNodeSelector(selector map[string]string) *Builder {
	if len(selector) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build deployment object: no node selector"),
		)
		return b
	}
	if b.deployment.object.Spec.Template.Spec.NodeSelector == nil {
		return b.WithNodeSelectorNew(selector)
	}

	for key, value := range selector {
		b.deployment.object.Spec.Template.Spec.NodeSelector[key] = value
	}
	return b
}

// WithNodeSelector Sets the node selector with the provided argument.
func (b *Builder) WithNodeSelectorNew(selector map[string]string) *Builder {
	if len(selector) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build deployment object: no new node selector"),
		)
		return b
	}

	b.deployment.object.Spec.Template.Spec.NodeSelector = selector
	return b
}

// WithOwnerReferenceNew sets ownerreference if any with
// ones that are provided here
func (b *Builder) WithOwnerReferenceNew(ownerRefernce []metav1.OwnerReference) *Builder {
	if len(ownerRefernce) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build deployment object: no new ownerRefernce"),
		)
		return b
	}

	b.deployment.object.OwnerReferences = ownerRefernce
	return b
}

// WithLabels merges existing labels if any
// with the ones that are provided here
func (b *Builder) WithLabels(labels map[string]string) *Builder {
	if len(labels) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build deployment object: missing labels"),
		)
		return b
	}

	if b.deployment.object.Labels == nil {
		return b.WithLabelsNew(labels)
	}

	for key, value := range labels {
		b.deployment.object.Labels[key] = value
	}
	return b
}

// WithLabelsNew resets existing labels if any with
// ones that are provided here
func (b *Builder) WithLabelsNew(labels map[string]string) *Builder {
	if len(labels) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build deployment object: no new labels"),
		)
		return b
	}

	// copy of original map
	newlbls := map[string]string{}
	for key, value := range labels {
		newlbls[key] = value
	}

	// override
	b.deployment.object.Labels = newlbls
	return b
}

// WithSelectorMatchLabels merges existing matchlabels if any
// with the ones that are provided here
func (b *Builder) WithSelectorMatchLabels(matchlabels map[string]string) *Builder {
	if len(matchlabels) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build deployment object: missing matchlabels"),
		)
		return b
	}

	if b.deployment.object.Spec.Selector == nil {
		return b.WithSelectorMatchLabelsNew(matchlabels)
	}

	for key, value := range matchlabels {
		b.deployment.object.Spec.Selector.MatchLabels[key] = value
	}
	return b
}

// WithSelectorMatchLabelsNew resets existing matchlabels if any with
// ones that are provided here
func (b *Builder) WithSelectorMatchLabelsNew(matchlabels map[string]string) *Builder {
	if len(matchlabels) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build deployment object: no new matchlabels"),
		)
		return b
	}

	// copy of original map
	newmatchlabels := map[string]string{}
	for key, value := range matchlabels {
		newmatchlabels[key] = value
	}

	newselector := &metav1.LabelSelector{
		MatchLabels: newmatchlabels,
	}

	// override
	b.deployment.object.Spec.Selector = newselector
	return b
}

// WithReplicas sets the replica field of deployment
func (b *Builder) WithReplicas(replicas *int32) *Builder {

	if replicas == nil {
		b.errors = append(
			b.errors,
			errors.New("failed to build deployment object: nil replicas"),
		)
		return b
	}

	newreplicas := *replicas

	if newreplicas < 0 {
		b.errors = append(
			b.errors,
			errors.Errorf(
				"failed to build deployment object: invalid replicas {%d}",
				newreplicas,
			),
		)
		return b
	}

	b.deployment.object.Spec.Replicas = &newreplicas
	return b
}

//WithStrategyType sets the strategy field of the deployment
func (b *Builder) WithStrategyType(
	strategytype appsv1.DeploymentStrategyType,
) *Builder {
	if len(strategytype) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build deployment object: missing strategytype"),
		)
		return b
	}

	b.deployment.object.Spec.Strategy.Type = strategytype
	return b
}

// WithPodTemplateSpecBuilder sets the template field of the deployment
func (b *Builder) WithPodTemplateSpecBuilder(
	tmplbuilder *templatespec.Builder,
) *Builder {
	if tmplbuilder == nil {
		b.errors = append(
			b.errors,
			errors.New("failed to build deployment: nil templatespecbuilder"),
		)
		return b
	}

	templatespecObj, err := tmplbuilder.Build()

	if err != nil {
		b.errors = append(
			b.errors,
			errors.Wrap(
				err,
				"failed to build deployment",
			),
		)
		return b
	}

	b.deployment.object.Spec.Template = *templatespecObj.Object
	return b
}

type deployBuildOption func(*Deploy)

// NewForAPIObject returns a new instance of builder
// for a given deployment Object
func NewForAPIObject(
	obj *appsv1.Deployment,
	opts ...deployBuildOption,
) *Deploy {
	d := &Deploy{object: obj}
	for _, o := range opts {
		o(d)
	}
	return d
}

// Build returns a deployment instance
func (b *Builder) Build() (*appsv1.Deployment, error) {
	err := b.validate()
	// TODO: err in Wrapf is not logged. Fix is required
	if err != nil {
		return nil, errors.Wrapf(err,
			"failed to build a deployment: %s",
			b.deployment.object.Name)
	}
	return b.deployment.object, nil
}

func (b *Builder) validate() error {
	if len(b.errors) != 0 {
		return errors.Errorf(
			"failed to validate: build errors were found: %+v",
			b.errors,
		)
	}
	return nil
}

// IsRollout range over rolloutChecks map and check status of each predicate
// also it generates status message from rolloutStatuses using predicate key
func (d *Deploy) IsRollout() (PredicateName, bool) {
	for pk, p := range rolloutChecks {
		if p(d) {
			return pk, false
		}
	}
	return "", true
}

// FailedRollout returns rollout status message for fail condition
func (d *Deploy) FailedRollout(name PredicateName) *RolloutOutput {
	return &RolloutOutput{
		Message:     rolloutStatuses[name](d),
		IsRolledout: false,
	}
}

// SuccessRollout returns rollout status message for success condition
func (d *Deploy) SuccessRollout() *RolloutOutput {
	return &RolloutOutput{
		Message:     "deployment successfully rolled out",
		IsRolledout: true,
	}
}

// RolloutStatus returns rollout message of deployment instance
func (d *Deploy) RolloutStatus() (op *RolloutOutput, err error) {
	pk, ok := d.IsRollout()
	if ok {
		return d.SuccessRollout(), nil
	}
	return d.FailedRollout(pk), nil
}

// RolloutStatusRaw returns rollout message of deployment instance
// in byte format
func (d *Deploy) RolloutStatusRaw() (op []byte, err error) {
	message, err := d.RolloutStatus()
	if err != nil {
		return nil, err
	}
	return NewRollout(
		withOutputObject(message)).
		Raw()
}

// AddCheck adds the predicate as a condition to be validated
// against the deployment instance
func (b *Builder) AddCheck(p Predicate) *Builder {
	b.checks = append(b.checks, p)
	return b
}

// AddChecks adds the provided predicates as conditions to be
// validated against the deployment instance
func (b *Builder) AddChecks(p []Predicate) *Builder {
	for _, check := range p {
		b.AddCheck(check)
	}
	return b
}

// IsProgressDeadlineExceeded is used to check update is timed out or not.
// If `Progressing` condition's reason is `ProgressDeadlineExceeded` then
// it is not rolled out.
func IsProgressDeadlineExceeded() Predicate {
	return func(d *Deploy) bool {
		return d.IsProgressDeadlineExceeded()
	}
}

// IsProgressDeadlineExceeded is used to check update is timed out or not.
// If `Progressing` condition's reason is `ProgressDeadlineExceeded` then
// it is not rolled out.
func (d *Deploy) IsProgressDeadlineExceeded() bool {
	for _, cond := range d.object.Status.Conditions {
		if cond.Type == appsv1.DeploymentProgressing &&
			cond.Reason == "ProgressDeadlineExceeded" {
			return true
		}
	}
	return false
}

// IsOlderReplicaActive check if older replica's are still active or not if
// Status.UpdatedReplicas < *Spec.Replicas then some of the replicas are
// updated and some of them are not.
func IsOlderReplicaActive() Predicate {
	return func(d *Deploy) bool {
		return d.IsOlderReplicaActive()
	}
}

// IsOlderReplicaActive check if older replica's are still active or not if
// Status.UpdatedReplicas < *Spec.Replicas then some of the replicas are
// updated and some of them are not.
func (d *Deploy) IsOlderReplicaActive() bool {
	return d.object.Spec.Replicas != nil &&
		d.object.Status.UpdatedReplicas < *d.object.Spec.Replicas
}

// IsTerminationInProgress checks for older replicas are waiting to
// terminate or not. If Status.Replicas > Status.UpdatedReplicas then
// some of the older replicas are in running state because newer
// replicas are not in running state. It waits for newer replica to
// come into running state then terminate.
func IsTerminationInProgress() Predicate {
	return func(d *Deploy) bool {
		return d.IsTerminationInProgress()
	}
}

// IsTerminationInProgress checks for older replicas are waiting to
// terminate or not. If Status.Replicas > Status.UpdatedReplicas then
// some of the older replicas are in running state because newer
// replicas are not in running state. It waits for newer replica to
// come into running state then terminate.
func (d *Deploy) IsTerminationInProgress() bool {
	return d.object.Status.Replicas > d.object.Status.UpdatedReplicas
}

// IsUpdateInProgress Checks if all the replicas are updated or not.
// If Status.AvailableReplicas < Status.UpdatedReplicas then all the
//older replicas are not there but there are less number of availableReplicas
func IsUpdateInProgress() Predicate {
	return func(d *Deploy) bool {
		return d.IsUpdateInProgress()
	}
}

// IsUpdateInProgress Checks if all the replicas are updated or not.
// If Status.AvailableReplicas < Status.UpdatedReplicas then all the
// older replicas are not there but there are less number of availableReplicas
func (d *Deploy) IsUpdateInProgress() bool {
	return d.object.Status.AvailableReplicas < d.object.Status.UpdatedReplicas
}

// IsNotSyncSpec compare generation in status and spec and check if
// deployment spec is synced or not. If Generation <= Status.ObservedGeneration
// then deployment spec is not updated yet.
func IsNotSyncSpec() Predicate {
	return func(d *Deploy) bool {
		return d.IsNotSyncSpec()
	}
}

// IsNotSyncSpec compare generation in status and spec and check if
// deployment spec is synced or not. If Generation <= Status.ObservedGeneration
// then deployment spec is not updated yet.
func (d *Deploy) IsNotSyncSpec() bool {
	return d.object.Generation > d.object.Status.ObservedGeneration
}
