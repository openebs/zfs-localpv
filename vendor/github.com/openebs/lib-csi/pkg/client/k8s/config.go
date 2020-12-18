/*
Copyright 2020 The OpenEBS Authors

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

package k8s

import (
	"strings"

	"github.com/openebs/lib-csi/pkg/common/env"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ConfigGetter abstracts fetching of kubernetes client config
type ConfigGetter interface {
	Get() (*rest.Config, error)
	Name() string
}

// configFromENV is an implementation of ConfigGetter
type configFromENV struct{}

// Name returns the name of this config getter instance
func (c *configFromENV) Name() string {
	return "k8s-config-from-env"
}

// Get returns kubernetes rest config based on kubernetes environment values
func (c *configFromENV) Get() (*rest.Config, error) {
	k8sMaster := env.Get(env.KubeMaster)
	kubeConfig := env.Get(env.KubeConfig)

	if len(strings.TrimSpace(k8sMaster)) == 0 && len(strings.TrimSpace(kubeConfig)) == 0 {
		return nil, errors.New("missing kubernetes master as well as kubeconfig: failed to get kubernetes client config")
	}

	return clientcmd.BuildConfigFromFlags(k8sMaster, kubeConfig)
}

// configFromREST is an implementation of ConfigGetter
type configFromREST struct{}

// Name returns the name of this config getter instance
func (c *configFromREST) Name() string {
	return "k8s-config-from-rest"
}

// Get returns kubernetes rest config based on in-cluster config implementation
func (c *configFromREST) Get() (*rest.Config, error) {
	return rest.InClusterConfig()
}

// ConfigGetters holds a list of ConfigGetter instances
//
// NOTE:
//  This is an implementation of ConfigGetter
type ConfigGetters []ConfigGetter

// Name returns the name of this config getter instance
func (c ConfigGetters) Name() string {
	return "list-of-k8s-config-getter"
}

// Get fetches the kubernetes client config that is used to make kubernetes API
// calls. It makes use of its list of getter instances to fetch kubernetes
// config.
func (c ConfigGetters) Get() (config *rest.Config, err error) {
	var errs []error
	for _, g := range c {
		config, err = g.Get()
		if err == nil {
			return
		}
		errs = append(errs, errors.Wrapf(err, "failed to get kubernetes client config via %s", g.Name()))
	}
	// at this point; all getters have failed
	err = errors.Errorf("%+v", errs)
	err = errors.Wrap(err, "failed to get kubernetes client config")
	return
}

// Config provides appropriate config getter instances that help in fetching
// kubernetes client config to invoke kubernetes API calls
func Config() ConfigGetter {
	return ConfigGetters{&configFromENV{}, &configFromREST{}}
}
