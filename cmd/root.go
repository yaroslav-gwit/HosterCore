package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hoster",
	Short: "HosterCore is a highly opinionated Bhyve automation platform written in Go",

	Run: func(cmd *cobra.Command, args []string) {
		hostMain()
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

	// VM command section
	rootCmd.AddCommand(vmCmd)

	// VM cmd -> list
	vmCmd.AddCommand(vmListCmd)
	vmListCmd.Flags().BoolVarP(&jsonOutputVm, "json", "j", false, "Output as JSON (useful for automation)")
	vmListCmd.Flags().BoolVarP(&jsonPrettyOutputVm, "json-pretty", "", false, "Pretty JSON Output")
	vmListCmd.Flags().BoolVarP(&tableUnixOutputVm, "unix-style", "u", false, "Show Unix style table (useful for bash scripting)")

	// VM cmd -> info
	vmCmd.AddCommand(vmInfoCmd)
	vmInfoCmd.Flags().BoolVarP(&jsonVmInfo, "json", "j", false, "Output as JSON (useful for automation)")
	vmInfoCmd.Flags().BoolVarP(&jsonPrettyVmInfo, "json-pretty", "", false, "Pretty JSON Output")

	// VM cmd -> start
	vmCmd.AddCommand(vmStartCmd)

	// VM cmd -> start all
	vmCmd.AddCommand(vmStartAllCmd)

	// VM cmd -> stop
	vmCmd.AddCommand(vmStopCmd)

	// VM cmd -> stop all
	vmCmd.AddCommand(vmStopAllCmd)

	// VM cmd -> snapshot
	vmCmd.AddCommand(vmZfsSnapshotCmd)
	vmZfsSnapshotCmd.Flags().StringVarP(&snapshotType, "stype", "t", "custom", "Snapshot type")
	vmZfsSnapshotCmd.Flags().IntVarP(&snapshotsToKeep, "keep", "k", 5, "Number of snapshots to keep for this specific snapshot type")

	// VM cmd -> snapshot all
	vmCmd.AddCommand(vmZfsSnapshotAllCmd)
	vmZfsSnapshotAllCmd.Flags().StringVarP(&snapshotAllType, "stype", "t", "custom", "Snapshot type")
	vmZfsSnapshotAllCmd.Flags().IntVarP(&snapshotsAllToKeep, "keep", "k", 5, "Number of snapshots to keep for this specific snapshot type")

	// VM cmd -> show log
	vmCmd.AddCommand(vmShowLogCmd)

	// VM cmd -> manually edit config
	vmCmd.AddCommand(vmEditConfigCmd)

	// VM cmd -> expand disk
	vmCmd.AddCommand(vmDistExpandCmd)
	vmDistExpandCmd.Flags().StringVarP(&diskImage, "image", "i", "disk0.img", "Disk image name, which should be expanded")
	vmDistExpandCmd.Flags().IntVarP(&expansionSize, "size", "s", 10, "How much size to add, in Gb")

	// VM cmd -> connect to the serial console
	vmCmd.AddCommand(vmSerialConsoleCmd)

	// VM cmd -> vm destroy
	vmCmd.AddCommand(vmDestroyCmd)

	// VM cmd -> vm deploy
	vmCmd.AddCommand(vmDeployCmd)
	vmDeployCmd.Flags().StringVarP(&vmName, "name", "n", "test-vm", "Set the VM name (automatically generated if left empty)")
	vmDeployCmd.Flags().IntVarP(&vmDeployCpus, "cpu-cores", "c", 2, "Number of CPU cores to assign to this VM")
	vmDeployCmd.Flags().StringVarP(&vmDeployRam, "ram", "r", "2G", "Amount of RAM to assign to this VM (ie 1500MB, 2GB, etc)")
	vmDeployCmd.Flags().StringVarP(&osType, "os-stype", "t", "debian11", "OS or type or distribution (ie: debian11, ubuntu2004, etc)")
	vmDeployCmd.Flags().StringVarP(&zfsDataset, "dataset", "d", "zroot/vm-encrypted", "Choose the parent dataset for the VM deployment")
	vmDeployCmd.Flags().BoolVarP(&vmDeployStartWhenReady, "start-now", "", false, "Whether to start the VM after it's deployed")

	// VM cmd -> vm deploy
	vmCmd.AddCommand(vmCiResetCmd)
	vmCiResetCmd.Flags().StringVarP(&newVmName, "new-name", "n", "", "Set a new VM name (if you'd like to rename the VM as well)")

	// VM cmd -> vm replicate
	vmCmd.AddCommand(vmZfsReplicateCmd)
	vmZfsReplicateCmd.Flags().StringVarP(&replicationEndpoint, "endpoint", "e", "", "Set the endpoint SSH address")
	vmZfsReplicateCmd.Flags().IntVarP(&endpointSshPort, "port", "p", 22, "Set the endpoint SSH port")
	vmZfsReplicateCmd.Flags().StringVarP(&sshKeyLocation, "key", "k", "/root/.ssh/id_rsa", "Set the absolute location for the SSH key")

	// VM cmd -> vm replicate all
	vmCmd.AddCommand(vmReplicateAllCmd)
	vmReplicateAllCmd.Flags().StringVarP(&replicationEndpointAll, "endpoint", "e", "", "Set the endpoint SSH address")
	vmReplicateAllCmd.Flags().IntVarP(&endpointSshPortAll, "port", "p", 22, "Set the endpoint SSH port")
	vmReplicateAllCmd.Flags().StringVarP(&sshKeyLocationAll, "key", "k", "/root/.ssh/id_rsa", "Set the absolute location for the SSH key")

	// API command section
	rootCmd.AddCommand(apiCmd)
	apiCmd.Flags().IntVarP(&apiServerPort, "port", "p", 3000, "Specify the port to listen on")
	apiCmd.Flags().StringVarP(&apiServerUser, "user", "u", "admin", "Username for API authentication")
	apiCmd.Flags().StringVarP(&apiServerPassword, "password", "", "123456", "Password for API authentication")

	// Init command section
	rootCmd.AddCommand(initCmd)

	// VM cmd -> image
	rootCmd.AddCommand(imageCmd)
	imageCmd.AddCommand(imageDownloadCmd)
	imageDownloadCmd.Flags().StringVarP(&imageOsType, "os-type", "t", "debian11", "Specify the OS/distribution image type")
	imageDownloadCmd.Flags().StringVarP(&imageDataset, "use-dataset", "d", "zroot/vm-encrypted", "Specify the dataset for this particular image")

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

	// Version command section
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of HosterCore",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("HosterCore v0.1, version based on Golang")
	},
}
