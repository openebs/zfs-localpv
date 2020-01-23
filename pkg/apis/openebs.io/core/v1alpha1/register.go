/*
Copyright © 2019 The OpenEBS Authors

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SchemeGroupVersion is group version used
// to register custom resources
//
// NOTE:
//  This variable name should not be changed
var SchemeGroupVersion = schema.GroupVersion{
	Group:   "openebs.io",
	Version: "v1alpha1",
}

// Resource takes an unqualified resource and
// returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.
		WithResource(resource).
		GroupResource()
}

var (
	// SchemeBuilder is the scheme builder
	// with scheme init functions to run
	// for this API package
	SchemeBuilder runtime.SchemeBuilder

	localSchemeBuilder = &SchemeBuilder

	// AddToScheme is a global function that
	// registers this API group & version to
	// a scheme
	AddToScheme = localSchemeBuilder.AddToScheme
)

func init() {
	// We only register manually written functions
	// here. This registration of generated functions
	// takes place in the generated files.
	//
	// NOTE:
	//  This separation makes the code compile even
	// when the generated files are missing.
	localSchemeBuilder.Register(addKnownTypes)
}

// Adds the list of known types to api.Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(
		SchemeGroupVersion,
		&ZFSVolume{},
		&ZFSVolumeList{},
		&ZFSSnapshot{},
		&ZFSSnapshotList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
