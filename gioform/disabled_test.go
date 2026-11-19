package gioform

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

// frame lays the form out once and hands the ops to the router.
func frame(r *input.Router, f *Form) {
	gtx := layout.Context{
		Ops:         new(op.Ops),
		Source:      r.Source(),
		Constraints: layout.Exact(image.Pt(500, 800)),
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Now:         time.Now(),
	}
	f.Layout(gtx)
	r.Frame(gtx.Ops)
}

func TestDisabledField(t *testing.T) {
	f := New(NewTheme(true), []Field{
		{Key: "pin", Label: "PIN", Type: Secret, Required: true},
		{Key: "name", Label: "Name"},
	})
	f.SetDisabled("pin", "Too many wrong PINs. Try again in 30s.")
	if f.Disabled("pin") == "" || f.Disabled("name") != "" || f.Disabled("missing") != "" {
		t.Fatal("Disabled does not report the reasons")
	}
	f.field("pin").ed.SetText("test-pin-not-real")

	var r input.Router
	frame(&r, f) // sets the first focus
	frame(&r, f)
	if r.Source().Focused(&f.field("pin").ed) {
		t.Fatal("a disabled field took the first focus")
	}
	if !r.Source().Focused(&f.field("name").ed) {
		t.Fatal("the first enabled field did not get the focus")
	}

	enter := func() {
		r.Queue(key.Event{Name: key.NameReturn, State: key.Press}, key.Event{Name: key.NameReturn, State: key.Release})
		frame(&r, f)
		frame(&r, f)
	}
	enter()
	if _, ok := f.Submitted(); ok {
		t.Fatal("submitted while a field is disabled")
	}

	f.SetDisabled("pin", "")
	enter()
	v, ok := f.Submitted()
	if !ok {
		t.Fatal("not submitted after the field was enabled again")
	}
	defer v.Wipe()
	if string(v.Secret("pin").Bytes()) != "test-pin-not-real" {
		t.Fatal("the text of a disabled field was not kept")
	}
}

func TestDisabledReleasesFocus(t *testing.T) {
	f := New(NewTheme(false), []Field{
		{Key: "pin", Label: "PIN", Type: NewSecret},
		{Key: "kind", Label: "Kind", Type: Choice, Choices: []Option{{Value: "a", Label: "A"}}},
	})
	var r input.Router
	frame(&r, f)
	frame(&r, f)
	if !r.Source().Focused(&f.field("pin").ed) {
		t.Fatal("the first field did not get the focus")
	}
	f.SetDisabled("pin", "Locked")
	f.SetDisabled("kind", "Not now") // a field without a text box is drawn disabled too
	frame(&r, f)
	frame(&r, f)
	if r.Source().Focused(&f.field("pin").ed) {
		t.Fatal("a field kept the focus after it was disabled")
	}
	f.Focus("pin") // still disabled: ignored
	frame(&r, f)
	frame(&r, f)
	if r.Source().Focused(&f.field("pin").ed) {
		t.Fatal("Focus moved the focus to a disabled field")
	}
	f.SetDisabled("pin", "")
	f.Focus("pin")
	frame(&r, f)
	frame(&r, f)
	if !r.Source().Focused(&f.field("pin").ed) {
		t.Fatal("Focus did not move the focus to the enabled field")
	}
}
