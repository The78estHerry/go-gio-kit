package gioform

import (
	"image"
	"image/color"
	"slices"
	"strings"
	"time"

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
	ed      widget.Editor // text fields, Combo, and a Duration being typed
	confirm widget.Editor // NewSecret only
	reveal  widget.Clickable
	choice  widget.Enum // Choice and Select: the chosen value
	err     string

	// Managing options (manage.go).
	add, remove, undo widget.Clickable
	addReq            bool
	adding            bool // Select: typing a new option in ed
	optsChanged       bool
	confirming        string  // the value whose removal waits for a second click
	removed           *Option // the last removed option, for Undo
	removedAt         int
	removedValue      bool // the removed option was the field's value

	// Select and Combo.
	box    widget.Clickable // Select: the closed field; Combo: the ▾ button
	open   bool
	filter bool // Combo: the list shows only options matching the text
	hl     int  // highlighted row of the open list, -1 for none
	items  []widget.Clickable
	list   widget.List
	closer bool   // event tag of the area outside an open list (a bool so its address is unique)
	picked string // Combo: the text just picked, so its change event does not reopen the list

	// Duration.
	dur                  time.Duration
	less, more, durBox   widget.Clickable
	tip                  widget.Clickable
	editing, editFocused bool
}

func (fs *fieldState) secret() bool { return fs.Type == Secret || fs.Type == NewSecret }

// value is the field's current text, the chosen value of a Choice or
// Select, or a Duration in Go's form.
func (fs *fieldState) value() string {
	switch fs.Type {
	case Choice, Select:
		return fs.choice.Value
	case Duration:
		return fs.dur.String()
	}
	return fs.ed.Text()
}

// hasChoice reports whether v is one of the field's choices.
func (fs *fieldState) hasChoice(v string) bool { return optionIndex(fs.Choices, v) >= 0 }

// initial sets the field to its default value.
func (fs *fieldState) initial() {
	fs.ed.SetText("")
	fs.confirm.SetText("")
	fs.choice.Value = fs.Default
	fs.err, fs.open, fs.editing, fs.picked = "", false, false, ""
	fs.confirming, fs.removed, fs.adding = "", nil, false
	switch fs.Type {
	case Combo:
		fs.ed.SetText(fs.Default)
		fs.picked = fs.Default // not typed: do not open the list
	case Duration:
		d, err := ParseDuration(fs.Default)
		if err != nil {
			d = fs.Min
		}
		fs.dur = clampDuration(d, fs.Min, fs.Max)
	}
}

// focusTag is what gets the keyboard focus when the field is focused.
func (fs *fieldState) focusTag() any {
	switch fs.Type {
	case Select:
		return &fs.box
	case Duration:
		return &fs.durBox
	}
	return &fs.ed
}

// Form lays out a list of fields with OK and Cancel buttons.
//
// Tab and Shift+Tab move the focus (Gio does this in layout order), Enter in
// any text field submits, Esc cancels (or closes an open list first). Check
// Submitted and Canceled after each Layout.
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
		if fs.Type == Duration {
			if fs.Max == 0 {
				fs.Max = DefaultDurationMax
			}
			fs.Max = max(fs.Max, fs.Min)
		}
		fs.Choices = slices.Clone(fd.Choices) // Add and Remove change it
		fs.items = make([]widget.Clickable, len(fd.Choices))
		fs.list.Axis = layout.Vertical
		fs.initial()
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

// SetText fills a text field, for example with saved settings, selects a
// Choice or Select by value, or sets a Duration ("90s", see ParseDuration;
// values out of range are clamped). It does nothing for secret fields.
func (f *Form) SetText(key, value string) {
	fs := f.field(key)
	switch {
	case fs == nil || fs.secret():
	case fs.Type == Choice || fs.Type == Select:
		if fs.hasChoice(value) {
			fs.choice.Value = value
		}
	case fs.Type == Duration:
		if d, err := ParseDuration(value); err == nil {
			fs.dur = clampDuration(d, fs.Min, fs.Max)
		}
	default:
		fs.ed.SetText(value)
	}
}

// Choice returns the value currently chosen in a Choice or Select field,
// so the application can adapt the rest of the form while the user picks.
func (f *Form) Choice(key string) string {
	if fs := f.field(key); fs != nil && (fs.Type == Choice || fs.Type == Select) {
		return fs.choice.Value
	}
	return ""
}

// Duration returns the current value of a Duration field.
func (f *Form) Duration(key string) time.Duration {
	if fs := f.field(key); fs != nil && fs.Type == Duration {
		return fs.dur
	}
	return 0
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

// Reset sets every field back to its default, clears secrets and error
// messages. The first field gets the focus again on the next Layout.
func (f *Form) Reset() {
	for _, fs := range f.fields {
		fs.initial()
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
		fs.err = ""
		if fs.Type == Duration {
			continue // always in range
		}
		v := fs.value()
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
	v := Values{text: map[string]string{}, secret: map[string]*SecretValue{}, dur: map[string]time.Duration{}}
	for _, fs := range f.fields {
		switch {
		case fs.secret():
			v.secret[fs.Key] = &SecretValue{b: []byte(fs.ed.Text())}
		case fs.Type == Duration:
			v.dur[fs.Key] = fs.dur
			v.text[fs.Key] = fs.value()
		default:
			v.text[fs.Key] = fs.value()
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
	if fs.Type == Combo && f.comboKeys(gtx, fs) {
		submit = true
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
			if fs.Type == Combo {
				fs.comboChanged()
			}
		case widget.SubmitEvent:
			submit = true
			fs.open = false
		}
	}
}

// Layout draws the form and handles its input.
func (f *Form) Layout(gtx layout.Context) layout.Dimensions {
	if !f.focusedOnce && len(f.fields) > 0 {
		gtx.Execute(key.FocusCmd{Tag: f.fields[0].focusTag()})
		f.focusedOnce = true
	}
	submit := f.submitBtn.Clicked(gtx)
	if f.cancelBtn.Clicked(gtx) {
		f.canceled = true
	}
	for _, fs := range f.fields {
		f.updateManage(gtx, fs)
		switch fs.Type {
		case Choice:
			if fs.choice.Update(gtx) {
				fs.err = ""
			}
			continue
		case Select:
			f.updateSelect(gtx, fs)
			continue
		case Duration:
			f.updateDuration(gtx, fs)
			continue
		case Combo:
			f.updateCombo(gtx, fs)
		}
		if f.update(gtx, fs, &fs.ed) {
			submit = true
		}
		if fs.Type == NewSecret && f.update(gtx, fs, &fs.confirm) {
			submit = true
		}
	}
	// Key events go to the first gtx.Event call whose filter matches, so the
	// form-wide Esc is read last: an open list or a time being typed takes
	// its Esc first.
	// Only the press counts: the release of an Esc that just closed a list
	// arrives later and must not cancel the form.
	for {
		ev, ok := gtx.Event(key.Filter{Name: key.NameEscape})
		if !ok {
			break
		}
		if k, isKey := ev.(key.Event); isKey && k.State == key.Press {
			f.canceled = true
		}
	}
	if submit {
		f.trySubmit()
	}

	th := f.Theme
	gap := layout.Spacer{Height: th.Spacing}
	var children []layout.FlexChild
	for _, fs := range f.fields {
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
	input := func(gtx layout.Context) layout.Dimensions {
		return f.layoutBox(gtx, fs, &fs.ed, fs.Placeholder)
	}
	switch fs.Type {
	case Choice:
		input = func(gtx layout.Context) layout.Dimensions { return f.layoutChoices(gtx, fs) }
	case Select:
		input = func(gtx layout.Context) layout.Dimensions { return f.layoutSelect(gtx, fs) }
	case Combo:
		input = func(gtx layout.Context) layout.Dimensions { return f.layoutCombo(gtx, fs) }
	case Duration:
		input = func(gtx layout.Context) layout.Dimensions { return f.layoutDuration(gtx, fs) }
	}
	if fs.manages() {
		field := input
		input = func(gtx layout.Context) layout.Dimensions {
			children := []layout.FlexChild{layout.Flexed(1, field)}
			if fs.Add != "" {
				children = append(children,
					layout.Rigid(layout.Spacer{Width: th.Spacing / 2}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions { return f.layoutAdd(gtx, fs) }))
			}
			if fs.Remove != "" {
				children = append(children,
					layout.Rigid(layout.Spacer{Width: th.Spacing / 2}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions { return f.layoutRemove(gtx, fs) }))
			}
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx, children...)
		}
	}
	rows := []layout.FlexChild{
		layout.Rigid(label(fs.Label, th.Palette.Fg)),
		layout.Rigid(small.Layout),
		layout.Rigid(input),
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
	caption := func(msg string, c color.NRGBA) {
		rows = append(rows, layout.Rigid(small.Layout), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			l := material.Caption(th.Material, msg)
			l.Color = c
			return l.Layout(gtx)
		}))
	}
	if msg != "" {
		caption(msg, c)
	}
	if fs.removed != nil {
		rows = append(rows, layout.Rigid(small.Layout), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return f.layoutUndo(gtx, fs)
		}))
	}
	if fs.Type == NewSecret && fs.err == "" {
		switch confirmState(fs.ed.Text(), fs.confirm.Text()) {
		case confirmMismatch:
			caption(f.Labels.Mismatch, th.Palette.Error)
		case confirmMatch:
			caption(f.Labels.Match, th.Palette.Muted)
		}
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, rows...)
}

// fieldBorder returns the border of a field box.
func (f *Form) fieldBorder(fs *fieldState, focused bool) widget.Border {
	th := f.Theme
	c, w := th.Palette.Border, unit.Dp(1)
	if focused {
		c, w = th.Palette.Accent, unit.Dp(2)
	}
	if fs.err != "" {
		c = th.Palette.Error
	}
	return widget.Border{Color: c, CornerRadius: th.Radius, Width: w}
}

// surface fills the area of a field box.
func (f *Form) surface(gtx layout.Context) layout.Dimensions {
	r := gtx.Dp(f.Theme.Radius)
	defer clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Min}, r).Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, f.Theme.Palette.Surface)
	return layout.Dimensions{Size: gtx.Constraints.Min}
}

// editor draws ed with the theme colors.
func (f *Form) editor(gtx layout.Context, ed *widget.Editor, hint string) layout.Dimensions {
	th := f.Theme
	es := material.Editor(th.Material, ed, hint)
	es.Color, es.HintColor = th.Palette.Fg, th.Palette.Muted
	sel := th.Palette.Accent
	sel.A = 0x60
	es.SelectionColor = sel
	return es.Layout(gtx)
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
	return f.fieldBorder(fs, gtx.Focused(ed)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Background{}.Layout(gtx, f.surface, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: 8, Bottom: 8, Left: 10, Right: 6}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				editor := func(gtx layout.Context) layout.Dimensions { return f.editor(gtx, ed, hint) }
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
		})
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

const (
	confirmPending = iota
	confirmMatch
	confirmMismatch
)

// confirmState compares a new secret with its confirmation while the user
// types. A confirmation that is still a prefix of the secret is pending, so
// the mismatch warning does not flash on every keystroke.
func confirmState(secret, confirm string) int {
	switch {
	case confirm == "":
		return confirmPending
	case confirm == secret:
		return confirmMatch
	case strings.HasPrefix(secret, confirm):
		return confirmPending
	}
	return confirmMismatch
}

// layoutChoices draws a Choice field as a row of radio buttons.
func (f *Form) layoutChoices(gtx layout.Context, fs *fieldState) layout.Dimensions {
	th := f.Theme
	var items []layout.FlexChild
	for _, c := range fs.Choices {
		items = append(items, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			r := material.RadioButton(th.Material, &fs.choice, c.Value, c.Label)
			r.Color, r.IconColor = th.Palette.Fg, th.Palette.Accent
			return layout.Inset{Right: unit.Dp(12)}.Layout(gtx, r.Layout)
		}))
	}
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx, items...)
}
