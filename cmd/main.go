//go:build freebsd
// +build freebsd

package cmd

import (
	"fmt"
	"os"

	HosterTables "HosterCore/internal/pkg/hoster/cli_tables"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hoster",
	Short: "HosterCore is a highly opinionated Bhyve automation platform written in Go",

	Run: func(cmd *cobra.Command, args []string) {
		checkInitFile()
		HosterTables.GenerateHostInfoTable(false)
		printZfsDatasetInfo()
		printNetworkInfoTable()
		HosterTables.GenerateVMsTable(false)
		HosterTables.GenerateJailsTable(false)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Host Command Section
	rootCmd.AddCommand(hostCmd)
	hostCmd.Flags().BoolVarP(&jsonHostInfoOutput, "json", "j", false, "Output as JSON (useful for automation)")
	hostCmd.Flags().BoolVarP(&jsonPrettyHostInfoOutput, "json-pretty", "", false, "Pretty JSON Output")

	// Host Network Section
	rootCmd.AddCommand(networkCmd)
	networkCmd.AddCommand(networkListCmd)
	networkListCmd.Flags().BoolVarP(&networkListUnixStyleTable, "unix-style", "u", false, "Show Unix style table (useful for scripting)")
	networkCmd.AddCommand(networkInitCmd)

	// Host Dataset Info
	rootCmd.AddCommand(datasetCmd)
	datasetCmd.AddCommand(datasetListCmd)
	datasetListCmd.Flags().BoolVarP(&datasetListUnixStyleTable, "unix-style", "u", false, "Show Unix style table (useful for scripting)")

	// Host Scheduler
	rootCmd.AddCommand(schedulerCmd)
	// Host Scheduler -> Start
	schedulerCmd.AddCommand(schedulerStartCmd)
	// Host Scheduler -> Status
	schedulerCmd.AddCommand(schedulerStatusCmd)
	// Host Scheduler -> Info
	schedulerCmd.AddCommand(schedulerInfoCmd)
	schedulerInfoCmd.Flags().BoolVarP(&schedulerInfoJsonPretty, "json-pretty", "", false, "Info for one of the scheduled jobs in a JSON-pretty format")
	// Host Scheduler -> List
	schedulerCmd.AddCommand(schedulerListCmd)
	schedulerListCmd.Flags().BoolVarP(&schedulerListUnix, "unix-style", "u", false, "Show Unix style table (useful for scripting)")
	schedulerListCmd.Flags().BoolVarP(&schedulerListJson, "json", "j", false, "List of scheduled jobs in a JSON format")
	schedulerListCmd.Flags().BoolVarP(&schedulerListJsonPretty, "json-pretty", "", false, "List of scheduled jobs in a JSON-pretty format")
	// Host Scheduler -> Show Log
	schedulerCmd.AddCommand(schedulerShowLogCmd)
	// Host Scheduler -> Stop
	schedulerCmd.AddCommand(schedulerStopCmd)
	// Host Scheduler -> Replication
	schedulerCmd.AddCommand(schedulerReplicateCmd)
	schedulerReplicateCmd.Flags().StringVarP(&schedulerReplicateEndpoint, "endpoint", "e", "", "SSH endpoint to send the replicated data to")
	schedulerReplicateCmd.Flags().StringVarP(&schedulerReplicateKey, "key", "k", "/root/.ssh/id_rsa", "SSH key location")
	schedulerReplicateCmd.Flags().IntVarP(&schedulerReplicatePort, "port", "p", 22, "Endpoint SSH port")
	schedulerReplicateCmd.Flags().IntVarP(&schedulerReplicateSpeedLimit, "speed-limit", "s", 50, "Replication speed limit")
	// Host Scheduler -> Snapshot
	schedulerCmd.AddCommand(schedulerSnapshotCmd)
	schedulerSnapshotCmd.Flags().StringVarP(&schedulerSnapshotType, "type", "t", "custom", "Snapshot type: custom, frequent, hourly, daily, weekly, monthly, yearly")
	schedulerSnapshotCmd.Flags().IntVarP(&schedulerSnapshotToKeep, "keep", "k", 5, "How many snapshots to keep")
	// Host Scheduler -> Snapshot All
	schedulerCmd.AddCommand(schedulerSnapshotAllCmd)
	schedulerSnapshotAllCmd.Flags().StringVarP(&schedulerSnapshotAllType, "type", "t", "custom", "Snapshot type: custom, frequent, hourly, daily, weekly, monthly, yearly")
	schedulerSnapshotAllCmd.Flags().IntVarP(&schedulerSnapshotAllToKeep, "keep", "k", 5, "How many snapshots to keep")

	// HA
	rootCmd.AddCommand(carpHaCmd)
	// HA -> start
	carpHaCmd.AddCommand(haStartCmd)
	// HA -> stop
	carpHaCmd.AddCommand(haStopCmd)
	// HA -> status
	carpHaCmd.AddCommand(haStatusCmd)

	// Jail Command Section
	rootCmd.AddCommand(jailCmd)
	// Jail -> start
	jailCmd.AddCommand(jailStartCmd)
	// Jail -> start-all
	jailCmd.AddCommand(jailStartAllCmd)
	// Jail -> stop-all
	jailCmd.AddCommand(jailStopAllCmd)
	// Jail -> stop
	jailCmd.AddCommand(jailStopCmd)
	// Jail -> list
	jailCmd.AddCommand(jailListCmd)
	jailListCmd.Flags().BoolVarP(&jailListCmdUnixStyle, "unix", "u", false, "Show Unix style table (useful for scripting)")
	// Jail -> destroy
	jailCmd.AddCommand(jailDestroyCmd)
	// Jail -> bootstrap
	jailCmd.AddCommand(jailBootstrapCmd)
	jailBootstrapCmd.Flags().StringVarP(&jailBootstrapCmdOsRelease, "release", "r", "", "Pick a FreeBSD OS Release version (your own OS release will be used by default)")
	jailBootstrapCmd.Flags().StringVarP(&jailBootstrapCmdDataset, "dataset", "d", "", "Specify a target dataset (first available DS in your config file will be used as a default)")
	jailBootstrapCmd.Flags().BoolVarP(&jailBootstrapCmdExcludeLib32, "exclude-lib32", "", false, "Exclude Lib32 from this Jail Template")
	// Jail -> deploy
	jailCmd.AddCommand(jailDeployCmd)
	jailDeployCmd.Flags().StringVarP(&jailDeployCmdJailName, "name", "n", "", "Jail name, test-jail-1 (2, 3 and so on) will be used by default")
	jailDeployCmd.Flags().StringVarP(&jailDeployCmdOsRelease, "release", "r", "", "Pick a FreeBSD OS Release version (your own OS release will be used by default)")
	jailDeployCmd.Flags().StringVarP(&jailDeployCmdDataset, "dataset", "d", "", "Specify a target dataset (first available DS in your config file will be used as a default)")
	jailDeployCmd.Flags().IntVarP(&jailDeployCmdCpuLimit, "cpu-limit-percent", "", 50, "CPU percentage execution limit")
	jailDeployCmd.Flags().StringVarP(&jailDeployCmdRamLimit, "ram-limit", "", "2G", "RAM limit")
	jailDeployCmd.Flags().StringVarP(&jailDeployCmdIpAddress, "ip", "", "", "IP address")
	jailDeployCmd.Flags().StringVarP(&jailDeployCmdNetwork, "network", "", "", "Network, eg: internal, external, etc")
	jailDeployCmd.Flags().StringVarP(&jailDeployCmdDnsServer, "dns-sever", "", "", "Specify a custom DNS server, eg: 1.1.1.1, etc")

	// VM command section
	rootCmd.AddCommand(vmCmd)

	// VM cmd -> unlock-all
	vmCmd.AddCommand(vmUnlockAllCmd)

	// VM cmd -> clone
	vmCmd.AddCommand(vmCloneCmd)
	vmCloneCmd.Flags().StringVarP(&vmCloneSnapshot, "snapshot", "s", "", "Specify a custom snapshot to be used as a clone source (latest one picked by default)")

	// VM cmd -> list
	vmCmd.AddCommand(vmListCmd)
	vmListCmd.Flags().BoolVarP(&jsonOutputVm, "json", "j", false, "Output as JSON (useful for automation)")
	vmListCmd.Flags().BoolVarP(&jsonPrettyOutputVm, "json-pretty", "", false, "Pretty JSON Output")
	vmListCmd.Flags().BoolVarP(&tableUnixOutputVm, "unix-style", "u", false, "Show Unix style table (useful for scripting)")

	// VM cmd -> info
	vmCmd.AddCommand(vmInfoCmd)
	vmInfoCmd.Flags().BoolVarP(&jsonVmInfo, "json", "j", false, "Output as JSON (useful for automation)")
	vmInfoCmd.Flags().BoolVarP(&jsonPrettyVmInfo, "json-pretty", "", false, "Pretty JSON Output")

	// VM cmd -> start
	vmCmd.AddCommand(vmStartCmd)
	vmStartCmd.Flags().BoolVarP(&vmStartCmdWaitForVnc, "wait-for-vnc", "", false, "Use this flag to wait for a VNC connection before booting the VM")
	vmStartCmd.Flags().BoolVarP(&vmStartCmdRestoreVmState, "restore-state", "", false, "Restore saved VM state (EXPERIMENTAL!)")
	vmStartCmd.Flags().BoolVarP(&vmStartCmdDebug, "debug-run", "", false, "Only console-print the start commands, but don't execute them")

	// VM cmd -> start all
	vmCmd.AddCommand(vmStartAllCmd)
	vmStartAllCmd.Flags().IntVarP(&waitTime, "wait-time", "t", 0, "Set a static wait time between each VM start")
	vmStartAllCmd.Flags().BoolVarP(&prodOnly, "production-only", "p", false, "Only start all production VMs")

	// VM cmd -> stop
	vmCmd.AddCommand(vmStopCmd)
	vmStopCmd.Flags().BoolVarP(&vmStopCmdForceStop, "force", "f", false, "Use -SIGKILL signal to forcefully kill the VM process")
	vmStopCmd.Flags().BoolVarP(&vmStopCmdCleanUp, "cleanup", "c", false, "Kill VM Supervisor as well as the VM itself (rarely needed)")

	// VM cmd -> stop all
	vmCmd.AddCommand(vmStopAllCmd)
	vmStopAllCmd.Flags().BoolVarP(&forceKill, "force", "f", false, "Use -SIGKILL signal to forcefully kill all of the VMs processes")
	vmStopAllCmd.Flags().BoolVarP(&forceCleanUp, "cleanup", "c", false, "Kill VM Supervisor as well as the VM itself (rarely needed)")

	// VM cmd -> show log
	vmCmd.AddCommand(vmShowLogCmd)

	// VM cmd -> manually edit config
	vmCmd.AddCommand(vmEditConfigCmd)

	// VM cmd -> expand disk
	vmCmd.AddCommand(vmDiskCmd)
	vmDiskCmd.AddCommand(vmDiskExpandCmd)
	vmDiskExpandCmd.Flags().StringVarP(&diskImage, "image", "i", "disk0.img", "Disk image name, which should be expanded")
	vmDiskExpandCmd.Flags().IntVarP(&expansionSize, "size", "s", 10, "How much size to add, in Gb")
	vmDiskCmd.AddCommand(vmDiskAddCmd)
	vmDiskAddCmd.Flags().IntVarP(&vmDiskAddSize, "size", "s", 10, "Initial size of the image, in Gb")

	// VM cmd -> connect to the serial console
	vmCmd.AddCommand(vmSerialConsoleCmd)

	// VM cmd -> vm destroy
	vmCmd.AddCommand(vmDestroyCmd)

	// VM cmd -> vm deploy
	vmCmd.AddCommand(vmDeployCmd)
	vmDeployCmd.Flags().StringVarP(&vmDeployCmdVmName, "name", "n", "test-vm", "Set the VM name (automatically generated if left empty)")
	vmDeployCmd.Flags().StringVarP(&vmDeployCmdNetworkName, "network-name", "", "", "Use this network for new VM deployment")
	vmDeployCmd.Flags().IntVarP(&vmDeployCmdCpus, "cpu-cores", "c", 2, "Number of CPU cores to assign to this VM")
	vmDeployCmd.Flags().StringVarP(&vmDeployCmdRam, "ram", "r", "2G", "Amount of RAM to assign to this VM (ie 1500MB, 2GB, etc)")
	vmDeployCmd.Flags().StringVarP(&vmDeployCmdOsType, "os-type", "t", "debian12", "OS type or distribution (ie: debian12, ubuntu2004, etc)")
	vmDeployCmd.Flags().StringVarP(&vmDeployCmdZfsDataset, "dataset", "d", "", "Choose the parent dataset for the VM deployment (first available dataset will be chosen as default)")
	vmDeployCmd.Flags().BoolVarP(&vmDeployCmdStartWhenReady, "start-now", "", false, "Whether to start the VM after it's deployed")
	vmDeployCmd.Flags().StringVarP(&vmDeployCmdIpAddress, "ip-address", "", "", "Set the IP address for your new VM manually")
	vmDeployCmd.Flags().StringVarP(&vmDeployCmdDnsServer, "dns-server", "", "", "Set a custom DNS server for your new VM")
	vmDeployCmd.Flags().StringVarP(&vmDeployCmdFromIso, "from-iso", "", "", "Deploy this VM using an ISO file (e.g. `/root/custom_os.iso`)")
	// vmDeployCmd.Flags().StringVarP(&vmDeployCmdIsoFilePath, "path-to-iso", "", "", "Path to the ISO file")

	// VM cmd -> vm cireset
	vmCmd.AddCommand(vmCiResetCmd)
	vmCiResetCmd.Flags().StringVarP(&ciResetCmdNewVmName, "new-name", "n", "", "Set a new VM name (if you'd like to rename the VM as well)")
	vmCiResetCmd.Flags().StringVarP(&ciResetCmdNetworkName, "network-name", "", "", "Use the specific network instead of a default choice")
	vmCiResetCmd.Flags().StringVarP(&ciResetCmdIpAddress, "ip-address", "", "", "Set the IP address for your VM manually")
	vmCiResetCmd.Flags().StringVarP(&ciResetCmdDnsServer, "dns-server", "", "", "Set a custom DNS server for your VM")

	// VM cmd -> vm replicate
	vmCmd.AddCommand(vmZfsReplicateCmd)
	vmZfsReplicateCmd.Flags().StringVarP(&replicationEndpoint, "endpoint", "e", "", "Set the endpoint SSH address, for example: `192.168.118.3`")
	vmZfsReplicateCmd.Flags().IntVarP(&endpointSshPort, "port", "p", 22, "Set the endpoint SSH port, for example `2202`")
	vmZfsReplicateCmd.Flags().IntVarP(&replicateSpeedLimit, "speed-limit", "", 50, "Set the replication speed limit in MB/s")
	vmZfsReplicateCmd.Flags().StringVarP(&sshKeyLocation, "key", "k", "/root/.ssh/id_rsa", "Set the absolute location for the SSH key, for example: `'/home/user-name/id_rsa'`")
	vmZfsReplicateCmd.Flags().StringVarP(&replicateScriptName, "script-name", "", "", "Set the replication script name (useful to run multiple jobs in parallel)")

	// VM cmd -> vm replicate all
	vmCmd.AddCommand(vmReplicateAllCmd)
	vmReplicateAllCmd.Flags().StringVarP(&vmReplicateAllFilter, "filter", "f", "", "Filter the VMs that will be included in the replication (uses coma separated VM names, or coma + space): `'test-vm-1,test-vm-2'`")
	vmReplicateAllCmd.Flags().StringVarP(&sshKeyLocationAll, "key", "k", "/root/.ssh/id_rsa", "Set the absolute location for the SSH key, for example: `'/home/user-name/id_rsa'`")
	vmReplicateAllCmd.Flags().StringVarP(&replicationEndpointAll, "endpoint", "e", "", "Set the endpoint SSH address, for example: `192.168.118.3`")
	vmReplicateAllCmd.Flags().IntVarP(&endpointSshPortAll, "port", "p", 22, "Set the endpoint SSH port, for example `2202`")
	vmReplicateAllCmd.Flags().IntVarP(&replicateAllSpeedLimit, "speed-limit", "", 50, "Set the replication speed limit in MB/s")
	vmReplicateAllCmd.Flags().StringVarP(&replicateAllScriptName, "script-name", "", "", "Set the replication script name (useful to run multiple jobs in parallel)")

	// Snapshot cmd
	rootCmd.AddCommand(snapshotCmd)

	// Snapshot cmd -> snapshot new
	snapshotCmd.AddCommand(snapshotNewCmd)
	snapshotNewCmd.Flags().StringVarP(&snapshotNewType, "stype", "t", "custom", "Snapshot type")
	snapshotNewCmd.Flags().IntVarP(&snapshotNewSnapsToKeep, "keep", "k", 5, "Number of snapshots to keep for this specific snapshot type")

	// Snapshot cmd -> snapshot list
	snapshotCmd.AddCommand(snapshotListCmd)
	snapshotListCmd.Flags().BoolVarP(&snapshotListUnixStyleTable, "unix", "u", false, "Output the table using `Unix` style for further processing")

	// Snapshot cmd -> snapshot list all
	snapshotCmd.AddCommand(snapshotListAllCmd)
	snapshotListAllCmd.Flags().BoolVarP(&snapshotListAllUnixStyleTable, "unix", "u", false, "Output the table using `Unix` style for further processing")

	// Snapshot cmd -> snapshot all
	// snapshotCmd.AddCommand(snapshotAllCmd)
	// snapshotAllCmd.Flags().StringVarP(&snapshotAllCmdType, "stype", "t", "custom", "Snapshot type")
	// snapshotAllCmd.Flags().IntVarP(&snapshotsAllCmdToKeep, "keep", "k", 5, "Number of snapshots to keep for this specific snapshot type")

	// Snapshot cmd -> snapshot destroy
	snapshotCmd.AddCommand(snapshotDestroyCmd)

	// Snapshot cmd -> snapshot rollback
	snapshotCmd.AddCommand(snapshotRollbackCmd)
	snapshotRollbackCmd.Flags().BoolVarP(&snapshotRollbackForceStop, "force-stop", "", false, "Automatically stop the VM using --force flag")
	snapshotRollbackCmd.Flags().BoolVarP(&snapshotRollbackForceStart, "force-start", "", false, "Automatically start the VM after roll-back operation")

	// Passthru command section
	rootCmd.AddCommand(passthruCmd)
	passthruCmd.AddCommand(passthruListCmd)

	// API command section
	rootCmd.AddCommand(apiCmd)
	apiCmd.AddCommand(apiStartCmd)
	// apiStartCmd.Flags().IntVarP(&apiStartPort, "port", "p", 3000, "Specify the port to listen on")
	// apiStartCmd.Flags().StringVarP(&apiStartUser, "user", "u", "admin", "Username for API authentication")
	// apiStartCmd.Flags().StringVarP(&apiStartPassword, "password", "", "123456", "Password for API authentication")
	// apiStartCmd.Flags().BoolVarP(&apiHaMode, "ha-mode", "", false, "Activate HA clustering mode")
	// apiStartCmd.Flags().BoolVarP(&apiHaDebug, "ha-debug", "", false, "Activate HA Debug mode, that only logs all actions, but doesn't execute anything")
	apiCmd.AddCommand(apiStatusCmd)
	apiCmd.AddCommand(apiStopCmd)
	apiCmd.AddCommand(apiShowLogCmd)

	// Node exporter command section
	rootCmd.AddCommand(nodeExporterCmd)
	nodeExporterCmd.AddCommand(nodeExporterStartCmd)
	nodeExporterCmd.AddCommand(nodeExporterStopCmd)
	nodeExporterCmd.AddCommand(nodeExporterStatusCmd)

	// Init command section
	rootCmd.AddCommand(initCmd)

	// Prometheus command section
	rootCmd.AddCommand(prometheusCmd)

	// Image command section
	rootCmd.AddCommand(imageCmd)
	imageCmd.AddCommand(imageDownloadCmd)
	imageDownloadCmd.Flags().StringVarP(&imageDataset, "dataset", "d", "", "Specify the dataset for this particular image (first available dataset is picked otherwise)")

	// VM cmd -> secrets
	vmCmd.AddCommand(vmSecretsCmd)
	vmSecretsCmd.Flags().BoolVarP(&vmSecretsUnixTable, "unix-style", "u", false, "Show Unix style table (useful for scripting and smaller screens)")

	// VM cmd -> ci-iso
	vmCmd.AddCommand(vmCiIsoCmd)
	vmCiIsoCmd.AddCommand(vmCiIsoMountCmd)
	vmCiIsoCmd.AddCommand(vmCiIsoUnmountCmd)

	// VM cmd -> set -> parent
	rootCmd.AddCommand(vmSetCmd)
	vmSetCmd.AddCommand(vmSetCmdParent)
	vmSetCmdParent.Flags().StringVarP(&vmSetNewParent, "new-parent", "p", "", "New parent name (optional, current hostname used by default)")

	// DNS server commands
	rootCmd.AddCommand(dnsCmd)
	dnsCmd.AddCommand(dnsStartCmd)
	dnsCmd.AddCommand(dnsStopCmd)
	dnsCmd.AddCommand(dnsReloadCmd)
	dnsCmd.AddCommand(dnsShowLogCmd)
	dnsCmd.AddCommand(dnsStatusCmd)

	// Version command section
	rootCmd.AddCommand(versionCmd)
}

var HosterVersion = "RELEASE"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of HosterCore",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(HosterVersion)
	},
}
