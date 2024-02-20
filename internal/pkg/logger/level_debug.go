// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterLogger

import (
	"HosterCore/internal/pkg/emojlog"
	"fmt"
)

// Test Debug Func
func (l *Log) Debug(value interface{}) {
	if l.Term {
		stringValue := fmt.Sprintf("%s", value)
		emojlog.PrintLogLine(emojlog.LEVEL_DEBUG, stringValue)
	}
	l.Logger.Debug(value)
}

// Test Debug Func
func (l *Log) DebugToFile(value interface{}) {
	l.Logger.Debug(value)
}
