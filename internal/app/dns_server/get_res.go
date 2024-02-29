package main

import (
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
)

type VmInfoStruct struct {
	vmName    string
	vmAddress string
}

func getVmsInfo() []VmInfoStruct {
	vmInfoVar := []VmInfoStruct{}
	allVms, err := HosterVmUtils.ListJsonApi()
	if err != nil {
		return vmInfoVar
	}

	for _, v := range allVms {
		vmInfoVar = append(vmInfoVar, VmInfoStruct{vmName: v.Name, vmAddress: v.Networks[0].IPAddress})
	}
	return vmInfoVar
}

type JailInfoStruct struct {
	JailName    string
	JailAddress string
}

func getJailsInfo() (r []JailInfoStruct) {
	jails, err := HosterJailUtils.ListAllExtendedTable()
	if err != nil {
		return
	}

	for _, v := range jails {
		r = append(r, JailInfoStruct{JailName: v.Name, JailAddress: v.MainIpAddress})
	}
	return
}
