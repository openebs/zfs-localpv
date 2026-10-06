/*
Copyright 2026 The OpenEBS Authors.

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

package driver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/openebs/lib-csi/pkg/common/env"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	zfsapi "github.com/openebs/zfs-localpv/v2/pkg/apis/openebs.io/zfs/v1"
	"github.com/openebs/zfs-localpv/v2/pkg/zfs"
)

const (
	testNamespace = "openebs"
	testPool      = "zfspv-pool"
	testCapacity  = "1073741824" // 1Gi, already rounded
)

// fakeAPIServer serves the zfs.openebs.io/v1 resources of one namespace from
// memory. The driver builds its clients from the environment, so the tests
// point OPENEBS_IO_K8S_MASTER here.
type fakeAPIServer struct {
	mu      sync.Mutex
	objects map[string]map[string]map[string]interface{} // resource -> name -> object
}

func newFakeAPIServer(t *testing.T) *fakeAPIServer {
	t.Helper()

	f := &fakeAPIServer{objects: map[string]map[string]map[string]interface{}{
		"zfsvolumes":   {},
		"zfssnapshots": {},
	}}
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)
	t.Setenv(env.KubeMaster, srv.URL)
	t.Setenv(env.KubeConfig, "")

	ns := zfs.OpenEBSNamespace
	zfs.OpenEBSNamespace = testNamespace
	t.Cleanup(func() { zfs.OpenEBSNamespace = ns })

	return f
}

func (f *fakeAPIServer) add(t *testing.T, resource string, obj interface{}) {
	t.Helper()

	raw, err := json.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.objects[resource][m["metadata"].(map[string]interface{})["name"].(string)] = m
}

func (f *fakeAPIServer) has(resource, name string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, ok := f.objects[resource][name]
	return ok
}

func (f *fakeAPIServer) get(t *testing.T, resource, name string, into interface{}) {
	t.Helper()

	f.mu.Lock()
	obj, ok := f.objects[resource][name]
	f.mu.Unlock()
	if !ok {
		t.Fatalf("%s %s not found", resource, name)
	}
	raw, _ := json.Marshal(obj)
	if err := json.Unmarshal(raw, into); err != nil {
		t.Fatal(err)
	}
}

func (f *fakeAPIServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	prefix := "/apis/zfs.openebs.io/v1/namespaces/" + testNamespace + "/"
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, prefix), "/")
	store, ok := f.objects[parts[0]]
	if !strings.HasPrefix(r.URL.Path, prefix) || !ok || len(parts) > 2 {
		http.Error(w, "unexpected path "+r.URL.Path, http.StatusBadRequest)
		return
	}
	resource := schema.GroupResource{Group: "zfs.openebs.io", Resource: parts[0]}

	w.Header().Set("Content-Type", "application/json")
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			f.list(w, store, r.URL.Query().Get("labelSelector"))
		case http.MethodPost:
			var obj map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&obj); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			name := obj["metadata"].(map[string]interface{})["name"].(string)
			if _, exists := store[name]; exists {
				writeStatus(w, k8serror.NewAlreadyExists(resource, name))
				return
			}
			if parts[0] == "zfsvolumes" {
				// stand in for the node agent creating the dataset
				obj["status"] = map[string]interface{}{"state": zfs.ZFSStatusReady}
			}
			store[name] = obj
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(obj)
		default:
			http.Error(w, "unexpected method "+r.Method, http.StatusMethodNotAllowed)
		}
		return
	}

	name := parts[1]
	obj, exists := store[name]
	if !exists {
		writeStatus(w, k8serror.NewNotFound(resource, name))
		return
	}
	switch r.Method {
	case http.MethodGet:
		_ = json.NewEncoder(w).Encode(obj)
	case http.MethodDelete:
		delete(store, name)
		_ = json.NewEncoder(w).Encode(metav1.Status{Status: metav1.StatusSuccess})
	default:
		http.Error(w, "unexpected method "+r.Method, http.StatusMethodNotAllowed)
	}
}

// list supports the single key=value selectors the driver uses
func (f *fakeAPIServer) list(w http.ResponseWriter, store map[string]map[string]interface{}, selector string) {
	key, value, _ := strings.Cut(selector, "=")
	items := []interface{}{}
	for _, obj := range store {
		labels, _ := obj["metadata"].(map[string]interface{})["labels"].(map[string]interface{})
		if selector == "" || labels[key] == value {
			items = append(items, obj)
		}
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"apiVersion": "zfs.openebs.io/v1",
		"metadata":   map[string]interface{}{},
		"items":      items,
	})
}

func writeStatus(w http.ResponseWriter, err *k8serror.StatusError) {
	w.WriteHeader(int(err.ErrStatus.Code))
	status := err.ErrStatus
	status.Kind = "Status"
	status.APIVersion = "v1"
	_ = json.NewEncoder(w).Encode(status)
}

// testSnapshot returns a ZFSSnapshot labelled with volume, or unlabelled if
// volume is empty
func testSnapshot(name, volume string) *zfsapi.ZFSSnapshot {
	snap := &zfsapi.ZFSSnapshot{
		TypeMeta: metav1.TypeMeta{APIVersion: "zfs.openebs.io/v1", Kind: "ZFSSnapshot"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: testNamespace,
		},
		Spec: zfsapi.VolumeInfo{
			PoolName:    testPool,
			Capacity:    testCapacity,
			OwnerNodeID: "node-1",
		},
		Status: zfsapi.SnapStatus{State: zfs.ZFSStatusReady},
	}
	if volume != "" {
		snap.Labels = map[string]string{zfs.ZFSVolKey: volume}
	}
	return snap
}

func TestCreateSnapCloneUsesSnapshotLabel(t *testing.T) {
	tests := map[string]struct {
		label    string
		wantSnap string
	}{
		// zfs promote pvc-b moved snapshot-1 from pvc-a to pvc-b
		"promoted":  {label: "pvc-b", wantSnap: "pvc-b@snapshot-1"},
		"unchanged": {label: "pvc-a", wantSnap: "pvc-a@snapshot-1"},
		"no label":  {label: "", wantSnap: "pvc-a@snapshot-1"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			api := newFakeAPIServer(t)
			api.add(t, "zfssnapshots", testSnapshot("snapshot-1", test.label))

			req := &csi.CreateVolumeRequest{
				Name:          "pvc-clone",
				CapacityRange: &csi.CapacityRange{RequiredBytes: Gi},
				Parameters:    map[string]string{"poolname": testPool},
			}
			if _, err := CreateSnapClone(context.Background(), req, "pvc-a@snapshot-1"); err != nil {
				t.Fatal(err)
			}

			var vol zfsapi.ZFSVolume
			api.get(t, "zfsvolumes", "pvc-clone", &vol)
			if vol.Spec.SnapName != test.wantSnap {
				t.Fatalf("clone source: want %s, got %s", test.wantSnap, vol.Spec.SnapName)
			}
		})
	}
}
