// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package zfsutils

import (
	FileExists "HosterCore/internal/pkg/file_exists"
	"encoding/json"
	"os"
	"time"
)

// Read the Snapshot List cache.
func ReadSnapshotCache() (r []SnapshotInfo, e error) {
	if !SnapshotCacheOutdated(SNAPSHOT_CACHE_FILE) {
		f, err := os.ReadFile(SNAPSHOT_CACHE_FILE)
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
		snapshots, err := WriteSnapshotCache()
		if err != nil {
			e = err
			return
		}
		r = snapshots
	}

	return
}

func WriteSnapshotCache() (r []SnapshotInfo, e error) {
	snapshots, err := SnapshotListWithDescriptions()
	if err != nil {
		e = err
		return
	}

	j, err := json.Marshal(snapshots)
	if err != nil {
		e = err
		return
	}

	err = os.WriteFile(SNAPSHOT_CACHE_FILE, j, 0644)
	if err != nil {
		e = err
		return
	}

	r = snapshots
	return
}

// If this function returns true, the cache is outdated or outright missing
func SnapshotCacheOutdated(filePath string) bool {
	if !FileExists.CheckUsingOsStat(SNAPSHOT_CACHE_FILE) {
		return true
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return true
	}

	modTime := fileInfo.ModTime()
	now := time.Now()

	// If the file is older than 10 minutes, it's outdated
	if now.Sub(modTime) < 10*time.Minute {
		return false
	} else {
		return true
	}
}
