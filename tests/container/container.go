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

package container

import (
	"github.com/openebs/lib-csi/pkg/common/errors"
	corev1 "k8s.io/api/core/v1"
)

type container struct {
	corev1.Container // kubernetes container type
}

// OptionFunc is a typed function that abstracts anykind of operation
// against the provided container instance
//
// This is the basic building block to create functional operations
// against the container instance
type OptionFunc func(*container)

// Predicate abstracts conditional logic w.r.t the container instance
//
// NOTE:
// Predicate is a functional approach versus traditional approach to mix
// conditions such as *if-else* within blocks of business logic
//
// NOTE:
// Predicate approach enables clear separation of conditionals from
// imperatives i.e. actions that form the business logic
type Predicate func(*container) (nameOrMsg string, ok bool)

// predicateFailedError returns the provided predicate as an error
func predicateFailedError(message string) error {
	return errors.Errorf("predicatefailed: %s", message)
}

var (
	errorvalidationFailed = errors.New("container validation failed")
)

// asContainer transforms this container instance into corresponding kubernetes
// container type
func (c *container) asContainer() corev1.Container {
	return corev1.Container{
		Name:                     c.Name,
		Image:                    c.Image,
		Command:                  c.Command,
		Args:                     c.Args,
		WorkingDir:               c.WorkingDir,
		Ports:                    c.Ports,
		EnvFrom:                  c.EnvFrom,
		Env:                      c.Env,
		Resources:                c.Resources,
		VolumeMounts:             c.VolumeMounts,
		VolumeDevices:            c.VolumeDevices,
		LivenessProbe:            c.LivenessProbe,
		ReadinessProbe:           c.ReadinessProbe,
		Lifecycle:                c.Lifecycle,
		TerminationMessagePath:   c.TerminationMessagePath,
		TerminationMessagePolicy: c.TerminationMessagePolicy,
		ImagePullPolicy:          c.ImagePullPolicy,
		SecurityContext:          c.SecurityContext,
		Stdin:                    c.Stdin,
		StdinOnce:                c.StdinOnce,
		TTY:                      c.TTY,
	}
}

// New returns a new kubernetes container
func New(opts ...OptionFunc) corev1.Container {
	c := &container{}
	for _, o := range opts {
		o(c)
	}
	return c.asContainer()
}

// Builder provides utilities required to build a kubernetes container type
type Builder struct {
	con    *container  // container instance
	checks []Predicate // validations to be done while building the container instance
	errors []error     // errors found while building the container instance
}

// NewBuilder returns a new instance of builder
func NewBuilder() *Builder {
	return &Builder{
		con: &container{},
	}
}

// validate will run checks against container instance
func (b *Builder) validate() error {
	for _, c := range b.checks {
		if m, ok := c(b.con); !ok {
			b.errors = append(b.errors, predicateFailedError(m))
		}
	}
	if len(b.errors) == 0 {
		return nil
	}
	return errorvalidationFailed
}

// Build returns the final kubernetes container
func (b *Builder) Build() (corev1.Container, error) {
	err := b.validate()
	if err != nil {
		return corev1.Container{}, err
	}
	return b.con.asContainer(), nil
}

// AddCheck adds the predicate as a condition to be validated against the
// container instance
func (b *Builder) AddCheck(p Predicate) *Builder {
	b.checks = append(b.checks, p)
	return b
}

// AddChecks adds the provided predicates as conditions to be validated against
// the container instance
func (b *Builder) AddChecks(p []Predicate) *Builder {
	for _, check := range p {
		b.AddCheck(check)
	}
	return b
}

// WithName sets the name of the container
func (b *Builder) WithName(name string) *Builder {
	if len(name) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: missing name"),
		)
		return b
	}
	WithName(name)(b.con)
	return b
}

// WithName sets the name of the container
func WithName(name string) OptionFunc {
	return func(c *container) {
		c.Name = name
	}
}

// WithImage sets the image of the container
func (b *Builder) WithImage(img string) *Builder {
	if len(img) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: missing image"),
		)
		return b
	}
	WithImage(img)(b.con)
	return b
}

// WithImage sets the image of the container
func WithImage(img string) OptionFunc {
	return func(c *container) {
		c.Image = img
	}
}

// WithCommandNew sets the command of the container
func (b *Builder) WithCommandNew(cmd []string) *Builder {
	if cmd == nil {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: nil command"),
		)
		return b
	}

	if len(cmd) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: missing command"),
		)
		return b
	}

	newcmd := []string{}
	newcmd = append(newcmd, cmd...)

	b.con.Command = newcmd
	return b
}

// WithArgumentsNew sets the command arguments of the container
func (b *Builder) WithArgumentsNew(args []string) *Builder {
	if args == nil {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: nil arguments"),
		)
		return b
	}

	if len(args) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: missing arguments"),
		)
		return b
	}

	newargs := []string{}
	newargs = append(newargs, args...)

	b.con.Args = newargs
	return b
}

// WithVolumeMountsNew sets the command arguments of the container
func (b *Builder) WithVolumeMountsNew(volumeMounts []corev1.VolumeMount) *Builder {
	if volumeMounts == nil {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: nil volumemounts"),
		)
		return b
	}

	if len(volumeMounts) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: missing volumemounts"),
		)
		return b
	}
	newvolumeMounts := []corev1.VolumeMount{}
	newvolumeMounts = append(newvolumeMounts, volumeMounts...)
	b.con.VolumeMounts = newvolumeMounts
	return b
}

// WithVolumeDevicesNew sets the command arguments of the container
func (b *Builder) WithVolumeDevicesNew(volumeDevices []corev1.VolumeDevice) *Builder {
	if volumeDevices == nil {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: nil volumeDevices"),
		)
		return b
	}

	if len(volumeDevices) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: missing volumeDevices"),
		)
		return b
	}
	newvolumeDevices := []corev1.VolumeDevice{}
	newvolumeDevices = append(newvolumeDevices, volumeDevices...)
	b.con.VolumeDevices = newvolumeDevices
	return b
}

// WithImagePullPolicy sets the image pull policy of the container
func (b *Builder) WithImagePullPolicy(policy corev1.PullPolicy) *Builder {
	if len(policy) == 0 {
		b.errors = append(
			b.errors,
			errors.New(
				"failed to build container object: missing imagepullpolicy",
			),
		)
		return b
	}

	b.con.ImagePullPolicy = policy
	return b
}

// WithPrivilegedSecurityContext sets securitycontext of the container
func (b *Builder) WithPrivilegedSecurityContext(privileged *bool) *Builder {
	if privileged == nil {
		b.errors = append(
			b.errors,
			errors.New(
				"failed to build container object: missing securitycontext",
			),
		)
		return b
	}

	newprivileged := *privileged
	newsecuritycontext := &corev1.SecurityContext{
		Privileged: &newprivileged,
	}

	b.con.SecurityContext = newsecuritycontext
	return b
}

// WithResources sets resources of the container
func (b *Builder) WithResources(
	resources *corev1.ResourceRequirements,
) *Builder {
	if resources == nil {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: missing resources"),
		)
		return b
	}

	newresources := *resources
	b.con.Resources = newresources
	return b
}

// WithResourcesByValue sets resources of the container
func (b *Builder) WithResourcesByValue(resources corev1.ResourceRequirements) *Builder {
	b.con.Resources = resources
	return b
}

// WithPortsNew sets ports of the container
func (b *Builder) WithPortsNew(ports []corev1.ContainerPort) *Builder {
	if ports == nil {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: nil ports"),
		)
		return b
	}

	if len(ports) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: missing ports"),
		)
		return b
	}

	newports := []corev1.ContainerPort{}
	newports = append(newports, ports...)

	b.con.Ports = newports
	return b
}

// WithEnvsNew sets the envs of the container
func (b *Builder) WithEnvsNew(envs []corev1.EnvVar) *Builder {
	if envs == nil {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: nil envs"),
		)
		return b
	}

	if len(envs) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: missing envs"),
		)
		return b
	}

	newenvs := []corev1.EnvVar{}
	newenvs = append(newenvs, envs...)

	b.con.Env = newenvs
	return b
}

// WithEnvs sets the envs of the container
func (b *Builder) WithEnvs(envs []corev1.EnvVar) *Builder {
	if envs == nil {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: nil envs"),
		)
		return b
	}

	if len(envs) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: missing envs"),
		)
		return b
	}

	if b.con.Env == nil {
		b.WithEnvsNew(envs)
		return b
	}

	b.con.Env = append(b.con.Env, envs...)
	return b
}

// WithLivenessProbe sets the liveness probe of the container
func (b *Builder) WithLivenessProbe(liveness *corev1.Probe) *Builder {
	if liveness == nil {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: nil liveness probe"),
		)
		return b
	}

	b.con.LivenessProbe = liveness
	return b
}

// WithLifeCycle sets the life cycle of the container
func (b *Builder) WithLifeCycle(lc *corev1.Lifecycle) *Builder {
	if lc == nil {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: nil lifecycle"),
		)
		return b
	}

	b.con.Lifecycle = lc
	return b
}
