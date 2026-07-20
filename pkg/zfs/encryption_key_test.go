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
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestValidateHexKey(t *testing.T) {
	valid := strings.Repeat("ab", 32) // 64 hex chars
	if err := validateHexKey(valid); err != nil {
		t.Errorf("valid key rejected: %v", err)
	}
	if err := validateHexKey(strings.Repeat("ab", 16)); err == nil {
		t.Error("short key accepted, want rejection")
	}
	if err := validateHexKey(strings.Repeat("zz", 32)); err == nil {
		t.Error("non-hex key accepted, want rejection")
	}
	if err := validateHexKey(strings.Repeat("ab", 31) + "aq"); err == nil || strings.Contains(err.Error(), "q") {
		t.Errorf("non-hex key: got %v, want an error that does not quote the key", err)
	}
}

func TestFetchKeyFromSecret(t *testing.T) {
	ns := "app"
	valid := strings.Repeat("ab", 32)

	t.Run("valid key (with trailing newline) is read and trimmed", func(t *testing.T) {
		client := fake.NewSimpleClientset(&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "k", Namespace: ns},
			Data:       map[string][]byte{encryptionSecretKeyField: []byte(valid + "\n")},
		})
		got, err := fetchKeyFromSecretWithClient(context.Background(), client, ns, "k")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != valid {
			t.Errorf("key = %q, want %q (trailing newline must be trimmed)", got, valid)
		}
	})

	t.Run("missing data key is rejected", func(t *testing.T) {
		client := fake.NewSimpleClientset(&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "k", Namespace: ns},
			Data:       map[string][]byte{"wrong": []byte(valid)},
		})
		if _, err := fetchKeyFromSecretWithClient(context.Background(), client, ns, "k"); err == nil {
			t.Error("expected error for missing data key, got nil")
		}
	})

	t.Run("invalid hex is rejected", func(t *testing.T) {
		client := fake.NewSimpleClientset(&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "k", Namespace: ns},
			Data:       map[string][]byte{encryptionSecretKeyField: []byte("not-a-valid-hex-key")},
		})
		if _, err := fetchKeyFromSecretWithClient(context.Background(), client, ns, "k"); err == nil {
			t.Error("expected error for invalid hex, got nil")
		}
	})

	t.Run("absent secret is an error", func(t *testing.T) {
		client := fake.NewSimpleClientset()
		if _, err := fetchKeyFromSecretWithClient(context.Background(), client, ns, "missing"); err == nil {
			t.Error("expected error for absent secret, got nil")
		}
	})
}
