// Copyright 2023s The Hoster Authors. All rights reserved.
// Use of this source code is governed by a Apache License 2.0
// license that can be found in the LICENSE fil

package console_login

import (
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/columns"
	"github.com/gcla/gowid/widgets/divider"
	"github.com/gcla/gowid/widgets/framed"
	"github.com/gcla/gowid/widgets/list"
	"github.com/gcla/gowid/widgets/palettemap"
	"github.com/gcla/gowid/widgets/pile"
	"github.com/gcla/gowid/widgets/progress"
	"github.com/gcla/gowid/widgets/selectable"
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

var (
	userName = "root"
	flow     = gowid.RenderFlow{}
	weight1  = gowid.RenderWithWeight{W: 1}
	weight5  = gowid.RenderWithWeight{W: 5}
)

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

func (v *HomeView) UpdateSessionTime(_ gowid.IApp) bool {
	v.sessionTimeText.SetText(MakeSessionInfoText(userName, v.sessionTime), app)

	v.sessionTime--
	if v.sessionTime < 0 {
		v.controller.StopAnimation()
		v.controller.ResetSessionTime(app)
	}

	return true
}

func MakeHomeWidget(sessionTimeText *text.Widget) *styled.Widget {
	widget := styled.New(framed.NewUnicode(pile.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{IWidget: getHeaderText(), D: flow},
		&gowid.ContainerWidget{IWidget: sessionTimeText, D: flow},
		&gowid.ContainerWidget{IWidget: divider.NewUnicode(), D: flow},
		&gowid.ContainerWidget{IWidget: Info(), D: flow},
		&gowid.ContainerWidget{IWidget: FirewallInfo(), D: flow},
		&gowid.ContainerWidget{IWidget: ListFrame(), D: flow},
	})),
		gowid.MakePaletteRef("background"),
	)

	return widget
}

func Info() *styled.Widget {
	widget := styled.New(
		columns.New([]gowid.IContainerWidget{
			&gowid.ContainerWidget{IWidget: CPUInfo(), D: weight1},
			&gowid.ContainerWidget{IWidget: RAMInfo(), D: weight1},
		}),
		gowid.MakePaletteRef("background"),
	)

	return widget
}

func CPUInfo() *styled.Widget {
	cpuUsageText := styled.New(text.New("CPU Usage", text.Options{Align: gowid.HAlignMiddle{}}), gowid.MakePaletteRef("info_text"))
	cpuInfoText := text.New("CPU Info (sockets,\n cores,\n threads)", text.Options{Align: gowid.HAlignMiddle{}})
	cpuProgress := progress.New(progress.Options{
		Normal:   gowid.MakePaletteRef("progress normal"),
		Complete: gowid.MakePaletteRef("progress complete"),
	})

	cpuProgress.SetProgress(app, 50)

	widget := styled.New(framed.NewUnicode(pile.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{IWidget: cpuUsageText, D: flow},
		&gowid.ContainerWidget{IWidget: cpuInfoText, D: flow},
		&gowid.ContainerWidget{IWidget: cpuProgress, D: flow},
	})),
		gowid.MakePaletteRef("background"),
	)

	return widget
}

func RAMInfo() *styled.Widget {
	ramUsageText := styled.New(text.New("RAM Usage", text.Options{Align: gowid.HAlignMiddle{}}), gowid.MakePaletteRef("info_text"))
	ramInfoText := text.New("RAM Info (Ram Free / Ram Overall)", text.Options{Align: gowid.HAlignMiddle{}})
	ramProgress := progress.New(progress.Options{
		Normal:   gowid.MakePaletteRef("progress normal"),
		Complete: gowid.MakePaletteRef("progress complete"),
	})
	swapInfoText := text.New("SWAP Info (Ram Free / Ram Overall)", text.Options{Align: gowid.HAlignMiddle{}})
	swapProgress := progress.New(progress.Options{
		Normal:   gowid.MakePaletteRef("progress normal"),
		Complete: gowid.MakePaletteRef("progress complete"),
	})

	widget := styled.New(framed.NewUnicode(pile.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{IWidget: ramUsageText, D: flow},
		&gowid.ContainerWidget{IWidget: ramInfoText, D: flow},
		&gowid.ContainerWidget{IWidget: ramProgress, D: flow},
		&gowid.ContainerWidget{IWidget: swapInfoText, D: flow},
		&gowid.ContainerWidget{IWidget: swapProgress, D: flow},
	})),
		gowid.MakePaletteRef("background"),
	)

	return widget
}

func FirewallInfo() *styled.Widget {
	infoText := text.New(MakeFirewallInfoText(true), text.Options{
		Align: gowid.HAlignMiddle{},
	})

	infoMessage := styled.New(
		infoText,
		gowid.MakePaletteRef("firewall_active_background"),
	)

	div := divider.NewBlank()
	inside := styled.New(div, gowid.MakePaletteRef("firewall_active_background"))

	warningWidget := styled.New(
		vpadding.New(
			pile.New([]gowid.IContainerWidget{
				&gowid.ContainerWidget{IWidget: inside, D: flow},
				&gowid.ContainerWidget{IWidget: infoMessage, D: flow},
				&gowid.ContainerWidget{IWidget: inside, D: flow},
			}),
			gowid.VAlignMiddle{},
			flow),
		gowid.MakePaletteRef("background"),
	)

	return warningWidget
}

func ListFrame() *styled.Widget {
	return styled.New(
		framed.NewUnicode(pile.New([]gowid.IContainerWidget{
			&gowid.ContainerWidget{IWidget: ListHeader(), D: flow},
			&gowid.ContainerWidget{IWidget: ListBody(), D: flow},
		})),
		gowid.MakePaletteRef("info_text"),
	)
}

func ListHeader() *styled.Widget {
	return styled.New(
		framed.NewUnicode(
			columns.New([]gowid.IContainerWidget{
				&gowid.ContainerWidget{IWidget: text.New("ID", text.Options{Align: gowid.HAlignMiddle{}}), D: weight1},
				&gowid.ContainerWidget{IWidget: text.New("NAME", text.Options{Align: gowid.HAlignMiddle{}}), D: weight5},
				&gowid.ContainerWidget{IWidget: text.New("STATUS", text.Options{Align: gowid.HAlignMiddle{}}), D: weight1},
				&gowid.ContainerWidget{IWidget: text.New("UPTIME", text.Options{Align: gowid.HAlignMiddle{}}), D: weight1},
			}),
		),
		gowid.MakePaletteRef("info_text"),
	)
}

func ListBody() *styled.Widget {
	widgets := make([]gowid.IWidget, 0)
	nl := gowid.MakePaletteRef

	for i := 0; i < 23; i++ {
		t := text.NewContent([]text.ContentSegment{
			text.StyledContent(fmt.Sprintf("abc%dd\t%d\t%d\t%d", i, i, i, i), nl("invred")),
		})
		mt := text.NewFromContent(t)
		mta := selectable.New(palettemap.New(mt, palettemap.Map{}, palettemap.Map{"invred": "red"}))
		widgets = append(widgets, mta)
	}

	walker := list.NewSimpleListWalker(widgets)
	lb := list.New(walker)

	return styled.New(
		framed.NewUnicode(
			columns.New([]gowid.IContainerWidget{
				&gowid.ContainerWidget{IWidget: lb, D: weight5},
			}),
		),
		gowid.MakePaletteRef("info_text"),
	)
}
