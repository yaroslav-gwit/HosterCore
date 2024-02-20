// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterVm

import HosterLogger "HosterCore/internal/pkg/logger"

var log = HosterLogger.New()

// Function that helps override the logger settings for this package
// and configure different logging settings from a higher-up function.
func SetLogger(l *HosterLogger.Log) {
	log = l
}
