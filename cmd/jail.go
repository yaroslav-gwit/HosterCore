//go:build freebsd
// +build freebsd

package cmd

import (
	"HosterCore/internal/pkg/emojlog"
	HosterTables "HosterCore/internal/pkg/hoster/cli_tables"
	HosterJail "HosterCore/internal/pkg/hoster/jail"
	"os"

	"github.com/spf13/cobra"
)

var (
	jailCmd = &cobra.Command{
		Use:   "jail",
		Short: "Jail related operations",
		Long:  `Jail related operations: deploy, stop, start, destroy, etc`,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
			cmd.Help()
		},
	}
)

var (
	jailStartCmd = &cobra.Command{
		Use:   "start [jailName]",
		Short: "Start a specific Jail",
		Long:  `Start a specific Jail using it's name`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := HosterJail.Start(args[0])
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

var (
	jailStartAllCmdWait     int
	jailStartAllCmdProdOnly bool
	jailStartAllCmd         = &cobra.Command{
		Use:   "start-all",
		Short: "Start all available Jails on this system",
		Long:  `Start all available Jails on this system.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := HosterJail.StartAll(jailStartAllCmdProdOnly, jailStartAllCmdWait)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

var (
	jailStopCmd = &cobra.Command{
		Use:   "stop [jailName]",
		Short: "Stop a specific Jail",
		Long:  `Stop a specific Jail using it's name`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := HosterJail.Stop(args[0])
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

var (
	jailStopAllCmd = &cobra.Command{
		Use:   "stop-all",
		Short: "Stop all online Jails on this system",
		Long:  `Stop all online Jails on this system.`,

		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := HosterJail.StopAll()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

var (
	jailListCmdUnixStyle bool

	jailListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all available Jails in a single table",
		Long:  `List all available Jails in a single table.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := HosterTables.GenerateJailsTable(jailListCmdUnixStyle)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

var (
	jailDestroyCmd = &cobra.Command{
		Use:   "destroy [jailName]",
		Short: "Destroy any existing Jail",
		Long:  `Destroy any existing Jail.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := HosterJail.Destroy(args[0])
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

var (
	jailDeployCmdOsRelease string
	jailDeployCmdDataset   string
	jailDeployCmdJailName  string
	jailDeployCmdCpuLimit  int
	jailDeployCmdRamLimit  string
	jailDeployCmdIpAddress string
	jailDeployCmdNetwork   string
	jailDeployCmdDnsServer string
	// jailDeployCmdTimezone    string
	// jailDeployCmdProduction  string
	// jailDeployCmdDescription string

	jailDeployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a new Jail",
		Long:  `Deploy a new Jail.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			i := HosterJail.DeployInput{}
			i.CpuLimit = jailDeployCmdCpuLimit
			i.DnsServer = jailDeployCmdDnsServer
			i.DsParent = jailDeployCmdDataset
			i.IpAddress = jailDeployCmdIpAddress
			i.JailName = jailDeployCmdJailName
			i.Network = jailDeployCmdNetwork
			i.RamLimit = jailDeployCmdRamLimit
			i.Release = jailDeployCmdOsRelease

			err := HosterJail.Deploy(i)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

var (
	jailBootstrapCmdOsRelease    string
	jailBootstrapCmdDataset      string
	jailBootstrapCmdExcludeLib32 bool

	jailBootstrapCmd = &cobra.Command{
		Use:   "bootstrap",
		Short: "Bootstrap a new Jail template",
		Long:  `Bootstrap a new Jail template`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := HosterJail.BootstrapOfficial(jailBootstrapCmdOsRelease, jailBootstrapCmdDataset, jailBootstrapCmdExcludeLib32)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

// type LiveJailStruct struct {
// 	ID         int
// 	Name       string
// 	Path       string
// 	Running    bool
// 	Ip4address string
// 	Ip6address string
// }

// type JailConfigFileStruct struct {
// 	CPULimitPercent  int    `json:"cpu_limit_percent"`
// 	RAMLimit         string `json:"ram_limit"`
// 	Production       bool   `json:"production"`
// 	StartupScript    string `json:"startup_script"`
// 	ShutdownScript   string `json:"shutdown_script"`
// 	ConsoleOutputLog string `json:"console_output_log"`
// 	ConfigFileAppend string `json:"config_file_append"`
// 	StartAfter       string `json:"start_after,omitempty"`
// 	StartupDelay     int    `json:"startup_delay,omitempty"`
// 	IPAddress        string `json:"ip_address"`
// 	Network          string `json:"network"`
// 	DnsServer        string `json:"dns_server"`
// 	Timezone         string `json:"timezone"`
// 	Parent           string `json:"parent"`
// 	Description      string `json:"description"`

// 	// Not a part of JSON config file
// 	JailName       string
// 	JailHostname   string
// 	JailRootPath   string // /zroot/vm-encrypted/jailDataset/root_folder
// 	JailFolder     string // /zroot/vm-encrypted/jailDataset/
// 	ZfsDatasetPath string // zroot/vm-encrypted/jailDataset
// 	Netmask        string
// 	Running        bool
// 	Backup         bool
// 	CpuLimitReal   int
// 	VnetInterfaceA string
// 	VnetInterfaceB string
// 	DefaultRouter  string
// }

// func GetAllJailsList() ([]string, error) {
// 	zfsDatasets, err := getZfsDatasetInfo()
// 	if err != nil {
// 		return []string{}, err
// 	}

// 	jails := []string{}
// 	for _, v := range zfsDatasets {
// 		mountPointDirWalk, err := os.ReadDir(v.MountPoint)
// 		if err != nil {
// 			return []string{}, err
// 		}

// 		for _, directory := range mountPointDirWalk {
// 			if directory.IsDir() {
// 				configFile := v.MountPoint + "/" + directory.Name() + "/jail_config.json"
// 				if FileExists(configFile) {
// 					jails = append(jails, directory.Name())
// 				}
// 			}
// 		}
// 	}

// 	natsort.Sort(jails)
// 	return jails, nil
// }
