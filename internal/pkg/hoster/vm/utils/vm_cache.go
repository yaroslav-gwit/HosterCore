// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterVmUtils

import (
	FileExists "HosterCore/internal/pkg/file_exists"
	"encoding/json"
	"os"
	"slices"
	"time"
)

// Read the VM cache. The functions contains 2 side effects:
//
// 1. If the cache file is outdated or missing, it will be updated.
//
// 2. The Running (VM online/offline) field will be updated with the current state of the VM.
func ReadCache() (r []VmApi, e error) {
	if !CacheOutdated(VM_CACHE_FILE) {
		f, err := os.ReadFile(VM_CACHE_FILE)
		if err != nil {
			e = err
			return
		}

		err = json.Unmarshal(f, &r)
		if err != nil {
			e = err
			return
		}
	} else {
		vms, err := WriteCache()
		if err != nil {
			e = err
			return
		}
		r = vms
	}

	online, err := GetRunningVms()
	if err != nil {
		e = err
		return
	}

	for i, v := range r {
		if slices.Contains(online, v.Name) {
			r[i].Running = true
		} else {
			r[i].Running = false
		}
	}

	return
}

// Write the VM cache
func WriteCache() (r []VmApi, e error) {
	vms, err := ListJsonApi()
	if err != nil {
		e = err
		return
	}

	j, err := json.Marshal(vms)
	if err != nil {
		e = err
		return
	}

	err = os.WriteFile(VM_CACHE_FILE, j, 0644)
	if err != nil {
		e = err
		return
	}

	return
}

// If this function returns true, the cache is outdated or outright missing
func CacheOutdated(filePath string) bool {
	if !FileExists.CheckUsingOsStat(VM_CACHE_FILE) {
		return true
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return true
	}

	modTime := fileInfo.ModTime()
	now := time.Now()

	// If the file is older than 5 minutes, it's outdated
	if now.Sub(modTime) < 5*time.Minute {
		return false
	} else {
		return true
	}
}
