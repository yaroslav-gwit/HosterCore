package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hoster",
	Short: "HosterCore is a highly opinionated Bhyve automation platform written in Go",

	Run: func(cmd *cobra.Command, args []string) {
		err := checkInitFile()
		if err != nil {
			log.Fatal(err.Error())
		}
		hostMain()
		printZfsDatasetInfo()
		printNetworkInfoTable()
		vmListMain()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Host command section
	rootCmd.AddCommand(hostCmd)
	hostCmd.Flags().BoolVarP(&jsonHostInfoOutput, "json", "j", false, "Output as JSON (useful for automation)")
	hostCmd.Flags().BoolVarP(&jsonPrettyHostInfoOutput, "json-pretty", "", false, "Pretty JSON Output")

	// Host network info
	rootCmd.AddCommand(networkCmd)
	networkCmd.AddCommand(networkListCmd)
	networkListCmd.Flags().BoolVarP(&networkListUnixStyleTable, "unix-style", "u", false, "Show Unix style table (useful for scripting)")

	// Host dataset info
	rootCmd.AddCommand(datasetCmd)
	datasetCmd.AddCommand(datasetListCmd)
	datasetListCmd.Flags().BoolVarP(&datasetListUnixStyleTable, "unix-style", "u", false, "Show Unix style table (useful for scripting)")

	// Jail command section
	rootCmd.AddCommand(jailCmd)
	jailCmd.AddCommand(jailStartCmd)

	// VM command section
	rootCmd.AddCommand(vmCmd)

	// VM cmd -> unlock-all
	vmCmd.AddCommand(vmUnlockAllCmd)

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

	// VM cmd -> start all
	vmCmd.AddCommand(vmStartAllCmd)
	vmStartAllCmd.Flags().IntVarP(&waitTime, "wait-time", "t", 5, "Set a static wait time between VM starts")

	// VM cmd -> stop
	vmCmd.AddCommand(vmStopCmd)
	vmStopCmd.Flags().BoolVarP(&forceStop, "force", "f", false, "Use -SIGKILL signal to forcefully kill the VM process")

	// VM cmd -> stop all
	vmCmd.AddCommand(vmStopAllCmd)
	vmStopAllCmd.Flags().BoolVarP(&forceStopAll, "force", "f", false, "Use -SIGKILL signal to forcefully kill all of the VMs processes")

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
	vmDeployCmd.Flags().StringVarP(&vmName, "name", "n", "test-vm", "Set the VM name (automatically generated if left empty)")
	vmDeployCmd.Flags().StringVarP(&networkName, "network-name", "", "", "Use this network for new VM deployment")
	vmDeployCmd.Flags().IntVarP(&vmDeployCpus, "cpu-cores", "c", 2, "Number of CPU cores to assign to this VM")
	vmDeployCmd.Flags().StringVarP(&vmDeployRam, "ram", "r", "2G", "Amount of RAM to assign to this VM (ie 1500MB, 2GB, etc)")
	vmDeployCmd.Flags().StringVarP(&osType, "os-type", "t", "debian12", "OS type or distribution (ie: debian12, ubuntu2004, etc)")
	vmDeployCmd.Flags().StringVarP(&osTypeAlias, "os-stype", "", "", "Alias for the os-type, because it was misspelled in the past as os-stype")
	vmDeployCmd.Flags().StringVarP(&zfsDataset, "dataset", "d", "zroot/vm-encrypted", "Choose the parent dataset for the VM deployment")
	vmDeployCmd.Flags().BoolVarP(&vmDeployStartWhenReady, "start-now", "", false, "Whether to start the VM after it's deployed")
	vmDeployCmd.Flags().StringVarP(&deployIpAddress, "ip-address", "", "", "Set the IP address for your new VM manually")
	vmDeployCmd.Flags().StringVarP(&deployDnsServer, "dns-server", "", "", "Set a custom DNS server for your new VM")
	vmDeployCmd.Flags().BoolVarP(&vmDeployFromIso, "from-iso", "", false, "Deploy this VM using an ISO file")
	vmDeployCmd.Flags().StringVarP(&vmDeployIsoFilePath, "path-to-iso", "", "", "Path to the ISO file")

	// VM cmd -> vm cireset
	vmCmd.AddCommand(vmCiResetCmd)
	vmCiResetCmd.Flags().StringVarP(&newVmName, "new-name", "n", "", "Set a new VM name (if you'd like to rename the VM as well)")
	vmCiResetCmd.Flags().StringVarP(&ciResetNetworkName, "network-name", "", "", "Use the specific network instead of a default choice")
	vmCiResetCmd.Flags().StringVarP(&ciResetIpAddress, "ip-address", "", "", "Set the IP address for your VM manually")
	vmCiResetCmd.Flags().StringVarP(&ciResetDnsServer, "dns-server", "", "", "Set a custom DNS server for your VM")

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
	snapshotCmd.AddCommand(snapshotAllCmd)
	snapshotAllCmd.Flags().StringVarP(&snapshotAllCmdType, "stype", "t", "custom", "Snapshot type")
	snapshotAllCmd.Flags().IntVarP(&snapshotsAllCmdToKeep, "keep", "k", 5, "Number of snapshots to keep for this specific snapshot type")

	// Snapshot cmd -> snapshot destroy
	snapshotCmd.AddCommand(snapshotDestroyCmd)

	// Snapshot cmd -> snapshot rollback
	snapshotCmd.AddCommand(snapshotRollbackCmd)
	snapshotRollbackCmd.Flags().BoolVarP(&snapshotRollbackForceStop, "force-stop", "", false, "Automatically stop the VM using --force flag")
	snapshotRollbackCmd.Flags().BoolVarP(&snapshotRollbackForceStart, "force-start", "", false, "Automatically start the VM after roll-back operation")

	// API command section
	rootCmd.AddCommand(apiCmd)
	apiCmd.AddCommand(apiStartCmd)
	apiStartCmd.Flags().IntVarP(&apiStartPort, "port", "p", 3000, "Specify the port to listen on")
	apiStartCmd.Flags().StringVarP(&apiStartUser, "user", "u", "admin", "Username for API authentication")
	apiStartCmd.Flags().StringVarP(&apiStartPassword, "password", "", "123456", "Password for API authentication")
	apiStartCmd.Flags().BoolVarP(&apiHaMode, "ha-mode", "", false, "Activate HA clustering mode")
	apiStartCmd.Flags().BoolVarP(&apiHaDebug, "ha-debug", "", false, "Activate HA Debug mode, that only logs all actions, but doesn't execute anything")
	apiCmd.AddCommand(apiStatusCmd)
	apiCmd.AddCommand(apiStopCmd)
	apiCmd.AddCommand(apiShowLogCmd)

	// Node exporter command section
	rootCmd.AddCommand(nodeExporterCmd)
	nodeExporterCmd.AddCommand(nodeExporterStartCmd)
	nodeExporterCmd.AddCommand(nodeExporterStopCmd)
	nodeExporterCmd.AddCommand(nodeExporterStatusCmd)

	// Traefik proxy command section
	rootCmd.AddCommand(proxyTraefikCmd)
	proxyTraefikCmd.AddCommand(proxyTraefikStartCmd)
	proxyTraefikCmd.AddCommand(proxyTraefikStopCmd)
	proxyTraefikCmd.AddCommand(proxyTraefikStatusCmd)

	// Init command section
	rootCmd.AddCommand(initCmd)

	// Image command section
	rootCmd.AddCommand(imageCmd)
	imageCmd.AddCommand(imageDownloadCmd)
	imageDownloadCmd.Flags().StringVarP(&imageDataset, "use-dataset", "d", "zroot/vm-encrypted", "Specify the dataset for this particular image")

	// VM cmd -> secrets
	vmCmd.AddCommand(vmSecretsCmd)
	vmSecretsCmd.Flags().BoolVarP(&vmSecretsUnixTable, "unix-style", "u", false, "Show Unix style table (useful for scripting and smaller screens)")

	// VM cmd -> ci-iso
	vmCmd.AddCommand(vmCiIsoCmd)
	vmCiIsoCmd.AddCommand(vmCiIsoMountCmd)
	vmCiIsoCmd.AddCommand(vmCiIsoUnmountCmd)

	// VM cmd -> change
	rootCmd.AddCommand(changeCmd)
	changeCmd.AddCommand(changeParentCmd)
	changeParentCmd.Flags().StringVarP(&changeParentVmName, "vm", "", "", "VM Name (mandatory flag)")
	changeParentCmd.Flags().StringVarP(&changeParentNewParent, "new-parent", "", "", "New parent name (optional, current hostname used by default)")

	// VM cmd -> nebula
	rootCmd.AddCommand(nebulaCmd)
	nebulaCmd.AddCommand(nebulaInitCmd)
	nebulaCmd.AddCommand(nebulaShowLogCmd)

	nebulaCmd.AddCommand(nebulaServiceCmd)
	nebulaServiceCmd.Flags().BoolVarP(&nebulaServiceStart, "start", "s", false, "Start Nebula service")
	nebulaServiceCmd.Flags().BoolVarP(&nebulaServiceStop, "stop", "k", false, "Stop/kill Nebula service")
	nebulaServiceCmd.Flags().BoolVarP(&nebulaServiceReload, "reload", "r", false, "Restart Nebula service")

	nebulaCmd.AddCommand(nebulaUpdateCmd)
	nebulaUpdateCmd.Flags().BoolVarP(&nebulaUpdateBinary, "binary", "b", false, "Download a fresh Nebula binary")
	nebulaUpdateCmd.Flags().BoolVarP(&nebulaUpdateConfig, "config", "c", false, "Request Nebula Control Plane to generate new config, and then download it")

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

var HosterVersion = "v0.2b-RELEASE"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of HosterCore",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(HosterVersion)
	},
}