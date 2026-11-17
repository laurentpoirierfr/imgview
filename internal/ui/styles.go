package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// ─── Styles ──────────────────────────────────────────────────────────────────

// palette Catppuccin : Latte (fond clair) / Mocha (fond sombre).
type palette struct {
	base, mantle, surface0, surface1, overlay0 color.Color
	text, subtext, accent, blue, green, red    color.Color
}

func newPalette(isDark bool) palette {
	ld := lipgloss.LightDark(isDark)
	c := func(light, dark string) color.Color { return ld(lipgloss.Color(light), lipgloss.Color(dark)) }
	return palette{
		base:     c("#eff1f5", "#1e1e2e"),
		mantle:   c("#e6e9ef", "#181825"),
		surface0: c("#ccd0da", "#313244"),
		surface1: c("#bcc0cc", "#45475a"),
		overlay0: c("#9ca0b0", "#6c7086"),
		text:     c("#4c4f69", "#cdd6f4"),
		subtext:  c("#6c6f85", "#a6adc8"),
		accent:   c("#8839ef", "#cba6f7"),
		blue:     c("#1e66f5", "#89b4fa"),
		green:    c("#40a02b", "#a6e3a1"),
		red:      c("#d20f39", "#f38ba8"),
	}
}

type styles struct {
	p palette

	brand, menuBar, menuTitle, menuTitleOpen, menuHint lipgloss.Style
	dropdown, dropItem, dropItemSel, dropKey, dropSep  lipgloss.Style

	panel, panelFocus, header, dim, text, errText lipgloss.Style
	treeSel, treeSelBlur, treeIcon                lipgloss.Style

	cell, cellSel, cellSelBlur, name, nameSel lipgloss.Style

	statusBar, badge, badgeAlt, statusInfo, statusErr, statusOK, hints, hintKey lipgloss.Style
	modal, modalTitle, modalKey, modalText                                      lipgloss.Style
}

func newStyles(isDark bool) styles {
	p := newPalette(isDark)
	s := styles{p: p}
	ns := lipgloss.NewStyle

	s.brand = ns().Bold(true).Foreground(p.base).Background(p.accent).Padding(0, 1)
	s.menuBar = ns().Background(p.mantle).Foreground(p.text)
	s.menuTitle = ns().Background(p.mantle).Foreground(p.text).Padding(0, 1)
	s.menuTitleOpen = ns().Background(p.surface1).Foreground(p.accent).Bold(true).Padding(0, 1)
	s.menuHint = ns().Background(p.mantle).Foreground(p.overlay0)

	s.dropdown = ns().Border(lipgloss.RoundedBorder()).BorderForeground(p.accent).
		BorderBackground(p.mantle).Background(p.mantle)
	s.dropItem = ns().Background(p.mantle).Foreground(p.text)
	s.dropItemSel = ns().Background(p.accent).Foreground(p.base).Bold(true)
	s.dropKey = ns().Background(p.mantle).Foreground(p.overlay0)
	s.dropSep = ns().Background(p.mantle).Foreground(p.surface1)

	s.panel = ns().Border(lipgloss.RoundedBorder()).BorderForeground(p.surface1)
	s.panelFocus = ns().Border(lipgloss.RoundedBorder()).BorderForeground(p.accent)
	s.header = ns().Bold(true).Foreground(p.accent)
	s.dim = ns().Foreground(p.overlay0)
	s.text = ns().Foreground(p.text)
	s.errText = ns().Foreground(p.red)

	s.treeSel = ns().Background(p.accent).Foreground(p.base).Bold(true)
	s.treeSelBlur = ns().Background(p.surface0).Foreground(p.text)
	s.treeIcon = ns().Foreground(p.blue)

	s.cell = ns().Border(lipgloss.RoundedBorder()).BorderForeground(p.surface1)
	s.cellSel = ns().Border(lipgloss.ThickBorder()).BorderForeground(p.accent)
	s.cellSelBlur = ns().Border(lipgloss.RoundedBorder()).BorderForeground(p.subtext)
	s.name = ns().Foreground(p.subtext)
	s.nameSel = ns().Foreground(p.accent).Bold(true)

	s.statusBar = ns().Background(p.mantle).Foreground(p.text)
	s.badge = ns().Bold(true).Foreground(p.base).Background(p.blue).Padding(0, 1)
	s.badgeAlt = ns().Bold(true).Foreground(p.base).Background(p.accent).Padding(0, 1)
	s.statusInfo = ns().Background(p.surface0).Foreground(p.text).Padding(0, 1)
	s.statusErr = ns().Background(p.red).Foreground(p.base).Bold(true).Padding(0, 1)
	s.statusOK = ns().Background(p.green).Foreground(p.base).Padding(0, 1)
	s.hints = ns().Foreground(p.overlay0)
	s.hintKey = ns().Foreground(p.subtext).Bold(true)

	s.modal = ns().Border(lipgloss.RoundedBorder()).BorderForeground(p.accent).
		BorderBackground(p.mantle).Background(p.mantle).Foreground(p.text).Padding(1, 3)
	s.modalTitle = ns().Bold(true).Foreground(p.accent).Background(p.mantle)
	s.modalKey = ns().Bold(true).Foreground(p.blue).Background(p.mantle)
	s.modalText = ns().Foreground(p.text).Background(p.mantle)
	return s
}
