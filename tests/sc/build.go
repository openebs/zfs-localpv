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

package sc

import (
	"github.com/openebs/lib-csi/pkg/common/errors"
	storagev1 "k8s.io/api/storage/v1"
)

// Builder enables building an instance of StorageClass
type Builder struct {
	sc   *StorageClass
	errs []error
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{sc: &StorageClass{object: &storagev1.StorageClass{}}}
}

// WithName sets the Name field of storageclass with provided argument.
func (b *Builder) WithName(name string) *Builder {
	if len(name) == 0 {
		b.errs = append(b.errs, errors.New("failed to build storageclass: missing storageclass name"))
		return b
	}
	b.sc.object.Name = name
	return b
}

// WithGenerateName appends a random string after the name
func (b *Builder) WithGenerateName(name string) *Builder {
	b.sc.object.GenerateName = name + "-"
	return b
}

// WithAnnotations sets the Annotations field of storageclass with provided value.
func (b *Builder) WithAnnotations(annotations map[string]string) *Builder {
	if len(annotations) == 0 {
		b.errs = append(b.errs, errors.New("failed to build storageclass: missing annotations"))
	}
	b.sc.object.Annotations = annotations
	return b
}

// WithParametersNew resets existing parameters if any with
// ones that are provided here
func (b *Builder) WithParametersNew(parameters map[string]string) *Builder {
	if len(parameters) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build storageclass object: no new parameters"),
		)
		return b
	}

	// copy of original map
	newparameters := map[string]string{}
	for key, value := range parameters {
		newparameters[key] = value
	}

	// override
	b.sc.object.Parameters = newparameters
	return b
}

// WithProvisioner sets the Provisioner field of storageclass with provided argument.
func (b *Builder) WithProvisioner(provisioner string) *Builder {
	if len(provisioner) == 0 {
		b.errs = append(b.errs, errors.New("failed to build storageclass: missing provisioner name"))
		return b
	}
	b.sc.object.Provisioner = provisioner
	return b
}

// WithVolumeExpansion sets the AllowedVolumeExpansion field of storageclass with provided argument.
func (b *Builder) WithVolumeExpansion(expansionAllowed bool) *Builder {
	b.sc.object.AllowVolumeExpansion = &expansionAllowed
	return b
}

// Build returns the StorageClass API instance
func (b *Builder) Build() (*storagev1.StorageClass, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}
	return b.sc.object, nil
}

// WithVolumeBindingMode sets the volume binding mode of storageclass with
// provided argument.
func (b *Builder) WithVolumeBindingMode(bindingMode storagev1.VolumeBindingMode) *Builder {
	if bindingMode == "" {
		b.errs = append(b.errs, errors.New("failed to build storageclass: missing volume binding mode"))
		return b
	}
	b.sc.object.VolumeBindingMode = &bindingMode
	return b
}
