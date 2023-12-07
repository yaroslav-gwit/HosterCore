// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package edit provides an editable text field widget with support for password hiding.
package edit

import (
	"fmt"
	"io"
	"strconv"
	"unicode"
	"unicode/utf8"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwutil"
	"github.com/gcla/gowid/widgets/text"
	tcell "github.com/gdamore/tcell/v2"
	"github.com/pkg/errors"
)

//======================================================================

// IEdit is an interface to be implemented by a text editing widget. A suitable implementation
// will be able to defer to RenderEdit() in its Render() function.
type IEdit interface {
	text.ISimple
	IMask
	INumeric
	text.ICursor
	Caption() string
	MakeText() text.IWidget
}

type IMask interface {
	UseMask() bool
	MaskChr() rune
}

type INumeric interface {
	UseNumeric() bool
	MaxNumberDigits() int
}

type IPaste interface {
	PasteState(...bool) bool
	AddKey(*tcell.EventKey)
	GetKeys() []*tcell.EventKey
}

type Mask struct {
	Chr    rune
	Enable bool
}

type Numeric struct {
	Enable bool
	Limit  int
}

// For callback registration
type Text struct{}
type Caption struct{}
type Cursor struct{}

func DisabledMask() Mask {
	return Mask{Chr: 'x', Enable: false}
}

func MakeMask(chr rune) Mask {
	return Mask{Chr: chr, Enable: true}
}

func (m Mask) UseMask() bool {
	return m.Enable
}

func (m Mask) MaskChr() rune {
	return m.Chr
}

func MakeNumeric(enable bool, limit int) Numeric {
	return Numeric{Enable: enable, Limit: limit}
}

func (n Numeric) UseNumeric() bool {
	return n.Enable
}

func (n Numeric) MaxNumberDigits() int {
	return n.Limit
}

type IWidget interface {
	IEdit
	text.IChrAt
	LinesFromTop() int
	SetLinesFromTop(int, gowid.IApp)
	UpLines(size gowid.IRenderSize, doPage bool, app gowid.IApp) bool
	DownLines(size gowid.IRenderSize, doPage bool, app gowid.IApp) bool
}

type IReadOnly interface {
	IsReadOnly() bool
}

type Widget struct {
	IMask
	INumeric
	caption      string
	text         string
	paste        bool
	readonly     bool
	pastedKeys   []*tcell.EventKey
	cursorPos    int
	linesFromTop int
	Callbacks    *gowid.Callbacks
	gowid.IsSelectable
}

var _ fmt.Stringer = (*Widget)(nil)
var _ io.Reader = (*Widget)(nil)
var _ gowid.IWidget = (*Widget)(nil)
var _ IPaste = (*Widget)(nil)
var _ IReadOnly = (*Widget)(nil)

// Writer embeds an EditWidget and provides the io.Writer interface. An gowid.IApp needs to
// be provided too because the widget's SetText() function requires it in order to issue
// callbacks when the text changes.
type Writer struct {
	*Widget
	gowid.IApp
}

type Options struct {
	Caption  string
	Text     string
	Mask     IMask
	Numeric  INumeric
	ReadOnly bool
}

func New(args ...Options) *Widget {
	var opt Options
	if len(args) > 0 {
		opt = args[0]
	}
	if opt.Mask == nil {
		opt.Mask = DisabledMask()
	}
	res := &Widget{
		IMask:        opt.Mask,
		INumeric:     opt.Numeric,
		caption:      opt.Caption,
		text:         opt.Text,
		readonly:     opt.ReadOnly,
		cursorPos:    len(opt.Text),
		pastedKeys:   make([]*tcell.EventKey, 0, 100),
		linesFromTop: 0,
		Callbacks:    gowid.NewCallbacks(),
	}
	return res
}

func (w *Widget) String() string {
	return fmt.Sprintf("edit")
}

func (w *Widget) IsReadOnly() bool {
	return w.readonly
}

func (w *Widget) SetReadOnly(v bool, _ gowid.IApp) {
	w.readonly = v
}

// Set content from array
func (w *Writer) Write(p []byte) (n int, err error) {
	w.SetText(string(p), w.IApp)
	w.cursorPos = 0
	w.linesFromTop = 0
	return len(p), nil
}

// Set array from widget content
func (w *Widget) Read(p []byte) (n int, err error) {
	pl := len(p)
	num := copy(p, w.text[:])
	if num < pl {
		return num, io.EOF
	} else {
		return num, nil
	}
}

func (w *Widget) Text() string {
	return w.text
}

var InvalidRuneIndex error = errors.New("Invalid rune index for string")

// TODO - this isn't ideal- if called in a loop, it would be quadratic.
func (w *Widget) ChrAt(i int) rune {
	j := 0
	for _, char := range w.caption {
		if j == i {
			return char
		}
		j++
	}
	for _, char := range w.text {
		if j == i {
			return char
		}
		j++
	}
	panic(errors.WithStack(gowid.WithKVs(InvalidRuneIndex, map[string]interface{}{"index": i, "text": w.text})))
}

func (w *Widget) SetText(text string, app gowid.IApp) {
	w.text = text
	wid := utf8.RuneCountInString(w.text)
	if w.cursorPos > wid {
		w.SetCursorPos(wid, app)
	}
	gowid.RunWidgetCallbacks(w.Callbacks, Text{}, app, w)
}

func (w *Widget) LinesFromTop() int {
	return w.linesFromTop
}

func (w *Widget) SetLinesFromTop(l int, app gowid.IApp) {
	w.linesFromTop = l
}

func (w *Widget) Caption() string {
	return w.caption
}

func (w *Widget) SetCaption(text string, app gowid.IApp) {
	w.caption = text
	gowid.RunWidgetCallbacks(w.Callbacks, Caption{}, app, w)
}

// func (w *Widget) PasteState(b ...bool) []*tcell.EventKey {
func (w *Widget) PasteState(b ...bool) bool {
	if len(b) > 0 {
		w.paste = b[0]
		//if w.paste {
		w.pastedKeys = w.pastedKeys[:0]
		//}
	}
	return w.paste
}

func (w *Widget) AddKey(ev *tcell.EventKey) {
	w.pastedKeys = append(w.pastedKeys, ev)
}

func (w *Widget) GetKeys() []*tcell.EventKey {
	return w.pastedKeys
}

func (w *Widget) CursorEnabled() bool {
	return w.cursorPos != -1
}

func (w *Widget) SetCursorDisabled() {
	w.cursorPos = -1
}

// TODO - weird that you could call set to 0, then get and it would be > 0...
func (w *Widget) CursorPos() int {
	if !w.CursorEnabled() {
		panic(errors.New("Cursor is disabled, cannot return!"))
	}
	return w.cursorPos
}

func (w *Widget) SetCursorPos(pos int, app gowid.IApp) {
	pos = gwutil.Min(pos, utf8.RuneCountInString(w.Text()))
	w.cursorPos = pos
	gowid.RunWidgetCallbacks(w.Callbacks, Cursor{}, app, w)
}

func (w *Widget) OnTextSet(cb gowid.IWidgetChangedCallback) {
	gowid.AddWidgetCallback(w.Callbacks, Text{}, cb)
}

func (w *Widget) RemoveOnTextSet(cb gowid.IIdentity) {
	gowid.RemoveWidgetCallback(w.Callbacks, Text{}, cb)
}

func (w *Widget) OnCaptionSet(cb gowid.IWidgetChangedCallback) {
	gowid.AddWidgetCallback(w.Callbacks, Caption{}, cb)
}

func (w *Widget) RemoveOnCaptionSet(cb gowid.IIdentity) {
	gowid.RemoveWidgetCallback(w.Callbacks, Caption{}, cb)
}

func (w *Widget) OnCursorPosSet(cb gowid.IWidgetChangedCallback) {
	gowid.AddWidgetCallback(w.Callbacks, Cursor{}, cb)
}

func (w *Widget) RemoveOnCursorPosSet(cb gowid.IIdentity) {
	gowid.RemoveWidgetCallback(w.Callbacks, Cursor{}, cb)
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return gowid.CalculateRenderSizeFallback(w, size, focus, app)
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return Render(w, size, focus, app)
}

func (w *Widget) MakeText() text.IWidget {
	return MakeText(w)
}

func (w *Widget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	return UserInput(w, ev, size, focus, app)
}

func (w *Widget) DownLines(size gowid.IRenderSize, doPage bool, app gowid.IApp) bool {
	return DownLines(w, size, doPage, app)
}

func (w *Widget) UpLines(size gowid.IRenderSize, doPage bool, app gowid.IApp) bool {
	return UpLines(w, size, doPage, app)
}

func (w *Widget) CalculateTopMiddleBottom(size gowid.IRenderSize) (int, int, int) {
	return CalculateTopMiddleBottom(w, size)
}

////////////////////////////////////////////////////////////////////////

func Render(w IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	twc := w.MakeText()
	c := twc.Render(size, focus, app)
	return c
}

func MakeText(w IWidget) text.IWidget {
	var txt string
	if w.UseMask() {
		arr := make([]rune, len(w.Text()))
		for i := 0; i < len(arr); i++ {
			arr[i] = w.MaskChr()
		}
		txt = string(arr)
	} else {
		txt = w.Text()
	}

	//txt = w.Caption() + "\u00A0" + txt
	txt = w.Caption() + txt

	tw := text.New(txt)
	tw.SetLinesFromTop(w.LinesFromTop(), nil)

	cu := &text.SimpleCursor{Pos: -1}
	cu.SetCursorPos(w.CursorPos()+utf8.RuneCountInString(w.Caption()), nil)
	twc := &text.WidgetWithCursor{Widget: tw, SimpleCursor: cu}

	return twc
}

func CalculateTopMiddleBottom(w IWidget, size gowid.IRenderSize) (int, int, int) {
	twc := w.MakeText()
	return text.CalculateTopMiddleBottom(twc, size)
}

// Return true if done
func DownLines(w IWidget, size gowid.IRenderSize, doPage bool, app gowid.IApp) bool {
	prev := w.CursorPos()

	twc := w.MakeText()
	caplen := utf8.RuneCountInString(w.Caption())
	// This incorporates the caption too
	cols, ok := size.(gowid.IColumns)
	if !ok {
		panic(gowid.WidgetSizeError{Widget: w, Size: size, Required: "gowid.IColumns"})
	}
	layout := text.MakeTextLayout(twc.Content(), cols.Columns(), text.WrapAny, gowid.HAlignLeft{})
	ccol, crow := text.GetCoordsFromCursorPos(w.CursorPos()+caplen, cols.Columns(), layout, w)
	offset := 1
	if rows, ok := size.(gowid.IRows); ok && doPage {
		if crow < w.LinesFromTop()+rows.Rows()-1 {
			// if the cursor is in the middle somewhere, jump to the bottom
			offset = w.LinesFromTop() + rows.Rows() - (crow + 1)
		} else {
			// otherwise jump a "page" worth
			offset = rows.Rows()
		}
	}

	targetRow := crow + offset
	newCursorPos := text.GetCursorPosFromCoords(ccol, targetRow, layout, w) - caplen
	if newCursorPos < 0 {
		return false
	} else {
		w.SetCursorPos(newCursorPos, app)
		// TODO - ugly to check for render type like this
		if box, ok := size.(gowid.IRenderBox); ok && (targetRow >= box.BoxRows()+w.LinesFromTop()) { // assumes we render fixed not flow
			w.SetLinesFromTop(w.LinesFromTop()+offset, app)
		}

		//w.linesFromTop += 1
		return (prev != w.CursorPos())
	}
}

// Return true if done
func UpLines(w IWidget, size gowid.IRenderSize, doPage bool, app gowid.IApp) bool {
	caplen := utf8.RuneCountInString(w.Caption())
	prev := w.CursorPos()
	twc := w.MakeText()
	cols, isColumns := size.(gowid.IColumns)
	if !isColumns {
		panic(gowid.WidgetSizeError{Widget: w, Size: size, Required: "gowid.IColumns"})
	}
	layout := text.MakeTextLayout(twc.Content(), cols.Columns(), text.WrapAny, gowid.HAlignLeft{})
	ccol, crow := text.GetCoordsFromCursorPos(w.CursorPos()+caplen, cols.Columns(), layout, w)

	if crow <= 0 {
		return false
	} else {
		offset := 1
		if rows, ok := size.(gowid.IRows); ok && doPage {
			if crow == w.LinesFromTop() {
				offset = rows.Rows()
			} else {
				offset = crow - w.LinesFromTop()
			}
		}
		targetCol := gwutil.Max(crow-offset, 0)

		newCursorPos := text.GetCursorPosFromCoords(ccol, targetCol, layout, w) - caplen
		if newCursorPos < 0 {
			return false
		} else {
			w.SetCursorPos(newCursorPos, app)

			if targetCol < w.LinesFromTop() {
				w.SetLinesFromTop(targetCol, app)
			}

			return (prev != w.CursorPos())
		}
	}
}

func keyIsPasteable(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyEnter, tcell.Key(' '), tcell.KeyRune:
		return true
	default:
		return false
	}
}

func isReadOnly(w interface{}) bool {
	readOnly := false
	if ro, ok := w.(IReadOnly); ok {
		readOnly = ro.IsReadOnly()
	}
	return readOnly
}

func pasteableKeyInput(w IWidget, ev *tcell.EventKey, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	if isReadOnly(w) {
		return false
	}

	handled := true
	switch ev.Key() {
	case tcell.KeyEnter:
		if !w.UseNumeric() {
			r := []rune(w.Text())
			w.SetText(string(r[0:w.CursorPos()])+string('\n')+string(r[w.CursorPos():]), app)
			w.SetCursorPos(w.CursorPos()+1, app)
		}
	case tcell.Key(' '):
		if !w.UseNumeric() {
			r := []rune(w.Text())
			w.SetText(string(r[0:w.CursorPos()])+" "+string(r[w.CursorPos():]), app)
			w.SetCursorPos(w.CursorPos()+1, app)
		}
	case tcell.KeyRune:
		// TODO: this is lame. Inserting a character is O(n) where n is length
		// of text. I should switch this to use the two stack model for edited
		// text.
		if w.UseNumeric() {
			if isNumeric(string(ev.Rune())) {
				txt := w.Text()
				if w.MaxNumberDigits() == 0 || len(txt) < w.MaxNumberDigits() {
					r := []rune(txt)
					cpos := w.CursorPos()
					rhs := make([]rune, len(r)-cpos)
					copy(rhs, r[cpos:])
					w.SetText(string(append(append(r[:cpos], ev.Rune()), rhs...)), app)
					w.SetCursorPos(w.CursorPos()+1, app)
				}
			}
		} else {
			txt := w.Text()
			r := []rune(txt)
			cpos := w.CursorPos()
			rhs := make([]rune, len(r)-cpos)
			copy(rhs, r[cpos:])
			w.SetText(string(append(append(r[:cpos], ev.Rune()), rhs...)), app)
			w.SetCursorPos(w.CursorPos()+1, app)
		}

	default:
		handled = false
	}

	return handled
}

func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func UserInput(w IWidget, ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	handled := true
	doup := false
	dodown := false
	recalcLinesFromTop := false
	readOnly := isReadOnly(w)

	switch ev := ev.(type) {
	case *tcell.EventMouse:
		switch ev.Buttons() {
		case tcell.WheelUp:
			doup = true
		case tcell.WheelDown:
			dodown = true
		case tcell.Button1:
			twc := w.MakeText()
			cols, isColumns := size.(gowid.IColumns)
			if !isColumns {
				panic(gowid.WidgetSizeError{Widget: w, Size: size, Required: "gowid.IColumns"})
			}
			layout := text.MakeTextLayout(twc.Content(), cols.Columns(), text.WrapAny, gowid.HAlignLeft{})
			mx, my := ev.Position()
			cursorPos := text.GetCursorPosFromCoords(mx, my+w.LinesFromTop(), layout, w) - (utf8.RuneCountInString(w.Caption()))
			if cursorPos < 0 {
				handled = false
			} else {
				w.SetCursorPos(cursorPos, app)
				handled = true
			}
		default:
			handled = false
		}

	case *tcell.EventPaste:
		if wp, ok := w.(IPaste); ok {
			if ev.Start() {
				wp.PasteState(true)
			} else {
				evs := wp.GetKeys()
				wp.PasteState(false)
				for _, ev := range evs {
					pasteableKeyInput(w, ev, size, focus, app)
				}
			}
		}

	case *tcell.EventKey:
		handled = false
		if wp, ok := w.(IPaste); ok {
			if wp.PasteState() && keyIsPasteable(ev) && !readOnly {
				wp.AddKey(ev)
				handled = true
			}
		}

		if !handled {
			handled = pasteableKeyInput(w, ev, size, focus, app)
		}

		if !handled {
			handled = true
			switch ev.Key() {
			case tcell.KeyPgUp:
				handled = w.UpLines(size, true, app)
			case tcell.KeyUp, tcell.KeyCtrlP:
				doup = true
			case tcell.KeyDown, tcell.KeyCtrlN:
				dodown = true
			case tcell.KeyPgDn:
				handled = w.DownLines(size, true, app)
			case tcell.KeyLeft, tcell.KeyCtrlB:
				if w.CursorPos() > 0 {
					w.SetCursorPos(w.CursorPos()-1, app)
				} else {
					handled = false
				}
			case tcell.KeyRight, tcell.KeyCtrlF:
				if w.CursorPos() < utf8.RuneCountInString(w.Text()) {
					w.SetCursorPos(w.CursorPos()+1, app)
				} else {
					handled = false
				}
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				if !readOnly {
					if w.CursorPos() > 0 {
						pos := w.CursorPos()
						w.SetCursorPos(w.CursorPos()-1, app)
						r := []rune(w.Text())
						w.SetText(string(r[0:pos-1])+string(r[pos:]), app)
					}
				}
			case tcell.KeyDelete, tcell.KeyCtrlD:
				if !readOnly {
					if w.CursorPos() < utf8.RuneCountInString(w.Text()) {
						r := []rune(w.Text())
						w.SetText(string(r[0:w.CursorPos()])+string(r[w.CursorPos()+1:]), app)
					}
				}
			case tcell.KeyCtrlK:
				if !readOnly {
					r := []rune(w.Text())
					w.SetText(string(r[0:w.CursorPos()]), app)
				}
			case tcell.KeyCtrlU:
				if !readOnly {
					r := []rune(w.Text())
					w.SetText(string(r[w.CursorPos():]), app)
					w.SetCursorPos(0, app)
				}
			case tcell.KeyHome:
				w.SetCursorPos(0, app)
				w.SetLinesFromTop(0, app)
			case tcell.KeyCtrlW:
				if !readOnly {
					txt := []rune(w.Text())
					origcp := w.CursorPos()
					cp := origcp
					for cp > 0 && unicode.IsSpace(txt[cp-1]) {
						cp--
					}
					for cp > 0 && !unicode.IsSpace(txt[cp-1]) {
						cp--
					}
					if cp != origcp {
						w.SetText(string(txt[0:cp])+string(txt[origcp:]), app)
						w.SetCursorPos(cp, app)
					}
				}
			case tcell.KeyCtrlA:
				// Would be nice to use a slice here, something that doesn't copy
				// TODO: terrible O(n) behavior :-(
				txt := w.Text()

				i := w.CursorPos()
				j := 0
				lastnl := false
				curstart := 0

				for _, ch := range txt {
					if lastnl {
						curstart = j
					}
					lastnl = (ch == '\n')

					if i == j {
						break
					}
					j += 1
				}

				w.SetCursorPos(curstart, app)
				recalcLinesFromTop = true

			case tcell.KeyEnd:
				w.SetCursorPos(utf8.RuneCountInString(w.Text()), app)
				recalcLinesFromTop = true

			case tcell.KeyCtrlE:
				// TODO: terrible O(n) behavior :-(
				txt := w.Text()
				i := w.CursorPos()
				j := 0
				checknl := false
				for _, ch := range txt {
					if i == j {
						checknl = true
					}
					j += 1
					if checknl {
						if ch == '\n' {
							break
						}
						i += 1
					}
				}
				w.SetCursorPos(i, app)
				recalcLinesFromTop = true

			default:
				handled = false
			}
		}
	}

	if doup {
		handled = w.UpLines(size, false, app)
	}
	if dodown {
		handled = w.DownLines(size, false, app)
	}

	box, ok := size.(gowid.IRenderBox)
	if recalcLinesFromTop && ok {
		twc := w.MakeText()
		caplen := utf8.RuneCountInString(w.Caption())
		layout := text.MakeTextLayout(twc.Content(), box.BoxColumns(), text.WrapAny, gowid.HAlignLeft{})
		_, crow := text.GetCoordsFromCursorPos(w.CursorPos()+caplen, box.BoxColumns(), layout, w)
		w.SetLinesFromTop(gwutil.Max(0, crow-(box.BoxRows()-1)), app)
	}

	return handled
}
