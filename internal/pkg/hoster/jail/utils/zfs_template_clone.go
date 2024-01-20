package HosterJailUtils

import (
	HosterZfs "HosterCore/internal/pkg/hoster/zfs"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func ZfsTemplateClone(jailName string, dsParent string, release string) error {
	// dsExists, err := doesDatasetExist(fmt.Sprintf("%s/jail-template-%s", dsParent, release))
	// if err != nil {
	// 	return err
	// }
	// if !dsExists {
	// 	return fmt.Errorf("parent dataset does not exist: %s/jail-template-%s", dsParent, release)
	// }
	templateDataset := dsParent + "/jail-template-" + release

	mountPoints, err := HosterZfs.ListMountPoints()
	if err != nil {
		return err
	}
	mpFound := false
	for _, v := range mountPoints {
		if templateDataset == v.DsName {
			mpFound = true
		}
	}
	if !mpFound {
		return fmt.Errorf("template dataset does not exist: %s", templateDataset)
	}

	timeNow := time.Now().Format("2006-01-02_15-04-05.000")
	jailSnapshotName := dsParent + "/jail-template-" + release + "@deployment_" + jailName + "_" + timeNow
	out, err := exec.Command("zfs", "snapshot", jailSnapshotName).CombinedOutput()
	if err != nil {
		return fmt.Errorf("could not execute zfs snapshot: %s; %s", strings.TrimSpace(string(out)), err.Error())
	}

	out, err = exec.Command("zfs", "clone", jailSnapshotName, dsParent+"/"+jailName).CombinedOutput()
	if err != nil {
		return fmt.Errorf("could not execute zfs clone: %s; %s", strings.TrimSpace(string(out)), err.Error())
	}

	return nil
}
