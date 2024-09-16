package HosterVm

import (
	HosterHostUtils "HosterCore/internal/pkg/hoster/host/utils"
	HosterNetwork "HosterCore/internal/pkg/hoster/network"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"errors"
)

func AddNewVmNetwork(vmName string, network HosterVmUtils.VmNetwork) error {
	if len(network.NetworkMac) < 1 {
		var err error
		network.NetworkMac, err = HosterVmUtils.GenerateMacAddress()
		if err != nil {
			return err
		}
	}

	if !HosterVmUtils.IsMacAddressValid(network.NetworkMac) {
		return errors.New("invalid MAC address")
	}
	if len(network.IPAddress) < 1 {
		var err error
		network.IPAddress, err = HosterHostUtils.GenerateNewRandomIp(network.NetworkBridge)
		if err != nil {
			return err
		}
	}

	if network.NetworkAdaptorType != "virtio-net" {
		if network.NetworkAdaptorType != "e1000" {
			return errors.New("invalid network driver type")
		}
	}

	if len(network.Comment) < 1 {
		network.Comment = "New VM network"
	}

	vm, err := HosterVmUtils.InfoJsonApi(vmName)
	if err != nil {
		return err
	}

	ipConflict := false
	macConflict := false

	vms, err := HosterVmUtils.ListJsonApi()
	if err != nil {
		return err
	}

	netConfig, err := HosterNetwork.GetNetworkConfig()
	if err != nil {
		return err
	}

	networkBridgeFound := false
	net := HosterNetwork.NetworkConfig{}
	for _, v := range netConfig {
		if v.NetworkName == network.NetworkBridge {
			networkBridgeFound = true
			net = v
			break
		}
	}
	if !networkBridgeFound {
		return errors.New("network bridge not found")
	}

OUTER:
	for _, v := range vms {
		for _, vv := range v.Networks {
			if vv.IPAddress == network.IPAddress {
				ipConflict = true
				break OUTER
			}
			if vv.NetworkMac == network.NetworkMac {
				macConflict = true
				break OUTER
			}
		}
	}

	if !HosterHostUtils.IsIpWithinRange(network.IPAddress, net.Subnet, net.RangeStart, net.RangeEnd) {
		return errors.New("IP address is not within the network range")
	}

	if ipConflict {
		return errors.New("IP address is already in use")
	}
	if macConflict {
		return errors.New("MAC address is already in use")
	}

	vm.VmConfig.Networks = append(vm.Networks, network)
	err = HosterVmUtils.ConfigFileWriter(vm.VmConfig, vm.Simple.Mountpoint+"/"+vm.Name+"/config.json")
	if err != nil {
		return err
	}

	return nil
}
