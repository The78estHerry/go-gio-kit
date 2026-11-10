<p align="center"><img src="assets/icon-256.png" width="160" alt="go-gio-kit のアイコン：積み木で作った GO の文字、入力欄、小さなウィンドウ、パスワードの点、左側にスピード線"></p>

# go-gio-kit

[繁體中文](README.md) | [English](README.en.md) | **日本語**

> 繁体字中国語版の README が正式版です。内容が異なる場合は繁体字中国語版を優先してください。

[![PayPal](https://img.shields.io/badge/PayPal-paypal.me%2FThe78ester-00457C?logo=paypal&logoColor=white)](https://paypal.me/The78ester) [![Ko-fi](https://img.shields.io/badge/Ko--fi-the78ester-FF5E5B?logo=kofi&logoColor=white)](https://ko-fi.com/the78ester)

💖 このツールが役に立った場合は、[PayPal](https://paypal.me/The78ester) または [Ko-fi](https://ko-fi.com/the78ester) から開発を支援していただけるとうれしいです。

[Gio](https://gioui.org) デスクトップアプリ向けの再利用可能な UI 部品です。トークンやパスワードを安全に入力できるフォーム、ダーク／ライトテーマ、そして Gio 自体が扱わないウィンドウの修正（Windows の日本語 IME、ダークタイトルバー）を提供します。

GUI に依存しないパッケージ（OS の資格情報ストア、暗号化ファイル、設定フォルダーなど）は別のリポジトリにあります：[go-desktop-kit](https://github.com/The78estHerry/go-desktop-kit)。

## パッケージ

| パッケージ | 用途 | 状態 |
|---|---|---|
| `gioform` | フォーム：通常のテキスト欄、Secret 欄（トークン、パスワード）、NewSecret 欄（新しいパスワードの設定、確認欄と強度表示つき）、検証、ダーク／ライトテーマ | v0.1.0 |
| `giowin` | ウィンドウの修正：Windows の日本語 IME、ダークタイトルバー、OS の配色設定の取得 | v0.1.0 |
| `giowidget` | 枠なし・最前面固定・ドラッグ可能な小さなウィンドウ | 予定 |

## インストール

```bash
go get github.com/The78estHerry/go-gio-kit@latest
```

Gio v0.10.3 が必要です。

## 使用例

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
    gw.HandleEvent(e) // すべてのイベントを giowin に渡す
    switch e := e.(type) {
    case app.FrameEvent:
        gtx := app.NewContext(&ops, e)
        form.Layout(gtx)
        if v, ok := form.Submitted(); ok {
            save(v.Text("host"), v.Text("user"), v.Secret("token").Bytes())
            v.Wipe() // 使い終わったらすぐ消去
            form.Reset()
        }
        e.Frame(gtx.Ops)
    }
}
```

完全な例（手動テスト用のプログラムを兼ねる）は [`gioform/example`](gioform/example/main.go) にあります：

```bash
go run ./gioform/example
```

UI の文字は既定で英語です。`form.Labels`（型 `gioform.Labels`）を置き換えると翻訳できます。

## セキュリティ設計

- **Secret 欄はコピーできない**：Gio のエディターはマスクしていても Ctrl+C／Ctrl+X で原文をクリップボードに入れてしまいます。gioform はエディターが処理する前にこれらのキーを横取りします。貼り付けは通常どおり使えます。
- **押している間だけ表示**：Secret 欄の Show ボタンを離すとマスクに戻ります。
- **値は消去できるバイト列**：`SecretValue.Wipe` は内容を 0 で上書きします。Go の文字列は消去できず、Gio のエディターは内部で文字列として保持するため、消去はベストエフォートです。秘密情報は長くメモリに置かず、OS の資格情報ストアに保存してください。
- **秘密情報を出力する API はない**：パッケージは値を呼び出し側に渡すだけで、記録も表示もしません。
- **新しいパスワードには任意の Unicode を使える**：強度は文字数で数えるので、中国語の一文でも強いパスワードになります。確認欄は入力中に比較されます。

## giowin が修正する問題

- **Windows の日本語 IME**：Gio v0.10.3 は `GCS_DELTASTART`（変更の開始位置）を変換の開始位置として扱うため、ローマ字入力の `chi`、`tsu`、`kya`、`kitte` などがほとんど入力できません（注音は影響なし）。giowin はメッセージが Gio に届く前にこのフラグを消します。上流で修正されたら取り除きます。
- **ダークタイトルバー**：Windows のタイトルバーはシステムが描画します。`SetDarkTitleBar` で UI と一緒に切り替えます。
- **OS の配色**：`SystemPrefersDark` は Windows、macOS、Linux（GNOME）のダーク／ライト設定を読み取ります。

## 対応プラットフォーム

実機テストは Windows のみです。macOS と Linux のコードは実装済みですが、実機ではまだ確認していません。結果の報告を歓迎します。

| パッケージ | Windows | macOS | Linux |
|---|---|---|---|
| `gioform` | 実機テスト済み | 未テスト | 未テスト |
| `giowin` | 実機テスト済み | 実装済み・未テスト（配色の取得のみ） | 実装済み・未テスト（配色の取得のみ） |

## ライセンス

[MIT](LICENSE)。本ソフトウェアは現状のまま提供され、いかなる保証もありません。使用によって生じたいかなる損害についても、作者は責任を負いません。

このツールが役に立った場合は、開発への支援（スポンサー）を歓迎します：[PayPal](https://paypal.me/The78ester) または [Ko-fi](https://ko-fi.com/the78ester)。
