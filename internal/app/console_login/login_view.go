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
	"github.com/gcla/gowid/widgets/vpadding"
)

type LoginView struct {
	*styled.Widget
	controller *LoginController
}

func NewLoginView(controller *LoginController) *LoginView {
	view := MakeLoginWidget()

	res := &LoginView{
		Widget:     view,
		controller: controller,
	}

	return res
}

func MakeLoginWidget() *styled.Widget {
	flow := gowid.RenderFlow{}
	weight_1 := gowid.RenderWithWeight{W: 1}

	wait_widget := vpadding.New(
		pile.New([]gowid.IContainerWidget{
			&gowid.ContainerWidget{IWidget: GetWaitText(), D: flow},
		}),
		gowid.VAlignMiddle{},
		flow)

	widget := styled.New(framed.NewUnicode(pile.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{IWidget: getHeaderText(), D: flow},
		&gowid.ContainerWidget{IWidget: divider.NewUnicode(), D: flow},
		&gowid.ContainerWidget{IWidget: wait_widget, D: weight_1},
	})),
		gowid.MakePaletteRef("background"),
	)

	return widget
}
