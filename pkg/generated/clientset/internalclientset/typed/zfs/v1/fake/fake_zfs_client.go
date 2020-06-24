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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1 "github.com/openebs/zfs-localpv/pkg/generated/clientset/internalclientset/typed/zfs/v1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeZfsV1 struct {
	*testing.Fake
}

func (c *FakeZfsV1) ZFSBackups(namespace string) v1.ZFSBackupInterface {
	return &FakeZFSBackups{c, namespace}
}

func (c *FakeZfsV1) ZFSRestores(namespace string) v1.ZFSRestoreInterface {
	return &FakeZFSRestores{c, namespace}
}

func (c *FakeZfsV1) ZFSSnapshots(namespace string) v1.ZFSSnapshotInterface {
	return &FakeZFSSnapshots{c, namespace}
}

func (c *FakeZfsV1) ZFSVolumes(namespace string) v1.ZFSVolumeInterface {
	return &FakeZFSVolumes{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeZfsV1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
