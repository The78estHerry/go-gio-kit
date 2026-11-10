package gioform

import (
	"image"
	"image/color"

	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

const maskRune = '•'

type fieldState struct {
	Field
	ed      widget.Editor
	confirm widget.Editor // NewSecret only
	reveal  widget.Clickable
	err     string
}

func (fs *fieldState) secret() bool { return fs.Type == Secret || fs.Type == NewSecret }

// Form lays out a list of fields with OK and Cancel buttons.
//
// Tab and Shift+Tab move the focus (Gio does this in layout order), Enter in
// any field submits, Esc cancels. Check Submitted and Canceled after each
// Layout.
type Form struct {
	Theme  *Theme
	Labels Labels

	fields        []*fieldState
	submitBtn     widget.Clickable
	cancelBtn     widget.Clickable
	focusedOnce   bool
	submitted     bool
	canceled      bool
	values        Values
	blockedCopies int
}

// New returns a form for fields.
func New(th *Theme, fields []Field) *Form {
	f := &Form{Theme: th, Labels: DefaultLabels}
	for _, fd := range fields {
		fs := &fieldState{Field: fd}
		fs.ed.SingleLine = true
		fs.ed.Submit = true
		fs.confirm.SingleLine = true
		fs.confirm.Submit = true
		if fs.secret() {
			fs.ed.Mask = maskRune
			fs.confirm.Mask = maskRune
		}
		f.fields = append(f.fields, fs)
	}
	return f
}

func (f *Form) field(key string) *fieldState {
	for _, fs := range f.fields {
		if fs.Key == key {
			return fs
		}
	}
	return nil
}

// SetText fills a text field, for example with saved settings. It does
// nothing for secret fields.
func (f *Form) SetText(key, value string) {
	if fs := f.field(key); fs != nil && !fs.secret() {
		fs.ed.SetText(value)
	}
}

// SetError shows msg under a field, for example when the service rejected
// a token after submission.
func (f *Form) SetError(key, msg string) {
	if fs := f.field(key); fs != nil {
		fs.err = msg
	}
}

// Submitted returns the values once after the user submitted a valid form.
// Wipe the returned secrets when done.
func (f *Form) Submitted() (Values, bool) {
	if !f.submitted {
		return Values{}, false
	}
	v := f.values
	f.submitted, f.values = false, Values{}
	return v, true
}

// Canceled reports once that the user pressed Esc or Cancel.
func (f *Form) Canceled() bool {
	c := f.canceled
	f.canceled = false
	return c
}

// Reset clears every field, including secrets, and the error messages. The
// first field gets the focus again on the next Layout.
func (f *Form) Reset() {
	for _, fs := range f.fields {
		fs.ed.SetText("")
		fs.confirm.SetText("")
		fs.err = ""
	}
	f.values.Wipe()
	f.submitted, f.canceled, f.values, f.focusedOnce = false, false, Values{}, false
}

// BlockedCopies counts the copy and cut attempts refused in secret fields.
func (f *Form) BlockedCopies() int { return f.blockedCopies }

// validate checks every field, sets the error messages and reports whether
// the form can be submitted.
func (f *Form) validate() bool {
	ok := true
	for _, fs := range f.fields {
		v := fs.ed.Text()
		fs.err = ""
		switch {
		case fs.Required && v == "":
			fs.err = f.Labels.Required
		case fs.Type == NewSecret && v != fs.confirm.Text():
			fs.err = f.Labels.Mismatch
		case fs.Validate != nil && v != "":
			if err := fs.Validate(v); err != nil {
				fs.err = err.Error()
			}
		}
		if fs.err != "" {
			ok = false
		}
	}
	return ok
}

func (f *Form) trySubmit() {
	if !f.validate() {
		return
	}
	v := Values{text: map[string]string{}, secret: map[string]*SecretValue{}}
	for _, fs := range f.fields {
		if fs.secret() {
			v.secret[fs.Key] = &SecretValue{b: []byte(fs.ed.Text())}
		} else {
			v.text[fs.Key] = fs.ed.Text()
		}
	}
	f.values, f.submitted = v, true
}

// blockCopy consumes Ctrl+C and Ctrl+X aimed at ed before the editor sees
// them: a masked Gio editor still copies its plain text otherwise.
func (f *Form) blockCopy(gtx layout.Context, ed *widget.Editor) {
	for {
		ev, ok := gtx.Event(
			key.Filter{Focus: ed, Name: "C", Required: key.ModShortcut},
			key.Filter{Focus: ed, Name: "X", Required: key.ModShortcut},
		)
		if !ok {
			return
		}
		if k, isKey := ev.(key.Event); isKey && k.State == key.Press {
			f.blockedCopies++
		}
	}
}

// update drains an editor's events.
func (f *Form) update(gtx layout.Context, fs *fieldState, ed *widget.Editor) (submit bool) {
	if fs.secret() {
		f.blockCopy(gtx, ed)
	}
	for {
		ev, ok := ed.Update(gtx)
		if !ok {
			return submit
		}
		switch ev.(type) {
		case widget.ChangeEvent:
			fs.err = ""
			if fs.Trim {
				if t := ed.Text(); trimSecret(t) != t {
					ed.SetText(trimSecret(t))
					n := ed.Len()
					ed.SetCaret(n, n)
				}
			}
		case widget.SubmitEvent:
			submit = true
		}
	}
}

// Layout draws the form and handles its input.
func (f *Form) Layout(gtx layout.Context) layout.Dimensions {
	if !f.focusedOnce && len(f.fields) > 0 {
		gtx.Execute(key.FocusCmd{Tag: &f.fields[0].ed})
		f.focusedOnce = true
	}
	for {
		if _, ok := gtx.Event(key.Filter{Name: key.NameEscape}); !ok {
			break
		}
		f.canceled = true
	}
	submit := f.submitBtn.Clicked(gtx)
	if f.cancelBtn.Clicked(gtx) {
		f.canceled = true
	}
	for _, fs := range f.fields {
		if f.update(gtx, fs, &fs.ed) {
			submit = true
		}
		if fs.Type == NewSecret && f.update(gtx, fs, &fs.confirm) {
			submit = true
		}
	}
	if submit {
		f.trySubmit()
	}

	th := f.Theme
	gap := layout.Spacer{Height: th.Spacing}
	var children []layout.FlexChild
	for _, fs := range f.fields {
		fs := fs
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return f.layoutField(gtx, fs)
		}), layout.Rigid(gap.Layout))
	}
	children = append(children, layout.Rigid(f.layoutButtons))
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

func (f *Form) layoutField(gtx layout.Context, fs *fieldState) layout.Dimensions {
	th := f.Theme
	label := func(text string, c color.NRGBA) layout.Widget {
		l := material.Body2(th.Material, text)
		l.Color = c
		return l.Layout
	}
	small := layout.Spacer{Height: unit.Dp(4)}
	rows := []layout.FlexChild{
		layout.Rigid(label(fs.Label, th.Palette.Fg)),
		layout.Rigid(small.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return f.layoutBox(gtx, fs, &fs.ed, fs.Placeholder)
		}),
	}
	if fs.Type == NewSecret {
		rows = append(rows,
			layout.Rigid(layout.Spacer{Height: th.Spacing / 2}.Layout),
			layout.Rigid(label(f.Labels.Confirm+" "+fs.Label, th.Palette.Fg)),
			layout.Rigid(small.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return f.layoutBox(gtx, fs, &fs.confirm, "")
			}),
		)
	}
	msg, c := fs.Help, th.Palette.Muted
	if fs.Type == NewSecret && fs.ed.Len() > 0 && fs.err == "" {
		msg = []string{f.Labels.Weak, f.Labels.Fair, f.Labels.Good}[Strength(fs.ed.Text())]
	}
	if fs.err != "" {
		msg, c = fs.err, th.Palette.Error
	}
	if msg != "" {
		rows = append(rows, layout.Rigid(small.Layout), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			l := material.Caption(th.Material, msg)
			l.Color = c
			return l.Layout(gtx)
		}))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, rows...)
}

// layoutBox draws one editor in a rounded box; secret fields get a button
// that shows the text while it is held down.
func (f *Form) layoutBox(gtx layout.Context, fs *fieldState, ed *widget.Editor, hint string) layout.Dimensions {
	th := f.Theme
	if fs.secret() {
		if fs.reveal.Pressed() {
			fs.ed.Mask, fs.confirm.Mask = 0, 0
		} else {
			fs.ed.Mask, fs.confirm.Mask = maskRune, maskRune
		}
	}
	borderColor, borderWidth := th.Palette.Border, unit.Dp(1)
	if gtx.Focused(ed) {
		borderColor, borderWidth = th.Palette.Accent, unit.Dp(2)
	}
	if fs.err != "" {
		borderColor = th.Palette.Error
	}
	border := widget.Border{Color: borderColor, CornerRadius: th.Radius, Width: borderWidth}
	return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Background{}.Layout(gtx,
			func(gtx layout.Context) layout.Dimensions {
				r := gtx.Dp(th.Radius)
				defer clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Min}, r).Push(gtx.Ops).Pop()
				paint.Fill(gtx.Ops, th.Palette.Surface)
				return layout.Dimensions{Size: gtx.Constraints.Min}
			},
			func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: 8, Bottom: 8, Left: 10, Right: 6}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					editor := func(gtx layout.Context) layout.Dimensions {
						es := material.Editor(th.Material, ed, hint)
						es.Color, es.HintColor = th.Palette.Fg, th.Palette.Muted
						sel := th.Palette.Accent
						sel.A = 0x60
						es.SelectionColor = sel
						return es.Layout(gtx)
					}
					if !fs.secret() || ed != &fs.ed {
						return editor(gtx)
					}
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1, editor),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							b := material.Button(th.Material, &fs.reveal, f.Labels.Show)
							b.Background, b.Color = th.Palette.Surface, th.Palette.Muted
							b.TextSize = th.Material.TextSize * 0.85
							b.Inset = layout.Inset{Top: 2, Bottom: 2, Left: 8, Right: 8}
							return b.Layout(gtx)
						}),
					)
				})
			},
		)
	})
}

func (f *Form) layoutButtons(gtx layout.Context) layout.Dimensions {
	th := f.Theme
	return layout.Flex{Spacing: layout.SpaceStart}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			b := material.Button(th.Material, &f.cancelBtn, f.Labels.Cancel)
			b.Background, b.Color = th.Palette.Surface, th.Palette.Fg
			return b.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: th.Spacing}.Layout),
		layout.Rigid(material.Button(th.Material, &f.submitBtn, f.Labels.Submit).Layout),
	)
}
