// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterJail

import (
	HosterHostUtils "HosterCore/internal/pkg/hoster/host/utils"
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"strings"
)

type CiResetInput struct {
	OldJailName string `json:"old_jail_name"`
	DeployInput
}

// This function is used to fully reset the Jail configuration.
// It's mostly used for node migration purposes and Jail clones, when you want to change the IP address, network name, etc etc.
func CiReset(input CiResetInput) error {
	var err error
	prod := input.Production

	// If the logger was already set, ignore this
	if !log.ConfigSet {
		log.SetFileLocation(HosterJailUtils.JAIL_AUDIT_LOG_LOCATION)
	}

	info, err := HosterJailUtils.InfoJsonApi(input.OldJailName)
	if err != nil {
		return err
	}

	// Validate new Jail name (if given)
	if len(input.JailName) > 0 {
		err = HosterVmUtils.ValidateResName(input.JailName)
		if err != nil {
			return err
		}
	}
	// EOF Validate new Jail name (if given)

	log.Info("Resetting Jail config: " + input.JailName)
	jailFolder := ""
	jailDataset := ""
	if len(input.JailName) > 0 {
		jailDataset = info.Simple.DsName + "/" + input.OldJailName
		jailFolder = input.DsParent + "/" + input.JailName
	} else {
		jailDataset = info.Simple.DsName + "/" + input.JailName
		jailFolder = input.DsParent + "/" + input.OldJailName
	}

	jailConfig, err := generateJailDeployConfig(input.CpuLimit, input.RamLimit, input.IpAddress, input.Network, input.DnsServer, prod)
	if err != nil {
		return err
	}

	if len(input.JailName) > 0 {
		out, err := exec.Command("zfs", "rename", info.Simple.DsName+"/"+input.OldJailName, jailDataset).CombinedOutput()
		if err != nil {
			e := fmt.Errorf("failed to rename ZFS dataset: %s", strings.TrimSpace(string(out)))
			return e
		}
	}

	// Create jail_config.json
	templateJail, err := template.New("templateJailConfigJson").Parse(HosterJailUtils.TemplateJailConfigJson)
	if err != nil {
		return err
	}
	fileTemplateJail, err := os.Create(fmt.Sprintf("%s/jail_config.json", jailFolder))
	if err != nil {
		return err
	}
	err = templateJail.Execute(fileTemplateJail, jailConfig)
	if err != nil {
		return err
	}
	// EOF Create jail_config.json

	err = HosterHostUtils.ReloadDns()
	if err != nil {
		return err
	}
	return nil
}
