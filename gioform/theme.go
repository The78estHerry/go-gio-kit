// Package gioform provides labeled input fields and generated forms for Gio:
// text fields, secret fields that never put their content on the
// clipboard, a new-password mode with confirmation, focus order, Enter to
// submit, Esc to cancel and validation.
//
// All user-visible words come from Labels, which callers can translate.
package gioform

import (
	"image/color"

	"gioui.org/font/gofont"
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
}

func rgb(c uint32) color.NRGBA {
	return color.NRGBA{R: uint8(c >> 16), G: uint8(c >> 8), B: uint8(c), A: 0xff}
}

// DarkPalette and LightPalette are the default color schemes.
var (
	DarkPalette = Palette{
		Bg: rgb(0x1e1f22), Surface: rgb(0x2b2d31), Fg: rgb(0xe6e6e6), Muted: rgb(0xa0a4ab),
		Border: rgb(0x4a4d55), Accent: rgb(0x4c8bf5), OnAccent: rgb(0xffffff), Error: rgb(0xf28b82),
	}
	LightPalette = Palette{
		Bg: rgb(0xffffff), Surface: rgb(0xf4f5f7), Fg: rgb(0x1e1f22), Muted: rgb(0x5f6368),
		Border: rgb(0xc4c7cc), Accent: rgb(0x3367d6), OnAccent: rgb(0xffffff), Error: rgb(0xc5221f),
	}
)

// Theme is the look of forms: colors, corner radius and spacing.
type Theme struct {
	Material *material.Theme
	Palette  Palette
	Radius   unit.Dp
	Spacing  unit.Dp
}

// NewTheme returns a theme for a dark or light UI. Its text shaper loads the
// system fonts, so any character a user types (Chinese, Japanese, ...) can
// be shown.
func NewTheme(dark bool) *Theme {
	t := &Theme{Material: material.NewTheme(), Radius: 6, Spacing: 12}
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
