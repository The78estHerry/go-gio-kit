package gioform

import (
	"fmt"
	"image"
	"image/color"
	"slices"
	"strings"

	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// Managing the options of Select and Combo fields: the Field.Add and
// Field.Remove buttons, confirmation, undo, and the reason an option cannot
// be removed.

// manages reports whether the field shows Add or Remove buttons.
func (fs *fieldState) manages() bool {
	return (fs.Type == Select || fs.Type == Combo) && (fs.Add != "" || fs.Remove != "")
}

// removable returns the index of the field's current option and whether it
// may be removed; if not, reason says why (empty when there is nothing to
// remove).
func (f *Form) removable(fs *fieldState) (i int, ok bool, reason string) {
	i = fs.current()
	switch {
	case i < 0:
		return i, false, ""
	case fs.Choices[i].Fixed:
		return i, false, f.Labels.Fixed
	case fs.CanRemove != nil:
		ok, reason = fs.CanRemove(fs.Choices[i])
		return i, ok, reason
	}
	return i, true, ""
}

// canAdd reports whether Add would do something: in a Combo the typed
// text must be new.
func (fs *fieldState) canAdd() bool {
	if fs.Type != Combo {
		return true
	}
	t := strings.TrimSpace(fs.ed.Text())
	return t != "" && !fs.hasChoice(t)
}

// updateManage handles the Add, Remove and Undo buttons.
func (f *Form) updateManage(gtx layout.Context, fs *fieldState) {
	if !fs.manages() {
		return
	}
	if fs.add.Clicked(gtx) && fs.canAdd() {
		switch {
		case fs.Type == Combo || fs.adding:
			f.addTyped(gtx, fs)
		case fs.AddExternal:
			fs.addReq = true
		default:
			fs.adding, fs.open, fs.editFocused, fs.err = true, false, false, ""
			fs.ed.SetText("")
			gtx.Execute(key.FocusCmd{Tag: &fs.ed})
		}
	}
	if fs.confirming != "" && fs.value() != fs.confirming {
		fs.confirming = "" // the value changed: ask again next time
	}
	if fs.remove.Clicked(gtx) {
		if i, ok, _ := f.removable(fs); ok {
			if v := fs.Choices[i].Value; fs.ConfirmRemove && fs.confirming != v {
				fs.confirming = v
			} else {
				f.removeOption(fs, i)
			}
		}
	}
	if fs.undo.Clicked(gtx) {
		f.undoRemove(fs)
	}
}

// setOptions replaces the options after an add, remove or undo.
func (fs *fieldState) setOptions(opts []Option) {
	fs.Choices = opts
	fs.items = make([]widget.Clickable, len(opts))
	fs.optsChanged, fs.open = true, false
	fs.removed, fs.confirming = nil, ""
}

// setValue makes v the field's value without opening the list.
func (fs *fieldState) setValue(v string) {
	if fs.Type == Select {
		fs.choice.Value = v
		return
	}
	fs.picked = v
	fs.ed.SetText(v)
	n := fs.ed.Len()
	fs.ed.SetCaret(n, n)
}

// addTyped adds the text typed in a Combo, or in a Select's add box, as an
// option and selects it.
func (f *Form) addTyped(gtx layout.Context, fs *fieldState) {
	t := strings.TrimSpace(fs.ed.Text())
	if fs.adding && (t == "" || fs.hasChoice(t)) {
		f.endAdding(gtx, fs) // nothing new: pick the existing one, if any
		if t != "" {
			fs.choice.Value = t
		}
		return
	}
	if fs.Validate != nil {
		if err := fs.Validate(t); err != nil {
			fs.err = err.Error()
			return
		}
	}
	fs.setOptions(append(slices.Clone(fs.Choices), Option{Value: t, Label: t}))
	if fs.adding {
		f.endAdding(gtx, fs)
	}
	fs.setValue(t)
	fs.err = ""
}

// endAdding closes a Select's add box and gives the focus back to it.
func (f *Form) endAdding(gtx layout.Context, fs *fieldState) {
	fs.adding = false
	fs.ed.SetText("")
	gtx.Execute(key.FocusCmd{Tag: &fs.box})
}

// updateAdding handles a Select's add box: Enter adds, Esc or moving the
// focus away cancels.
func (f *Form) updateAdding(gtx layout.Context, fs *fieldState) {
	for {
		ev, ok := gtx.Event(key.Filter{Focus: &fs.ed, Name: key.NameEscape})
		if !ok {
			break
		}
		if k, isKey := ev.(key.Event); isKey && k.State == key.Press {
			f.endAdding(gtx, fs)
			return
		}
	}
	for {
		ev, ok := fs.ed.Update(gtx)
		if !ok {
			break
		}
		switch ev.(type) {
		case widget.ChangeEvent:
			fs.err = ""
		case widget.SubmitEvent:
			f.addTyped(gtx, fs)
			if !fs.adding {
				return
			}
		}
	}
	if gtx.Focused(&fs.ed) {
		fs.editFocused = true
	} else if fs.editFocused && !gtx.Focused(&fs.add) {
		fs.adding, fs.err = false, ""
		fs.ed.SetText("")
	}
}

// removeOption removes option i, which is the field's value.
func (f *Form) removeOption(fs *fieldState, i int) {
	o := fs.Choices[i]
	fs.setOptions(slices.Delete(slices.Clone(fs.Choices), i, i+1))
	fs.setValue("")
	if fs.UndoRemove {
		fs.removed, fs.removedAt = &o, i
	}
}

// undoRemove puts the last removed option back and selects it again.
func (f *Form) undoRemove(fs *fieldState) {
	o, at := fs.removed, fs.removedAt
	if o == nil {
		return
	}
	fs.setOptions(slices.Insert(slices.Clone(fs.Choices), min(at, len(fs.Choices)), *o))
	fs.setValue(o.Value)
}

// AddRequested reports once that the Add button of a Select was clicked;
// the application then adds options with SetOptions.
func (f *Form) AddRequested(key string) bool {
	fs := f.field(key)
	if fs == nil || !fs.addReq {
		return false
	}
	fs.addReq = false
	return true
}

// Options returns the current options of a field.
func (f *Form) Options(key string) []Option {
	if fs := f.field(key); fs != nil {
		return slices.Clone(fs.Choices)
	}
	return nil
}

// OptionsChanged reports once that the user added, removed or restored an
// option, so the application can save Options.
func (f *Form) OptionsChanged(key string) bool {
	fs := f.field(key)
	if fs == nil || !fs.optsChanged {
		return false
	}
	fs.optsChanged = false
	return true
}

// SetOptions replaces the options of a field, for example after
// AddRequested. A Select whose value is no longer an option is cleared.
// It does not count as a change for OptionsChanged.
func (f *Form) SetOptions(key string, opts []Option) {
	fs := f.field(key)
	if fs == nil {
		return
	}
	fs.setOptions(slices.Clone(opts))
	fs.optsChanged = false
	if fs.Type != Combo && !fs.hasChoice(fs.choice.Value) {
		fs.choice.Value = ""
	}
}

// layoutAdd draws the Add button.
func (f *Form) layoutAdd(gtx layout.Context, fs *fieldState) layout.Dimensions {
	return f.smallButton(gtx, &fs.add, fs.Add, fs.canAdd(), f.Theme.Palette.Accent)
}

// layoutRemove draws the Remove button: grayed out when the option cannot
// be removed, with the reason floating below while hovered.
func (f *Form) layoutRemove(gtx layout.Context, fs *fieldState) layout.Dimensions {
	th := f.Theme
	i, ok, reason := f.removable(fs)
	label, col := fs.Remove, th.Palette.Accent
	if ok && fs.confirming != "" && fs.Choices[i].Value == fs.confirming {
		label, col = fmt.Sprintf(f.Labels.ConfirmRemove, fs.Choices[i].Label), th.Palette.Error
	}
	dims := f.smallButton(gtx, &fs.remove, label, ok && !fs.adding, col)
	if !ok && reason != "" && (fs.remove.Hovered() || gtx.Focused(&fs.remove)) {
		f.layoutFloat(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if th.Icons.Warning == nil {
						return layout.Dimensions{}
					}
					return layout.Inset{Right: unit.Dp(6)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return th.Icons.Warning(gtx, unit.Dp(18), th.Palette.Warning)
					})
				}),
				layout.Flexed(1, f.caption(reason, th.Palette.Fg)),
			)
		}, func(size image.Point) image.Point {
			return image.Pt(dims.Size.X-size.X, dims.Size.Y+gtx.Dp(4)) // right-aligned under the button
		})
	}
	return dims
}

// layoutUndo draws "Removed X  Undo".
func (f *Form) layoutUndo(gtx layout.Context, fs *fieldState) layout.Dimensions {
	th := f.Theme
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(f.caption(fmt.Sprintf(f.Labels.Removed, fs.removed.Label), th.Palette.Muted)),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.Clickable(gtx, &fs.undo, func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Left: 4, Right: 4}.Layout(gtx, f.caption(f.Labels.Undo, th.Palette.Accent))
			})
		}),
	)
}

// smallButton draws a text button tinted with col. A disabled one is gray
// and faded as a whole and shows no press effect, but still reports
// hovering for a tooltip.
func (f *Form) smallButton(gtx layout.Context, c *widget.Clickable, text string, enabled bool, col color.NRGBA) layout.Dimensions {
	th := f.Theme
	if !enabled {
		col = th.Palette.Muted
	}
	fill, edge := col, col
	fill.A, edge.A = 0x26, 0x99
	body := func(gtx layout.Context) layout.Dimensions {
		b := widget.Border{Color: edge, CornerRadius: th.Radius, Width: unit.Dp(1)}
		return b.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Background{}.Layout(gtx,
				func(gtx layout.Context) layout.Dimensions {
					r := gtx.Dp(th.Radius)
					defer clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Min}, r).Push(gtx.Ops).Pop()
					paint.Fill(gtx.Ops, fill)
					return layout.Dimensions{Size: gtx.Constraints.Min}
				},
				func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Top: 7, Bottom: 7, Left: 10, Right: 10}.Layout(gtx, f.label(text, col))
				})
		})
	}
	if enabled {
		return material.Clickable(gtx, c, body)
	}
	return c.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		defer paint.PushOpacity(gtx.Ops, 0.45).Pop()
		return body(gtx)
	})
}

// caption returns a small text widget.
func (f *Form) caption(text string, c color.NRGBA) layout.Widget {
	l := material.Body2(f.Theme.Material, text)
	l.Color = c
	return l.Layout
}

// layoutFloat draws w in a box on top of everything else, at the offset
// place returns for the box size. It does not take space in the layout.
func (f *Form) layoutFloat(gtx layout.Context, w layout.Widget, place func(size image.Point) image.Point) {
	th := f.Theme
	content := op.Record(gtx.Ops)
	fg := gtx
	fg.Constraints = layout.Constraints{Max: image.Pt(gtx.Dp(340), gtx.Dp(400))}
	dims := widget.Border{Color: th.Palette.Border, CornerRadius: th.Radius, Width: unit.Dp(1)}.Layout(fg, func(gtx layout.Context) layout.Dimensions {
		return layout.Background{}.Layout(gtx, f.surface, func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(8)).Layout(gtx, w)
		})
	})
	call := content.Stop()
	macro := op.Record(gtx.Ops)
	op.Offset(place(dims.Size)).Add(gtx.Ops)
	call.Add(gtx.Ops)
	op.Defer(gtx.Ops, macro.Stop())
}
