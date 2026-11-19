package gioform

import (
	"image"

	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// disabledOpacity is how much of a disabled control shows through.
const disabledOpacity = 0.45

// SetDisabled turns the field key off, showing why, or back on when reason
// is "". Call it again to change the reason, for example to count down.
//
// A disabled field is faded as a whole and takes no focus or input; its box
// shows the reason with Theme.Icons.Warning instead of its text, which is
// kept. While any field is disabled the form cannot be submitted: the
// submit button is faded too and shows the reason while hovered.
func (f *Form) SetDisabled(key, reason string) {
	if fs := f.field(key); fs != nil {
		fs.disabled = reason
	}
}

// Focus moves the keyboard focus to the field key in the next Layout,
// for example after it is enabled again. A disabled field is skipped.
func (f *Form) Focus(key string) {
	f.focusKey = key
	f.focusedOnce = true
}

// Disabled returns why the field key is disabled, or "" if it is not.
func (f *Form) Disabled(key string) string {
	if fs := f.field(key); fs != nil {
		return fs.disabled
	}
	return ""
}

// disabledReason returns the reason of the first disabled field, or "".
func (f *Form) disabledReason() string {
	for _, fs := range f.fields {
		if fs.disabled != "" {
			return fs.disabled
		}
	}
	return ""
}

// releaseFocus moves the focus away from a field that was just disabled.
func (f *Form) releaseFocus(gtx layout.Context, fs *fieldState) {
	for _, tag := range []any{fs.focusTag(), &fs.ed, &fs.confirm} {
		if gtx.Focused(tag) {
			gtx.Execute(key.FocusCmd{Tag: nil})
			return
		}
	}
}

// layoutDisabledField draws a disabled field: its label, a faded box with
// the reason (and an empty one for the confirmation of a NewSecret), and
// its help. Errors are hidden; the reason says what matters now.
func (f *Form) layoutDisabledField(gtx layout.Context, fs *fieldState) layout.Dimensions {
	th := f.Theme
	small := layout.Spacer{Height: unit.Dp(4)}
	label := func(text string) layout.FlexChild {
		return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			defer paint.PushOpacity(gtx.Ops, disabledOpacity).Pop()
			return f.caption(text, th.Palette.Fg)(gtx)
		})
	}
	box := func(reason string) layout.FlexChild {
		return layout.Rigid(func(gtx layout.Context) layout.Dimensions { return f.disabledBox(gtx, reason) })
	}
	rows := []layout.FlexChild{label(fs.Label), layout.Rigid(small.Layout), box(fs.disabled)}
	if fs.Type == NewSecret {
		rows = append(rows,
			layout.Rigid(layout.Spacer{Height: th.Spacing / 2}.Layout),
			label(f.Labels.Confirm+" "+fs.Label), layout.Rigid(small.Layout), box(""))
	}
	if fs.Help != "" {
		rows = append(rows, layout.Rigid(small.Layout), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			l := material.Caption(th.Material, fs.Help)
			l.Color = th.Palette.Muted
			return l.Layout(gtx)
		}))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, rows...)
}

// disabledBox draws a faded field box with the reason inside. The box and
// its border are faded; the reason is not, so it stays easy to read.
func (f *Form) disabledBox(gtx layout.Context, reason string) layout.Dimensions {
	th := f.Theme
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	return layout.Background{}.Layout(gtx,
		func(gtx layout.Context) layout.Dimensions {
			defer paint.PushOpacity(gtx.Ops, disabledOpacity).Pop()
			size := gtx.Constraints.Min
			r := gtx.Dp(th.Radius)
			func() {
				defer clip.UniformRRect(image.Rectangle{Max: size}, r).Push(gtx.Ops).Pop()
				paint.Fill(gtx.Ops, th.Palette.Surface)
			}()
			b := widget.Border{Color: th.Palette.Border, CornerRadius: th.Radius, Width: unit.Dp(1)}
			return b.Layout(gtx, func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{Size: size} })
		},
		func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: 8, Bottom: 8, Left: 10, Right: 6}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				// Keep the height of an empty box the same as a text line.
				line := f.label(" ", th.Palette.Fg)
				if reason == "" {
					return line(gtx)
				}
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions { return f.warnIcon(gtx) }),
					layout.Flexed(1, f.label(reason, th.Palette.Fg)),
				)
			})
		})
}

// warnIcon draws Theme.Icons.Warning with a gap after it, or nothing.
func (f *Form) warnIcon(gtx layout.Context) layout.Dimensions {
	th := f.Theme
	if th.Icons.Warning == nil {
		return layout.Dimensions{}
	}
	return layout.Inset{Right: unit.Dp(6)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return th.Icons.Warning(gtx, unit.Dp(18), th.Palette.Warning)
	})
}

// warnTip floats the reason with the warning icon under a control whose
// size is dims, right-aligned to it.
func (f *Form) warnTip(gtx layout.Context, reason string, dims layout.Dimensions) {
	f.layoutFloat(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions { return f.warnIcon(gtx) }),
			layout.Flexed(1, f.caption(reason, f.Theme.Palette.Fg)),
		)
	}, func(size image.Point) image.Point {
		return image.Pt(dims.Size.X-size.X, dims.Size.Y+gtx.Dp(4))
	})
}

// layoutSubmit draws the submit button; while a field is disabled it is
// faded, ignores clicks and shows the reason while hovered.
func (f *Form) layoutSubmit(gtx layout.Context) layout.Dimensions {
	th := f.Theme
	b := material.Button(th.Material, &f.submitBtn, f.Labels.Submit)
	reason := f.disabledReason()
	if reason == "" {
		return b.Layout(gtx)
	}
	// submitTip only reports hovering; the button itself gets no input.
	dims := f.submitTip.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		defer paint.PushOpacity(gtx.Ops, disabledOpacity).Pop()
		return b.Layout(gtx.Disabled())
	})
	if f.submitTip.Hovered() {
		f.warnTip(gtx, reason, dims)
	}
	return dims
}
