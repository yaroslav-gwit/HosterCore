// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterLogger

import (
	"HosterCore/internal/pkg/emojlog"
	"fmt"
)

// Test Info Func
func (l *Log) Info(value interface{}) {
	if l.Term {
		stringValue := fmt.Sprintf("%s", value)
		emojlog.PrintLogLine(emojlog.Info, stringValue)
	}
	l.Logger.Info(value)
}

// Test Info Func
func (l *Log) InfoToFile(value interface{}) {
	l.Logger.Info(value)
}
