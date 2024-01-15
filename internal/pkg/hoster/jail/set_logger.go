// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterJail

import HosterLogger "HosterCore/internal/pkg/logger"

var log = HosterLogger.New()

// Override the logger for this package
func SetLogger(l *HosterLogger.Log) {
	log = l
}
