//go:build freebsd
// +build freebsd

package cmd

import (
	HosterPrometheus "HosterCore/internal/pkg/hoster/prometheus"
	"encoding/json"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var (
	prometheusCmd = &cobra.Command{
		Use:   "prometheus",
		Short: "Prometheus related operations",
		Long:  `Prometheus related operations: generate the config, reload the service, etc`,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
			cmd.Help()

			// r, e := HosterPrometheus.GenerateVmTargets(false)
			// if e != nil {
			// 	log.Fatal(e)
			// }

			// result, _ := json.MarshalIndent(r, "", "   ")
			// fmt.Println(string(result))
		},
	}
)

var (
	prometheusVmsUseIps bool
	prometheusVmsCmd    = &cobra.Command{
		Use:   "vms",
		Short: "",
		Long:  `Generate Prometheus autodiscovery JSON output for all Hoster VMs`,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			// cmd.Help()
			r, e := HosterPrometheus.GenerateVmTargets(prometheusVmsUseIps)
			if e != nil {
				log.Fatal(e)
			}

			result, _ := json.MarshalIndent(r, "", "   ")
			fmt.Println(string(result))
		},
	}
)
