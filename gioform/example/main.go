// Command example shows gioform and giowin and doubles as a manual test:
// switch the color scheme, type Chinese or Japanese in the Note field, try
// Ctrl+C in the secret fields, press Enter or Esc.
package main

import (
	"fmt"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/The78estHerry/go-gio-kit/gioform"
	"github.com/The78estHerry/go-gio-kit/giowin"
)

func main() {
	go func() {
		w := new(app.Window)
		w.Option(app.Title("gioform example"), app.Size(unit.Dp(520), unit.Dp(720)))
		if err := run(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run(w *app.Window) error {
	var gw giowin.Window
	mode := gioform.System
	systemDark, _ := giowin.SystemPrefersDark()
	th := gioform.NewTheme(mode.Dark(systemDark))
	gw.SetDarkTitleBar(mode.Dark(systemDark))

	form := gioform.New(th, []gioform.Field{
		{Key: "service", Label: "Service", Type: gioform.Choice, Required: true, Default: "gitea",
			Choices: []gioform.Option{{Value: "github", Label: "GitHub"}, {Value: "gitea", Label: "Gitea"}},
			Help:    "Choice field: the Server field is only needed for Gitea"},
		{Key: "host", Label: "Server", Placeholder: "git.example.com", Required: true,
			Help: "Host name only, without https://", Validate: gioform.Host("")},
		{Key: "user", Label: "User name", Placeholder: "your-name", Required: true},
		{Key: "token", Label: "Token", Type: gioform.Secret, Required: true, Trim: true,
			Help: "Paste works; copy and cut are blocked. Hold Show to see it."},
		{Key: "export", Label: "Export password", Type: gioform.NewSecret,
			Help: "Any text, Chinese sentences too"},
		{Key: "note", Label: "Note (IME test)", Placeholder: "注音、日本語 chi / kitte"},
	})

	var scheme widget.Enum
	scheme.Value = "system"
	var list widget.List
	list.Axis = layout.Vertical
	status := "Fill in the form, then press Enter or OK."

	var ops op.Ops
	for {
		e := w.Event()
		gw.HandleEvent(e)
		switch e := e.(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			if scheme.Update(gtx) {
				mode = map[string]gioform.Mode{"system": gioform.System, "dark": gioform.Dark, "light": gioform.Light}[scheme.Value]
				systemDark, _ = giowin.SystemPrefersDark()
				th.SetDark(mode.Dark(systemDark))
				gw.SetDarkTitleBar(mode.Dark(systemDark))
			}
			if v, ok := form.Submitted(); ok {
				status = fmt.Sprintf("Submitted: service=%s server=%s user=%s token=%d bytes export=%d bytes. The form was reset.",
					v.Text("service"), v.Text("host"), v.Text("user"), v.Secret("token").Len(), v.Secret("export").Len())
				v.Wipe()
				form.Reset()
			}
			if form.Canceled() {
				status = "Canceled."
				form.Reset()
			}

			paint.Fill(gtx.Ops, th.Palette.Bg)
			items := []layout.Widget{
				func(gtx layout.Context) layout.Dimensions {
					l := material.H6(th.Material, "gioform example")
					l.Color = th.Palette.Fg
					return l.Layout(gtx)
				},
				func(gtx layout.Context) layout.Dimensions {
					radio := func(key, label string) layout.FlexChild {
						return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							r := material.RadioButton(th.Material, &scheme, key, label)
							r.Color, r.IconColor = th.Palette.Fg, th.Palette.Accent
							return r.Layout(gtx)
						})
					}
					return layout.Flex{}.Layout(gtx, radio("system", "System"), radio("dark", "Dark"), radio("light", "Light"))
				},
				form.Layout,
				func(gtx layout.Context) layout.Dimensions {
					l := material.Body2(th.Material, fmt.Sprintf("%s\nBlocked copy/cut attempts: %d", status, form.BlockedCopies()))
					l.Color = th.Palette.Muted
					return l.Layout(gtx)
				},
			}
			material.List(th.Material, &list).Layout(gtx, len(items), func(gtx layout.Context, i int) layout.Dimensions {
				return layout.UniformInset(unit.Dp(16)).Layout(gtx, items[i])
			})
			e.Frame(gtx.Ops)
		}
	}
}
