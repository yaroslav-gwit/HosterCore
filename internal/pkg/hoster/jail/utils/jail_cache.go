// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterJailUtils

import (
	FileExists "HosterCore/internal/pkg/file_exists"
	"encoding/json"
	"os"
	"time"
)

// Read the Jails cache. The functions contains 2 side effects:
//
// 1. If the cache file is outdated or missing, it will be updated.
//
// 2. The Running (Jail online/offline) field will be updated with the current state of the Jail.
func ReadCache() (r []JailApi, e error) {
	if !CacheOutdated(JAIL_CACHE_FILE) {
		f, err := os.ReadFile(JAIL_CACHE_FILE)
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
		jails, err := WriteCache()
		if err != nil {
			e = err
			return
		}
		r = jails
	}

	online, err := GetRunningJails()
	if err != nil {
		e = err
		return
	}

	for i, v := range r {
		found := false
		for _, vv := range online {
			if v.Name == vv.Name {
				r[i].Running = true
				found = true
			}
		}
		if !found {
			r[i].Running = false
		}
	}

	return
}

// Write the VM cache
func WriteCache() (r []JailApi, e error) {
	jails, err := ListJsonApi()
	if err != nil {
		e = err
		return
	}

	j, err := json.Marshal(jails)
	if err != nil {
		e = err
		return
	}

	err = os.WriteFile(JAIL_CACHE_FILE, j, 0644)
	if err != nil {
		e = err
		return
	}

	r = jails
	return
}

// If this function returns true, the cache is outdated or outright missing
func CacheOutdated(filePath string) bool {
	if !FileExists.CheckUsingOsStat(JAIL_CACHE_FILE) {
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
