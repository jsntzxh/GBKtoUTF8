package main

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const (
	regRoot      = registry.LOCAL_MACHINE
	regPath      = `SYSTEM\CurrentControlSet\Control\Nls\CodePage`
	backupRoot   = registry.CURRENT_USER
	backupPath   = `Software\UTF8Switcher`
	utf8CodePage = "65001"

	hwndBroadcast   = uintptr(0xFFFF)
	wmSettingChange = uint32(0x001A)
	smtoAbortIfHung = uint32(0x0002)

	// Default code pages for Chinese (Simplified) Windows
	defaultACP   = "936"
	defaultOEMCP = "437"
	defaultMACCP = "10008"

	// KEY_WOW64_64KEY (0x0100) ensures the 64-bit registry view is used.
	// Without this, a 32-bit binary writes to the 32-bit view and the
	// actual system setting is never changed.
	_reg64 = 0x0100

	regReadAccess  = registry.QUERY_VALUE | _reg64
	regWriteAccess = registry.WRITE | _reg64
)

var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	procSendMessageTimeoutW = user32.NewProc("SendMessageTimeoutW")
)

// ToggleResult carries the outcome of a toggle operation back to the UI.
type ToggleResult struct {
	OK      bool
	Message string
	IsError bool
	MsgKey  string
	MsgArgs []any
}

type codePageValues struct {
	ACP   string
	OEMCP string
	MACCP string
}

// isAdmin checks whether the current process has administrator privileges.
func isAdmin() bool {
	token := windows.Token(0)
	err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token)
	if err != nil {
		return false
	}
	defer token.Close()

	return token.IsElevated()
}

// isUTF8Enabled reads the ACP registry value and checks if it's set to 65001.
func isUTF8Enabled() (bool, error) {
	k, err := registry.OpenKey(regRoot, regPath, regReadAccess)
	if err != nil {
		return false, fmt.Errorf("failed to open registry key: %w", err)
	}
	defer k.Close()

	acp, _, err := k.GetStringValue("ACP")
	if err != nil {
		return false, fmt.Errorf("failed to read ACP value: %w", err)
	}

	return acp == utf8CodePage, nil
}

// getCurrentCodePages reads the current ACP, OEMCP, and MACCP values from the registry.
func getCurrentCodePages() (*codePageValues, error) {
	k, err := registry.OpenKey(regRoot, regPath, regReadAccess)
	if err != nil {
		return nil, fmt.Errorf("failed to open registry key: %w", err)
	}
	defer k.Close()

	acp, _, err := k.GetStringValue("ACP")
	if err != nil {
		return nil, fmt.Errorf("failed to read ACP: %w", err)
	}

	oemcp, _, err := k.GetStringValue("OEMCP")
	if err != nil {
		oemcp = defaultOEMCP
	}

	maccp, _, err := k.GetStringValue("MACCP")
	if err != nil {
		maccp = defaultMACCP
	}

	return &codePageValues{ACP: acp, OEMCP: oemcp, MACCP: maccp}, nil
}

// saveOriginalValues backs up the current code page values before enabling UTF-8.
func saveOriginalValues(vals *codePageValues) error {
	k, _, err := registry.CreateKey(backupRoot, backupPath, regWriteAccess)
	if err != nil {
		return fmt.Errorf("failed to create backup key: %w", err)
	}
	defer k.Close()

	if err := k.SetStringValue("ACP", vals.ACP); err != nil {
		return fmt.Errorf("failed to backup ACP: %w", err)
	}
	if err := k.SetStringValue("OEMCP", vals.OEMCP); err != nil {
		return fmt.Errorf("failed to backup OEMCP: %w", err)
	}
	if err := k.SetStringValue("MACCP", vals.MACCP); err != nil {
		return fmt.Errorf("failed to backup MACCP: %w", err)
	}
	return nil
}

// getOriginalValues reads previously saved code page values from the backup location.
// Returns defaults if no backup exists.
func getOriginalValues() *codePageValues {
	k, err := registry.OpenKey(backupRoot, backupPath, regReadAccess)
	if err != nil {
		return &codePageValues{ACP: defaultACP, OEMCP: defaultOEMCP, MACCP: defaultMACCP}
	}
	defer k.Close()

	acp, _, err := k.GetStringValue("ACP")
	if err != nil {
		acp = defaultACP
	}
	oemcp, _, err := k.GetStringValue("OEMCP")
	if err != nil {
		oemcp = defaultOEMCP
	}
	maccp, _, err := k.GetStringValue("MACCP")
	if err != nil {
		maccp = defaultMACCP
	}

	return &codePageValues{ACP: acp, OEMCP: oemcp, MACCP: maccp}
}

// setCodePages writes the given values to the Nls\CodePage registry key.
func setCodePages(vals *codePageValues) error {
	k, err := registry.OpenKey(regRoot, regPath, regWriteAccess)
	if err != nil {
		return fmt.Errorf("open registry key for writing: %w", err)
	}
	defer k.Close()

	if err := k.SetStringValue("ACP", vals.ACP); err != nil {
		return fmt.Errorf("set ACP: %w", err)
	}
	if err := k.SetStringValue("OEMCP", vals.OEMCP); err != nil {
		return fmt.Errorf("set OEMCP: %w", err)
	}
	if err := k.SetStringValue("MACCP", vals.MACCP); err != nil {
		return fmt.Errorf("set MACCP: %w", err)
	}
	return nil
}

// broadcastSettingChange sends WM_SETTINGCHANGE to all top-level windows,
// notifying them that the system locale/code page configuration has changed.
// This allows the change to take effect without a system restart.
func broadcastSettingChange() error {
	envStr, _ := syscall.UTF16PtrFromString("intl")
	ret, _, lastErr := procSendMessageTimeoutW.Call(
		hwndBroadcast,
		uintptr(wmSettingChange),
		0,
		uintptr(unsafe.Pointer(envStr)),
		uintptr(smtoAbortIfHung),
		5000, // 5 second timeout
		0,
	)
	if ret == 0 {
		if lastErr != nil {
			return fmt.Errorf("SendMessageTimeout failed: %w", lastErr)
		}
		return fmt.Errorf("SendMessageTimeout failed (unknown error)")
	}
	return nil
}

// enableUTF8 backs up current values and sets all code pages to UTF-8 (65001).
// Returns a ToggleResult with a human-readable outcome.
func enableUTF8() ToggleResult {
	current, err := getCurrentCodePages()
	if err != nil {
		return ToggleResult{OK: false, IsError: true, Message: fmt.Sprintf("Read current values: %v", err), MsgKey: "op.read_error", MsgArgs: []any{err}}
	}

	// If already UTF-8, nothing to do.
	if current.ACP == utf8CodePage {
		return ToggleResult{OK: true, Message: "UTF-8 is already enabled.", MsgKey: "op.already_enabled"}
	}

	if err := saveOriginalValues(current); err != nil {
		return ToggleResult{OK: false, IsError: true, Message: fmt.Sprintf("Save backup: %v", err), MsgKey: "op.save_error", MsgArgs: []any{err}}
	}

	utf8Vals := &codePageValues{ACP: utf8CodePage, OEMCP: utf8CodePage, MACCP: utf8CodePage}
	if err := setCodePages(utf8Vals); err != nil {
		return ToggleResult{OK: false, IsError: true, Message: fmt.Sprintf("Write registry: %v", err), MsgKey: "op.write_error", MsgArgs: []any{err}}
	}

	// Verify the write took effect.
	confirm, err := getCurrentCodePages()
	if err != nil {
		return ToggleResult{OK: false, IsError: true, Message: fmt.Sprintf("Verify write: %v", err), MsgKey: "op.verify_error", MsgArgs: []any{err}}
	}
	if confirm.ACP != utf8CodePage {
		return ToggleResult{OK: false, IsError: true, Message: fmt.Sprintf(
			"Registry write did not take effect. ACP is still %s (expected %s).\n"+
				"Check: is the binary 32-bit? Use 'go env GOARCH' to verify.",
			confirm.ACP, utf8CodePage), MsgKey: "op.write_no_effect", MsgArgs: []any{confirm.ACP, utf8CodePage}}
	}

	// Broadcast — a failure here does NOT roll back the registry change.
	_ = broadcastSettingChange()

	return ToggleResult{OK: true, Message: "UTF-8 has been enabled. New programs use UTF-8 immediately.", MsgKey: "op.enabled_ok"}
}

// disableUTF8 restores the original code page values from backup.
func disableUTF8() ToggleResult {
	original := getOriginalValues()

	// If already at original values, nothing to do.
	current, _ := getCurrentCodePages()
	if current != nil && current.ACP == original.ACP {
		return ToggleResult{OK: true, Message: "UTF-8 is already disabled.", MsgKey: "op.already_disabled"}
	}

	if err := setCodePages(original); err != nil {
		return ToggleResult{OK: false, IsError: true, Message: fmt.Sprintf("Write registry: %v", err), MsgKey: "op.write_error", MsgArgs: []any{err}}
	}

	// Verify the write took effect.
	confirm, err := getCurrentCodePages()
	if err != nil {
		return ToggleResult{OK: false, IsError: true, Message: fmt.Sprintf("Verify write: %v", err), MsgKey: "op.verify_error", MsgArgs: []any{err}}
	}
	if confirm.ACP != original.ACP {
		return ToggleResult{OK: false, IsError: true, Message: fmt.Sprintf(
			"Registry write did not take effect. ACP is still %s (expected %s).",
			confirm.ACP, original.ACP), MsgKey: "op.write_no_effect_short", MsgArgs: []any{confirm.ACP, original.ACP}}
	}

	// Broadcast — failure here is non-fatal.
	_ = broadcastSettingChange()

	return ToggleResult{OK: true, Message: "UTF-8 has been disabled. Original code page restored.", MsgKey: "op.disabled_ok"}
}

// toggleUTF8 switches between UTF-8 and original code pages.
func toggleUTF8(enable bool) ToggleResult {
	if enable {
		return enableUTF8()
	}
	return disableUTF8()
}
