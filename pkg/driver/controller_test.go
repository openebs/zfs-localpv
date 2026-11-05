/*
Copyright 2020 The OpenEBS Authors.

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
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	zfsapi "github.com/openebs/zfs-localpv/v2/pkg/apis/openebs.io/zfs/v1"
	"github.com/openebs/zfs-localpv/v2/pkg/zfs"
)

func TestRoundOff(t *testing.T) {

	tests := map[string]struct {
		input    int64
		expected int64
	}{
		"Minimum allocatable is 1Mi": {input: 1, expected: Mi},
		"roundOff to same Mi size":   {input: Mi, expected: Mi},
		"roundOff to nearest Mi":     {input: Mi + 1, expected: Mi * 2},
		"roundOff to same Gi size":   {input: Gi, expected: Gi},
		"roundOff to nearest Gi":     {input: Gi + 1, expected: Gi * 2},
		"roundOff MB size":           {input: 5 * MB, expected: 5 * Mi},
		"roundOff GB size":           {input: 5 * GB, expected: 5 * Gi},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, getRoundedCapacity(test.input))
		})
	}
}

func TestVerifyExpandCapacity(t *testing.T) {
	pools := []zfsapi.Pool{
		{Name: "zpool", Free: *resource.NewQuantity(10*Gi, resource.BinarySI)},
		{Name: "other", Free: *resource.NewQuantity(100*Gi, resource.BinarySI)},
	}
	newVol := func(pool, vtype, thin string) *zfsapi.ZFSVolume {
		return &zfsapi.ZFSVolume{
			ObjectMeta: metav1.ObjectMeta{Name: "pvc-1"},
			Spec:       zfsapi.VolumeInfo{PoolName: pool, VolumeType: vtype, ThinProvision: thin},
		}
	}
	zvol, dataset := zfs.VolTypeZVol, zfs.VolTypeDataset

	tests := map[string]struct {
		vol        *zfsapi.ZFSVolume
		additional int64
		code       codes.Code
	}{
		"fits in the pool":                       {vol: newVol("zpool", zvol, "no"), additional: 5 * Gi, code: codes.OK},
		"exceeds the free capacity":              {vol: newVol("zpool", zvol, "no"), additional: 10*Gi + 1, code: codes.ResourceExhausted},
		"dataset below the pool":                 {vol: newVol("zpool/k8s", zvol, "no"), additional: 20 * Gi, code: codes.ResourceExhausted},
		"thick dataset is checked":               {vol: newVol("zpool", dataset, "no"), additional: 20 * Gi, code: codes.ResourceExhausted},
		"zvol with default thinprovision":        {vol: newVol("zpool", zvol, ""), additional: 20 * Gi, code: codes.ResourceExhausted},
		"dataset with default thinprovision":     {vol: newVol("zpool", dataset, ""), additional: 20 * Gi, code: codes.OK},
		"thin provisioned zvol is not checked":   {vol: newVol("zpool", zvol, "yes"), additional: 20 * Gi, code: codes.OK},
		"thin provisioned dataset isn't checked": {vol: newVol("zpool", dataset, "yes"), additional: 20 * Gi, code: codes.OK},
		"unknown pool is not checked":            {vol: newVol("missing", zvol, "no"), additional: 20 * Gi, code: codes.OK},
		"nothing additional is not checked":      {vol: newVol("zpool", zvol, "no"), additional: 0, code: codes.OK},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := verifyExpandCapacity(pools, test.vol, test.additional)
			assert.Equal(t, test.code, status.Code(err))
		})
	}
}
