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

package v1alpha1

import (
	"testing"

	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func fakeGetClientsetOk(c *rest.Config) (*kubernetes.Clientset, error) {
	return &kubernetes.Clientset{}, nil
}

func fakeGetClientsetErr(c *rest.Config) (*kubernetes.Clientset, error) {
	return nil, errors.New("fake error")
}

func fakeInClusterConfigOk() (*rest.Config, error) {
	return &rest.Config{}, nil
}

func fakeInClusterConfigErr() (*rest.Config, error) {
	return nil, errors.New("fake error")
}

func fakeBuildConfigFromFlagsOk(kubemaster string, kubeconfig string) (*rest.Config, error) {
	return &rest.Config{}, nil
}

func fakeBuildConfigFromFlagsErr(kubemaster string, kubeconfig string) (*rest.Config, error) {
	return nil, errors.New("fake error")
}

func fakeGetKubeConfigPathOk(e string) string {
	return "fake"
}

func fakeGetKubeConfigPathNil(e string) string {
	return ""
}

func fakeGetKubeMasterIPOk(e string) string {
	return "fake"
}

func fakeGetKubeMasterIPNil(e string) string {
	return ""
}

func fakeGetDynamicClientSetOk(c *rest.Config) (dynamic.Interface, error) {
	return dynamic.NewForConfig(c)
}

func fakeGetDynamicClientSetNil(c *rest.Config) (dynamic.Interface, error) {
	return nil, nil
}

func fakeGetDynamicClientSetErr(c *rest.Config) (dynamic.Interface, error) {
	return nil, errors.New("fake error")
}

func TestNewInCluster(t *testing.T) {
	c := New(InCluster())
	if !c.IsInCluster {
		t.Fatalf("test failed: expected IsInCluster as 'true' actual '%t'", c.IsInCluster)
	}
}

func TestConfig(t *testing.T) {
	tests := map[string]struct {
		isInCluster        bool
		kubeConfigPath     string
		getInClusterConfig getInClusterConfigFunc
		getKubeMasterIP    getKubeMasterIPFunc
		getKubeConfigPath  getKubeConfigPathFunc
		getConfigFromENV   buildConfigFromFlagsFunc
		isErr              bool
	}{
		"t1": {true, "", fakeInClusterConfigOk, nil, nil, nil, false},
		"t2": {true, "", fakeInClusterConfigErr, nil, nil, nil, true},
		"t3": {false, "", fakeInClusterConfigErr, fakeGetKubeMasterIPNil, fakeGetKubeConfigPathNil, nil, true},
		"t4": {false, "", fakeInClusterConfigOk, fakeGetKubeMasterIPNil, fakeGetKubeConfigPathNil, nil, false},
		"t5": {false, "fakeKubeConfigPath", nil, fakeGetKubeMasterIPOk, fakeGetKubeConfigPathNil, fakeBuildConfigFromFlagsOk, false},
		"t6": {false, "", nil, fakeGetKubeMasterIPNil, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsOk, false},
		"t7": {false, "", nil, fakeGetKubeMasterIPOk, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsOk, false},
		"t8": {false, "fakeKubeConfigPath", nil, fakeGetKubeMasterIPOk, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsErr, true},
		"t9": {false, "fakeKubeConfigpath", nil, fakeGetKubeMasterIPOk, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsOk, false},
	}
	for name, mock := range tests {
		name, mock := name, mock // pin It
		t.Run(name, func(t *testing.T) {
			c := &Client{
				IsInCluster:          mock.isInCluster,
				KubeConfigPath:       mock.kubeConfigPath,
				getInClusterConfig:   mock.getInClusterConfig,
				getKubeMasterIP:      mock.getKubeMasterIP,
				getKubeConfigPath:    mock.getKubeConfigPath,
				buildConfigFromFlags: mock.getConfigFromENV,
			}
			_, err := c.Config()
			if mock.isErr && err == nil {
				t.Fatalf("test '%s' failed: expected no error actual '%s'", name, err)
			}
		})
	}
}

func TestGetConfigFromENV(t *testing.T) {
	tests := map[string]struct {
		getKubeMasterIP   getKubeMasterIPFunc
		getKubeConfigPath getKubeConfigPathFunc
		getConfigFromENV  buildConfigFromFlagsFunc
		isErr             bool
	}{
		"t1": {fakeGetKubeMasterIPNil, fakeGetKubeConfigPathNil, nil, true},
		"t2": {fakeGetKubeMasterIPNil, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsOk, false},
		"t3": {fakeGetKubeMasterIPOk, fakeGetKubeConfigPathNil, fakeBuildConfigFromFlagsOk, false},
		"t4": {fakeGetKubeMasterIPOk, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsOk, false},
		"t5": {fakeGetKubeMasterIPNil, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsErr, true},
		"t6": {fakeGetKubeMasterIPOk, fakeGetKubeConfigPathNil, fakeBuildConfigFromFlagsErr, true},
		"t7": {fakeGetKubeMasterIPOk, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsErr, true},
	}
	for name, mock := range tests {
		name, mock := name, mock // pin It
		t.Run(name, func(t *testing.T) {
			c := &Client{
				getKubeMasterIP:      mock.getKubeMasterIP,
				getKubeConfigPath:    mock.getKubeConfigPath,
				buildConfigFromFlags: mock.getConfigFromENV,
			}
			_, err := c.getConfigFromENV()
			if mock.isErr && err == nil {
				t.Fatalf("test '%s' failed: expected error actual no error", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("test '%s' failed: expected no error actual '%s'", name, err)
			}
		})
	}
}

func TestGetConfigFromPathOrDirect(t *testing.T) {
	tests := map[string]struct {
		kubeConfigPath     string
		getConfigFromFlags buildConfigFromFlagsFunc
		getInClusterConfig getInClusterConfigFunc
		isErr              bool
	}{
		"T1": {"", fakeBuildConfigFromFlagsErr, fakeInClusterConfigOk, false},
		"T2": {"fake-path", fakeBuildConfigFromFlagsOk, fakeInClusterConfigErr, false},
		"T3": {"fake-path", fakeBuildConfigFromFlagsErr, fakeInClusterConfigOk, true},
		"T4": {"", fakeBuildConfigFromFlagsOk, fakeInClusterConfigErr, true},
		"T5": {"fake-path", fakeBuildConfigFromFlagsErr, fakeInClusterConfigErr, true},
	}
	for name, mock := range tests {
		name, mock := name, mock // pin It
		t.Run(name, func(t *testing.T) {
			c := &Client{
				KubeConfigPath:       mock.kubeConfigPath,
				buildConfigFromFlags: mock.getConfigFromFlags,
				getInClusterConfig:   mock.getInClusterConfig,
				getKubeMasterIP:      fakeGetKubeMasterIPNil,
				getKubeConfigPath:    fakeGetKubeConfigPathNil,
			}
			_, err := c.GetConfigForPathOrDirect()
			if mock.isErr && err == nil {
				t.Fatalf("test '%s' failed: expected error actual no error", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("test '%s' failed: expected no error actual '%s'", name, err)
			}
		})
	}
}

func TestClientset(t *testing.T) {
	tests := map[string]struct {
		isInCluster            bool
		kubeConfigPath         string
		getInClusterConfig     getInClusterConfigFunc
		getKubeMasterIP        getKubeMasterIPFunc
		getKubeConfigPath      getKubeConfigPathFunc
		getConfigFromENV       buildConfigFromFlagsFunc
		getKubernetesClientset getKubernetesClientsetFunc
		isErr                  bool
	}{
		"t10": {true, "", fakeInClusterConfigOk, nil, nil, nil, fakeGetClientsetOk, false},
		"t11": {true, "", fakeInClusterConfigOk, nil, nil, nil, fakeGetClientsetErr, true},
		"t12": {true, "", fakeInClusterConfigErr, nil, nil, nil, fakeGetClientsetOk, true},

		"t21": {false, "", nil, fakeGetKubeMasterIPOk, fakeGetKubeConfigPathNil, fakeBuildConfigFromFlagsOk, fakeGetClientsetOk, false},
		"t22": {false, "", nil, fakeGetKubeMasterIPNil, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsOk, fakeGetClientsetOk, false},
		"t23": {false, "", nil, fakeGetKubeMasterIPOk, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsOk, fakeGetClientsetOk, false},
		"t24": {false, "fake-path", nil, fakeGetKubeMasterIPOk, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsErr, fakeGetClientsetOk, true},
		"t25": {false, "", nil, fakeGetKubeMasterIPOk, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsOk, fakeGetClientsetErr, true},
		"t26": {false, "fakePath", nil, fakeGetKubeMasterIPOk, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsErr, fakeGetClientsetOk, true},

		"t30": {false, "", fakeInClusterConfigOk, fakeGetKubeMasterIPNil, fakeGetKubeConfigPathNil, nil, fakeGetClientsetOk, false},
		"t31": {false, "", fakeInClusterConfigOk, fakeGetKubeMasterIPNil, fakeGetKubeConfigPathNil, nil, fakeGetClientsetErr, true},
		"t32": {false, "", fakeInClusterConfigErr, fakeGetKubeMasterIPNil, fakeGetKubeConfigPathNil, nil, nil, true},
		"t33": {false, "fakePath", nil, fakeGetKubeMasterIPOk, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsOk, fakeGetClientsetOk, false},
	}
	for name, mock := range tests {
		name, mock := name, mock // pin It
		t.Run(name, func(t *testing.T) {
			c := &Client{
				IsInCluster:            mock.isInCluster,
				KubeConfigPath:         mock.kubeConfigPath,
				getInClusterConfig:     mock.getInClusterConfig,
				getKubeMasterIP:        mock.getKubeMasterIP,
				getKubeConfigPath:      mock.getKubeConfigPath,
				buildConfigFromFlags:   mock.getConfigFromENV,
				getKubernetesClientset: mock.getKubernetesClientset,
			}
			_, err := c.Clientset()
			if mock.isErr && err == nil {
				t.Fatalf("test '%s' failed: expected error actual no error", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("test '%s' failed: expected no error actual '%s'", name, err)
			}
		})
	}
}

func TestDynamic(t *testing.T) {
	tests := map[string]struct {
		getKubeMasterIP               getKubeMasterIPFunc
		getInClusterConfig            getInClusterConfigFunc
		getKubernetesDynamicClientSet getKubernetesDynamicClientFunc
		kubeConfigPath                string
		getConfigFromENV              buildConfigFromFlagsFunc
		getKubeConfigPath             getKubeConfigPathFunc
		isErr                         bool
	}{
		"t1": {fakeGetKubeMasterIPNil, fakeInClusterConfigErr, fakeGetDynamicClientSetOk, "fake-path", fakeBuildConfigFromFlagsOk, fakeGetKubeConfigPathNil, false},
		"t2": {fakeGetKubeMasterIPNil, fakeInClusterConfigErr, fakeGetDynamicClientSetErr, "fake-path", fakeBuildConfigFromFlagsOk, fakeGetKubeConfigPathOk, true},
		"t3": {fakeGetKubeMasterIPNil, fakeInClusterConfigErr, fakeGetDynamicClientSetOk, "fake-path", fakeBuildConfigFromFlagsErr, fakeGetKubeConfigPathOk, true},
		"t4": {fakeGetKubeMasterIPOk, fakeInClusterConfigOk, fakeGetDynamicClientSetOk, "", fakeBuildConfigFromFlagsOk, fakeGetKubeConfigPathOk, false},
		"t5": {fakeGetKubeMasterIPOk, fakeInClusterConfigErr, fakeGetDynamicClientSetErr, "", fakeBuildConfigFromFlagsOk, fakeGetKubeConfigPathOk, true},
		"t6": {fakeGetKubeMasterIPNil, fakeInClusterConfigOk, fakeGetDynamicClientSetErr, "", fakeBuildConfigFromFlagsErr, fakeGetKubeConfigPathNil, true},
		"t7": {fakeGetKubeMasterIPNil, fakeInClusterConfigErr, fakeGetDynamicClientSetOk, "", fakeBuildConfigFromFlagsErr, fakeGetKubeConfigPathNil, true},
		"t8": {fakeGetKubeMasterIPNil, fakeInClusterConfigErr, fakeGetDynamicClientSetErr, "", fakeBuildConfigFromFlagsErr, fakeGetKubeConfigPathNil, true},
	}
	for name, mock := range tests {
		name, mock := name, mock // pin It
		t.Run(name, func(t *testing.T) {
			c := &Client{
				getKubeMasterIP:            mock.getKubeMasterIP,
				KubeConfigPath:             mock.kubeConfigPath,
				getInClusterConfig:         mock.getInClusterConfig,
				buildConfigFromFlags:       mock.getConfigFromENV,
				getKubeConfigPath:          mock.getKubeConfigPath,
				getKubernetesDynamicClient: mock.getKubernetesDynamicClientSet,
			}
			_, err := c.Dynamic()
			if mock.isErr && err == nil {
				t.Fatalf("test '%s' failed: expected error actual no error", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("test '%s' failed: expected no error but got '%v'", name, err)
			}
		})
	}
}

func TestConfigForPath(t *testing.T) {
	tests := map[string]struct {
		kubeConfigPath    string
		getConfigFromPath buildConfigFromFlagsFunc
		isErr             bool
	}{
		"T1": {"", fakeBuildConfigFromFlagsErr, true},
		"T2": {"fake-path", fakeBuildConfigFromFlagsOk, false},
	}
	for name, mock := range tests {
		name, mock := name, mock // pin It
		t.Run(name, func(t *testing.T) {
			c := &Client{
				KubeConfigPath:       mock.kubeConfigPath,
				buildConfigFromFlags: mock.getConfigFromPath,
			}
			_, err := c.ConfigForPath(mock.kubeConfigPath)
			if mock.isErr && err == nil {
				t.Fatalf("test '%s' failed: expected error actual no error", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("test '%s' failed: expected no error but got '%v'", name, err)
			}
		})
	}
}
