// Command example shows gioform and giowin and doubles as a manual test:
// switch the color scheme, type Chinese or Japanese in the Note field, try
// Ctrl+C in the secret fields, press Enter or Esc.
package main

import (
	"fmt"
	"log"
	"os"
	"time"

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
		w.Option(app.Title("gioform example"), app.Size(unit.Dp(560), unit.Dp(860)))
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
		{Key: "service", Label: "Service", Type: gioform.Select, Required: true,
			Add: "+ Add", Remove: "− Remove", ConfirmRemove: true,
			Choices: []gioform.Option{
				{Value: "github", Label: "GitHub", Fixed: true}, {Value: "gitea", Label: "Gitea", Fixed: true},
				{Value: "gitlab", Label: "GitLab"}, {Value: "forgejo", Label: "Forgejo"},
				{Value: "bitbucket", Label: "Bitbucket"}, {Value: "codeberg", Label: "Codeberg"},
				{Value: "azure", Label: "Azure DevOps"}, {Value: "sourcehut", Label: "SourceHut"},
				{Value: "gogs", Label: "Gogs"}, {Value: "onedev", Label: "OneDev"},
			},
			Help: "Select: Up/Down, Enter/Space, Esc. + Add types a new one; GitHub and Gitea are built in; Remove asks first"},
		{Key: "protocol", Label: "Protocol", Type: gioform.Choice, Default: "https",
			Choices: []gioform.Option{{Value: "https", Label: "HTTPS"}, {Value: "ssh", Label: "SSH"}}},
		{Key: "host", Label: "Server", Type: gioform.Combo, Placeholder: "git.example.com", Required: true,
			Add: "+", Remove: "−", UndoRemove: true,
			CanRemove: func(o gioform.Option) (bool, string) {
				if o.Value == "code.example.com" {
					return false, "A saved token uses this server. Remove that token first."
				}
				return true, ""
			},
			Choices: []gioform.Option{
				{Value: "git.example.com", Label: "git.example.com"},
				{Value: "code.example.com", Label: "code.example.com"},
				{Value: "gitea.example.org", Label: "gitea.example.org"},
			},
			Help: "Combo: type to filter; + adds the typed host, − removes it (code.example.com is in use)", Validate: gioform.Host("")},
		{Key: "user", Label: "User name", Placeholder: "your-name", Required: true},
		{Key: "token", Label: "Token", Type: gioform.Secret, Required: true, Trim: true,
			Help: "Paste works; copy and cut are blocked. Hold Show to see it."},
		{Key: "clear", Label: "Clear the clipboard after", Type: gioform.Duration, Default: "30s",
			Min: 5 * time.Second, Max: 10 * time.Minute,
			Help: "Duration field: arrows (or Left/Right), click the time to type it, hover ? for help"},
		{Key: "lock", Label: "Lock after idle", Type: gioform.Duration, Default: "15m", Max: 24 * time.Hour},
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
				status = fmt.Sprintf("Submitted: service=%s protocol=%s server=%s user=%s token=%d bytes clear=%v lock=%v export=%d bytes. The form was reset.",
					v.Text("service"), v.Text("protocol"), v.Text("host"), v.Text("user"), v.Secret("token").Len(),
					v.Duration("clear"), v.Duration("lock"), v.Secret("export").Len())
				v.Wipe()
				form.Reset()
			}
			for _, k := range []string{"service", "host"} {
				if form.OptionsChanged(k) {
					status = fmt.Sprintf("Options of %s changed: %d now (an application would save them).", k, len(form.Options(k)))
				}
			}
			if form.Canceled() {
				status = "Canceled (an application would close the window; the example keeps the form)."
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
