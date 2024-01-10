// Copyright 2023s The Hoster Authors. All rights reserved.
// Use of this source code is governed by a Apache License 2.0
// license that can be found in the LICENSE fil

package console_login

import (
	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/divider"
	"github.com/gcla/gowid/widgets/framed"
	"github.com/gcla/gowid/widgets/pile"
	"github.com/gcla/gowid/widgets/styled"
	"github.com/gcla/gowid/widgets/text"
	"github.com/gcla/gowid/widgets/vpadding"
)

type LockView struct {
	*styled.Widget
	controller *LockController
	timeout    int
	text       *text.Widget
}

func NewLockView(controller *LockController) *LockView {
	lockTimeout := lockTimeoutDefault
	if controller != nil {
		lockTimeout = controller.GetLockTimeout()
	}

	text := text.New(MakeWarningTring(lockTimeout), text.Options{
		Align: gowid.HAlignMiddle{},
	})

	view := MakeLockWidget(text)

	res := &LockView{
		Widget:     view,
		controller: controller,
		timeout:    lockTimeout,
		text:       text,
	}

	return res
}

func MakeLockWidget(lock_timeout_text *text.Widget) *styled.Widget {
	warning_message := styled.New(
		lock_timeout_text,
		gowid.MakePaletteRef("warning_background"),
	)

	flow := gowid.RenderFlow{}

	div := divider.NewBlank()
	outside := styled.New(div, gowid.MakePaletteRef("warning_background"))
	inside := styled.New(div, gowid.MakePaletteRef("warning_background"))

	warning_widget := styled.New(
		vpadding.New(
			pile.New([]gowid.IContainerWidget{
				&gowid.ContainerWidget{IWidget: outside, D: flow},
				&gowid.ContainerWidget{IWidget: inside, D: flow},
				&gowid.ContainerWidget{IWidget: warning_message, D: flow},
				&gowid.ContainerWidget{IWidget: inside, D: flow},
				&gowid.ContainerWidget{IWidget: outside, D: flow},
			}),
			gowid.VAlignMiddle{},
			flow),
		gowid.MakePaletteRef("background"),
	)

	weight_1 := gowid.RenderWithWeight{W: 1}

	widget := styled.New(framed.NewUnicode(pile.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{IWidget: getHeaderText(), D: flow},
		&gowid.ContainerWidget{IWidget: divider.NewUnicode(), D: flow},
		&gowid.ContainerWidget{IWidget: warning_widget, D: weight_1},
	})),
		gowid.MakePaletteRef("background"),
	)

	return widget
}

func (v *LockView) UpdateLock(_ gowid.IApp) bool {
	v.text.SetText(MakeWarningTring(v.timeout), app)

	v.timeout--
	if v.timeout < 0 {
		v.controller.StopAnimation()
		v.controller.ResetLock(app)
		showLoginWidget(app)
	}

	return true
}
