// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package SchedulerClient

import (
	SchedulerUtils "HosterCore/internal/app/scheduler/utils"
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	HosterLocations "HosterCore/internal/pkg/hoster/locations"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	zfsutils "HosterCore/internal/pkg/zfs_utils"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

func AddReplicationJob(vmName string, endpoint string, key string, speedLimit int) error {
	c, err := net.Dial("unix", SchedulerUtils.SockAddr)
	if err != nil {
		return err
	}

	defer c.Close()

	job := SchedulerUtils.Job{}
	job.JobType = SchedulerUtils.JOB_TYPE_REPLICATION
	job.Replication.ResName = vmName
	job.Replication.SshEndpoint = endpoint
	job.Replication.SshKey = key
	job.Replication.SpeedLimit = speedLimit

	jsonJob, err := json.Marshal(job)
	if err != nil {
		return err
	}

	_, err = c.Write(jsonJob)
	if err != nil {
		return err
	}

	return nil
}

type RemoteDs struct {
	Name       string // normal ZFS name, aka zroot/vm-encrypted/test-vm-1, or zroot/vm-encrypted/test-vm-1@clone_test-vm-100_2023-11-29_19-55-58.222
	MountPoint string // ZFS mountpoint, e.g. /zroot/vm-encrypted/test-vm-1
}

func Replicate(job SchedulerUtils.ReplicationJob) error {
	localDs := ""
	if len(job.ResName) < 1 {
		return fmt.Errorf("resource name cannot be empty")
	}

	mbufferBinary, err := HosterLocations.LocateBinary(HosterLocations.MBUFFER_BINARY_NAME)
	if err != nil {
		return err
	}

	vms, err := HosterVmUtils.ListAllSimple()
	if err != nil {
		return err
	}
	jails, err := HosterJailUtils.ListAllSimple()
	if err != nil {
		return err
	}

	for _, v := range vms {
		if v.VmName == job.ResName {
			localDs = v.DsName + "/" + v.VmName
		}
	}
	if len(localDs) < 1 {
		for _, v := range jails {
			if v.JailName == job.ResName {
				localDs = v.DsName + "/" + v.JailName
			}
		}
	}
	if len(localDs) < 1 {
		return fmt.Errorf("could not find resource specified")
	}

	out, err := exec.Command("ssh", "-oBatchMode=yes", "-i", job.SshKey, fmt.Sprintf("-p%d", job.SshPort), job.SshEndpoint, "zfs", "list", "-t", "all", "-o", "name,mountpoint").CombinedOutput()
	if err != nil {
		return fmt.Errorf("could not a list of remote ZFS snapshots: %s; %s", strings.TrimSpace(string(out)), err.Error())
	}

	reSplitSpace := regexp.MustCompile(`\s+`)
	var remoteDs []string
	var toRemove []string
	var localSnaps []string
	var toReplicate []string
	var commonSnaps []string
	for i, v := range strings.Split(string(out), "\n") {
		if i == 0 {
			continue
		}
		if len(strings.TrimSpace(v)) < 1 {
			continue
		}

		split := reSplitSpace.Split(v, -1)
		if split[0] == localDs || strings.Contains(split[0], localDs+"@") {
			// if len(split) > 1 {
			// 	remoteDs = append(remoteDs, RemoteDs{Name: strings.TrimSpace(split[0]), MountPoint: strings.TrimSpace(split[1])})
			// } else {
			// 	remoteDs = append(remoteDs, RemoteDs{Name: strings.TrimSpace(split[0])})
			// }
			remoteDs = append(remoteDs, split[0])
		}
	}

	if len(remoteDs) == 1 {
		return fmt.Errorf("remote dataset exists")
	}

	_, _, err = zfsutils.TakeScheduledSnapshot(localDs, zfsutils.TYPE_REPLICATION, 5)
	if err != nil {
		return err
	}

	snaps, err := zfsutils.SnapshotListAll()
	if err != nil {
		return err
	}

	for _, v := range snaps {
		if v.Dataset == localDs {
			localSnaps = append(localSnaps, v.Name)
		}
	}

	for _, v := range localSnaps {
		if !slices.Contains(remoteDs, v) {
			if strings.Contains(v, "@") {
				toReplicate = append(toReplicate, v)
			}
		}
	}
	for _, v := range remoteDs {
		if !slices.Contains(localSnaps, v) {
			if strings.Contains(v, "@") {
				toRemove = append(toRemove, v)
			}
		} else {
			commonSnaps = append(commonSnaps, v)
		}
	}

	// jsonOut, err := json.MarshalIndent(remoteDs, "", "   ")
	// if err != nil {
	// 	return err
	// }

	// fmt.Println(string(jsonOut))
	if len(remoteDs) > 1 && len(commonSnaps) < 1 {
		return fmt.Errorf("could not find any common snapshots, remote resource must have a conflicting name with our local one")
	}

	// fmt.Printf("%s: %v\n", "To Remove", toRemove)
	// fmt.Printf("%s: %v\n", "To Replicate", toReplicate)
	// fmt.Printf("%s: %v\n", "Common", commonSnaps)

	var replicateCmds []string
	var removeCmds []string
	// Remove the old snaps first
	for _, v := range toRemove {
		cmd := fmt.Sprintf("ssh -oBatchMode=yes -i %s -p%d %s zfs destroy %s", job.SshKey, job.SshPort, job.SshEndpoint, v)
		removeCmds = append(removeCmds, cmd)
	}

	// Send initial snapshot
	if len(remoteDs) < 1 {
		if job.SpeedLimit > 0 {
			os.Setenv("SPEED_LIMIT_MB_PER_SECOND", strconv.Itoa(job.SpeedLimit))
		}
		cmd := fmt.Sprintf("zfs send -P -v %s | %s | ssh -oBatchMode=yes -i %s -p%d %s zfs receive %s", toReplicate[0], mbufferBinary, job.SshKey, job.SshPort, job.SshEndpoint, localDs)
		replicateCmds = append(replicateCmds, cmd)

		for _, v := range replicateCmds {
			fmt.Println(v)
		}
		return nil
	}

	// Prepend the first common snapshot to the replication list
	var tmp []string
	tmp = append(tmp, commonSnaps[len(commonSnaps)-1])
	tmp = append(tmp, toReplicate...)
	copy(toReplicate, tmp)

	// Send incremental snapshots
	for i, v := range toReplicate {
		if i >= len(toReplicate)-1 {
			break
		}

		if job.SpeedLimit > 0 {
			os.Setenv("SPEED_LIMIT_MB_PER_SECOND", strconv.Itoa(job.SpeedLimit))
		}
		cmd := fmt.Sprintf("zfs send -P -vi %s %s | %s | ssh -i %s -p%d %s zfs -F receive %s", v, toReplicate[i+1], mbufferBinary, job.SshKey, job.SshPort, job.SshEndpoint, localDs)
		replicateCmds = append(replicateCmds, cmd)
	}

	for _, v := range removeCmds {
		fmt.Println("Remote Snaps to remove")
		fmt.Println(v)
	}
	for _, v := range replicateCmds {
		fmt.Println("Snaps to replicate")
		fmt.Println(v)
	}
	return nil
}
