package gioform

import (
	"errors"
	"fmt"
	"image"

	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// A Duration field is "◀ 5 min ▶ ?": the arrows (or the arrow keys with
// the value focused) step through common times, a click on the value (or
// Enter or Space) lets the user type one, and "?" shows how to type it.

// updateDuration handles the arrows, the value button and typing.
func (f *Form) updateDuration(gtx layout.Context, fs *fieldState) {
	stops := durationStops(fs.Min, fs.Max, fs.Steps)
	step := func(dir int) {
		fs.dur, fs.err = stepDuration(stops, fs.dur, dir), ""
	}
	if fs.less.Clicked(gtx) {
		step(-1)
	}
	if fs.more.Clicked(gtx) {
		step(1)
	}
	for {
		ev, ok := gtx.Event(
			key.Filter{Focus: &fs.durBox, Name: key.NameLeftArrow},
			key.Filter{Focus: &fs.durBox, Name: key.NameRightArrow},
			key.Filter{Focus: &fs.durBox, Name: key.NameUpArrow},
			key.Filter{Focus: &fs.durBox, Name: key.NameDownArrow},
		)
		if !ok {
			break
		}
		if k, isKey := ev.(key.Event); isKey && k.State == key.Press {
			if k.Name == key.NameLeftArrow || k.Name == key.NameDownArrow {
				step(-1)
			} else {
				step(1)
			}
		}
	}
	if fs.durBox.Clicked(gtx) && !fs.editing {
		fs.editing, fs.editFocused, fs.err = true, false, ""
		fs.ed.SetText("")
		gtx.Execute(key.FocusCmd{Tag: &fs.ed})
	}
	if !fs.editing {
		return
	}
	for {
		ev, ok := gtx.Event(key.Filter{Focus: &fs.ed, Name: key.NameEscape})
		if !ok {
			break
		}
		if k, isKey := ev.(key.Event); isKey && k.State == key.Press {
			fs.editing = false
			fs.ed.SetText("")
			gtx.Execute(key.FocusCmd{Tag: &fs.durBox})
		}
	}
	for fs.editing {
		ev, ok := fs.ed.Update(gtx)
		if !ok {
			break
		}
		switch ev.(type) {
		case widget.ChangeEvent:
			fs.err = ""
		case widget.SubmitEvent:
			f.commitDuration(fs)
			gtx.Execute(key.FocusCmd{Tag: &fs.durBox})
		}
	}
	if !fs.editing {
		return
	}
	if gtx.Focused(&fs.ed) {
		fs.editFocused = true
	} else if fs.editFocused {
		f.commitDuration(fs) // the focus moved away: keep what was typed
	}
}

// commitDuration ends typing and takes the typed time if it is valid.
func (f *Form) commitDuration(fs *fieldState) {
	fs.editing = false
	t := fs.ed.Text()
	fs.ed.SetText("")
	if trimSecret(t) == "" {
		fs.err = "" // nothing typed: keep the value
		return
	}
	d, err := ParseDurationIn(t, fs.Min, fs.Max)
	switch {
	case err == nil:
		fs.dur, fs.err = d, ""
	case errors.Is(err, ErrDurationTooLong):
		fs.err = fmt.Sprintf(f.Labels.DurationTooLong, FormatDuration(fs.Max, f.Labels))
	case errors.Is(err, ErrDurationTooShort):
		fs.err = fmt.Sprintf(f.Labels.DurationTooShort, FormatDuration(fs.Min, f.Labels))
	default:
		fs.err = f.Labels.DurationFormat
	}
}

// layoutDuration draws "◀ value ▶ ?" and, while "?" is hovered or
// focused, the typing hint floating below without moving the layout.
func (f *Form) layoutDuration(gtx layout.Context, fs *fieldState) layout.Dimensions {
	th := f.Theme
	stops := durationStops(fs.Min, fs.Max, fs.Steps)
	arrow := func(c *widget.Clickable, s string, dir int) layout.FlexChild {
		col := th.Palette.Accent
		if stepDuration(stops, fs.dur, dir) == fs.dur {
			col = th.Palette.Muted // nothing further this way
		}
		return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return f.textButton(gtx, c, s, col)
		})
	}
	width := gtx.Sp(th.Material.TextSize * 9)
	value := func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Min.X, gtx.Constraints.Max.X = width, width
		shown := FormatDuration(fs.dur, f.Labels)
		inset := layout.Inset{Top: 6, Bottom: 6, Left: 8, Right: 8}
		if fs.editing {
			return f.fieldBorder(fs, true).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Background{}.Layout(gtx, f.surface, func(gtx layout.Context) layout.Dimensions {
					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return f.editor(gtx, &fs.ed, shown)
					})
				})
			})
		}
		return fs.durBox.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return f.fieldBorder(fs, gtx.Focused(&fs.durBox)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Background{}.Layout(gtx, f.surface, func(gtx layout.Context) layout.Dimensions {
					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Min.X = gtx.Constraints.Max.X
						l := material.Body1(th.Material, shown)
						l.Color, l.Alignment = th.Palette.Fg, text.Middle
						return l.Layout(gtx)
					})
				})
			})
		})
	}
	tip := func(gtx layout.Context) layout.Dimensions {
		return fs.tip.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			b := widget.Border{Color: th.Palette.Muted, CornerRadius: unit.Dp(10), Width: unit.Dp(1)}
			return b.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: 1, Bottom: 1, Left: 7, Right: 7}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					l := material.Body2(th.Material, "?")
					l.Color = th.Palette.Muted
					return l.Layout(gtx)
				})
			})
		})
	}
	dims := layout.Flex{Alignment: layout.Middle}.Layout(gtx,
		arrow(&fs.less, "◀", -1),
		layout.Rigid(value),
		arrow(&fs.more, "▶", 1),
		layout.Rigid(layout.Spacer{Width: unit.Dp(6)}.Layout),
		layout.Rigid(tip),
	)
	if (fs.tip.Hovered() || gtx.Focused(&fs.tip)) && f.Labels.DurationHint != "" {
		f.layoutFloat(gtx, f.caption(f.Labels.DurationHint, th.Palette.Fg), func(image.Point) image.Point {
			return image.Pt(0, dims.Size.Y+gtx.Dp(4))
		})
	}
	return dims
}
