// Copyright 2023s The Hoster Authors. All rights reserved.
// Use of this source code is governed by a Apache License 2.0
// license that can be found in the LICENSE fil

package console_login

import (
	"HosterCore/internal/pkg/encryption"
	"HosterCore/internal/pkg/host"
	"HosterCore/internal/pkg/widgets/dialog"
	"HosterCore/internal/pkg/widgets/edit"
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/divider"
	"github.com/gcla/gowid/widgets/framed"
	"github.com/gcla/gowid/widgets/hpadding"
	"github.com/gcla/gowid/widgets/pile"
	"github.com/gcla/gowid/widgets/styled"
	"github.com/gcla/gowid/widgets/text"
	"github.com/gcla/gowid/widgets/vpadding"
)

var (
	pinAttempts  = 1
	pin_edit     *edit.Widget
	login_dialog *dialog.Widget
)

type LoginController struct {
	view *LoginView
}

func NewLoginController() *LoginController {
	res := &LoginController{nil}
	view := NewLoginView(res)
	res.view = view

	return res
}

// ======================================================================
// Dialogs

func (c *LoginController) CreateLoginDialog(holder *styled.Widget) {
	if holder == nil {
		return
	}

	login_button := dialog.Button{
		Msg:    "Login",
		Action: gowid.MakeWidgetCallback("login", gowid.WidgetChangedFunction(c.PinVerification)),
	}

	flow := gowid.RenderFlow{}
	msg := text.New("Enter PIN to login: ")
	title := hpadding.New(msg, gowid.HAlignMiddle{}, gowid.RenderFixed{})
	pin_edit = edit.New(
		edit.Options{
			Mask:    edit.MakeMask('*'),
			Numeric: edit.MakeNumeric(true, c.GetMaxPINLength()),
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

func (c *LoginController) CreateLoginDialogWithError(holder *styled.Widget) {
	if holder == nil {
		return
	}

	login_button := dialog.Button{
		Msg:    "Login",
		Action: gowid.MakeWidgetCallback("login", gowid.WidgetChangedFunction(c.PinVerification)),
	}

	flow := gowid.RenderFlow{}
	msg := text.New("Enter PIN to login: ")
	title := hpadding.New(msg, gowid.HAlignMiddle{}, gowid.RenderFixed{})
	pin_edit = edit.New(
		edit.Options{
			Mask:    edit.MakeMask('*'),
			Numeric: edit.MakeNumeric(true, c.GetMaxPINLength()),
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

func (c *LoginController) ShowLoginDialog(holder *styled.Widget, app *gowid.App) {
	if holder == nil {
		return
	}

	c.CreateLoginDialog(holder)
	if login_dialog != nil {
		login_dialog.Open(holder, gowid.RenderWithRatio{R: 0.2}, app)
	}
}

func (c *LoginController) ShowLoginDialogWithError(holder *styled.Widget, app *gowid.App) {
	if holder == nil {
		return
	}

	c.CreateLoginDialogWithError(holder)
	if login_dialog != nil {
		login_dialog.Open(holder, gowid.RenderWithRatio{R: 0.2}, app)
	}
}

func (c *LoginController) CloseLoginDialog(app *gowid.App) {
	pin_edit = nil

	if login_dialog != nil {
		login_dialog.Close(app)
	}
}

// ======================================================================
// Verification

func (c *LoginController) PinVerification(_ gowid.IApp, widget gowid.IWidget) {
	// Read PIN from edit
	pin := c.GetPinFromDialog()

	login_controller.CloseLoginDialog(app)

	// Load host config
	hostConfig, err := host.GetHostConfig()
	if err != nil {
		fmt.Println(err)
	}

	// check password hash
	pin_hash := hostConfig.ConsolePanel.PIN
	match := encryption.CheckPasswordHash(pin, pin_hash)

	if match {
		pinAttempts = 1
		showHomeWidget(app)
	} else {
		pinAttempts++
		login_controller.ShowLoginDialogWithError(main_widget, app)
	}

	// Check the number of pin attempts
	if pinAttempts > c.GetMaximumPINAttempts() {
		showWarningWidget(app)
		pinAttempts = 1
	}
}

func (c *LoginController) GetPinFromDialog() string {
	if pin_edit == nil {
		return ""
	}

	return pin_edit.Text()
}

func (c *LoginController) GetMaximumPINAttempts() int {
	maximumPINAttempts := host_config.ConsolePanel.MaximumPINAttempts

	if maximumPINAttempts == 0 {
		return maximumPINAttemptsDefault
	}

	return maximumPINAttempts
}

func (c *LoginController) GetMaxPINLength() int {
	maxPINLength := host_config.ConsolePanel.MaxPINLength

	if maxPINLength == 0 {
		return maxPINLengthDefault
	}

	return maxPINLength
}
