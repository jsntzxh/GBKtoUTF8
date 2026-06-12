package main

import (
	"image"
	"image/color"
	"log"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

const (
	windowWidth  = unit.Dp(480)
	windowHeight = unit.Dp(520)
	cardRadius   = 8
)

var (
	bgDark  = rgb(0x0F1117)
	bgCard  = rgb(0x161B22)
	border  = rgb(0x30363D)
	green   = rgb(0x3FB950)
	red     = rgb(0xF85149)
	blue    = rgb(0x58A6FF)
	gray    = rgb(0x8B949E)
	yellow  = rgb(0xD2991D)
	success = rgb(0x7EE787)
	errFg   = rgb(0xF0883E)
)

func rgb(c uint32) color.NRGBA {
	return color.NRGBA{
		R: uint8((c >> 16) & 0xFF),
		G: uint8((c >> 8) & 0xFF),
		B: uint8(c & 0xFF),
		A: 255,
	}
}

type applyBtnTag struct{}
type langTag struct{ locale lang }

type uiState struct {
	enabled bool
	isAdmin bool

	lang       lang
	langZhTag  langTag
	langEnTag  langTag
	invalidate func()

	toggleSwitch widget.Bool
	applyTag     applyBtnTag

	msgText string
	msgType string

	pendingClick bool
}

func newUIState() *uiState {
	enabled, err := isUTF8Enabled()
	if err != nil {
		log.Printf("WARNING: could not read UTF-8 status: %v", err)
		enabled = false
	}

	s := &uiState{
		enabled:    enabled,
		isAdmin:    isAdmin(),
		lang:       currentLang,
		langZhTag:  langTag{locale: langZH},
		langEnTag:  langTag{locale: langEN},
	}
	s.toggleSwitch.Value = enabled
	return s
}

func (s *uiState) applyToggle() {
	if !s.isAdmin {
		s.msgText = T(s.lang, "result.no_admin")
		s.msgType = "error"
		return
	}

	newState := s.toggleSwitch.Value
	if newState == s.enabled {
		if s.enabled {
			s.msgText = T(s.lang, "result.already_enabled")
		} else {
			s.msgText = T(s.lang, "result.already_disabled")
		}
		s.msgType = ""
		return
	}

	result := toggleUTF8(newState)
	if result.MsgKey != "" {
		s.msgText = T(s.lang, result.MsgKey, result.MsgArgs...)
	} else {
		s.msgText = result.Message
	}
	if result.IsError {
		s.msgType = "error"
		s.toggleSwitch.Value = s.enabled
		return
	}
	if !result.OK {
		// Non-error but didn't toggle (e.g., already in desired state)
		s.msgType = ""
		return
	}

	s.enabled = newState
	s.msgType = "success"
}

func runUI() {
	go func() {
		var w app.Window
		w.Option(
			app.Title(T(currentLang, "header.title")),
			app.Size(windowWidth, windowHeight),
		)

		th := material.NewTheme()
		th.Bg = bgDark
		th.Fg = rgb(0xE1E4E8)

		state := newUIState()
		state.invalidate = w.Invalidate
		var ops op.Ops

		for {
			e := w.Event()
			switch e := e.(type) {
			case app.FrameEvent:
				gtx := app.NewContext(&ops, e)
				// Clear ops each frame
				ops.Reset()

				// Layout the UI — this processes widget events
				layoutUI(gtx, th, state)

				// Finalize the frame
				e.Frame(gtx.Ops)

				// Handle button click (set during layout in applyBtn)
				if state.pendingClick {
					state.pendingClick = false
					state.applyToggle()
					w.Invalidate()
				}
			case app.DestroyEvent:
				return
			}
		}
	}()
	app.Main()
}

func layoutUI(gtx layout.Context, th *material.Theme, s *uiState) layout.Dimensions {
	paint.Fill(gtx.Ops, bgDark)

	return layout.UniformInset(unit.Dp(18)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions { return header(gtx, th) }),
			layout.Rigid(vSpacer(8)),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions { return langSwitcher(gtx, th, s) }),
			layout.Rigid(vSpacer(8)),
			layout.Rigid(hSeparator),
			layout.Rigid(vSpacer(14)),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions { return statusRow(gtx, th, s) }),
			layout.Rigid(vSpacer(12)),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions { return toggleRow(gtx, th, s) }),
			layout.Rigid(vSpacer(10)),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions { return applyBtn(gtx, th, s) }),
			layout.Rigid(vSpacer(8)),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions { return msgBox(gtx, th, s) }),
			layout.Rigid(vSpacer(10)),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions { return adminRow(gtx, th, s) }),
			layout.Rigid(vSpacer(14)),
			layout.Rigid(hSeparator),
			layout.Rigid(vSpacer(10)),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions { return infoSection(gtx, th, s) }),
		)
	})
}

func vSpacer(dp int) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Spacer{Height: unit.Dp(dp)}.Layout(gtx)
	}
}

func hSeparator(gtx layout.Context) layout.Dimensions {
	s := image.Pt(gtx.Constraints.Max.X, gtx.Dp(1))
	defer clip.Rect{Max: s}.Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, border)
	return layout.Dimensions{Size: s}
}

// ---- Header ----

func header(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			l := material.Body1(th, T(currentLang, "header.title"))
			l.Font.Weight = font.Bold
			l.Alignment = text.Middle
			l.TextSize = th.TextSize * 1.2
			return l.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			l := material.Caption(th, T(currentLang, "header.subtitle"))
			l.Alignment = text.Middle
			l.Color = gray
			return l.Layout(gtx)
		}),
	)
}

// ---- Language Switcher ----

func langSwitcher(gtx layout.Context, th *material.Theme, s *uiState) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return langLabel(gtx, th, s, &s.langZhTag)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				l := material.Body1(th, " | ")
				l.Color = gray
				return l.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return langLabel(gtx, th, s, &s.langEnTag)
			}),
		)
	})
}

func langLabel(gtx layout.Context, th *material.Theme, s *uiState, tag *langTag) layout.Dimensions {
	// Determine the label text (language's own name for itself)
	var labelText string
	if tag.locale == langZH {
		labelText = T(langZH, "lang.zh_label")
	} else {
		labelText = T(langEN, "lang.en_label")
	}

	macro := op.Record(gtx.Ops)
	l := material.Body1(th, labelText)
	if s.lang == tag.locale {
		l.Font.Weight = font.Bold
		l.Color = rgb(0xE1E4E8)
	} else {
		l.Color = gray
	}
	lblDims := l.Layout(gtx)
	labelOps := macro.Stop()

	btnW := lblDims.Size.X + gtx.Dp(8)
	btnH := lblDims.Size.Y + gtx.Dp(4)
	btnSize := image.Pt(btnW, btnH)

	defer clip.Rect{Max: btnSize}.Push(gtx.Ops).Pop()
	event.Op(gtx.Ops, tag)

	for {
		ev, ok := gtx.Event(pointer.Filter{
			Target: tag,
			Kinds:  pointer.Press | pointer.Release,
		})
		if !ok {
			break
		}
		if e, ok := ev.(pointer.Event); ok {
			if e.Kind == pointer.Release {
				if s.lang != tag.locale {
					s.lang = tag.locale
					currentLang = tag.locale
					s.invalidate()
				}
			}
		}
	}

	offsetX := (btnW - lblDims.Size.X) / 2
	offsetY := (btnH - lblDims.Size.Y) / 2
	offStack := op.Offset(image.Pt(offsetX, offsetY)).Push(gtx.Ops)
	labelOps.Add(gtx.Ops)
	offStack.Pop()

	return layout.Dimensions{Size: btnSize}
}

// ---- Status Row ----

func statusRow(gtx layout.Context, th *material.Theme, s *uiState) layout.Dimensions {
	return card(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return statusDot(gtx, s.enabled)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				txt := T(s.lang, "status.disabled")
				if s.enabled {
					txt = T(s.lang, "status.enabled")
				}
				return material.Body1(th, txt).Layout(gtx)
			}),
		)
	})
}

func statusDot(gtx layout.Context, enabled bool) layout.Dimensions {
	dotSize := gtx.Dp(14)
	dotColor := red
	if enabled {
		dotColor = green
	}
	defer clip.Ellipse(image.Rect(0, 0, dotSize, dotSize)).Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, dotColor)
	return layout.Dimensions{Size: image.Pt(dotSize, dotSize)}
}

// ---- Card helper ----

type cardOpt func(*cardConfig)

type cardConfig struct {
	bgOverride *color.NRGBA
}

func withBg(c color.NRGBA) cardOpt {
	return func(cfg *cardConfig) { cfg.bgOverride = &c }
}

func card(gtx layout.Context, w layout.Widget, opts ...cardOpt) layout.Dimensions {
	cfg := cardConfig{}
	for _, o := range opts {
		o(&cfg)
	}

	bgColor := bgCard
	if cfg.bgOverride != nil {
		bgColor = *cfg.bgOverride
	}

	macro := op.Record(gtx.Ops)
	dims := layout.UniformInset(unit.Dp(14)).Layout(gtx, w)
	call := macro.Stop()

	r := gtx.Dp(cardRadius)
	defer clip.UniformRRect(image.Rectangle{Max: dims.Size}, r).Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, bgColor)

	call.Add(gtx.Ops)
	return dims
}

// ---- Toggle Row ----

func toggleRow(gtx layout.Context, th *material.Theme, s *uiState) layout.Dimensions {
	return card(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.Body1(th, T(s.lang, "toggle.label")).Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						l := material.Caption(th, T(s.lang, "toggle.desc"))
						l.Color = gray
						return l.Layout(gtx)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.Switch(th, &s.toggleSwitch, "").Layout(gtx)
			}),
		)
	})
}

// ---- Apply Button ----

func applyBtn(gtx layout.Context, th *material.Theme, s *uiState) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Measure label (recorded so it doesn't draw at wrong position)
		macro := op.Record(gtx.Ops)
		lblDims := material.Body1(th, T(s.lang, "button.apply")).Layout(gtx)
		labelOps := macro.Stop()

		btnW := lblDims.Size.X + gtx.Dp(24)
		btnH := lblDims.Size.Y + gtx.Dp(12)
		btnSize := image.Pt(btnW, btnH)

		// Clip the button area (for both drawing and input)
		defer clip.Rect{Max: btnSize}.Push(gtx.Ops).Pop()

		// Register tag at current clip area for event routing (Gio v0.10.0 API)
		event.Op(gtx.Ops, &s.applyTag)

		// Process click events from previous frame
		for {
			ev, ok := gtx.Event(pointer.Filter{
				Target: &s.applyTag,
				Kinds:  pointer.Press | pointer.Release,
			})
			if !ok {
				break
			}
			if e, ok := ev.(pointer.Event); ok {
				if e.Kind == pointer.Release {
					s.pendingClick = true
				}
			}
		}

		// Draw button background
		paint.Fill(gtx.Ops, blue)

		// Manually offset to center the label within the button
		offsetX := (btnW - lblDims.Size.X) / 2
		offsetY := (btnH - lblDims.Size.Y) / 2
		offStack := op.Offset(image.Pt(offsetX, offsetY)).Push(gtx.Ops)
		labelOps.Add(gtx.Ops)
		offStack.Pop()

		return layout.Dimensions{Size: btnSize}
	})
}

// ---- Message Box ----

func msgBox(gtx layout.Context, th *material.Theme, s *uiState) layout.Dimensions {
	if s.msgText == "" {
		return layout.Dimensions{}
	}

	var bgCol, fgCol color.NRGBA
	switch s.msgType {
	case "success":
		bgCol = rgb(0x033012)
		fgCol = success
	case "error":
		bgCol = rgb(0x3D100C)
		fgCol = errFg
	default:
		bgCol = bgCard
		fgCol = gray
	}

	return card(gtx, func(gtx layout.Context) layout.Dimensions {
		l := material.Body2(th, s.msgText)
		l.Color = fgCol
		return l.Layout(gtx)
	}, withBg(bgCol))
}

// ---- Admin Status ----

func adminRow(gtx layout.Context, th *material.Theme, s *uiState) layout.Dimensions {
	var bgCol, fgCol color.NRGBA
	var txt string
	if s.isAdmin {
		bgCol = rgb(0x033012)
		fgCol = success
		txt = T(s.lang, "admin.active")
	} else {
		bgCol = rgb(0x332805)
		fgCol = yellow
		txt = T(s.lang, "admin.required")
	}

	return card(gtx, func(gtx layout.Context) layout.Dimensions {
		l := material.Body2(th, txt)
		l.Color = fgCol
		return l.Layout(gtx)
	}, withBg(bgCol))
}

// ---- Info Section ----

func infoSection(gtx layout.Context, th *material.Theme, s *uiState) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			l := material.Body2(th, T(s.lang, "info.how_it_works"))
			l.Font.Weight = font.Bold
			return l.Layout(gtx)
		}),
		layout.Rigid(vSpacer(6)),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return bullet(gtx, th, T(s.lang, "info.bullet_toggle"))
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return bullet(gtx, th, T(s.lang, "info.bullet_broadcast"))
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return bullet(gtx, th, T(s.lang, "info.bullet_immediate"))
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return bullet(gtx, th, T(s.lang, "info.bullet_no_restart"))
			}),
	)
}

func bullet(gtx layout.Context, th *material.Theme, txt string) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Baseline}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			l := material.Caption(th, "\u2022")
			l.Color = blue
			return l.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(6)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			l := material.Caption(th, txt)
			l.Color = gray
			return l.Layout(gtx)
		}),
	)
}
