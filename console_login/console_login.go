// Copyright 2023s The Hoster Authors. All rights reserved.
// Use of this source code is governed by a Apache License 2.0
// license that can be found in the LICENSE fil

package main

import (
	"HosterCore/utils/encryption"
	"HosterCore/utils/host"
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
const (
	maxPINLength       = 6
	maximumPINAttempts = 3
)

// ======================================================================
// variables
// widgets
var (
	main_widget  *styled.Widget
	login_dialog *dialog.Widget
	pin_edit     *edit.Widget
)

var (
	pinAttempts = 1
)

// strings
var (
	welcome_string = "Welcome to Hoster"
	wait_string    = "Please wait..."
	warning_string = "You've entered an incorrect PIN too many times.\n\n Try again in 100s."
	incorrect_pin  = "Incorrect PIN entered"
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

func createLoginDialog(holder *styled.Widget) {
	if holder == nil {
		return
	}

	login_button := dialog.Button{
		Msg:    "Login",
		Action: gowid.MakeWidgetCallback("login", gowid.WidgetChangedFunction(pinVerification)),
	}

	flow := gowid.RenderFlow{}
	msg := text.New("Enter PIN to login: ")
	title := hpadding.New(msg, gowid.HAlignMiddle{}, gowid.RenderFixed{})
	pin_edit = edit.New(
		edit.Options{
			Mask:    edit.MakeMask('*'),
			Numeric: edit.MakeNumeric(true, maxPINLength),
		})
	edit := styled.New(
		framed.NewUnicode(pin_edit),
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
}

func createLoginDialogWithError(holder *styled.Widget) {
	if holder == nil {
		return
	}

	login_button := dialog.Button{
		Msg:    "Login",
		Action: gowid.MakeWidgetCallback("login", gowid.WidgetChangedFunction(pinVerification)),
	}

	flow := gowid.RenderFlow{}
	msg := text.New("Enter PIN to login: ")
	title := hpadding.New(msg, gowid.HAlignMiddle{}, gowid.RenderFixed{})
	pin_edit = edit.New(
		edit.Options{
			Mask:    edit.MakeMask('*'),
			Numeric: edit.MakeNumeric(true, maxPINLength),
		})
	edit := styled.New(
		framed.NewUnicode(pin_edit),
		gowid.MakePaletteRef("edit"))
	spacer := divider.NewBlank()
	warning_message :=
		text.NewFromContentExt(
			text.NewContent([]text.ContentSegment{
				text.StyledContent(incorrect_pin, gowid.MakePaletteRef("warning_text")),
			}),
			text.Options{
				Align: gowid.HAlignMiddle{},
			},
		)
	login_dialog = dialog.New(
		framed.NewSpace(vpadding.New(
			pile.New([]gowid.IContainerWidget{
				&gowid.ContainerWidget{IWidget: title, D: flow},
				&gowid.ContainerWidget{IWidget: edit, D: flow},
				&gowid.ContainerWidget{IWidget: spacer, D: flow},
				&gowid.ContainerWidget{IWidget: warning_message, D: flow},
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
}

func showLoginDialog(holder *styled.Widget) {
	if holder == nil {
		return
	}

	app := &gowid.App{}

	createLoginDialog(holder)
	login_dialog.Open(holder, gowid.RenderWithRatio{R: 0.2}, app)
}

func showLoginDialogWithError(holder *styled.Widget) {
	if holder == nil {
		return
	}

	app := &gowid.App{}

	createLoginDialogWithError(holder)
	login_dialog.Open(holder, gowid.RenderWithRatio{R: 0.2}, app)
}

func closeLoginDialog() {
	app := &gowid.App{}

	login_dialog.Close(app)
}

// ======================================================================
// Warning widget

func warningWidget() *styled.Widget {
	warning_message := styled.New(
		text.NewFromContentExt(
			text.NewContent([]text.ContentSegment{
				text.StyledContent(warning_string, gowid.MakePaletteRef("warning_background")),
			}),
			text.Options{
				Align: gowid.HAlignMiddle{},
			},
		),
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

func showWarningWidget() {
	app := &gowid.App{}

	main_widget.SetSubWidget(warningWidget(), app)
}

// ======================================================================
// Main widget

func mainWidget() *styled.Widget {
	widget := loginWidget()

	return widget
}

//======================================================================

func pinVerification(app gowid.IApp, widget gowid.IWidget) {
	// Read PIN from edit
	if pin_edit == nil {
		return
	}

	pin := pin_edit.Text()

	closeLoginDialog()

	// Check the number of pin attempts
	if pinAttempts >= maximumPINAttempts {
		showWarningWidget()
		return
	}

	// Load host config
	hostConfig, err := host.GetHostConfig()
	if err != nil {
		fmt.Println(err)
	}

	// check password hash
	pin_hash := hostConfig.ConsolePanelPin
	match := encryption.CheckPasswordHash(pin, pin_hash)

	if match {
		pinAttempts = 1
		showHomeWidget()
	} else {
		pinAttempts++
		showLoginDialogWithError(main_widget)
	}
}

func showHomeWidget() {
	app := &gowid.App{}

	main_widget.SetSubWidget(homeWidget(), app)
}

//======================================================================

func main() {
	styles := gowid.Palette{
		"body":               gowid.MakeStyledPaletteEntry(gowid.NewUrwidColor("black"), gowid.NewUrwidColor("light gray"), gowid.StyleBold),
		"background":         gowid.MakePaletteEntry(gowid.NewUrwidColor("white"), gowid.NewUrwidColor("blue")),
		"warning_background": gowid.MakeStyledPaletteEntry(gowid.ColorNone, gowid.NewUrwidColor("dark red"), gowid.StyleBold),
		"warning_text":       gowid.MakeStyledPaletteEntry(gowid.NewUrwidColor("dark red"), gowid.ColorNone, gowid.StyleBold),
		"edit":               gowid.MakePaletteEntry(gowid.NewUrwidColor("white"), gowid.NewUrwidColor("dark blue")),
		"banner":             gowid.MakePaletteEntry(gowid.ColorWhite, gowid.MakeRGBColor("#60d")),
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

	// createLoginDialog(main_widget, app)
	showLoginDialog(main_widget)

	app.SimpleMainLoop()
}
