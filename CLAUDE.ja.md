[繁體中文](CLAUDE.md) | [English](CLAUDE.en.md) | **日本語**

# CLAUDE.md — go-gio-kit

Claude 向けのプロジェクト説明です。このリポジトリは初日から GitHub で公開しているため、公開しても問題のない内容だけを記載します。

このファイルは翻訳版です。主版は `CLAUDE.md`（繁体字中国語）で、Claude Code が自動的に読み込むのはそのファイルだけです。3 つの版は同じコミットで更新してください。

## 目的

Gio に依存する再利用可能なパッケージ：`gioform`（TextField、SecretField、SelectField、Form、Theme）と `giowidget`（枠なし・最前面固定可・ドラッグ可能な小さなウィンドウ）。依存は `gioui.org` のみ。

最初の利用者は go-credential-porter。go-virtual-display も後でこれらのパッケージに切り替える。

## 設計方針

- パッケージは「トークン」や「仮想ディスプレイ」を知らない。業務上の文言はすべて呼び出し側から渡す。
- `internal/` は置かず、どのリポジトリからも import できるようにする。
- パッケージ名はすべて小文字の 1 単語。Gio 関連には `gio` の接頭辞を付け、Gio 自身のパッケージとの名前の衝突を避ける。
- このリポジトリ内のパッケージはまとめてリリースし、タグは通常の `v0.1.0` 形式。後でパッケージを独立した module に分けるときは、そのフォルダに go.mod を追加し、ambiguous import を避けるため直ちにルート module の新バージョンをリリースする。
- 機密内容を出力する API を追加しない。

## GitHub に関する決定

- **GitHub に関する決定**（2026-11-09）：公開の GitHub を主な開発場所とし（`github.com/The78estHerry/go-gio-kit`）、Gitea（`The78estWorld/go-gio-kit`）にも置いて両方にプッシュする。remote は `origin` が Gitea、`github` が GitHub。ブランチとタグは両方にプッシュし、リリースも両方で作成する（手順は `RELEASING.md`）。
- 準備済み：MIT LICENSE（The78ester）、`.gitignore`。
- テストとサンプルに本物の資格情報、アカウント、内部 URL を入れない（`example.com`、`your-name`、`test-token-not-real` を使う）。

## プラットフォーム

- Windows では CGO 不要。Gio は macOS と Linux で CGO が必要なため、GitHub Actions でビルドする。
- 現在は Windows でのみ実機テストを行う。macOS と Linux は「実装済み・未テスト」と表示し、プラットフォーム固有のファイルの先頭に `// Status: implemented, not yet tested on real hardware.` を書く。

## 言語の規則

- コードコメントと Godoc はすべて英語。
- PR の説明、issue の本文、コードレビューのコメントはすべて英語。
- コミットメッセージは 3 言語（英語が主、繁体字中国語が次、日本語が補足）。形式はグローバル CLAUDE.md の「Commit message language (all projects)」を参照。
- `README.md` が主版で繁体字中国語。`README.en.md`（英語）と `README.ja.md`（日本語）は同じコミットで同期する。`CLAUDE.md` の 3 つの版も同様。
- 開発者への返信は常に繁体字中国語。

## 進捗記録

- 2026-11-09：リポジトリの骨組みを作成（go.mod、LICENSE、.gitignore、3 言語版の CLAUDE.md）。
- 2026-11-09：アイコン決定（積み木 B3：GO・入力欄・小さなウィンドウ・パスワードの点・スピード線、`assets/icon.svg`、`assets/make-icon.ps1` で PNG に変換）。
- 2026-11-09 18:08：ハングアップ（Kernel-Power 41＋6008、BugCheck なし）。ここで `go vet`／`go test` を実行中、別プロジェクトで Tauri アプリをコンパイル中だった。熱やメモリ不足ではなく、ディスプレイドライバーのタイムアウト（LiveKernelEvent 141／117）が主な容疑。`gioform/`、`giowin/`、`go.mod`、`go.sum` は無事だったが未コミット。
- 2026-11-10：クラッシュ後に `go vet`、`go test -p 1 ./...`（Windows）と Linux／macOS のクロス vet（`gioform`、`giowin` のみ。`gioform/example` は `gioui.org/app` を使い、Linux では CGO が必要）を再実行し、すべて成功。giowin はウィンドウ破棄（`WM_NCDESTROY`）時に subclass の記録を消すようにした。gioform／giowin をコミット。次：M3 の仕上げ（実機で example を実行、README、v0.1.0 リリース）。
- 2026-11-10：**gioform の実機テスト合格**（ユーザーが example を操作）：Secret 欄は貼り付け可・コピー／切り取り不可・Show を押している間だけ表示；注音と日本語ローマ字（chi、tsu、kya、kitte）正常；必須と検証、Enter／Esc、ダーク／ライトとタイトルバー、日中文字の表示も正常。ユーザーの指摘により、NewSecret の確認欄を入力中に即時比較するようにした（入力途中は警告なし、異なれば即赤字、一致で match；`Labels.Match` を追加）。次：3 言語の README、その後 GitHub への初回 push と v0.1.0 リリース（ユーザーの同意が必要）。
- 2026-11-10：3 言語の README、`.github/FUNDING.yml`、`RELEASING.md`、CI（GitHub：Windows と macOS は全体をビルド、Linux は `gioform`、`giowin` のみ。`.gitea/workflows` は Linux のみで、Windows／macOS のジョブが NAS で待ち続けないようにする）。次：ユーザーの同意を得て GitHub へ初回 push（リポジトリ作成済み）、`v0.1.0` タグ、両方でリリース作成。
