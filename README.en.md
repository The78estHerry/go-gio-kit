<p align="center"><img src="assets/icon-256.png" width="160" alt="go-gio-kit icon: blocks showing the letters GO, an input box, a small window and password dots, with speed lines on the left"></p>

# go-gio-kit

[繁體中文](README.md) | **English** | [日本語](README.ja.md)

> The Traditional Chinese README is the authoritative version. If the versions differ, it takes precedence.

[![PayPal](https://img.shields.io/badge/PayPal-paypal.me%2FThe78ester-00457C?logo=paypal&logoColor=white)](https://paypal.me/The78ester) [![Ko-fi](https://img.shields.io/badge/Ko--fi-the78ester-FF5E5B?logo=kofi&logoColor=white)](https://ko-fi.com/the78ester)

💖 If this tool helps you, please consider supporting development via [PayPal](https://paypal.me/The78ester) or [Ko-fi](https://ko-fi.com/the78ester).

Reusable UI building blocks for [Gio](https://gioui.org) desktop apps: forms for entering tokens and passwords safely, dark and light themes, and window fixes that Gio itself does not handle (Japanese IME on Windows, dark title bar).

Packages that do not depend on a GUI (OS credential store, encrypted files, config folders and more) live in a separate repo: [go-desktop-kit](https://github.com/The78estHerry/go-desktop-kit).

## Packages

| Package | Purpose | Status |
|---|---|---|
| `gioform` | Forms: plain text fields, Secret fields (tokens, passwords), NewSecret fields (setting a new password, with confirmation and strength hint), validation, dark and light themes | v0.1.0 |
| `giowin` | Window fixes: Japanese IME on Windows, dark title bar, reading the OS color scheme | v0.1.0 |
| `giowidget` | Desktop widget window: borderless, optionally on top, dragged by its background, kept out of the taskbar; remembers its position, jumps next to the pointer, moves to the primary monitor while its monitor is unplugged and back afterwards | In development (main) |

## Install

```bash
go get github.com/The78estHerry/go-gio-kit@latest
```

Requires Gio v0.10.3.

## Example

```go
th := gioform.NewTheme(gioform.System.Dark(systemDark))
form := gioform.New(th, []gioform.Field{
    {Key: "host", Label: "Server", Placeholder: "git.example.com",
        Required: true, Validate: gioform.Host("")},
    {Key: "user", Label: "User name", Placeholder: "your-name", Required: true},
    {Key: "token", Label: "Token", Type: gioform.Secret, Required: true, Trim: true},
})

var gw giowin.Window
for {
    e := w.Event()
    gw.HandleEvent(e) // giowin must see every event
    switch e := e.(type) {
    case app.FrameEvent:
        gtx := app.NewContext(&ops, e)
        form.Layout(gtx)
        if v, ok := form.Submitted(); ok {
            save(v.Text("host"), v.Text("user"), v.Secret("token").Bytes())
            v.Wipe() // clear as soon as you are done
            form.Reset()
        }
        e.Frame(gtx.Ops)
    }
}
```

A complete example, which also serves as a manual test, is in [`gioform/example`](gioform/example/main.go):

```bash
go run ./gioform/example
```

The UI text is English by default; replace `form.Labels` (type `gioform.Labels`) to translate it.

## Security design

- **Secret fields cannot be copied**: a Gio editor still puts its plain text on the clipboard on Ctrl+C/Ctrl+X even when masked; gioform intercepts those keys before the editor sees them. Paste works as usual.
- **Shown only while held**: the Show button of a Secret field masks the text again when released.
- **Values are wipeable bytes**: `SecretValue.Wipe` overwrites the content with zeros. Go strings cannot be cleared and the Gio editor keeps its text as a string, so wiping is best effort; keep secrets in the OS credential store rather than in memory.
- **No API that outputs secrets**: the package only hands values to the caller and never logs or prints them.
- **New passwords may use any Unicode**: strength counts characters, so a sentence in Chinese can be a strong password. The confirmation is compared while typing.

## What giowin fixes

- **Japanese IME on Windows**: Gio v0.10.3 treats `GCS_DELTASTART` (where the change begins) as the start of the composition, so romaji such as `chi`, `tsu`, `kya` and `kitte` can hardly be typed (Zhuyin is not affected). giowin clears that flag before the message reaches Gio. The fix will be removed once upstream is fixed.
- **Dark title bar**: Windows draws the title bar itself; `SetDarkTitleBar` switches it along with the UI.
- **OS color scheme**: `SystemPrefersDark` reads the dark/light setting on Windows, macOS and Linux (GNOME).

## Platform support

Tested on real hardware on Windows only. The macOS and Linux code is implemented but not yet verified on real machines; reports are welcome.

| Package | Windows | macOS | Linux |
|---|---|---|---|
| `gioform` | Tested | Not tested | Not tested |
| `giowin` | Tested | Implemented, not tested (color scheme only) | Implemented, not tested (color scheme only) |
| `giowidget` | Tested | Implemented, not tested (borderless and on top; the window manager places it) | Implemented, not tested (same) |

## License

[MIT](LICENSE). This software is provided as is, without warranty of any kind; the author is not liable for any damage caused by its use.

If this tool helps you, sponsorship to support development is welcome via [PayPal](https://paypal.me/The78ester) or [Ko-fi](https://ko-fi.com/the78ester).
