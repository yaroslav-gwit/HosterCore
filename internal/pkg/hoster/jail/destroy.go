package HosterJail

import (
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

func Destroy(jailName string) error {
	// If the logger was already set, ignore this
	if !log.ConfigSet {
		log.SetFileLocation(HosterJailUtils.JAIL_AUDIT_LOG_LOCATION)
	}
	log.Info("Destroying the Jail: " + jailName)

	// Check if the Jail is running/online
	running, err := isJailRunning(jailName)
	if err != nil {
		return err
	}
	if running {
		errorValue := "Can't destroy - the Jail is still running: " + jailName
		log.ErrorToFile(errorValue)
		return errors.New(errorValue)
	}
	// EOF Check if the Jail is running/online

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
	jailDs := jailDsInfo.DsName + "/" + jailName
	// EOF Check if Jail exists and get it's dataset configuration

	// Get the parent dataset
	out, err := exec.Command("zfs", "get", "-H", "origin", jailDs).CombinedOutput()
	// Correct value:
	// NAME                           PROPERTY  VALUE                                                                           SOURCE
	//    [0]                             [1]           [2]                                                                     [3]
	// zroot/vm-encrypted/twelveFour	origin	zroot/vm-encrypted/jail-template-12.4-RELEASE@deployment_twelveFour_qc5q7u6khy	-
	// Empty value
	// zroot/vm-encrypted/wordpress-one	origin	         -	                                                                     -
	if err != nil {
		errorValue := "could not find a parent DS: " + strings.TrimSpace(string(out)) + "; " + err.Error()
		return fmt.Errorf("%s", errorValue)
	}

	reSpaceSplit := regexp.MustCompile(`\s+`)
	parentDataset := reSpaceSplit.Split(strings.TrimSpace(string(out)), -1)[2]
	// EOF Get the parent dataset

	// Remove the Jail dataset
	out, err = exec.Command("zfs", "destroy", "-r", jailDs).CombinedOutput()
	if err != nil {
		errorValue := "could not remove the dataset " + jailDs + ": " + strings.TrimSpace(string(out)) + "; " + err.Error()
		return fmt.Errorf("%s", errorValue)
	}
	log.Warn("Jail dataset has been destroyed: " + jailDs)
	// EOF Remove the Jail dataset

	// Remove the parent dataset if it exists
	reMatch := regexp.MustCompile(`deployment_`)
	if len(parentDataset) > 1 && reMatch.MatchString(parentDataset) {
		out, err := exec.Command("zfs", "destroy", parentDataset).CombinedOutput()
		if err != nil {
			errorValue := "FATAL: " + strings.TrimSpace(string(out)) + "; " + err.Error()
			return fmt.Errorf("%s", errorValue)
		}
		log.Warn("Jail parent dataset has been destroyed: " + parentDataset)
	}
	// EOF Remove the parent dataset if it exists

	return nil
}
