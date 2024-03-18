package main

import (
	SchedulerUtils "HosterCore/internal/app/scheduler/utils"
	"fmt"
	"os/exec"
	"strings"
)

func Replicate(job SchedulerUtils.ReplicationJob) error {
	out, err := exec.Command("ssh", "-oBatchMode=yes", "-i", job.SshKey, fmt.Sprintf("-p%d", job.SshPort), job.SshEndpoint, "zfs", "list", "-t", "all").CombinedOutput()
	if err != nil {
		return fmt.Errorf("could not a list of remote ZFS snapshots: %s; %s", strings.TrimSpace(string(out)), err.Error())
	}

	fmt.Println(string(out))
	return nil
}
