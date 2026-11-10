// Command example shows a giowidget and doubles as a manual test: drag it
// by its background, toggle "on top", jump next to the pointer, close it
// and run the example again to see the remembered position.
package main

import (
	"encoding/json"
	"image"
	"image/color"
	"log"
	"os"
	"path/filepath"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/The78estHerry/go-gio-kit/giowidget"
)

// The position file lives in the temp directory; it only serves this demo.
var posFile = filepath.Join(os.TempDir(), "giowidget-example.json")

func loadPos() *image.Point {
	b, err := os.ReadFile(posFile)
	if err != nil {
		return nil
	}
	var p image.Point
	if json.Unmarshal(b, &p) != nil {
		return nil
	}
	return &p
}

func main() {
	go func() {
		if err := run(); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run() error {
	w := giowidget.New(giowidget.Options{
		Title: "giowidget example", Width: 300, Height: 120,
		TopMost: true, Position: loadPos(),
	})
	th := material.NewTheme()
	th.Palette.Fg = color.NRGBA{0xf2, 0xf2, 0xf5, 0xff}
	var (
		ops                     op.Ops
		topBtn, cursorBtn, quit widget.Clickable
		top                     = true
	)
	for {
		e := w.Window().Event()
		w.HandleEvent(e)
		switch e := e.(type) {
		case app.DestroyEvent:
			if p, ok := w.Position(); ok {
				b, _ := json.Marshal(p)
				os.WriteFile(posFile, b, 0o600)
			}
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			if topBtn.Clicked(gtx) {
				top = !top
				w.SetTopMost(top)
			}
			if cursorBtn.Clicked(gtx) {
				w.MoveToCursor()
			}
			if quit.Clicked(gtx) {
				w.Window().Perform(system.ActionClose)
			}
			paint.Fill(gtx.Ops, color.NRGBA{0x22, 0x25, 0x2b, 0xff})
			giowidget.DragArea(gtx)
			layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(material.Body1(th, "Drag me by the background").Layout),
					layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := "On top: on"
						if !top {
							label = "On top: off"
						}
						btn := func(c *widget.Clickable, s string) layout.FlexChild {
							return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return layout.Inset{Right: unit.Dp(6)}.Layout(gtx, material.Button(th, c, s).Layout)
							})
						}
						return layout.Flex{}.Layout(gtx, btn(&topBtn, label), btn(&cursorBtn, "To cursor"), btn(&quit, "Close"))
					}),
				)
			})
			e.Frame(gtx.Ops)
		}
	}
}
