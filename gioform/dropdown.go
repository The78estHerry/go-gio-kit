package gioform

import (
	"image"
	"image/color"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// The drop-down list of Select and Combo fields. It is drawn with op.Defer
// on top of everything else; an invisible area around it catches clicks
// outside and closes it.

// visible returns the option indexes the open list shows.
func (fs *fieldState) visible() []int {
	if fs.Type == Combo && fs.filter {
		return filterOptions(fs.Choices, fs.ed.Text())
	}
	return filterOptions(fs.Choices, "")
}

// shown reports whether the list is open and has rows.
func (fs *fieldState) shown() bool { return fs.open && len(fs.visible()) > 0 }

// current returns the option index of the field's value, or -1.
func (fs *fieldState) current() int {
	if fs.Type == Combo {
		return optionIndex(fs.Choices, fs.ed.Text())
	}
	return optionIndex(fs.Choices, fs.choice.Value)
}

// openList opens the list with the current value highlighted.
func (fs *fieldState) openList(filter bool) {
	fs.open, fs.filter, fs.hl, fs.editFocused = true, filter, -1, false
	cur := fs.current()
	for i, oi := range fs.visible() {
		if oi == cur {
			fs.hl = i
		}
	}
	if fs.hl < 0 && fs.Type == Select && len(fs.Choices) > 0 {
		fs.hl = 0
	}
	if fs.hl >= 0 {
		fs.list.ScrollTo(fs.hl)
	}
}

// moveHL moves the highlight and scrolls it into view.
func (fs *fieldState) moveHL(dir int) {
	fs.hl = moveHighlight(fs.hl, len(fs.visible()), dir)
	if fs.hl < 0 {
		return
	}
	p := fs.list.Position
	switch {
	case fs.hl < p.First:
		fs.list.ScrollTo(fs.hl)
	case p.Count > 0 && fs.hl >= p.First+p.Count-1:
		fs.list.ScrollTo(fs.hl - max(p.Count-2, 0))
	}
}

// pick takes option i as the field's value and closes the list.
func (f *Form) pick(gtx layout.Context, fs *fieldState, i int) {
	o := fs.Choices[i]
	fs.err, fs.open = "", false
	if fs.Type == Select {
		fs.choice.Value = o.Value
	} else {
		fs.picked = o.Value
		fs.ed.SetText(o.Value)
		n := fs.ed.Len()
		fs.ed.SetCaret(n, n)
	}
	gtx.Execute(key.FocusCmd{Tag: fs.focusTag()})
}

// pickHighlighted picks the highlighted row, if any.
func (f *Form) pickHighlighted(gtx layout.Context, fs *fieldState) bool {
	vis := fs.visible()
	if fs.hl < 0 || fs.hl >= len(vis) {
		return false
	}
	f.pick(gtx, fs, vis[fs.hl])
	return true
}

// updateList handles clicks on the open list and outside it, and closes
// the list when the keyboard focus leaves the field.
func (f *Form) updateList(gtx layout.Context, fs *fieldState) {
	for {
		ev, ok := gtx.Event(pointer.Filter{Target: &fs.closer, Kinds: pointer.Press})
		if !ok {
			break
		}
		if _, ok := ev.(pointer.Event); ok {
			fs.open = false
		}
	}
	for i := range fs.items {
		if fs.items[i].Clicked(gtx) {
			f.pick(gtx, fs, i)
		}
	}
	if !fs.open {
		return
	}
	// editFocused records that the field had the focus while open, so the
	// frame that opens the list does not close it before the focus moves.
	if gtx.Focused(fs.focusTag()) || gtx.Focused(&fs.box) {
		fs.editFocused = true
	} else if fs.editFocused {
		fs.open = false
	}
}

// updateSelect handles a Select field: a click opens or closes the list;
// with the focus on the field, Up and Down change the choice (or the
// highlight while open), Enter or Space opens the list or picks, Esc
// closes it.
func (f *Form) updateSelect(gtx layout.Context, fs *fieldState) {
	if fs.adding {
		f.updateAdding(gtx, fs)
		return
	}
	f.updateList(gtx, fs)
	filters := []event.Filter{
		key.Filter{Focus: &fs.box, Name: key.NameUpArrow},
		key.Filter{Focus: &fs.box, Name: key.NameDownArrow},
		key.Filter{Focus: &fs.box, Name: key.NameReturn},
		key.Filter{Focus: &fs.box, Name: key.NameSpace},
	}
	if fs.open {
		filters = append(filters, key.Filter{Focus: &fs.box, Name: key.NameEscape})
	}
	for {
		ev, ok := gtx.Event(filters...)
		if !ok {
			break
		}
		k, isKey := ev.(key.Event)
		if !isKey || k.State != key.Press {
			continue
		}
		switch k.Name {
		case key.NameUpArrow, key.NameDownArrow:
			dir := 1
			if k.Name == key.NameUpArrow {
				dir = -1
			}
			if fs.open {
				fs.moveHL(dir)
			} else if i := moveHighlight(fs.current(), len(fs.Choices), dir); i >= 0 {
				fs.choice.Value, fs.err = fs.Choices[i].Value, ""
			}
		case key.NameReturn, key.NameSpace:
			if !fs.open || !f.pickHighlighted(gtx, fs) {
				fs.openList(false)
			}
		case key.NameEscape:
			if fs.open {
				fs.open = false
			} else {
				f.canceled = true
			}
		}
	}
	if fs.box.Clicked(gtx) {
		if fs.open {
			fs.open = false
		} else {
			fs.openList(false)
			gtx.Execute(key.FocusCmd{Tag: &fs.box})
		}
	}
}

// updateCombo handles the ▾ button and the list of a Combo field; its
// keys are read by comboKeys before the editor sees them.
func (f *Form) updateCombo(gtx layout.Context, fs *fieldState) {
	f.updateList(gtx, fs)
	if fs.box.Clicked(gtx) {
		if fs.open {
			fs.open = false
		} else {
			fs.openList(false)
		}
		gtx.Execute(key.FocusCmd{Tag: &fs.ed})
	}
}

// comboKeys reads the keys of a Combo editor that drive its list: Down
// opens it or moves down, Up moves up, Enter picks the highlighted row
// (or submits the form when none is), Esc closes it.
func (f *Form) comboKeys(gtx layout.Context, fs *fieldState) (submit bool) {
	filters := []event.Filter{
		key.Filter{Focus: &fs.ed, Name: key.NameUpArrow},
		key.Filter{Focus: &fs.ed, Name: key.NameDownArrow},
	}
	if fs.shown() {
		filters = append(filters,
			key.Filter{Focus: &fs.ed, Name: key.NameReturn},
			key.Filter{Focus: &fs.ed, Name: key.NameEscape},
		)
	}
	for {
		ev, ok := gtx.Event(filters...)
		if !ok {
			return submit
		}
		k, isKey := ev.(key.Event)
		if !isKey || k.State != key.Press {
			continue
		}
		switch k.Name {
		case key.NameDownArrow:
			if fs.shown() {
				fs.moveHL(1)
			} else {
				fs.openList(false)
			}
		case key.NameUpArrow:
			if fs.shown() {
				fs.moveHL(-1)
			}
		case key.NameReturn:
			if !fs.shown() || !f.pickHighlighted(gtx, fs) {
				fs.open, submit = false, true
			}
		case key.NameEscape:
			if fs.shown() {
				fs.open = false
			} else {
				f.canceled = true
			}
		}
	}
}

// comboChanged opens the list filtered by what the user typed.
func (fs *fieldState) comboChanged() {
	if fs.ed.Text() == fs.picked {
		return
	}
	fs.picked = ""
	fs.open, fs.filter, fs.hl, fs.editFocused = true, true, -1, false
}

// layoutSelect draws a Select field and, when open, its list.
func (f *Form) layoutSelect(gtx layout.Context, fs *fieldState) layout.Dimensions {
	th := f.Theme
	if fs.adding {
		return f.layoutBox(gtx, fs, &fs.ed, f.Labels.AddHint)
	}
	dims := fs.box.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return f.fieldBorder(fs, gtx.Focused(&fs.box) || fs.open).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Background{}.Layout(gtx, f.surface, func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: 8, Bottom: 8, Left: 10, Right: 10}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					text, c := f.Labels.SelectNone, th.Palette.Muted
					if i := fs.current(); i >= 0 {
						text, c = fs.Choices[i].Label, th.Palette.Fg
					}
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1, f.label(text, c)),
						layout.Rigid(f.label("▾", th.Palette.Muted)),
					)
				})
			})
		})
	})
	if fs.open {
		f.layoutList(gtx, fs, dims.Size)
	}
	return dims
}

// layoutCombo draws a Combo field (an editor with a ▾ button) and, when
// open, its list.
func (f *Form) layoutCombo(gtx layout.Context, fs *fieldState) layout.Dimensions {
	dims := f.fieldBorder(fs, gtx.Focused(&fs.ed) || fs.shown()).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Background{}.Layout(gtx, f.surface, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: 8, Bottom: 8, Left: 10, Right: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return f.editor(gtx, &fs.ed, fs.Placeholder)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return f.textButton(gtx, &fs.box, "▾", f.Theme.Palette.Muted)
					}),
				)
			})
		})
	})
	if fs.shown() {
		f.layoutList(gtx, fs, dims.Size)
	}
	return dims
}

// layoutList draws the open list under a field of the given size, on top
// of everything else.
func (f *Form) layoutList(gtx layout.Context, fs *fieldState, field image.Point) {
	vis := fs.visible()
	if len(vis) == 0 {
		return
	}
	th := f.Theme
	macro := op.Record(gtx.Ops)
	func() {
		const far = 1 << 20
		defer clip.Rect(image.Rect(-far, -far, far, far)).Push(gtx.Ops).Pop()
		event.Op(gtx.Ops, &fs.closer)
	}()
	op.Offset(image.Pt(0, field.Y+gtx.Dp(2))).Add(gtx.Ops)
	lg := gtx
	lg.Constraints = layout.Constraints{Min: image.Pt(field.X, 0), Max: image.Pt(field.X, gtx.Dp(240))}
	widget.Border{Color: th.Palette.Border, CornerRadius: th.Radius, Width: unit.Dp(1)}.Layout(lg, func(gtx layout.Context) layout.Dimensions {
		return layout.Background{}.Layout(gtx, f.surface, func(gtx layout.Context) layout.Dimensions {
			return material.List(th.Material, &fs.list).Layout(gtx, len(vis), func(gtx layout.Context, i int) layout.Dimensions {
				oi := vis[i]
				item := &fs.items[oi]
				return item.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					return layout.Background{}.Layout(gtx,
						func(gtx layout.Context) layout.Dimensions {
							if i == fs.hl || item.Hovered() {
								defer clip.Rect{Max: gtx.Constraints.Min}.Push(gtx.Ops).Pop()
								hl := th.Palette.Accent
								hl.A = 0x38
								paint.Fill(gtx.Ops, hl)
							}
							return layout.Dimensions{Size: gtx.Constraints.Min}
						},
						func(gtx layout.Context) layout.Dimensions {
							c := th.Palette.Fg
							if oi == fs.current() {
								c = th.Palette.Accent
							}
							return layout.Inset{Top: 6, Bottom: 6, Left: 10, Right: 10}.Layout(gtx, f.label(fs.Choices[oi].Label, c))
						},
					)
				})
			})
		})
	})
	op.Defer(gtx.Ops, macro.Stop())
}

// label returns a body text widget.
func (f *Form) label(text string, c color.NRGBA) layout.Widget {
	l := material.Body1(f.Theme.Material, text)
	l.Color = c
	return l.Layout
}

// textButton draws a small flat button showing text, such as ▾ or ◀.
func (f *Form) textButton(gtx layout.Context, c *widget.Clickable, text string, col color.NRGBA) layout.Dimensions {
	return material.Clickable(gtx, c, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Top: 2, Bottom: 2, Left: 6, Right: 6}.Layout(gtx, f.label(text, col))
	})
}
