// Package gioform provides labeled input fields and generated forms for Gio:
// text fields, secret fields that never put their content on the
// clipboard, a new-password mode with confirmation, focus order, Enter to
// submit, Esc to cancel and validation.
//
// All user-visible words come from Labels, which callers can translate.
package gioform

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

// Mode is the UI color scheme setting.
type Mode int

const (
	// System follows the operating system's dark or light setting.
	System Mode = iota
	Dark
	Light
)

// Dark reports whether the mode means a dark UI, given the system setting.
func (m Mode) Dark(systemDark bool) bool {
	switch m {
	case Dark:
		return true
	case Light:
		return false
	}
	return systemDark
}

// Palette holds the colors of one color scheme.
type Palette struct {
	Bg       color.NRGBA // window background
	Surface  color.NRGBA // field background
	Fg       color.NRGBA // text
	Muted    color.NRGBA // labels, hints, help text
	Border   color.NRGBA // field border
	Accent   color.NRGBA // focus ring, primary button
	OnAccent color.NRGBA // text on Accent
	Error    color.NRGBA // validation messages
	Warning  color.NRGBA // warning icons
}

func rgb(c uint32) color.NRGBA {
	return color.NRGBA{R: uint8(c >> 16), G: uint8(c >> 8), B: uint8(c), A: 0xff}
}

// DarkPalette and LightPalette are the default color schemes.
var (
	DarkPalette = Palette{
		Bg: rgb(0x1e1f22), Surface: rgb(0x2b2d31), Fg: rgb(0xe6e6e6), Muted: rgb(0xa0a4ab),
		Border: rgb(0x4a4d55), Accent: rgb(0x4c8bf5), OnAccent: rgb(0xffffff), Error: rgb(0xf28b82),
		Warning: rgb(0xfdd663),
	}
	LightPalette = Palette{
		Bg: rgb(0xffffff), Surface: rgb(0xf4f5f7), Fg: rgb(0x1e1f22), Muted: rgb(0x5f6368),
		Border: rgb(0xc4c7cc), Accent: rgb(0x3367d6), OnAccent: rgb(0xffffff), Error: rgb(0xc5221f),
		Warning: rgb(0xf9ab00),
	}
)

// Theme is the look of forms: colors, corner radius and spacing.
type Theme struct {
	Material *material.Theme
	Palette  Palette
	Radius   unit.Dp
	Spacing  unit.Dp
	Icons    Icons
}

// Icon draws an icon of the given size in color c. Replace the icons in
// Theme.Icons to change how forms look.
type Icon func(gtx layout.Context, size unit.Dp, c color.NRGBA) layout.Dimensions

// Icons are the icons forms draw.
type Icons struct {
	// Warning marks why something cannot be done, such as an option that
	// cannot be removed. It is drawn in Palette.Warning.
	Warning Icon
}

// WarningIcon is the default warning icon: a triangle with "!".
func WarningIcon(gtx layout.Context, size unit.Dp, c color.NRGBA) layout.Dimensions {
	s := float32(gtx.Dp(size))
	var p clip.Path
	p.Begin(gtx.Ops)
	p.MoveTo(f32.Pt(s/2, s*0.06))
	p.LineTo(f32.Pt(s*0.98, s*0.92))
	p.LineTo(f32.Pt(s*0.02, s*0.92))
	p.Close()
	paint.FillShape(gtx.Ops, c, clip.Outline{Path: p.End()}.Op())
	mark := color.NRGBA{R: 0x20, G: 0x21, B: 0x24, A: 0xff}
	bar := func(y0, y1 float32) {
		r := image.Rect(int(s*0.45), int(s*y0), int(s*0.55+0.999), int(s*y1))
		paint.FillShape(gtx.Ops, mark, clip.Rect(r).Op())
	}
	bar(0.36, 0.66)
	bar(0.73, 0.83)
	n := int(s)
	return layout.Dimensions{Size: image.Pt(n, n)}
}

// NewTheme returns a theme for a dark or light UI. Its text shaper loads the
// system fonts, so any character a user types (Chinese, Japanese, ...) can
// be shown.
func NewTheme(dark bool) *Theme {
	t := &Theme{Material: material.NewTheme(), Radius: 6, Spacing: 12, Icons: Icons{Warning: WarningIcon}}
	t.Material.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	t.SetDark(dark)
	return t
}

// SetDark switches between DarkPalette and LightPalette.
func (t *Theme) SetDark(dark bool) {
	if dark {
		t.SetPalette(DarkPalette)
	} else {
		t.SetPalette(LightPalette)
	}
}

// SetPalette applies a custom palette.
func (t *Theme) SetPalette(p Palette) {
	t.Palette = p
	t.Material.Palette = material.Palette{Bg: p.Bg, Fg: p.Fg, ContrastBg: p.Accent, ContrastFg: p.OnAccent}
}
