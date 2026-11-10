**繁體中文** | [English](CLAUDE.en.md) | [日本語](CLAUDE.ja.md)

# CLAUDE.md — go-gio-kit

給 Claude 的專案說明。這個 repo 從第一天起就在 GitHub 公開，內容只寫公開也沒關係的事。

本檔是主要版本，Claude Code 只會自動讀取這個檔案；`CLAUDE.en.md`（英文）與 `CLAUDE.ja.md`（日文）是翻譯，修改本檔時在同一個 commit 裡一起更新。

## 目的

依賴 Gio 的可重用套件：`gioform`（TextField、SecretField、SelectField、Form、Theme）與 `giowidget`（無邊框、可固定最上層、可拖曳的小視窗）。只依賴 `gioui.org`。

第一個使用者是 go-credential-porter；go-virtual-display 之後也改用這裡的套件。

## 設計原則

- 套件不知道「Token」或「虛擬螢幕」是什麼，任何業務文字都由呼叫端傳入。
- 不放 `internal/`，讓其他 repo 都能 import。
- 套件名稱全小寫、單一單字；Gio 相關的加 `gio` 前綴，避免和 Gio 自己的套件撞名。
- 同一個 repo 內的套件一起發佈版本，標籤是普通的 `v0.1.0`。某個套件之後要拆成獨立 module 時，在該資料夾加 go.mod，並立刻從根 module 發新版，避免 ambiguous import。
- 不得新增會輸出機密內容的 API。

## GitHub 決定

- **GitHub 決定**（2026-11-09）：GitHub 公開為主要開發位置（`github.com/The78estHerry/go-gio-kit`）；Gitea（`The78estWorld/go-gio-kit`）也保留一份，兩邊都推：remote `origin` 是 Gitea、`github` 是 GitHub，分支與 tag 兩邊都推，Release 兩邊都建（步驟見 `RELEASING.md`）。
- 已準備：MIT LICENSE（The78ester）、`.gitignore`。
- 測試與範例不放真實憑證、帳號或內部網址（用 `example.com`、`your-name`、`test-token-not-real`）。

## 平台

- Windows 免 CGO；Gio 在 macOS、Linux 需要 CGO，由 GitHub Actions 編譯。
- 目前只在 Windows 實機測試；macOS、Linux 標示「已實作、未測試」：平台專屬檔案開頭加 `// Status: implemented, not yet tested on real hardware.`。

## 語言規則

- 程式碼註解與 Godoc 一律用英文。
- PR 說明、issue 內容、程式碼審查留言一律用英文。
- commit 訊息三語：英文為主、繁體中文次之、日文為輔，格式見全域 CLAUDE.md 的「Commit message language (all projects)」。
- `README.md` 是主要版本，用繁體中文；`README.en.md`（英文）與 `README.ja.md`（日文）在同一個 commit 裡同步更新。`CLAUDE.md` 三個版本也一樣。
- 回覆開發者一律用繁體中文。

## 進度紀錄

- 2026-11-09：建立 repo 骨架（go.mod、LICENSE、.gitignore、CLAUDE.md 三語版本）。
- 2026-11-09：圖示定案（積木 B3：GO＋輸入框＋小視窗＋密碼點＋速度線，`assets/icon.svg`，`assets/make-icon.ps1` 轉 PNG）。
- 2026-11-09 18:08 當機（無 BugCheck，硬當機；Kernel-Power 41＋6008）：當時 go-credential-porter 的對話正在這裡跑 `go vet`／`go test -p 1 ./...` 與跨平台 vet，同時 icon-theme-park 在 cargo 編譯並啟動 Tauri 程式。當機前溫度 75°C、記憶體 46%，不是過熱或記憶體不足；開機後有多筆 LiveKernelEvent 141／117（顯示卡驅動程式逾時），顯示卡驅動是主要嫌疑。檢查結果：`gioform/`、`giowin/`、`go.mod`、`go.sum` 都完整（沒有截斷、gofmt 與 go vet 通過），仍未 commit。下次從：重跑 `go test -p 1 ./...`，確認後 commit gioform／giowin。
- 2026-11-10：當機後重跑 `go vet`、`go test -p 1 ./...`（Windows）與 Linux／macOS 交叉 vet（只檢查 `gioform`、`giowin`；`gioform/example` 用到 `gioui.org/app`，在 Linux 需要 CGO，不能交叉 vet），全部通過。giowin 補上視窗關閉（`WM_NCDESTROY`）時清除 subclass 紀錄。commit gioform／giowin。下一步：M3 收尾（實機跑範例、README、發佈 v0.1.0）。
- 2026-11-10：**gioform 實機測試通過**（使用者操作範例）：Secret 欄位可貼上、擋住複製／剪下、按住 Show 才顯示；注音與日文羅馬拼音（chi、tsu、kya、kitte）正常；必填與驗證、Enter／Esc、深淺色與標題列切換、中日文顯示都正常。依使用者回饋：NewSecret 的確認欄改成打字時即時比對（還沒打完不報錯，不一樣立刻紅字，一致顯示 match；新增 `Labels.Match`）。下一步：三語 README，之後第一次推上 GitHub 並發佈 v0.1.0（要使用者同意）。
- 2026-11-10：三語 README、`.github/FUNDING.yml`、`RELEASING.md`、CI（GitHub：Windows、macOS 編譯全部，Linux 只檢查 `gioform`、`giowin`；`.gitea/workflows` 只跑 Linux，避免 Windows／macOS 工作在 NAS 上一直排隊）。下一步：使用者同意後第一次推上 GitHub（repo 已建好），打 `v0.1.0` tag、兩邊建 Release。
- 2026-11-10（暫停點）：main 與 `v0.1.0` tag 已推上 GitHub（`github` remote 已加入）與 Gitea。**下次從這裡接續**：確認 `v0.1.0` 的 CI（GitHub 三平台、Gitea Linux），通過後照 `RELEASING.md` 第 6、7 步兩邊建 Release 並確認 module proxy；失敗則修正後發 `v0.1.1`（已推送的 tag 不可移動）。
- 2026-11-10：**v0.1.0 已發佈**：GitHub 三平台與 Gitea CI 通過，兩邊都有 Release，module proxy 可取得。下一步：M4 的 giowidget（從 go-virtual-display 抽出）。
