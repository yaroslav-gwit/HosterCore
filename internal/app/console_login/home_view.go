// Copyright 2023s The Hoster Authors. All rights reserved.
// Use of this source code is governed by a Apache License 2.0
// license that can be found in the LICENSE fil

package console_login

import (
	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/columns"
	"github.com/gcla/gowid/widgets/divider"
	"github.com/gcla/gowid/widgets/framed"
	"github.com/gcla/gowid/widgets/pile"
	"github.com/gcla/gowid/widgets/styled"
	"github.com/gcla/gowid/widgets/text"
	"github.com/gcla/gowid/widgets/vpadding"
)

type HomeView struct {
	*styled.Widget
	controller      *HomeController
	sessionTime     int
	sessionTimeText *text.Widget
}

var userName = "root"

func NewHomeView(controller *HomeController) *HomeView {
	sessionTime := sessionTimeDefault
	if controller != nil {
		sessionTime = controller.GetSessionTime()
	}

	sessionTimeText := text.New(MakeSessionInfoText(userName, sessionTime), text.Options{
		Align: gowid.HAlignMiddle{},
	})

	view := MakeHomeWidget(sessionTimeText)

	res := &HomeView{
		Widget:          view,
		controller:      controller,
		sessionTime:     sessionTime,
		sessionTimeText: sessionTimeText,
	}

	return res
}

func MakeHomeWidget(session_time_text *text.Widget) *styled.Widget {
	flow := gowid.RenderFlow{}

	widget := styled.New(framed.NewUnicode(pile.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{IWidget: getHeaderText(), D: flow},
		&gowid.ContainerWidget{IWidget: session_time_text, D: flow},
		&gowid.ContainerWidget{IWidget: divider.NewUnicode(), D: flow},
	})),
		gowid.MakePaletteRef("background"),
	)

	return widget
}

func (v *HomeView) UpdateSessionTime(_ gowid.IApp) bool {
	v.sessionTimeText.SetText(MakeSessionInfoText(userName, v.sessionTime), app)

	v.sessionTime--
	if v.sessionTime < 0 {
		v.controller.StopAnimation()
		v.controller.ResetSessionTime(app)
	}

	return true
}

func getSessionInfoText() *text.Widget {
	info_text := text.NewFromContentExt(
		text.NewContent([]text.ContentSegment{
			text.StringContent("Logged in as"),
			text.StringContent(" "),
			text.StringContent("ROOT"),
			text.StringContent(" "),
			text.StringContent("(automatic logout in"),
			text.StringContent(" "),
			text.StringContent("100"),
			text.StringContent(" "),
			text.StringContent("seconds)"),
		}), text.Options{Align: gowid.HAlignMiddle{}})

	return info_text
}

func getHeaderView() *styled.Widget {
	flow := gowid.RenderFlow{}

	header_view := styled.New(
		framed.NewSpace(
			styled.New(
				framed.NewUnicode(
					vpadding.New(
						pile.New([]gowid.IContainerWidget{
							&gowid.ContainerWidget{IWidget: getHeaderText(), D: flow},
							&gowid.ContainerWidget{IWidget: getSessionInfoText(), D: flow},
						}),
						gowid.VAlignMiddle{},
						flow),
				),
				gowid.MakePaletteRef("body"),
			),
		),
		gowid.MakePaletteRef("screen edge"),
	)

	return header_view
}

// ======================================================================
// Info section
func createInfoView(name string) *framed.Widget {
	if len(name) == 0 {
		name = "-/-"
	}

	flow := gowid.RenderFlow{}
	info_view := framed.NewSpace(
		styled.New(
			framed.NewUnicode(
				vpadding.New(
					text.New(name, text.Options{Align: gowid.HAlignMiddle{}}),
					gowid.VAlignMiddle{},
					flow),
			),
			gowid.MakePaletteRef("body"),
		),
	)
	return info_view
}

func getInfoSectionView() *framed.Widget {
	weight_1 := gowid.RenderWithWeight{W: 1}

	infoSectionView := framed.NewSpace(
		styled.New(
			framed.NewUnicode(
				columns.New([]gowid.IContainerWidget{
					&gowid.ContainerWidget{IWidget: createInfoView("Press \"F\" to toggle a firewall"), D: weight_1},
					&gowid.ContainerWidget{IWidget: createInfoView("Up/Down to select a VM"), D: weight_1},
					&gowid.ContainerWidget{IWidget: createInfoView("\"S\" to start or stop a VM"), D: weight_1},
					&gowid.ContainerWidget{IWidget: createInfoView("\"I\" to display more info"), D: weight_1},
					&gowid.ContainerWidget{IWidget: createInfoView("\"R\" to enter a command mode"), D: weight_1},
				}),
			),
			gowid.MakePaletteRef("body"),
		),
	)

	return infoSectionView
}
