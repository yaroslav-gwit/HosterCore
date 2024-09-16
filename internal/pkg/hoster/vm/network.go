package HosterVm

import (
	HosterHostUtils "HosterCore/internal/pkg/hoster/host/utils"
	HosterNetwork "HosterCore/internal/pkg/hoster/network"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"errors"
)

func AddNewVmNetwork(vmName string, network HosterVmUtils.VmNetwork) error {
	if !HosterVmUtils.IsMacAddressValid(network.NetworkMac) {
		return errors.New("invalid MAC address")
	}

	vm, err := HosterVmUtils.InfoJsonApi(vmName)
	if err != nil {
		return err
	}

	ipConflict := false
	macConflict := false
	networkBridgeFound := false

	vms, err := HosterVmUtils.ListJsonApi()
	if err != nil {
		return err
	}

	netConfig, err := HosterNetwork.GetNetworkConfig()
	if err != nil {
		return err
	}

	net := HosterNetwork.NetworkConfig{}
	for _, v := range netConfig {
		if v.BridgeInterface == network.NetworkBridge {
			networkBridgeFound = true
			net = v
			break
		}
	}
	if !networkBridgeFound {
		return errors.New("network bridge not found")
	}
	if len(network.IPAddress) < 1 {
		HosterHostUtils.GenerateNewRandomIp(network.NetworkBridge)
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

	vm.Networks = append(vm.Networks, network)
	err = HosterVmUtils.ConfigFileWriter(vm.VmConfig, vm.Simple.Mountpoint+"/"+vm.Name+"/config.json")
	if err != nil {
		return err
	}

	return nil
}
