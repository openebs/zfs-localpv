/*
Copyright 2020 The OpenEBS Authors

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

// TODO
// Move this file to pkg/k8sresource/v1alpha1
package v1alpha1

// verify if resource struct is an implementation of ResourceGetter
var _ ResourceGetter = &ResourceStruct{}

// verify if resource struct is an implementation of ResourceCreator
var _ ResourceCreator = &ResourceStruct{}

// verify if resource struct is an implementation of ResourceUpdater
var _ ResourceUpdater = &ResourceStruct{}

// verify if createOrUpdate struct is an implementation of ResourceApplier
var _ ResourceApplier = &ResourceCreateOrUpdater{}
