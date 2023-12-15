// Copyright 2023s The Hoster Authors. All rights reserved.
// Use of this source code is governed by a Apache License 2.0
// license that can be found in the LICENSE fil

package main

import (
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/text"
)

var (
	welcome_string = "Welcome to Hoster"
	wait_string    = "Please wait..."
	warning_string = "You've entered an incorrect PIN too many times.\n\n Try again in %vs."
	incorrect_pin  = "Incorrect PIN entered"
)

func MakeWarningTring(timeout int) string {
	return fmt.Sprintf(warning_string, timeout)
}

func GetWaitText() *text.Widget {
	wait_message := text.NewFromContentExt(
		text.NewContent([]text.ContentSegment{
			text.StringContent(wait_string),
		}),
		text.Options{
			Align: gowid.HAlignMiddle{},
		},
	)
	return wait_message
}
