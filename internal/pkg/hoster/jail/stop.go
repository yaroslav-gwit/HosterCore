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

// Stops a Jail using it's name. Has it's own logging configured.
func Stop(jailName string) error {
	// If the logger was already set, ignore this
	if !log.ConfigSet {
		log.SetFileLocation(HosterJailUtils.JAIL_AUDIT_LOG_LOCATION)
	}
	log.Info("Stopping the Jail: " + jailName)

	running, err := isJailRunning(jailName)
	if err != nil {
		return err
	}
	if !running {
		errorValue := "Jail is already offline: " + jailName
		log.ErrorToFile(errorValue)
		return errors.New(errorValue)
	}

	// Check if Jail exists and get it's dataset configuration
	jails, err := HosterJailUtils.ListAllSimple()
	if err != nil {
		return err
	}
	jailDsInfo := HosterJailUtils.JailListSimple{}
	jailFound := false
	for _, v := range jails {
		if v.JailName == jailName {
			jailFound = true
			jailDsInfo = v
		}
	}
	if !jailFound {
		errorValue := fmt.Sprintf("Jail doesn't exist: %s", jailName)
		log.ErrorToFile(errorValue)
		return errors.New(errorValue)
	}
	jailTempRuntimeLocation := jailDsInfo.Mountpoint + "/" + jailName + "/" + HosterJailUtils.JAIL_TEMP_RUNTIME
	// EOF Check if Jail exists and get it's dataset configuration

	out, err := exec.Command("jail", "-f", jailTempRuntimeLocation, "-r", jailName).CombinedOutput()
	if err != nil {
		errorValue := fmt.Sprintf("%s; %s", strings.TrimSpace(string(out)), err.Error())
		log.ErrorToFile(errorValue)
		return errors.New(errorValue)
	}

	err = HosterJailUtils.RemoveUptimeStateFile(jailName)
	if err != nil {
		log.ErrorToFile(err.Error())
		return err
	}

	log.Info("Jail has been stopped: " + jailName)
	return nil
}
