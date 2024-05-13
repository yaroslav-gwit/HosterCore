//go:build freebsd
// +build freebsd

package cmd

import (
	"HosterCore/internal/pkg/emojlog"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	replicationEndpointAll string
	endpointSshPortAll     int
	sshKeyLocationAll      string
	vmReplicateAllFilter   string
	replicateAllSpeedLimit int
	replicateAllScriptName string

	vmReplicateAllCmd = &cobra.Command{
		Use:   "replicate-all",
		Short: "Replicate all live and production VMs to a backup node",
		Long:  `Replicate all live and production VMs to a backup node.`,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
			if len(replicationEndpointAll) < 1 {
				emojlog.PrintLogMessage("Please, specify an endpoint", emojlog.Error)
				os.Exit(1)
			}
			replicateAllProdVms(replicationEndpointAll, endpointSshPortAll, sshKeyLocationAll)
		},
	}
)

func replicateAllProdVms(replicationEndpoint string, endpointSshPort int, sshKeyLocation string) {
	replicationScriptLocation := "/tmp/replication.sh"
	_, err := os.Stat(replicationScriptLocation)
	if err == nil {
		log.Fatal("another replication process is already running (lock file exists): " + replicationScriptLocation)
	}

	var filteredVmList []string
	var filteredVmListTemp []string
	if len(vmReplicateAllFilter) > 0 {
		filteredVmListTemp = strings.Split(vmReplicateAllFilter, ",")
		for _, v := range filteredVmListTemp {
			v = strings.TrimSpace(v)
			filteredVmList = append(filteredVmList, v)
		}
	}

	allVms, _ := HosterVmUtils.ListJsonApi()
	if len(filteredVmList) > 0 {
		for _, v := range filteredVmList {
			vmFound := false
			for _, vv := range allVms {
				if vv.Name == v {
					vmFound = true
				}
			}
			if !vmFound {
				emojlog.PrintLogMessage("VM from the filtered list was not found: "+v, emojlog.Warning)
				continue
			}
			err := replicateVm(v, replicationEndpoint, endpointSshPort, sshKeyLocation, replicateAllSpeedLimit, replicateAllScriptName)
			if err != nil {
				emojlog.PrintLogMessage("Replication failed for a VM: "+v+" || Exact error: "+err.Error(), emojlog.Error)
			}
		}
	} else {
		for _, v := range allVms {
			if v.Backup {
				continue
			}
			if !v.Running {
				continue
			}

			if v.Production {
				err := replicateVm(v.Name, replicationEndpoint, endpointSshPort, sshKeyLocation, replicateAllSpeedLimit, replicateAllScriptName)
				if err != nil {
					emojlog.PrintLogMessage("Replication failed for a VM: "+v.Name+" || Exact error: "+err.Error(), emojlog.Error)
				}
			}
		}
	}
}
