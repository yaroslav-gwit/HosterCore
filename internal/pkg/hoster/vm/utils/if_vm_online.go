package HosterVmUtils

import "slices"

// This function checks if the VM is online.
// It's not the "cheapest" version available, but it's not very heavy either.
func IsVmOnline(vmName string) (bool, error) {
	running, err := GetRunningVms()
	if err != nil {
		return false, err
	}

	if len(running) == 0 {
		return false, nil
	}

	if slices.Contains(running, vmName) {
		return true, nil
	}

	return false, nil
}
