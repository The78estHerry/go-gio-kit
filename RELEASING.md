# Releasing go-gio-kit

go-gio-kit is a Go library: a release is a git tag plus release notes.
There are no binaries to attach. Every release goes to both remotes:

| Remote | Host | Role |
|---|---|---|
| `github` | `github.com/The78estHerry/go-gio-kit` | Primary, public; the module path |
| `origin` | Gitea `The78estWorld/go-gio-kit` | Copy |

## Version numbers

Semantic versioning. While the version is `v0.x.y`, a minor bump (`v0.2.0`)
may change APIs; a patch bump (`v0.1.1`) only fixes bugs. Never move or
delete a tag that has been pushed: the Go module proxy caches it forever.

## Steps

1. Make sure the working tree is clean and `main` is up to date on both remotes.
2. Check the code on Windows and cross-compile the other platforms:

   ```bash
   go vet ./...
   go test ./...
   GOOS=linux go vet ./gioform/ ./giowin/
   GOOS=darwin go vet ./gioform/ ./giowin/
   ```

   `gioform/example` imports `gioui.org/app`, which needs CGO on Linux and
   macOS, so it cannot be cross-checked from Windows; CI builds it on macOS.
   Also run the example on real hardware and try the secret fields and the
   IME (see the comment at the top of `gioform/example/main.go`).

3. Update the package status column in `README.md`, `README.en.md` and
   `README.ja.md` for packages added or changed in this release.
4. Tag and push to both remotes:

   ```bash
   git tag -a v0.1.0 -m "v0.1.0"
   git push origin main --tags
   git push github main --tags
   ```

5. Wait for the GitHub Actions run on the tag to pass on all three platforms,
   and the Gitea Actions run (Linux only) to pass.
6. Create the release on both sides with the same notes. Write the notes in
   Traditional Chinese first, with English and Japanese in `<details>` blocks.
   - GitHub: `POST https://api.github.com/repos/The78estHerry/go-gio-kit/releases`
     with `tag_name`, `name` and `body`.
   - Gitea: `POST /api/v1/repos/The78estWorld/go-gio-kit/releases` with the same fields.
   - Read tokens with `git credential fill`; never print them.
7. Check that the version resolves through the module proxy:

   ```bash
   GOPROXY=https://proxy.golang.org go list -m github.com/The78estHerry/go-gio-kit@v0.1.0
   ```
