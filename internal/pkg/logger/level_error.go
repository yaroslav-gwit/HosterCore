package HosterLogger

import (
	"HosterCore/internal/pkg/emojlog"
	"fmt"
)

// Test Error Func
func (l *Log) Error(value interface{}) {
	if l.Term {
		stringValue := fmt.Sprintf("%s", value)
		emojlog.PrintLogLine(emojlog.Error, stringValue)
	}
	l.Logger.Error(value)
}

// Test Error Func
func (l *Log) ErrorToFile(value interface{}) {
	l.Logger.Error(value)
}
