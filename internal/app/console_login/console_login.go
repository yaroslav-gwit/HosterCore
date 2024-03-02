// Copyright 2023s The Hoster Authors. All rights reserved.
// Use of this source code is governed by a Apache License 2.0
// license that can be found in the LICENSE fil

package console_login

import (
	hostconfig "HosterCore/internal/models/host_config"
	"HosterCore/internal/pkg/host"
	"fmt"

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
	sessionTimeDefault        = 300  // 5 m
)

// ======================================================================
// variables

var (
	host_config     hostconfig.Config
	app             *gowid.App
	mainWidget      *styled.Widget
	loginController *LoginController
	lockController  *LockController
	homeController  *HomeController
)

// ======================================================================
// Header

func getHeaderText() *styled.Widget {
	header_text := styled.New(
		text.New(welcomeString, text.Options{Align: gowid.HAlignMiddle{}}),
		gowid.MakePaletteRef("info_text"),
	)
	return header_text
}

// ======================================================================
// Widgets

func showLoginWidget(app *gowid.App) {
	loginController = NewLoginController()

	mainWidget.SetSubWidget(loginController.view, app)
	loginController.ShowLoginDialog(loginController.view.Widget, app)
}

func showWarningWidget(app *gowid.App) {
	lockController = NewLockController()

	mainWidget.SetSubWidget(lockController.view, app)
	lockController.AnimateLock(app)
}

func showHomeWidget(app *gowid.App) {
	homeController = NewHomeController()

	mainWidget.SetSubWidget(homeController.view, app)
	homeController.AnimateSessionTime(app)
}

// ======================================================================
func New() error {
	var err error

	styles := gowid.Palette{
		"body":                         gowid.MakeStyledPaletteEntry(gowid.NewUrwidColor("black"), gowid.NewUrwidColor("light gray"), gowid.StyleBold),
		"background":                   gowid.MakePaletteEntry(gowid.NewUrwidColor("white"), gowid.NewUrwidColor("light blue")),
		"warning_background":           gowid.MakeStyledPaletteEntry(gowid.ColorNone, gowid.NewUrwidColor("dark red"), gowid.StyleBold),
		"firewall_active_background":   gowid.MakeStyledPaletteEntry(gowid.ColorNone, gowid.NewUrwidColor("dark green"), gowid.StyleBold),
		"firewall_inactive_background": gowid.MakeStyledPaletteEntry(gowid.ColorNone, gowid.NewUrwidColor("dark red"), gowid.StyleBold),
		"info_text":                    gowid.MakeStyledPaletteEntry(gowid.NewUrwidColor("white"), gowid.ColorNone, gowid.StyleBold),
		"warning_text":                 gowid.MakeStyledPaletteEntry(gowid.NewUrwidColor("dark red"), gowid.ColorNone, gowid.StyleBold),
		"edit":                         gowid.MakePaletteEntry(gowid.NewUrwidColor("white"), gowid.NewUrwidColor("dark blue")),
		"progress normal":              gowid.MakeStyledPaletteEntry(gowid.NewUrwidColor("black"), gowid.NewUrwidColor("light gray"), gowid.StyleBold),
		"progress complete":            gowid.MakeStyleMod(gowid.MakePaletteRef("progress normal"), gowid.MakeBackground(gowid.NewUrwidColor("light green"))),
		"red":                          gowid.MakePaletteEntry(gowid.ColorRed, gowid.ColorBlack),
		"invred":                       gowid.MakePaletteEntry(gowid.ColorBlack, gowid.ColorRed),
	}

	host_config, err = host.GetHostConfig()
	if err != nil {
		fmt.Println(err)
	}

	loginController = NewLoginController()
	mainWidget = loginController.view.Widget

	app, err = gowid.NewApp(gowid.AppArgs{
		View:    mainWidget,
		Palette: &styles,
	})

	if err != nil {
		return err
	}

	loginController.ShowLoginDialog(mainWidget, app)

	app.SimpleMainLoop()

	return nil
}
