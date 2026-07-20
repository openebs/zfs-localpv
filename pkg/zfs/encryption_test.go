/*
Copyright 2024 The OpenEBS Authors

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

package zfs

import (
	"errors"
	"strings"
	"testing"

	apis "github.com/openebs/zfs-localpv/v2/pkg/apis/openebs.io/zfs/v1"
)

func hasArg(args []string, want string) bool {
	for _, a := range args {
		if a == want {
			return true
		}
	}
	return false
}

func TestUsesManagedKey(t *testing.T) {
	tests := []struct {
		name string
		vol  apis.VolumeInfo
		want bool
	}{
		{"secret ref", apis.VolumeInfo{EncryptionKeyRef: &apis.EncryptionKeyReference{Name: "k"}}, true},
		{"legacy keylocation", apis.VolumeInfo{KeyLocation: "file:///etc/key"}, false},
		{"none", apis.VolumeInfo{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vol := &apis.ZFSVolume{Spec: tt.vol}
			if got := UsesManagedKey(vol); got != tt.want {
				t.Errorf("UsesManagedKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildCreateArgs_ManagedKey(t *testing.T) {
	base := func() *apis.ZFSVolume {
		v := &apis.ZFSVolume{}
		v.Name = "pvc-1"
		v.Spec.PoolName = "pool"
		v.Spec.Capacity = "1073741824"
		v.Spec.Encryption = "aes-256-gcm"
		v.Spec.KeyFormat = "hex"
		return v
	}

	t.Run("zvol with secret ref uses keylocation=prompt", func(t *testing.T) {
		v := base()
		v.Spec.EncryptionKeyRef = &apis.EncryptionKeyReference{Name: "my-key", Namespace: "app"}
		args := buildZvolCreateArgs(v)
		if !hasArg(args, "keylocation=prompt") {
			t.Errorf("expected keylocation=prompt, got %v", args)
		}
		if !hasArg(args, "encryption=aes-256-gcm") || !hasArg(args, "keyformat=hex") {
			t.Errorf("expected encryption/keyformat args, got %v", args)
		}
		for _, a := range args {
			if strings.HasPrefix(a, "keylocation=file") {
				t.Errorf("managed key must not use a keylocation file, got %v", args)
			}
		}
	})

	t.Run("dataset with secret ref uses keylocation=prompt", func(t *testing.T) {
		v := base()
		v.Spec.VolumeType = VolTypeDataset
		v.Spec.EncryptionKeyRef = &apis.EncryptionKeyReference{Name: "my-key"}
		args := buildDatasetCreateArgs(v)
		if !hasArg(args, "keylocation=prompt") {
			t.Errorf("expected keylocation=prompt, got %v", args)
		}
	})

	t.Run("legacy keylocation file is preserved", func(t *testing.T) {
		v := base()
		v.Spec.KeyLocation = "file:///etc/zfs/key"
		args := buildZvolCreateArgs(v)
		if !hasArg(args, "keylocation=file:///etc/zfs/key") {
			t.Errorf("expected legacy keylocation file, got %v", args)
		}
		if hasArg(args, "keylocation=prompt") {
			t.Errorf("legacy mode must not use prompt, got %v", args)
		}
	})
}

func TestLoadKeyError(t *testing.T) {
	exit := errors.New("exit status 255")
	tests := []struct {
		name    string
		out     string
		err     error
		wantErr string
	}{
		{"loaded", "", nil, ""},
		{"loaded concurrently", "Key load error: Key already loaded for 'pool/pvc-a'.", exit, ""},
		{"wrong key", "Key load error: Incorrect key provided for 'pool/pvc-a'.", exit, "is incorrect"},
		{"other failure", "cannot open 'pool/pvc-a': dataset does not exist", exit, "zfs load-key"},
	}
	for _, tt := range tests {
		err := loadKeyError("pool/pvc-a", []byte(tt.out), tt.err)
		switch {
		case tt.wantErr == "" && err != nil:
			t.Errorf("%s: loadKeyError = %v, want nil", tt.name, err)
		case tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)):
			t.Errorf("%s: loadKeyError = %v, want an error containing %q", tt.name, err, tt.wantErr)
		}
	}
}

func TestEncryptedMarker(t *testing.T) {
	ref := &apis.EncryptionKeyReference{Name: "db-key", Namespace: "apps"}
	tests := []struct {
		name string
		spec apis.VolumeInfo
		want string
	}{
		{"not encrypted", apis.VolumeInfo{}, ""},
		{"per-volume key", apis.VolumeInfo{Encryption: "aes-256-gcm", EncryptionKeyRef: ref}, EncryptedVolume},
		{"storageclass keylocation", apis.VolumeInfo{Encryption: "aes-256-gcm", KeyLocation: "file:///keys/k"}, EncryptedStorageClass},
		{"encryption off", apis.VolumeInfo{Encryption: "off"}, ""},
		{"encryption off with a keylocation", apis.VolumeInfo{Encryption: "off", KeyLocation: "file:///keys/k"}, ""},
		{"key inherited from the parent dataset", apis.VolumeInfo{Encryption: "aes-256-gcm"}, ""},
	}
	for _, tt := range tests {
		vol := &apis.ZFSVolume{Spec: tt.spec}
		if got := EncryptedMarker(vol); got != tt.want {
			t.Errorf("%s: EncryptedMarker = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestEncryptionEnabled(t *testing.T) {
	for value, want := range map[string]bool{"": false, "off": false, "on": true, "aes-256-gcm": true} {
		if got := EncryptionEnabled(value); got != want {
			t.Errorf("EncryptionEnabled(%q) = %v, want %v", value, got, want)
		}
	}
}
