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

package usage

import (
	k8sapi "github.com/openebs/lib-csi/pkg/client/k8s"
)

// Usage struct represents all information about a usage metric sent to
// Google Analytics with respect to the application
type Usage struct {
	// Embedded Event struct as we are currently only sending hits of type
	// 'event'
	Event

	// https://developers.google.com/analytics/devguides/collection/protocol/v1/parameters#an
	// use-case: cstor or jiva volume, or m-apiserver application
	// Embedded field for application
	Application

	// Embedded Gclient struct
	Gclient
}

// Event is a represents usage of OpenEBS
// Event contains all the query param fields when hits is of type='event'
// Ref: https://developers.google.com/analytics/devguides/collection/protocol/v1/parameters#ec
type Event struct {
	// (Required) Event Category, ec
	category string
	// (Required) Event Action, ea
	action string
	// (Optional) Event Label, el
	label string
	// (Optional) Event vallue, ev
	// Non negative
	value int64
}

// NewEvent returns an Event struct with eventCategory, eventAction,
// eventLabel, eventValue fields
func (u *Usage) NewEvent(c, a, l string, v int64) *Usage {
	u.category = c
	u.action = a
	u.label = l
	u.value = v
	return u
}

// Application struct holds details about the Application
type Application struct {
	// eg. project version
	appVersion string

	// eg. kubernetes version
	appInstallerID string

	// Name of the application, usage(OpenEBS/NDM)
	appID string

	// eg. usage(os-type/architecture) of system or volume's CASType
	appName string
}

// Gclient struct represents a Google Analytics hit
type Gclient struct {
	// constant tracking-id used to send a hit
	trackID string

	// anonymous client-id
	clientID string

	// anonymous campaign source
	campaignSource string

	// anonymous campaign name
	campaignName string

	// https://developers.google.com/analytics/devguides/collection/protocol/v1/parameters#ds
	// (usecase) node-detail
	dataSource string

	// Document-title property in Google Analytics
	// https://developers.google.com/analytics/devguides/collection/protocol/v1/parameters#dt
	// use-case: uuid of the volume objects or a uuid to anonymously tell objects apart
	documentTitle string
}

// New returns an instance of Usage
func New() *Usage {
	return &Usage{}
}

// SetDataSource : usage(os-type, kernel)
func (u *Usage) SetDataSource(dataSource string) *Usage {
	u.dataSource = dataSource
	return u
}

// SetTrackingID Sets the GA-code for the project
func (u *Usage) SetTrackingID(track string) *Usage {
	u.trackID = track
	return u
}

// SetCampaignSource : source of openebs installater like:
// helm or operator etc. This will have to be configured
// via ENV variable OPENEBS_IO_INSTALLER_TYPE
func (u *Usage) SetCampaignSource(campaignSrc string) *Usage {
	u.campaignSource = campaignSrc
	return u
}

// SetDocumentTitle : usecase(anonymous-id)
func (u *Usage) SetDocumentTitle(documentTitle string) *Usage {
	u.documentTitle = documentTitle
	return u
}

// SetApplicationName : usecase(os-type/arch, volume CASType)
func (u *Usage) SetApplicationName(appName string) *Usage {
	u.appName = appName
	return u
}

// SetCampaignName : set the name of the PVC or will be empty.
func (u *Usage) SetCampaignName(campaignName string) *Usage {
	u.campaignName = campaignName
	return u
}

// SetApplicationID : usecase(OpenEBS/NDM)
func (u *Usage) SetApplicationID(appID string) *Usage {
	u.appID = appID
	return u
}

// SetApplicationVersion : usecase(project-version)
func (u *Usage) SetApplicationVersion(appVersion string) *Usage {
	u.appVersion = appVersion
	return u
}

// SetApplicationInstallerID : usecase(k8s-version)
func (u *Usage) SetApplicationInstallerID(appInstallerID string) *Usage {
	u.appInstallerID = appInstallerID
	return u
}

// SetClientID sets the anonymous user id
func (u *Usage) SetClientID(userID string) *Usage {
	u.clientID = userID
	return u
}

// SetCategory sets the category of an event
func (u *Usage) SetCategory(c string) *Usage {
	u.category = c
	return u
}

// SetAction sets the action of an event
func (u *Usage) SetAction(a string) *Usage {
	u.action = a
	return u
}

// SetLabel sets the label for an event
func (u *Usage) SetLabel(l string) *Usage {
	u.label = l
	return u
}

// SetValue sets the value for an event's label
func (u *Usage) SetValue(v int64) *Usage {
	u.value = v
	return u
}

// Build is a builder method for Usage struct
func (u *Usage) Build() *Usage {
	// Default ApplicationID for openebs project is OpenEBS
	v := NewVersion()
	v.getVersion(false)
	u.SetApplicationID(AppName).
		SetTrackingID(GAclientID).
		SetClientID(v.id).
		SetCampaignSource(v.installerType)
	// TODO: Add condition for version over-ride
	// Case: CAS/Jiva version, etc
	return u
}

// ApplicationBuilder Application builder is used for adding k8s&openebs environment detail
// for non install events
func (u *Usage) ApplicationBuilder() *Usage {
	v := NewVersion()
	v.getVersion(false)
	u.SetApplicationVersion(v.openebsVersion).
		SetApplicationName(v.k8sArch).
		SetApplicationInstallerID(v.k8sVersion).
		SetDataSource(v.nodeType)
	return u
}

// SetVolumeCapacity sets the storage capacity of the volume for a volume event
func (u *Usage) SetVolumeCapacity(volCapG string) *Usage {
	s, _ := toGigaUnits(volCapG)
	u.SetValue(s)
	return u
}

// SetVolumeType Wrapper for setting the default storage-engine for volume-provision event
func (u *Usage) SetVolumeType(volType, method string) *Usage {
	if method == VolumeProvision && volType == "" {
		// Set the default storage engine, if not specified in the request
		u.SetApplicationName(DefaultCASType)
	} else {
		u.SetApplicationName(volType)
	}
	return u
}

// SetReplicaCount Wrapper for setting replica count for volume events
// NOTE: This doesn't get the replica count in a volume de-provision event.
// TODO: Pick the current value of replica-count from the CAS-engine
func (u *Usage) SetReplicaCount(count, method string) *Usage {
	if method == VolumeProvision && count == "" {
		// Case: When volume-provision the replica count isn't specified
		// it is set to three by default by the m-apiserver
		u.SetAction(DefaultReplicaCount)
	} else {
		// Catch all case for volume-deprovision event and
		// volume-provision event with an overridden replica-count
		u.SetAction(Replica + count)
	}
	return u
}

// InstallBuilder is a concrete builder for install events
func (u *Usage) InstallBuilder(override bool) *Usage {
	v := NewVersion()
	clusterSize, _ := k8sapi.NumberOfNodes()
	v.getVersion(override)
	u.SetApplicationVersion(v.openebsVersion).
		SetApplicationName(v.k8sArch).
		SetApplicationInstallerID(v.k8sVersion).
		SetDataSource(v.nodeType).
		SetDocumentTitle(v.id).
		SetApplicationID(AppName).
		NewEvent(InstallEvent, RunningStatus, EventLabelNode, int64(clusterSize))
	return u
}
