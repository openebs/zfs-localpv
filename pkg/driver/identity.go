/*
Copyright © 2019 The OpenEBS Authors

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
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/openebs/zfs-localpv/pkg/version"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// identity is the server implementation
// for CSI IdentityServer
type identity struct {
	driver *CSIDriver
}

// NewIdentity returns a new instance of CSI
// IdentityServer
func NewIdentity(d *CSIDriver) csi.IdentityServer {
	return &identity{
		driver: d,
	}
}

// GetPluginInfo returns the version and name of
// this service
//
// This implements csi.IdentityServer
func (id *identity) GetPluginInfo(
	ctx context.Context,
	req *csi.GetPluginInfoRequest,
) (*csi.GetPluginInfoResponse, error) {

	if id.driver.config.DriverName == "" {
		return nil, status.Error(codes.Unavailable, "missing driver name")
	}

	if id.driver.config.Version == "" {
		return nil, status.Error(codes.Unavailable, "missing driver version")
	}

	return &csi.GetPluginInfoResponse{
		Name: id.driver.config.DriverName,
		// TODO
		// verify which version needs to be used:
		// config.version or version.Current()
		VendorVersion: version.Current(),
	}, nil
}

// TODO
// Need to implement this
//
// # Probe checks if the plugin is running or not
//
// This implements csi.IdentityServer
func (id *identity) Probe(
	ctx context.Context,
	req *csi.ProbeRequest,
) (*csi.ProbeResponse, error) {

	return &csi.ProbeResponse{}, nil
}

// GetPluginCapabilities returns supported capabilities
// of this plugin
//
// Currently it reports whether this plugin can serve
// the Controller interface. Controller interface methods
// are called dependant on this
//
// This implements csi.IdentityServer
func (id *identity) GetPluginCapabilities(
	ctx context.Context,
	req *csi.GetPluginCapabilitiesRequest,
) (*csi.GetPluginCapabilitiesResponse, error) {

	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_VOLUME_ACCESSIBILITY_CONSTRAINTS,
					},
				},
			},
		},
	}, nil
}
