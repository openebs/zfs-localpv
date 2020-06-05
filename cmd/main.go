package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Sirupsen/logrus"
	config "github.com/openebs/zfs-localpv/pkg/config"
	"github.com/openebs/zfs-localpv/pkg/driver"
	"github.com/openebs/zfs-localpv/pkg/version"
	zfs "github.com/openebs/zfs-localpv/pkg/zfs"
	"github.com/spf13/cobra"
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
	var config = config.Default()

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
		&config.NodeID, "nodeid", zfs.NodeID, "NodeID to identify the node running this driver",
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

	err := cmd.Execute()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
}

func run(config *config.Config) {
	if config.Version == "" {
		config.Version = version.Current()
	}

	logrus.Infof("ZFS Driver Version :- %s - commit :- %s", version.Current(), version.GetGitCommit())
	logrus.Infof(
		"DriverName: %s Plugin: %s EndPoint: %s NodeID: %s",
		config.DriverName,
		config.PluginType,
		config.Endpoint,
		config.NodeID,
	)

	err := driver.New(config).Run()
	if err != nil {
		log.Fatalln(err)
	}
	os.Exit(0)
}
