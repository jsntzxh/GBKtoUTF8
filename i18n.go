package main

import "fmt"

type lang string

const (
	langZH lang = "zh"
	langEN lang = "en"
)

var currentLang lang = langZH

var messages = map[string]map[lang]string{
	// Header
	"header.title": {
		langZH: "UTF-8 语言支持切换器",
		langEN: "UTF-8 Language Support Switcher",
	},
	"header.subtitle": {
		langZH: "Beta：使用 Unicode UTF-8 提供全球语言支持",
		langEN: "Beta: Use Unicode UTF-8 for worldwide language support",
	},

	// Language labels
	"lang.zh_label": {
		langZH: "中文",
		langEN: "中文",
	},
	"lang.en_label": {
		langZH: "EN",
		langEN: "EN",
	},

	// Status
	"status.enabled": {
		langZH: "UTF-8 已启用",
		langEN: "UTF-8 is ENABLED",
	},
	"status.disabled": {
		langZH: "UTF-8 已禁用",
		langEN: "UTF-8 is DISABLED",
	},

	// Toggle
	"toggle.label": {
		langZH: "启用 UTF-8 支持",
		langEN: "Enable UTF-8 Support",
	},
	"toggle.desc": {
		langZH: "将系统代码页设置为 UTF-8 (65001)",
		langEN: "Set system code page to UTF-8 (65001)",
	},

	// Button
	"button.apply": {
		langZH: "应用更改",
		langEN: "Apply Changes",
	},

	// Admin
	"admin.active": {
		langZH: "已获取管理员权限",
		langEN: "Administrator privileges active",
	},
	"admin.required": {
		langZH: "需要管理员权限。\n请以管理员身份重新启动以切换。",
		langEN: "Administrator privileges required.\nRestart as Administrator to toggle.",
	},

	// Info section
	"info.how_it_works": {
		langZH: "工作原理",
		langEN: "How it works",
	},
	"info.bullet_toggle": {
		langZH: "将 Windows 代码页切换为 UTF-8 (65001)",
		langEN: "Toggles Windows code page to UTF-8 (65001)",
	},
	"info.bullet_broadcast": {
		langZH: "广播系统设置更改通知",
		langEN: "Broadcasts system settings change notification",
	},
	"info.bullet_immediate": {
		langZH: "新程序立即使用 UTF-8",
		langEN: "New programs use UTF-8 immediately",
	},
	"info.bullet_no_restart": {
		langZH: "无需重启",
		langEN: "No restart is required",
	},

	// Result messages (applyToggle)
	"result.no_admin": {
		langZH: "需要管理员权限。请以管理员身份运行。",
		langEN: "Administrator privileges required. Run as Administrator.",
	},
	"result.already_enabled": {
		langZH: "UTF-8 已经启用。请先关闭开关。",
		langEN: "UTF-8 is already enabled. Toggle the switch off first.",
	},
	"result.already_disabled": {
		langZH: "UTF-8 已经禁用。请先打开开关。",
		langEN: "UTF-8 is already disabled. Toggle the switch on first.",
	},

	// Operation messages (switcher.go)
	"op.read_error": {
		langZH: "读取当前值失败: %v",
		langEN: "Read current values: %v",
	},
	"op.already_enabled": {
		langZH: "UTF-8 已经启用。",
		langEN: "UTF-8 is already enabled.",
	},
	"op.save_error": {
		langZH: "保存备份失败: %v",
		langEN: "Save backup: %v",
	},
	"op.write_error": {
		langZH: "写入注册表失败: %v",
		langEN: "Write registry: %v",
	},
	"op.verify_error": {
		langZH: "验证写入失败: %v",
		langEN: "Verify write: %v",
	},
	"op.write_no_effect": {
		langZH: "注册表写入未生效。ACP 仍为 %s（期望 %s）。\n检查：是否为 32 位程序？使用 'go env GOARCH' 验证。",
		langEN: "Registry write did not take effect. ACP is still %s (expected %s).\nCheck: is the binary 32-bit? Use 'go env GOARCH' to verify.",
	},
	"op.enabled_ok": {
		langZH: "UTF-8 已启用。新程序将立即使用 UTF-8。",
		langEN: "UTF-8 has been enabled. New programs use UTF-8 immediately.",
	},
	"op.already_disabled": {
		langZH: "UTF-8 已经禁用。",
		langEN: "UTF-8 is already disabled.",
	},
	"op.write_no_effect_short": {
		langZH: "注册表写入未生效。ACP 仍为 %s（期望 %s）。",
		langEN: "Registry write did not take effect. ACP is still %s (expected %s).",
	},
	"op.disabled_ok": {
		langZH: "UTF-8 已禁用。已恢复原始代码页。",
		langEN: "UTF-8 has been disabled. Original code page restored.",
	},

	// Main.go errors
	"err.windows_only": {
		langZH: "错误：此应用程序仅可在 Windows 上运行。",
		langEN: "Error: This application only runs on Windows.",
	},
	"err.windows_only_hint": {
		langZH: "它管理 Windows 'Beta: 使用 Unicode UTF-8 提供全球语言支持' 系统设置。",
		langEN: "It manages the Windows 'Beta: Use Unicode UTF-8' system setting.",
	},
}

func T(l lang, key string, args ...any) string {
	m, ok := messages[key]
	if !ok {
		return key
	}
	s, ok := m[l]
	if !ok {
		s = m[langEN]
	}
	if len(args) > 0 {
		return fmt.Sprintf(s, args...)
	}
	return s
}
