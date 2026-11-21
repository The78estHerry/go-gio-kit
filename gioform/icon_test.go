package gioform

import (
	"image/color"
	"testing"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/unit"
)

func TestOptionIcon(t *testing.T) {
	var drawn []color.NRGBA
	icon := func(gtx layout.Context, size unit.Dp, c color.NRGBA) layout.Dimensions {
		drawn = append(drawn, c)
		return layout.Dimensions{Size: gtx.Constraints.Min}
	}
	f := New(NewTheme(false), []Field{{Key: "svc", Label: "Service", Type: Select, Default: "a",
		Choices: []Option{{Value: "a", Label: "A", Icon: icon}, {Value: "b", Label: "B"}}}})
	var r input.Router
	frame(&r, f)
	if len(drawn) != 1 || drawn[0] != f.Theme.Palette.Fg {
		t.Fatalf("closed Select drew the icon %d times with %v", len(drawn), drawn)
	}
	f.field("svc").choice.Value = "b"
	drawn = nil
	frame(&r, f)
	if len(drawn) != 0 {
		t.Fatal("an option without an icon drew one")
	}
}
