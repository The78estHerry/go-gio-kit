[繁體中文](CLAUDE.md) | **English** | [日本語](CLAUDE.ja.md)

# CLAUDE.md — go-gio-kit

Project notes for Claude. This repository is public on GitHub from day one, so it only contains information that is safe to publish.

This file is a translation. `CLAUDE.md` (Traditional Chinese) is the primary version and the only one Claude Code loads automatically; update all three versions in the same commit.

## Purpose

Reusable packages that depend on Gio: `gioform` (TextField, SecretField, SelectField, Form, Theme) and `giowidget` (a borderless, always-on-top, draggable small window). Depends only on `gioui.org`.

The first user is go-credential-porter; go-virtual-display will switch to these packages later.

## Design principles

- Packages know nothing about "tokens" or "virtual displays"; all domain text is passed in by the caller.
- No `internal/` directory, so any repository can import the packages.
- Package names are lowercase single words; Gio-related ones get a `gio` prefix to avoid clashing with Gio's own packages.
- All packages in this repository are released together with plain tags such as `v0.1.0`. To split a package into its own module later, add a go.mod in its directory and immediately release a new version of the root module to avoid ambiguous imports.
- Never add an API that outputs secret content.

## GitHub decision

- **GitHub decision** (2026-11-09): GitHub (public) is the primary development location (`github.com/The78estHerry/go-gio-kit`); a copy is also kept on Gitea (`The78estWorld/go-gio-kit`) and every push goes to both: remote `origin` is Gitea and `github` is GitHub; push branches and tags to both and create releases on both (steps in `RELEASING.md`).
- Prepared: MIT LICENSE (The78ester), `.gitignore`.
- No real credentials, accounts or internal URLs in tests and examples (use `example.com`, `your-name`, `test-token-not-real`).

## Platforms

- Windows builds need no CGO; Gio needs CGO on macOS and Linux, which GitHub Actions builds.
- Only Windows is tested on real hardware for now; macOS and Linux are marked "implemented, not yet tested": platform-specific files start with `// Status: implemented, not yet tested on real hardware.`

## Language rules

- Write all code comments and Godoc in English.
- Write all PR descriptions, issue text and review comments in English.
- Commit messages are trilingual (English primary, Traditional Chinese secondary, Japanese supplementary); see "Commit message language (all projects)" in the global CLAUDE.md.
- `README.md` is the primary README, written in Traditional Chinese. Keep `README.en.md` (English) and `README.ja.md` (Japanese) in sync in the same commit. The same applies to the three versions of `CLAUDE.md`.
- Always reply to the developer in Traditional Chinese.

## Progress log

- 2026-11-09: Created the repository skeleton (go.mod, LICENSE, .gitignore, CLAUDE.md in three languages).
- 2026-11-09: Icon finalized (blocks B3: GO + input box + small window + password dots + speed lines, `assets/icon.svg`, rendered by `assets/make-icon.ps1`).
- 2026-11-09 18:08: hard crash (Kernel-Power 41 + 6008, no BugCheck) while `go vet`/`go test` ran here and another project compiled a Tauri app. Not heat or memory; display driver timeouts (LiveKernelEvent 141/117) are the main suspect. `gioform/`, `giowin/`, `go.mod`, `go.sum` were intact but uncommitted.
- 2026-11-10: After the crash, re-ran `go vet`, `go test -p 1 ./...` (Windows) and cross-vet for Linux/macOS (only `gioform` and `giowin`; `gioform/example` imports `gioui.org/app`, which needs CGO on Linux), all passing. giowin now forgets the subclass entry on `WM_NCDESTROY`. Committed gioform/giowin. Next: finish M3 (run the example on real hardware, README, release v0.1.0).
- 2026-11-10: **gioform passed real-hardware testing** (the user drove the example): secret fields accept paste, block copy/cut and reveal only while Show is held; Zhuyin and Japanese romaji (chi, tsu, kya, kitte) work; required/validation, Enter/Esc, dark/light with the title bar and CJK text all work. From user feedback: the NewSecret confirmation is now compared while typing (no warning while it is still a prefix, red mismatch at once, "match" when equal; new `Labels.Match`). Next: README in three languages, then the first push to GitHub and release v0.1.0 (needs the user's approval).
- 2026-11-10: README in three languages, `.github/FUNDING.yml`, `RELEASING.md`, CI (GitHub: Windows and macOS build everything, Linux checks only `gioform` and `giowin`; `.gitea/workflows` runs Linux only so Windows/macOS jobs do not queue forever on the NAS). Next: with the user's approval, first push to GitHub (the repo already exists), tag `v0.1.0`, create releases on both sides.
- 2026-11-10 (pause point): main and the `v0.1.0` tag are on GitHub (`github` remote added) and Gitea. **Resume here**: check CI for `v0.1.0` (GitHub three platforms, Gitea Linux); if it passes, follow steps 6 and 7 of `RELEASING.md` (releases on both sides, module proxy); if it fails, fix and release `v0.1.1` (pushed tags must not move).
- 2026-11-10: **v0.1.0 released**: GitHub three-platform and Gitea CI passed, releases on both sides, available through the module proxy. Next: giowidget for M4 (extracted from go-virtual-display).
- TODO: Gitea Actions created no jobs for `8b2d4f9` and `8d2bc9d` (CLAUDE.md-only changes); only `fe29bb3` ran (once for main, once for the tag). GitHub ran all of them. Cause not investigated.
- 2026-11-11: added `giowidget` (extracted from go-virtual-display and split into layers): the core `place.go` (default bottom-right corner, `besideCursor`, `visibleEnough`, the `placer` for unplugged monitors) works only with `image.Rectangle` and asks the platform through the `screen` interface; unit tests use fake monitor layouts. Windows implementation (tool-window style, Windows 11 rounded corners, one shared `EnumDisplayMonitors` callback); other platforms get borderless and on-top only. `giowidget/example` passed all real-hardware checks (driven by the user, including unplugging and replugging a second monitor and dragging while displaced). The Linux CI part now also skips `giowidget` (`gioui.org/app` needs CGO). Next: switch go-virtual-display to the new packages, then release both kits.
- 2026-11-11: M6 preparation: gioform gained `Choice` fields (radio buttons; `Field.Choices`/`Default`, `Form.Choice(key)` so the app can adapt the form while the user picks; `Option` type), and the example has a "Service" choice; giowin gained `AuthWindow()`, which returns the owner window for desktop/userauth (depends on go-desktop-kit v0.3.0). Unit tests pass; **the choice field has not been looked at on real hardware yet** (open the example next time).
- 2026-11-11: gioform gained `Select` (drop-down list), `Combo` (drop-down you can type in), `Duration` (time; design follows the time input of go-screen-burn-in-fixer and icon-theme-park: ◀ ▶ step through common values, click the value to type it, min/max, "?" help) and adding/removing options (`Field.Add`/`Remove`, `Option.Fixed`, `CanRemove`, `ConfirmRemove`, `UndoRemove`, `AddExternal`; when removing is not allowed the whole button fades and shows ⚠ with the reason; the icon can be replaced in `Theme.Icons`). The earlier `Field.Action` became `Add` (never released, so no compatibility issue). The pure logic (`ParseDuration`, option filtering) has unit tests. **All passed on real hardware** (the user ran the example, Choice fields included). Bugs fixed: Gio gives a key event to the first reader, and the form read the form-wide Esc first, so lists never got it; the Esc release also canceled the form (now only the press counts). Next: release go-gio-kit v0.3.0, then the M6 GUI skeleton in credential-porter.
