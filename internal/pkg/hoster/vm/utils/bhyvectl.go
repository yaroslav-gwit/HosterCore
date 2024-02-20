// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterVmUtils

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// A very simple wrapper for a bhyvectl --destroy command.
// Takes in a VM name as a parameter, and returns an error if something went wrong.
func BhyveCtlDestroy(vmName string) error {
	out, err := exec.Command("bhyvectl", "--destroy", "--vm="+vmName).CombinedOutput()
	if err != nil {
		message := fmt.Sprintf("bhyvectl failed: %s; %s", strings.TrimSpace(string(out)), err.Error())
		return errors.New(message)
	}

	return nil
}
