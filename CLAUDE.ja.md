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
