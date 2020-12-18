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
	"fmt"
	"time"

	"github.com/openebs/lib-csi/pkg/common/env"
)

// OpenEBSPingPeriod  ping interval of volume io analytics
var OpenEBSPingPeriod = "OPENEBS_IO_ANALYTICS_PING_INTERVAL"

const (
	// defaultPingPeriod sets the default ping heartbeat interval
	defaultPingPeriod time.Duration = 24 * time.Hour
	// minimumPingPeriod sets the minimum possible configurable
	// heartbeat period, if a value lower than this will be set, the
	// defaultPingPeriod will be used
	minimumPingPeriod time.Duration = 1 * time.Hour
)

// PingCheck sends ping events to Google Analytics
func PingCheck() {
	// Create a new usage field
	u := New()
	duration := getPingPeriod()
	ticker := time.NewTicker(duration)
	for range ticker.C {
		u.Build().
			InstallBuilder(true).
			SetCategory(Ping).
			Send()
	}
}

// getPingPeriod sets the duration of health events, defaults to 24
func getPingPeriod() time.Duration {
	value := env.GetOrDefault(OpenEBSPingPeriod, fmt.Sprint(defaultPingPeriod))
	duration, _ := time.ParseDuration(value)
	// Sanitychecks for setting time duration of health events
	// This way, we are checking for negative and zero time duration and we
	// also have a minimum possible configurable time duration between health events
	if duration < minimumPingPeriod {
		// Avoid corner case when the ENV value is undesirable
		return time.Duration(defaultPingPeriod)
	}

	return time.Duration(duration)

}
