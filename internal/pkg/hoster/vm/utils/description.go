package HosterVmUtils

import "errors"

func UpdateDescription(vmName string, description string) error {
	vm, err := InfoJsonApi(vmName)
	if err != nil {
		return err
	}

	vmConfLoc := vm.Simple.Mountpoint + "/" + vm.Name + "/" + VM_CONFIG_NAME
	if len(description) < 1 {
		return errors.New("description is empty")
	}
	if len(description) > 255 {
		return errors.New("description is too long")
	}

	vm.VmConfig.Description = description
	err = ConfigFileWriter(vm.VmConfig, vmConfLoc)
	if err != nil {
		return err
	}

	return nil
}
