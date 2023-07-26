package patcher

import (
	"context"
	"fmt"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"os"
)

type Patchable struct {
	inner dynamic.NamespaceableResourceInterface
}

func PatchCrdOrIgnore(gvr schema.GroupVersionResource, namespace string) *Patchable {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate in-cluster config: %s", err.Error())
		return nil
	}

	client, err := dynamic.NewForConfig(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate in-cluster kubernetes client: %s", err.Error())
		return nil
	}

	return &Patchable{client.Resource(gvr)}
}

func (p *Patchable) WithMergePatchFrom(name string, original, modified []byte) {
	if p == nil {
		return
	}

	patch, err := strategicpatch.CreateTwoWayMergePatch(original, modified, apiextensions.CustomResourceDefinition{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create merge patch for %s: %s", name, err.Error())
		return
	}
	_, err = p.inner.Patch(context.TODO(), name, k8stypes.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to patch resource %s: %s", name, err.Error())
		return
	}

	fmt.Printf("successfully patched CRD %s\n", name)
}
