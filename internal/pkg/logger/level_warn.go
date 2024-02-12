// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterLogger

import (
	"HosterCore/internal/pkg/emojlog"
	"fmt"
)

// Test Error Func
func (l *Log) Warn(value interface{}) {
	if l.Term {
		stringValue := fmt.Sprintf("%s", value)
		emojlog.PrintLogLine(emojlog.Warning, stringValue)
	}
	l.Logger.Warn(value)
}

// Test Error Func
func (l *Log) WarnToFile(value interface{}) {
	l.Logger.Warn(value)
}
