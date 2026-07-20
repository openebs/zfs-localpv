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
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWaitForDevice(t *testing.T) {
	t.Run("returns immediately when the device already exists", func(t *testing.T) {
		dev := filepath.Join(t.TempDir(), "zd0")
		if err := os.WriteFile(dev, nil, 0600); err != nil {
			t.Fatal(err)
		}
		if err := waitForDevice(context.Background(), dev, time.Second); err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("returns once the device appears during the wait", func(t *testing.T) {
		dev := filepath.Join(t.TempDir(), "zd1")
		go func() {
			time.Sleep(150 * time.Millisecond)
			_ = os.WriteFile(dev, nil, 0600)
		}()
		if err := waitForDevice(context.Background(), dev, 3*time.Second); err != nil {
			t.Errorf("expected nil once device appears, got %v", err)
		}
	})

	t.Run("times out when the device never appears", func(t *testing.T) {
		dev := filepath.Join(t.TempDir(), "absent")
		if err := waitForDevice(context.Background(), dev, 150*time.Millisecond); err == nil {
			t.Error("expected a timeout error, got nil")
		}
	})

	t.Run("honors context cancellation instead of blocking for the full timeout", func(t *testing.T) {
		dev := filepath.Join(t.TempDir(), "absent")
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		done := make(chan error, 1)
		go func() { done <- waitForDevice(ctx, dev, 10*time.Second) }()
		select {
		case err := <-done:
			if err == nil {
				t.Error("expected a context error, got nil")
			}
		case <-time.After(2 * time.Second):
			t.Error("waitForDevice ignored context cancellation")
		}
	})
}

func TestZvolDeviceWaitTimeout(t *testing.T) {
	t.Setenv(ZvolDeviceWaitTimeoutEnv, "250ms")
	if got := zvolDeviceWaitTimeout(); got != 250*time.Millisecond {
		t.Errorf("timeout = %v, want 250ms", got)
	}
	t.Setenv(ZvolDeviceWaitTimeoutEnv, "not-a-duration")
	if got := zvolDeviceWaitTimeout(); got != zvolDeviceWaitDefaultTimeout {
		t.Errorf("invalid env should fall back to %v, got %v", zvolDeviceWaitDefaultTimeout, got)
	}
}
