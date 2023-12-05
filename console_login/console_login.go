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
var main_widget *styled.Widget
var login_dialog *dialog.Widget

//======================================================================
// Header

func GetHeaderText() *text.Widget {
	header_text := text.New("Welcome to Hoster", text.Options{Align: gowid.HAlignMiddle{}})
	return header_text
}

func GetSessionInfoText() *text.Widget {
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

func GetHeaderView() *styled.Widget {
	flow := gowid.RenderFlow{}

	header_view := styled.New(
		framed.NewSpace(
			styled.New(
				framed.NewUnicode(
					vpadding.New(
						pile.New([]gowid.IContainerWidget{
							&gowid.ContainerWidget{IWidget: GetHeaderText(), D: flow},
							&gowid.ContainerWidget{IWidget: GetSessionInfoText(), D: flow},
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
func CreateInfoView(name string) *framed.Widget {
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

func GetInfoSectionView() *framed.Widget {
	weight_1 := gowid.RenderWithWeight{W: 1}

	infoSectionView := framed.NewSpace(
		styled.New(
			framed.NewUnicode(
				columns.New([]gowid.IContainerWidget{
					&gowid.ContainerWidget{IWidget: CreateInfoView("Press \"F\" to toggle a firewall"), D: weight_1},
					&gowid.ContainerWidget{IWidget: CreateInfoView("Up/Down to select a VM"), D: weight_1},
					&gowid.ContainerWidget{IWidget: CreateInfoView("\"S\" to start or stop a VM"), D: weight_1},
					&gowid.ContainerWidget{IWidget: CreateInfoView("\"I\" to display more info"), D: weight_1},
					&gowid.ContainerWidget{IWidget: CreateInfoView("\"R\" to enter a command mode"), D: weight_1},
				}),
			),
			gowid.MakePaletteRef("body"),
		),
	)

	return infoSectionView
}

// ======================================================================
// Home widget

func HomeWidget() *styled.Widget {
	weight_1 := gowid.RenderWithWeight{W: 1}

	widget := styled.New(pile.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{IWidget: GetHeaderView(), D: weight_1},
		&gowid.ContainerWidget{IWidget: GetInfoSectionView(), D: weight_1},
	}),
		gowid.MakePaletteRef("background"),
	)

	return widget
}

// ======================================================================
// Login widget

func LoginWidget() *styled.Widget {
	flow := gowid.RenderFlow{}
	widget := styled.New(framed.NewUnicode(pile.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{IWidget: GetHeaderText(), D: flow},
		&gowid.ContainerWidget{IWidget: divider.NewUnicode(), D: flow},
	})),
		gowid.MakePaletteRef("background"),
	)

	return widget
}

func ShowLoginDialog(holder *styled.Widget, app *gowid.App) {
	if holder == nil || app == nil {
		return
	}

	login_button := dialog.Button{
		Msg:    "Login",
		Action: gowid.MakeWidgetCallback("login", gowid.WidgetChangedFunction(PINVerification)),
	}

	flow := gowid.RenderFlow{}
	msg := text.New("Enter PIN to login: ")
	title := hpadding.New(msg, gowid.HAlignMiddle{}, gowid.RenderFixed{})
	edit := styled.New(framed.NewUnicode(edit.New(edit.Options{Mask: edit.MakeMask('*'), Numeric: edit.MakeNumeric(true, MaxPINLength)})), gowid.MakePaletteRef("edit"))
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
			Modal:         true,
		},
	)
	login_dialog.Open(holder, gowid.RenderWithRatio{R: 0.2}, app)
}

// ======================================================================
// Warning widget

func WarningWidget() *styled.Widget {
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

func MainWidget() *styled.Widget {
	widget := LoginWidget()

	return widget
}

//======================================================================

func PINVerification(app gowid.IApp, widget gowid.IWidget) {
	ShowHomeWidget(app, widget)
}

func ShowHomeWidget(app gowid.IApp, widget gowid.IWidget) {
	login_dialog.Close(app)
	main_widget.SetSubWidget(HomeWidget(), app)
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

	main_widget = LoginWidget()
	app, err := gowid.NewApp(gowid.AppArgs{
		View:    main_widget,
		Palette: &styles,
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	ShowLoginDialog(main_widget, app)

	app.SimpleMainLoop()
}
