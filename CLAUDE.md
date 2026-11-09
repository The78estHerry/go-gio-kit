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

- **GitHub 決定**（2026-11-09）：GitHub 公開為主要開發位置（`github.com/The78estHerry/go-gio-kit`），Gitea 用拉取鏡像，不在鏡像上提交。
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
