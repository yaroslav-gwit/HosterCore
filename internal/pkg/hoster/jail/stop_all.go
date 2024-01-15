// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterJail

import (
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

func StopAll() error {
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
		if err != nil {
			log.Error(err.Error())
			continue
		}
		if !running {
			continue
		}

		// Insert an empty spacer on every one but first iteration
		if first {
			first = false
		} else {
			log.Spacer()
		}

		log.Info("Stopping the Jail: " + v.JailName)
		jailDsInfo := HosterJailUtils.JailListSimple{}
		for _, vv := range jails {
			if v.JailName == vv.JailName {
				jailDsInfo = v
			}
		}

		jailTempRuntimeLocation := jailDsInfo.MountPoint.Mountpoint + "/" + v.JailName + "/" + HosterJailUtils.JAIL_TEMP_RUNTIME
		out, err := exec.Command("jail", "-f", jailTempRuntimeLocation, "-r", v.JailName).CombinedOutput()
		if err != nil {
			errorValue := fmt.Sprintf("%s; %s", strings.TrimSpace(string(out)), err.Error())
			log.ErrorToFile(errorValue)
			return errors.New(errorValue)
		}

		err = HosterJailUtils.RemoveUptimeStateFile(v.JailName)
		if err != nil {
			log.ErrorToFile(err.Error())
			return err
		}

		log.Info("Jail has been stopped: " + v.JailName)
	}

	return nil
}
