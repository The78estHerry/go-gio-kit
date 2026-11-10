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
