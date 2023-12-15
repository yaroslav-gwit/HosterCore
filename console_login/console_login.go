// Copyright 2023s The Hoster Authors. All rights reserved.
// Use of this source code is governed by a Apache License 2.0
// license that can be found in the LICENSE fil

package main

import (
	"fmt"
	"os"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/columns"
	"github.com/gcla/gowid/widgets/framed"
	"github.com/gcla/gowid/widgets/pile"
	"github.com/gcla/gowid/widgets/styled"
	"github.com/gcla/gowid/widgets/text"
	"github.com/gcla/gowid/widgets/vpadding"
)

// ======================================================================
// Constants
const (
	maxPINLength       = 6
	maximumPINAttempts = 3
	pinTimeout         = 18 // 30 m
)

// ======================================================================
// variables

var (
	app              *gowid.App
	main_widget      *styled.Widget
	login_controller *LoginController
	lock_controller  *LockController
)

// ======================================================================
// Header

func getHeaderText() *text.Widget {
	header_text := text.New(welcome_string, text.Options{Align: gowid.HAlignMiddle{}})
	return header_text
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

// ======================================================================
// Home widget

func homeWidget() *styled.Widget {
	weight_1 := gowid.RenderWithWeight{W: 1}

	widget := styled.New(pile.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{IWidget: getHeaderView(), D: weight_1},
		&gowid.ContainerWidget{IWidget: getInfoSectionView(), D: weight_1},
	}),
		gowid.MakePaletteRef("background"),
	)

	return widget
}

// ======================================================================
// Widgets

func showLoginWidget(app *gowid.App) {
	login_controller = NewLoginController()

	main_widget.SetSubWidget(login_controller.view, app)
	login_controller.ShowLoginDialog(login_controller.view.Widget, app)
}

func showWarningWidget(app *gowid.App) {
	lock_controller = NewLockController()

	main_widget.SetSubWidget(lock_controller.view, app)
	lock_controller.AnimateLock(app)
}

func showHomeWidget(app *gowid.App) {
	main_widget.SetSubWidget(homeWidget(), app)
}

//======================================================================

//======================================================================

func main() {
	var err error

	styles := gowid.Palette{
		"body":               gowid.MakeStyledPaletteEntry(gowid.NewUrwidColor("black"), gowid.NewUrwidColor("light gray"), gowid.StyleBold),
		"background":         gowid.MakePaletteEntry(gowid.NewUrwidColor("white"), gowid.NewUrwidColor("blue")),
		"warning_background": gowid.MakeStyledPaletteEntry(gowid.ColorNone, gowid.NewUrwidColor("dark red"), gowid.StyleBold),
		"warning_text":       gowid.MakeStyledPaletteEntry(gowid.NewUrwidColor("dark red"), gowid.ColorNone, gowid.StyleBold),
		"edit":               gowid.MakePaletteEntry(gowid.NewUrwidColor("white"), gowid.NewUrwidColor("dark blue")),
		"banner":             gowid.MakePaletteEntry(gowid.ColorWhite, gowid.MakeRGBColor("#60d")),
	}

	login_controller = NewLoginController()
	main_widget = login_controller.view.Widget

	app, err = gowid.NewApp(gowid.AppArgs{
		View:    main_widget,
		Palette: &styles,
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	login_controller.ShowLoginDialog(main_widget, app)

	app.SimpleMainLoop()
}
