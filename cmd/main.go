/*
Copyright © 2020 The OpenEBS Authors

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

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	configs "github.com/openebs/zfs-localpv/pkg/config"
	"github.com/openebs/zfs-localpv/pkg/driver"
	"github.com/openebs/zfs-localpv/pkg/version"
	zfs "github.com/openebs/zfs-localpv/pkg/zfs"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

/*
 * main routine to start the zfs-driver. The same
 * binary is used to controller and agent deployment.
 * they both are differentiated via plugin command line
 * argument. To start the controller, we have to pass
 * --plugin=controller and to start it as agent, we have
 * to pass --plugin=agent.
 */
func main() {
	_ = flag.CommandLine.Parse([]string{})
	var config = configs.Default()

	cmd := &cobra.Command{
		Use:   "zfs-driver",
		Short: "driver for provisioning zfs volume",
		Long: `provisions and deprovisions the volume
		    on the node which has zfs pool configured.`,
		Run: func(cmd *cobra.Command, args []string) {
			run(config)
		},
	}

	cmd.Flags().AddGoFlagSet(flag.CommandLine)

	cmd.PersistentFlags().StringVar(
		&config.Nodename, "nodename", zfs.NodeID, "Nodename to identify the node running this driver",
	)

	cmd.PersistentFlags().StringVar(
		&config.Version, "version", "", "Displays driver version",
	)

	cmd.PersistentFlags().StringVar(
		&config.Endpoint, "endpoint", "unix://csi/csi.sock", "CSI endpoint",
	)

	cmd.PersistentFlags().StringVar(
		&config.DriverName, "name", "zfs.csi.openebs.io", "Name of this driver",
	)

	cmd.PersistentFlags().StringVar(
		&config.PluginType, "plugin", "csi-plugin", "Type of this driver i.e. controller or node",
	)

	cmd.PersistentFlags().StringVar(
		&configs.QuotaType, "quota-type", "quota", "quota type: refquota or quota",
	)

	err := cmd.Execute()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
}

func run(config *configs.Config) {
	if config.Version == "" {
		config.Version = version.Current()
	}

	if configs.QuotaType != configs.Quota && configs.QuotaType != configs.RefQuota {
		log.Fatalln(fmt.Errorf("quota-type should be quota or refquota"))
	}

	klog.Infof("ZFS Driver Version :- %s - commit :- %s", version.Current(), version.GetGitCommit())
	klog.Infof(
		"DriverName: %s Plugin: %s EndPoint: %s Node Name: %s",
		config.DriverName,
		config.PluginType,
		config.Endpoint,
		config.Nodename,
	)

	err := driver.New(config).Run()
	if err != nil {
		log.Fatalln(err)
	}
	os.Exit(0)
}
