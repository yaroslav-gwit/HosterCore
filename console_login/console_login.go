// Copyright 2023s The Hoster Authors. All rights reserved.
// Use of this source code is governed by a Apache License 2.0
// license that can be found in the LICENSE fil

package main

import (
	dialog "HosterCore/widgets/dialog"
	edit "HosterCore/widgets/edit"
	"fmt"
	"os"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/columns"
	"github.com/gcla/gowid/widgets/divider"
	"github.com/gcla/gowid/widgets/framed"
	"github.com/gcla/gowid/widgets/hpadding"
	"github.com/gcla/gowid/widgets/pile"
	"github.com/gcla/gowid/widgets/styled"
	"github.com/gcla/gowid/widgets/text"
	"github.com/gcla/gowid/widgets/vpadding"
)

// ======================================================================
// Constants
const MaxPINLength = 6

// variables
var (
	main_widget  *styled.Widget
	login_dialog *dialog.Widget
)

// strings
var (
	welcome_string = "Welcome to Hoster"
	wait_string    = "Please wait..."
)

//======================================================================
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
// Login widget

func getWaitText() *text.Widget {
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

func loginWidget() *styled.Widget {
	flow := gowid.RenderFlow{}
	weight_1 := gowid.RenderWithWeight{W: 1}

	wait_widget := vpadding.New(
		pile.New([]gowid.IContainerWidget{
			&gowid.ContainerWidget{IWidget: getWaitText(), D: flow},
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

func showLoginDialog(holder *styled.Widget, app *gowid.App) {
	if holder == nil || app == nil {
		return
	}

	login_button := dialog.Button{
		Msg:    "Login",
		Action: gowid.MakeWidgetCallback("login", gowid.WidgetChangedFunction(pinVerification)),
	}

	flow := gowid.RenderFlow{}
	msg := text.New("Enter PIN to login: ")
	title := hpadding.New(msg, gowid.HAlignMiddle{}, gowid.RenderFixed{})
	edit := styled.New(
		framed.NewUnicode(
			edit.New(
				edit.Options{
					Mask:    edit.MakeMask('*'),
					Numeric: edit.MakeNumeric(true, MaxPINLength),
				})),
		gowid.MakePaletteRef("edit"))
	login_dialog = dialog.New(
		framed.NewSpace(vpadding.New(
			pile.New([]gowid.IContainerWidget{
				&gowid.ContainerWidget{IWidget: title, D: flow},
				&gowid.ContainerWidget{IWidget: edit, D: flow},
			}),
			gowid.VAlignMiddle{},
			flow)),
		dialog.Options{
			Buttons:       []dialog.Button{login_button},
			NoEscapeClose: true,
			FocusOnWidget: true,
			AutoFocusOn:   true,
			Modal:         true,
		},
	)
	login_dialog.Open(holder, gowid.RenderWithRatio{R: 0.2}, app)
}

// ======================================================================
// Warning widget

func warningWidget() *styled.Widget {
	warning_message := text.NewFromContentExt(
		text.NewContent([]text.ContentSegment{
			text.StyledContent("Too many login attempts wait 100 s to unlock", gowid.MakePaletteRef("warning_text")),
		}),
		text.Options{
			Align: gowid.HAlignMiddle{},
		},
	)

	flow := gowid.RenderFlow{}

	widget := styled.New(
		vpadding.New(
			pile.New([]gowid.IContainerWidget{
				&gowid.ContainerWidget{IWidget: warning_message, D: flow},
			}),
			gowid.VAlignMiddle{},
			flow),
		gowid.MakePaletteRef("warning_background"),
	)

	return widget
}

// ======================================================================
// Main widget

func mainWidget() *styled.Widget {
	widget := loginWidget()

	return widget
}

//======================================================================

func pinVerification(app gowid.IApp, widget gowid.IWidget) {
	showHomeWidget(app, widget)
}

func showHomeWidget(app gowid.IApp, widget gowid.IWidget) {
	login_dialog.Close(app)
	main_widget.SetSubWidget(homeWidget(), app)
}

//======================================================================

func main() {
	styles := gowid.Palette{
		"body":       gowid.MakeStyledPaletteEntry(gowid.NewUrwidColor("black"), gowid.NewUrwidColor("light gray"), gowid.StyleBold),
		"background": gowid.MakePaletteEntry(gowid.NewUrwidColor("white"), gowid.NewUrwidColor("blue")),
		// "text_info":          gowid.MakePaletteEntry(gowid.NewUrwidColor("white"), gowid.ColorNone),
		"warning_background": gowid.MakePaletteEntry(gowid.ColorNone, gowid.NewUrwidColor("dark red")),
		// "warning_text":       gowid.MakePaletteEntry(gowid.NewUrwidColor("white"), gowid.ColorNone),
		"edit": gowid.MakePaletteEntry(gowid.NewUrwidColor("white"), gowid.NewUrwidColor("dark blue")),
	}

	main_widget = mainWidget()
	app, err := gowid.NewApp(gowid.AppArgs{
		View:    main_widget,
		Palette: &styles,
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	showLoginDialog(main_widget, app)

	app.SimpleMainLoop()
}
