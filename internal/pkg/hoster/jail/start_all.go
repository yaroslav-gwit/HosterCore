// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterJail

import (
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	HosterNetwork "HosterCore/internal/pkg/hoster/network"
	"fmt"
	"os/exec"
	"strings"
)

func StartAll() error {
	// If the logger was already set, ignore this
	if !log.ConfigSet {
		log.SetFileLocation(HosterJailUtils.JAIL_AUDIT_LOG_LOCATION)
	}

	jails, err := HosterJailUtils.ListAllSimple()
	if err != nil {
		return err
	}

	first := true
	for _, v := range jails {
		running, err := isJailRunning(v.JailName)
		if err != nil || running {
			log.Error(err.Error())
			continue
		}
		if running {
			continue
		}

		// Insert an empty spacer on every one but first iteration
		if first {
			first = false
		} else {
			log.Spacer()
		}

		log.Info("Starting the Jail: " + v.JailName)

		jailDsInfo := HosterJailUtils.JailListSimple{}
		for _, vv := range jails {
			if vv.JailName == v.JailName {
				jailDsInfo = v
			}
		}
		jailDsFolder := jailDsInfo.MountPoint.Mountpoint + "/" + v.JailName

		jailConfig, err := HosterJailUtils.GetJailConfig(jailDsFolder)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		ifaces, err := HosterNetwork.CreateEpairInterface(v.JailName, jailConfig.Network)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		err = createMissingConfigFiles(jailConfig, jailDsFolder+"/"+HosterJailUtils.JAIL_ROOT_FOLDER)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		jailStartConf, err := setJailStartValues(v.JailName, jailDsFolder, jailConfig, ifaces)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		jailTempRuntimeLocation, err := generatePartialTemplate(jailStartConf, jailConfig, jailDsFolder)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		out, err := exec.Command("jail", "-f", jailTempRuntimeLocation, "-c").CombinedOutput()
		if err != nil {
			errorValue := fmt.Sprintf("%s; %s", strings.TrimSpace(string(out)), err.Error())
			log.Error(errorValue)
			continue
		}

		err = HosterJailUtils.CreateUptimeStateFile(v.JailName)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		log.Info("The Jail is now running: " + v.JailName)
	}

	return nil
}
