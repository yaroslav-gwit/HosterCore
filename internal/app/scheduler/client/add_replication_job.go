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
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"time"
)

func AddReplicationJob(replJob SchedulerUtils.ReplicationJob) error {
	c, err := net.Dial("unix", SchedulerUtils.SockAddr)
	if err != nil {
		return err
	}
	defer c.Close()

	job := SchedulerUtils.Job{}
	output, resType, err := Replicate(replJob)
	if err != nil {
		return err
	}

	job.JobType = SchedulerUtils.JOB_TYPE_REPLICATION
	job.Replication = output
	job.ResType = resType
	job.Replication.ResName = replJob.ResName
	job.Replication.SpeedLimit = replJob.SpeedLimit

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

func AddReplicationByTagJob(tag string, sshKeyFile string, sshEndpoint string, sshPort int, speedLimit int) error {
	vms, err := HosterVmUtils.ListJsonApi()
	if err != nil {
		return err
	}

	jails, err := HosterJailUtils.ListJsonApi()
	if err != nil {
		return err
	}

	// Validate input fields
	if len(tag) < 1 {
		return fmt.Errorf("tag cannot be empty")
	}
	if len(sshKeyFile) < 1 {
		return fmt.Errorf("ssh key file cannot be empty")
	}
	if len(sshEndpoint) < 1 {
		return fmt.Errorf("ssh endpoint cannot be empty")
	}
	if sshPort < 1 {
		return fmt.Errorf("ssh port cannot be less than 1")
	}
	if speedLimit < 1 {
		return fmt.Errorf("speed limit cannot be less than 1")
	}

	for _, v := range vms {
		if v.Backup {
			continue
		}

		for _, vv := range v.Tags {
			if vv == tag {
				job := SchedulerUtils.ReplicationJob{}
				job.ResName = v.Name
				job.SshKey = sshKeyFile
				job.SshEndpoint = sshEndpoint
				job.SshPort = sshPort
				job.SpeedLimit = speedLimit

				err = AddReplicationJob(job)
				if err != nil {
					return err
				}
				time.Sleep(1500 * time.Millisecond)
			}
		}
	}

	for _, v := range jails {
		if v.Backup {
			continue
		}

		for _, vv := range v.Tags {
			if vv == tag {
				job := SchedulerUtils.ReplicationJob{}
				job.ResName = v.Name
				job.SshKey = sshKeyFile
				job.SshEndpoint = sshEndpoint
				job.SshPort = sshPort
				job.SpeedLimit = speedLimit

				err = AddReplicationJob(job)
				if err != nil {
					return err
				}
				time.Sleep(1500 * time.Millisecond)
			}
		}
	}

	return nil
}

func Replicate(job SchedulerUtils.ReplicationJob) (r SchedulerUtils.ReplicationJob, resType string, e error) {
	localDs := ""
	if len(job.ResName) < 1 {
		e = fmt.Errorf("resource name cannot be empty")
		return
	}

	mbufferBinary, err := HosterLocations.LocateBinary(HosterLocations.MBUFFER_BINARY_NAME)
	if err != nil {
		e = err
		return
	}

	vms, err := HosterVmUtils.ListAllSimple()
	if err != nil {
		e = err
		return
	}
	jails, err := HosterJailUtils.ListAllSimple()
	if err != nil {
		e = err
		return
	}

	for _, v := range vms {
		if v.VmName == job.ResName {
			localDs = v.DsName + "/" + v.VmName
			resType = "VM"
		}
	}
	if len(localDs) < 1 {
		for _, v := range jails {
			if v.JailName == job.ResName {
				localDs = v.DsName + "/" + v.JailName
				resType = "Jail"
			}
		}
	}
	if len(localDs) < 1 {
		e = fmt.Errorf("could not find resource specified")
		return
	}

	// rsName, _, err := zfsutils.TakeScheduledSnapshot(localDs, zfsutils.TYPE_REPLICATION, 5)
	_, _, err = zfsutils.TakeScheduledSnapshot(localDs, zfsutils.TYPE_REPLICATION, 5)
	if err != nil {
		e = err
		return
	}
	// fmt.Println("Took a new snapshot: " + rsName)

	out, err := exec.Command("ssh", "-oStrictHostKeyChecking=accept-new", "-oBatchMode=yes", "-i", job.SshKey, fmt.Sprintf("-p%d", job.SshPort), job.SshEndpoint, "zfs", "list", "-t", "all", "-o", "name,mountpoint").CombinedOutput()
	if err != nil {
		e = fmt.Errorf("could not a list of remote ZFS snapshots: %s; %s", strings.TrimSpace(string(out)), err.Error())
		return
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
			remoteDs = append(remoteDs, split[0])
		}
	}

	if len(remoteDs) == 1 {
		e = fmt.Errorf("remote dataset exists")
		return
	}

	snaps, err := zfsutils.SnapshotListAll()
	if err != nil {
		e = err
		return
	}

	for _, v := range snaps {
		if v.Dataset == localDs {
			localSnaps = append(localSnaps, v.Name)
		}
	}

	customSnapExists := false
	for _, v := range localSnaps {
		if strings.Contains(v, "@custom") {
			customSnapExists = true
		}
	}
	if !customSnapExists {
		_, _, err := zfsutils.TakeScheduledSnapshot(localDs, zfsutils.TYPE_CUSTOM, 5000)
		if err != nil {
			e = err
			return
		}

		snaps, err = zfsutils.SnapshotListAll()
		if err != nil {
			e = err
			return
		}

		localSnaps = []string{}
		for _, v := range snaps {
			if v.Dataset == localDs {
				localSnaps = append(localSnaps, v.Name)
			}
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

	if len(remoteDs) > 1 && len(commonSnaps) < 1 {
		e = fmt.Errorf("could not find any common snapshots for %s on the remote endpoint", job.ResName)
		return
	}

	var replicateCmds []string
	var removeCmds []string
	// Remove the old snaps first
	if len(commonSnaps) > 1 {
		for _, v := range toRemove {
			cmd := fmt.Sprintf("ssh -oStrictHostKeyChecking=accept-new -oBatchMode=yes -i %s -p%d %s zfs destroy %s", job.SshKey, job.SshPort, job.SshEndpoint, v)
			removeCmds = append(removeCmds, cmd)
		}
	}

	// Send initial snapshot
	if len(remoteDs) < 1 {
		// if job.SpeedLimit > 0 {
		// 	os.Setenv("SPEED_LIMIT_MB_PER_SECOND", strconv.Itoa(job.SpeedLimit))
		// }
		cmd := fmt.Sprintf("zfs send -P -v %s | %s | ssh -oStrictHostKeyChecking=accept-new -oBatchMode=yes -i %s -p%d %s zfs receive %s", toReplicate[0], mbufferBinary, job.SshKey, job.SshPort, job.SshEndpoint, localDs)
		replicateCmds = append(replicateCmds, cmd)
	} else {
		// Prepend the first common snapshot to the replication list
		var tmp []string
		tmp = append(tmp, commonSnaps[len(commonSnaps)-1])
		tmp = append(tmp, toReplicate...)
		toReplicate = []string{}
		toReplicate = append(toReplicate, tmp...)

		// Send incremental snapshots
		for i, v := range toReplicate {
			if i+1 >= len(toReplicate) {
				break
			}

			// if job.SpeedLimit > 0 {
			// 	os.Setenv("SPEED_LIMIT_MB_PER_SECOND", strconv.Itoa(job.SpeedLimit))
			// }
			cmd := fmt.Sprintf("zfs send -P -vi %s %s | %s | ssh -oStrictHostKeyChecking=accept-new -i %s -p%d %s zfs receive -F %s", v, toReplicate[i+1], mbufferBinary, job.SshKey, job.SshPort, job.SshEndpoint, localDs)
			replicateCmds = append(replicateCmds, cmd)
		}
	}

	r.ScriptsRemove = append(r.ScriptsRemove, removeCmds...)
	r.ScriptsReplicate = append(r.ScriptsReplicate, replicateCmds...)

	return
}
