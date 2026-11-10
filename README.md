<p align="center"><img src="assets/icon-256.png" width="160" alt="go-gio-kit 圖示：積木組成的 GO 字樣、輸入框、小視窗與密碼圓點，左側有速度線"></p>

# go-gio-kit

**繁體中文** | [English](README.en.md) | [日本語](README.ja.md)

[![PayPal](https://img.shields.io/badge/PayPal-paypal.me%2FThe78ester-00457C?logo=paypal&logoColor=white)](https://paypal.me/The78ester) [![Ko-fi](https://img.shields.io/badge/Ko--fi-the78ester-FF5E5B?logo=kofi&logoColor=white)](https://ko-fi.com/the78ester)

💖 如果這個工具對你有幫助，歡迎透過 [PayPal](https://paypal.me/The78ester) 或 [Ko-fi](https://ko-fi.com/the78ester) 贊助支持開發。

[Gio](https://gioui.org) 桌面程式的可重用介面元件：可以安全輸入 Token 與密碼的表單、深淺色主題，以及 Gio 本身沒有處理的視窗修正（Windows 的日文輸入法、深色標題列）。

不依賴 GUI 的套件（系統保管庫、加密檔案、設定資料夾等）放在另一個 repo：[go-desktop-kit](https://github.com/The78estHerry/go-desktop-kit)。

## 套件

| 套件 | 用途 | 狀態 |
|---|---|---|
| `gioform` | 表單：一般文字欄、Secret 欄（Token、密碼）、NewSecret 欄（設定新密碼，含確認欄與強度提示）、驗證、深淺色主題 | v0.1.0 |
| `giowin` | 視窗修正：Windows 日文輸入法、深色標題列、讀取系統的深淺色設定 | v0.1.0 |
| `giowidget` | 無邊框、可固定在最上層、可拖曳的小視窗 | 規劃中 |

## 安裝

```bash
go get github.com/The78estHerry/go-gio-kit@latest
```

需要 Gio v0.10.3。

## 使用範例

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
    gw.HandleEvent(e) // 每個事件都要交給 giowin
    switch e := e.(type) {
    case app.FrameEvent:
        gtx := app.NewContext(&ops, e)
        form.Layout(gtx)
        if v, ok := form.Submitted(); ok {
            save(v.Text("host"), v.Text("user"), v.Secret("token").Bytes())
            v.Wipe() // 用完立刻清除
            form.Reset()
        }
        e.Frame(gtx.Ops)
    }
}
```

完整的範例（也是手動測試用的程式）在 [`gioform/example`](gioform/example/main.go)：

```bash
go run ./gioform/example
```

介面文字預設是英文，替換 `form.Labels`（型別 `gioform.Labels`）就能翻譯。

## 安全設計

- **Secret 欄位不能複製**：Gio 的編輯框就算開啟遮罩，Ctrl+C／Ctrl+X 仍會把原文放進剪貼簿；gioform 在編輯框處理之前把這兩個按鍵攔下來。貼上照常可用。
- **只在按住時顯示**：Secret 欄的 Show 按鈕放開就恢復遮罩。
- **取值是可清除的位元組**：`SecretValue.Wipe` 會把內容清成 0。Go 的字串無法清除，而 Gio 的編輯框內部用字串保存文字，所以只能盡量清除；機密請存在系統保管庫，不要長時間留在記憶體裡。
- **沒有任何會輸出機密的 API**：套件只把值交給呼叫端，不記錄也不印出。
- **新密碼可以用任何 Unicode**：強度以字元數計算，一句中文也可以是強密碼。確認欄在打字時即時比對。

## giowin 修正的問題

- **Windows 日文輸入法**：Gio v0.10.3 把 `GCS_DELTASTART`（變動的起點）當成組字起點，羅馬拼音的 `chi`、`tsu`、`kya`、`kitte` 等幾乎打不出來（注音不受影響）。giowin 在訊息交給 Gio 之前清掉這個旗標。上游修好後會移除這個修正。
- **深色標題列**：Windows 的標題列由系統繪製，`SetDarkTitleBar` 讓它跟著介面切換深淺色。
- **系統配色**：`SystemPrefersDark` 讀取 Windows、macOS 與 Linux（GNOME）的深淺色設定。

## 平台支援

目前只在 Windows 上實機測試；macOS 與 Linux 的程式已實作，但還沒有實機驗證，歡迎回報結果。

| 套件 | Windows | macOS | Linux |
|---|---|---|---|
| `gioform` | 實機測試 | 未測試 | 未測試 |
| `giowin` | 實機測試 | 已實作、未測試（只有讀取系統配色） | 已實作、未測試（只有讀取系統配色） |

## 授權

[MIT](LICENSE)。本軟體依現狀提供，不提供任何保證；使用造成的任何損害，作者不負責任。

如果這個工具對你有幫助，歡迎透過 [PayPal](https://paypal.me/The78ester) 或 [Ko-fi](https://ko-fi.com/the78ester) 贊助支持開發。
