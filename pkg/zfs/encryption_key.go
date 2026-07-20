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
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// encryptionSecretKeyField is the Secret data key holding a volume key.
	encryptionSecretKeyField = "key"
	// hexKeyLen is the length of a 32-byte key encoded as hex (keyformat=hex).
	hexKeyLen = 64
)

// validateHexKey checks that key is 64 hex characters (32 bytes).
func validateHexKey(key string) error {
	if len(key) != hexKeyLen {
		return fmt.Errorf("encryption key must be %d hex chars (32 bytes), got %d", hexKeyLen, len(key))
	}
	// the decoder error quotes the offending byte: never echo key material
	if _, err := hex.DecodeString(key); err != nil {
		return errors.New("encryption key is not valid hex")
	}
	return nil
}

// fetchKeyFromSecret reads and validates the key held by a Secret.
func fetchKeyFromSecret(ctx context.Context, namespace, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("zfs: no encryption key secret referenced")
	}
	client, err := getKubeClient()
	if err != nil {
		return "", err
	}
	return fetchKeyFromSecretWithClient(ctx, client, namespace, name)
}

// fetchKeyFromSecretWithClient is fetchKeyFromSecret with a given client.
func fetchKeyFromSecretWithClient(ctx context.Context, client kubernetes.Interface, namespace, name string) (string, error) {
	if namespace == "" {
		return "", fmt.Errorf("zfs: encryption key secret %q has no namespace", name)
	}
	secret, err := client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("zfs: get encryption key secret %s/%s: %w", namespace, name, err)
	}
	raw, ok := secret.Data[encryptionSecretKeyField]
	if !ok {
		return "", fmt.Errorf("zfs: encryption key secret %s/%s has no %q data key",
			namespace, name, encryptionSecretKeyField)
	}
	key := strings.TrimSpace(string(raw))
	if err := validateHexKey(key); err != nil {
		return "", fmt.Errorf("zfs: encryption key secret %s/%s: %w (check for a trailing newline or extra characters)",
			namespace, name, err)
	}
	return key, nil
}
