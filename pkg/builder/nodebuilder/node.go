/*
Copyright 2021 The OpenEBS Authors

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

package nodebuilder

import (
	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/zfs/v1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Builder is the builder object for ZFSNode
type Builder struct {
	node *ZFSNode
	errs []error
}

// ZFSNode is a wrapper over
// ZFSNode API instance
type ZFSNode struct {
	// ZFSVolume object
	Object *apis.ZFSNode
}

// From returns a new instance of
// zfs volume
func From(node *apis.ZFSNode) *ZFSNode {
	return &ZFSNode{
		Object: node,
	}
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{
		node: &ZFSNode{
			Object: &apis.ZFSNode{},
		},
	}
}

// BuildFrom returns new instance of Builder
// from the provided api instance
func BuildFrom(node *apis.ZFSNode) *Builder {
	if node == nil {
		b := NewBuilder()
		b.errs = append(
			b.errs,
			errors.New("failed to build zfs node object: nil node"),
		)
		return b
	}
	return &Builder{
		node: &ZFSNode{
			Object: node,
		},
	}
}

// WithNamespace sets the namespace of ZFSNode
func (b *Builder) WithNamespace(namespace string) *Builder {
	if namespace == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build zfs node object: missing namespace",
			),
		)
		return b
	}
	b.node.Object.Namespace = namespace
	return b
}

// WithName sets the name of ZFSNode
func (b *Builder) WithName(name string) *Builder {
	if name == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build zfs node object: missing name",
			),
		)
		return b
	}
	b.node.Object.Name = name
	return b
}

// WithPools sets the pools of ZFSNode
func (b *Builder) WithPools(pools []apis.Pool) *Builder {
	b.node.Object.Pools = pools
	return b
}

// WithOwnerReferences sets the owner references of ZFSNode
func (b *Builder) WithOwnerReferences(ownerRefs ...metav1.OwnerReference) *Builder {
	b.node.Object.OwnerReferences = ownerRefs
	return b
}

// Build returns ZFSNode API object
func (b *Builder) Build() (*apis.ZFSNode, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}

	return b.node.Object, nil
}
