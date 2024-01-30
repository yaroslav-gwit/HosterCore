// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterVmUtils

import (
	"os/exec"
	"strings"
)

// Returns a simple, flat list of running `bhyve` VMs, using the `ls /dev/vmm/`.
//
// Error handling is still TBD. Function simply returns an empty list if there was an error of any sort.
func GetRunningVms() (r []string, e error) {
	out, err := exec.Command("ls", "-1", "/dev/vmm/").CombinedOutput()
	if err != nil {
		// Need to implement some error value matching here.
		// For example, if the `/dev/vmm/` folder doesn't exist - we can safely ignore that,
		// because it means that there are zero VMs online and `bhyve` simply hasn't created the folder yet.
		//
		// But that is totally different from the file/directory access permission issues, which should be matched and returned.
		//
		// e = fmt.Errorf("could not list the running VMs in /dev/vmm/: %s; %s", strings.TrimSpace(string(out)), err.Error())
		return
	}

	r = append(r, strings.Split(string(out), "\n")...)
	return
}
