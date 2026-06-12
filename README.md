# UTF-8 语言支持切换器

> [English Version](#english)

一个 Windows GUI 工具，用于一键切换"Beta: 使用 Unicode UTF-8 提供全球语言支持"系统设置，无需重启。

---

## 目录

- [功能](#功能)
- [截图](#截图)
- [构建](#构建)
- [使用方法](#使用方法)
- [工作原理](#工作原理)
- [故障排除](#故障排除)
- [技术栈](#技术栈)

---

## 功能

- **一键切换** — 通过开关控件启用或禁用 UTF-8 系统代码页
- **即时生效** — 无需重启，更改立即广播到所有正在运行的窗口
- **自动备份** — 首次启用 UTF-8 时自动备份原始代码页设置，禁用时可恢复
- **中文/英文界面** — 运行时实时切换语言，默认中文
- **管理员检测** — 自动检测并以可视化方式提示是否以管理员身份运行
- **写入验证** — 每次修改注册表后立即验证是否生效
- **64 位安全** — 强制使用 64 位注册表视图，避免 32 位重定向陷阱
- **深色主题** — GitHub 风格深色配色方案

---

## 构建

### 前置要求

- Go 1.26+
- Windows 操作系统（构建目标平台）

### 编译

```bash
cd D:\GBKtoUTF8
go build -o utf8-switcher.exe .
```

### 静态检查

```bash
go vet ./...
```

---

## 使用方法

### 启动

1. 右键点击 `utf8-switcher.exe`
2. 选择 **"以管理员身份运行"**
3. 界面顶部显示管理员权限状态：
   - 🟢 **管理员权限已激活** — 可以正常切换
   - 🟡 **需要管理员权限，请以管理员身份重新启动** — 只读模式

### 切换 UTF-8 设置

1. 查看状态指示器确认当前 UTF-8 状态
2. 点击开关切换到想要的设置（开/关）
3. 点击 **"应用更改"** 按钮
4. 看到绿色成功消息即表示完成

### 语言切换

- 界面顶部点击 **"中文"** 或 **"EN"** 即可实时切换界面语言

### 默认代码页值

如果从未启用过 UTF-8（没有备份值），禁用时将恢复到以下默认值：

| 代码页 | 值   | 说明         |
|--------|------|-------------|
| ACP    | 936  | 简体中文 GBK  |
| OEMCP  | 437  | 美语 OEM     |
| MACCP  | 10008 | 简体中文 Mac  |

---

## 工作原理

1. **读取**注册表 `HKLM\SYSTEM\CurrentControlSet\Control\Nls\CodePage` 下的 `ACP`、`OEMCP`、`MACCP` 值
2. **启用 UTF-8 时**：先备份当前值到 `HKCU\Software\UTF8Switcher`，再将三个值全部设为 `65001`（UTF-8）
3. **禁用 UTF-8 时**：从备份中恢复原始值
4. **验证写入**：修改后立即重新读取注册表确认值已生效
5. **广播变更**：通过 `WM_SETTINGCHANGE` 消息（`"intl"` 环境字符串）通知系统所有窗口，使新程序立即使用新的代码页设置

---

## 故障排除

| 问题 | 可能原因 | 解决方案 |
|------|----------|----------|
| "需要管理员权限" | 未以管理员身份运行 | 右键 exe → 以管理员身份运行 |
| "注册表写入未生效" | 32 位二进制被重定向 | 确认使用 `GOARCH=amd64` 构建 64 位版本 |
| 非 Windows 系统无法运行 | 应用限制为 Windows | 本工具仅支持 Windows 平台 |
| 更改未立即生效 | 某些程序需手动刷新 | 重启相关应用程序 |

### 注册表路径

- **系统设置**：`HKLM\SYSTEM\CurrentControlSet\Control\Nls\CodePage`
- **备份位置**：`HKCU\Software\UTF8Switcher`
- **Windows 设置位置**：设置 → 时间和语言 → 语言和区域 → 管理语言设置 → 更改系统区域设置 → Beta: 使用 Unicode UTF-8

---

## 技术栈

| 技术 | 用途 |
|------|------|
| [Go](https://go.dev/) | 编程语言 |
| [Gio](https://gioui.org/) | 跨平台 GUI 框架 |
| `golang.org/x/sys/windows/registry` | Windows 注册表 API |
| `golang.org/x/sys/windows` | Windows 系统调用 |

---

## 项目结构

```
GBKtoUTF8/
├── main.go          # 入口，Windows 平台检查
├── switcher.go      # 核心逻辑：注册表操作、管理员检测
├── ui.go            # GUI 界面（Gio 框架）
├── i18n.go          # 中英文国际化
├── go.mod           # Go 模块定义
├── go.sum           # 依赖校验
└── README.md        # 本文档
```

---

---

<h1 id="english">UTF-8 Language Support Switcher</h1>

> [中文版](#utf-8-语言支持切换器)

A Windows GUI tool for one-click toggling of the "Beta: Use Unicode UTF-8 for worldwide language support" system setting—no reboot required.

---

## Table of Contents

- [Features](#features)
- [Build](#build)
- [Usage](#usage)
- [How It Works](#how-it-works)
- [Troubleshooting](#troubleshooting-1)
- [Tech Stack](#tech-stack)

---

## Features

- **One-Click Toggle** — Enable or disable the UTF-8 system code page with a single switch
- **Immediate Effect** — No reboot required; changes are broadcast to all running windows instantly
- **Auto Backup** — Original code page values are backed up on first enable, restored when disabling
- **Bilingual UI** — Switch between Chinese and English at runtime (Chinese by default)
- **Admin Detection** — Automatically detects and visually indicates administrator status
- **Write Verification** — Verifies registry changes take effect immediately after writing
- **64-bit Safe** — Forces 64-bit registry view to avoid 32-bit redirection pitfalls
- **Dark Theme** — GitHub-style dark color scheme

---

## Build

### Prerequisites

- Go 1.26+
- Windows OS (target platform)

### Compile

```bash
cd D:\GBKtoUTF8
go build -o utf8-switcher.exe .
```

### Static Analysis

```bash
go vet ./...
```

---

## Usage

### Launch

1. Right-click `utf8-switcher.exe`
2. Select **"Run as Administrator"**
3. The admin status is displayed at the top:
   - 🟢 **Administrator privileges active** — Ready to toggle
   - 🟡 **Administrator privileges required. Restart as Administrator to toggle.** — Read-only mode

### Toggle UTF-8 Setting

1. Check the status indicator for the current UTF-8 state
2. Click the switch to select the desired state (On/Off)
3. Click **"Apply Changes"**
4. A green success message confirms the operation

### Language Switching

- Click **"中文"** or **"EN"** at the top of the window to switch the UI language in real time

### Default Code Page Values

If UTF-8 has never been enabled (no backup exists), disabling will restore these defaults:

| Code Page | Value | Description           |
|-----------|-------|-----------------------|
| ACP       | 936   | Simplified Chinese GBK |
| OEMCP     | 437   | US English OEM         |
| MACCP     | 10008 | Simplified Chinese Mac  |

---

## How It Works

1. **Reads** the `ACP`, `OEMCP`, `MACCP` values from `HKLM\SYSTEM\CurrentControlSet\Control\Nls\CodePage`
2. **Enabling UTF-8**: backs up current values to `HKCU\Software\UTF8Switcher`, then sets all three to `65001` (UTF-8)
3. **Disabling UTF-8**: restores original values from the backup
4. **Verifies writes**: re-reads the registry immediately after modification to confirm the values took effect
5. **Broadcasts changes**: sends a `WM_SETTINGCHANGE` message with `"intl"` environment string so new programs immediately use the updated code page

---

## Troubleshooting

| Issue | Likely Cause | Solution |
|-------|-------------|----------|
| "Administrator privileges required" | Not running as admin | Right-click exe → Run as Administrator |
| "Registry write had no effect" | 32-bit binary redirected | Ensure 64-bit build with `GOARCH=amd64` |
| Won't run on non-Windows | Platform restriction | This tool is Windows-only |
| Change not immediately visible | Some apps need manual refresh | Restart the affected application |

### Registry Paths

- **System setting**: `HKLM\SYSTEM\CurrentControlSet\Control\Nls\CodePage`
- **Backup location**: `HKCU\Software\UTF8Switcher`
- **Windows Settings UI**: Settings → Time & Language → Language & Region → Administrative language settings → Change system locale → Beta: Use Unicode UTF-8

---

## Tech Stack

| Technology | Purpose |
|------------|---------|
| [Go](https://go.dev/) | Programming language |
| [Gio](https://gioui.org/) | Cross-platform GUI framework |
| `golang.org/x/sys/windows/registry` | Windows Registry API |
| `golang.org/x/sys/windows` | Windows syscalls |

---

## Project Structure

```
GBKtoUTF8/
├── main.go          # Entry point, Windows platform check
├── switcher.go      # Core logic: registry operations, admin detection
├── ui.go            # GUI interface (Gio framework)
├── i18n.go          # Chinese/English internationalization
├── go.mod           # Go module definition
├── go.sum           # Dependency checksums
└── README.md        # This document
```
