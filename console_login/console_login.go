// Copyright 2023s The Hoster Authors. All rights reserved.
// Use of this source code is governed by a Apache License 2.0
// license that can be found in the LICENSE fil

package main

import (
	hostconfig "HosterCore/models/host_config"
	"HosterCore/utils/host"
	"fmt"
	"os"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/styled"
	"github.com/gcla/gowid/widgets/text"
)

// ======================================================================
// Constants
const (
	maxPINLengthDefault       = 6
	maximumPINAttemptsDefault = 3
	lockTimeoutDefault        = 1800 // 30 m
)

// ======================================================================
// variables

var (
	host_config      hostconfig.Config
	app              *gowid.App
	main_widget      *styled.Widget
	login_controller *LoginController
	lock_controller  *LockController
	home_controller  *HomeController
)

// ======================================================================
// Header

func getHeaderText() *text.Widget {
	header_text := text.New(welcome_string, text.Options{Align: gowid.HAlignMiddle{}})
	return header_text
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
	home_controller = NewHomeController()

	main_widget.SetSubWidget(home_controller.view, app)
}

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

	host_config, err = host.GetHostConfig()
	if err != nil {
		fmt.Println(err)
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
