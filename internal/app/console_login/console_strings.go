// Copyright 2023s The Hoster Authors. All rights reserved.
// Use of this source code is governed by a Apache License 2.0
// license that can be found in the LICENSE fil

package console_login

import (
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/text"
)

var (
	welcomeString      = "Welcome to Hoster"
	waitString         = "Please wait..."
	warningString      = "You've entered an incorrect PIN too many times.\n\n Try again in %vs."
	incorrectPinString = "Incorrect PIN entered"
	sessionInfoString  = "Logged in as %v (automatic logout in %v seconds)"
	firewallInfoString = "Firewall status (%v)\n!! Be careful, your VMs will lose access to the network !!"
	activeString       = "ACTIVE"
	inactiveString     = "INACTIVE"
)

func MakeWarningTring(timeout int) string {
	return fmt.Sprintf(warningString, timeout)
}

func MakeSessionInfoText(user_name string, time int) string {
	return fmt.Sprintf(sessionInfoString, user_name, time)
}

func MakeFirewallInfoText(active bool) string {
	status := inactiveString
	if active {
		status = activeString
	}
	return fmt.Sprintf(firewallInfoString, status)
}

func GetWaitText() *text.Widget {
	wait_message := text.NewFromContentExt(
		text.NewContent([]text.ContentSegment{
			text.StringContent(waitString),
		}),
		text.Options{
			Align: gowid.HAlignMiddle{},
		},
	)
	return wait_message
}
