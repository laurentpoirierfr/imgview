package ui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/laurentpoirierfr/imgview/internal/menu"
)

// ─── Clavier ─────────────────────────────────────────────────────────────────

func (m *Model) handleKey(k tea.KeyPressMsg) tea.Cmd {
	s := k.String()
	if s == "ctrl+c" {
		return m.do(menu.ActQuit)
	}

	if m.modal != modalNone {
		switch s {
		case "esc", "q", "enter", "?", "space", " ":
			m.modal = modalNone
		}
		return nil
	}

	if m.menuOpen >= 0 {
		return m.menuKey(s)
	}

	switch s {
	case "f10", "alt+f":
		m.openMenu(0)
		return nil
	case "alt+a":
		m.openMenu(1)
		return nil
	case "alt+h":
		m.openMenu(2)
		return nil
	case "q":
		return m.do(menu.ActQuit)
	case "?":
		return m.do(menu.ActHelp)
	case "f5":
		return m.do(menu.ActRefresh)
	case "+", "=":
		return m.do(menu.ActZoomIn)
	case "-":
		return m.do(menu.ActZoomOut)
	case ".":
		return m.do(menu.ActToggleHidden)
	case "r":
		return m.do(menu.ActToggleRenderer)
	case "f":
		return m.do(menu.ActFullscreen)
	case "o":
		return m.do(menu.ActOpenExternal)
	case "d":
		return m.do(menu.ActToggleHD)
	case "tab", "shift+tab":
		if !m.preview {
			if m.focus == focusTree {
				m.focus = focusGallery
			} else {
				m.focus = focusTree
			}
		}
		return nil
	}

	switch {
	case m.preview:
		m.previewKey(s)
	case m.focus == focusTree:
		m.treeKey(s)
	default:
		return m.galleryKey(s)
	}
	return nil
}

func (m *Model) treeKey(s string) {
	t := m.tree
	switch s {
	case "up", "k":
		t.Move(-1)
	case "down", "j":
		t.Move(1)
	case "pgup":
		t.Move(-m.layout().treeIH)
	case "pgdown":
		t.Move(m.layout().treeIH)
	case "home", "g":
		t.Cursor = 0
	case "end", "G":
		t.Cursor = len(t.Rows) - 1
	case "right", "l":
		t.Right()
	case "left", "h":
		t.Left()
	case "enter", "space", " ":
		t.Toggle()
	}
}

func (m *Model) galleryKey(s string) tea.Cmd {
	g, gl := m.gal, m.gridLayout()
	switch s {
	case "left", "h":
		g.Move(-1)
	case "right", "l":
		g.Move(1)
	case "up", "k":
		if g.Cursor >= gl.Cols {
			g.Move(-gl.Cols)
		}
	case "down", "j":
		if g.Cursor+gl.Cols < len(g.Files) {
			g.Move(gl.Cols)
		}
	case "pgup":
		g.Move(-gl.Cols * gl.Rows)
	case "pgdown":
		g.Move(gl.Cols * gl.Rows)
	case "home", "g":
		g.Cursor = 0
	case "end", "G":
		g.Move(len(g.Files))
	case "enter", "space", " ":
		return m.do(menu.ActPreview)
	case "esc", "backspace":
		m.focus = focusTree
	}
	return nil
}

func (m *Model) previewKey(s string) {
	g := m.gal
	switch s {
	case "esc", "backspace", "enter":
		m.preview = false
	case "left", "h", "up", "k", "pgup", "p":
		g.Move(-1)
	case "right", "l", "down", "j", "pgdown", "n", "space", " ":
		g.Move(1)
	case "home", "g":
		g.Cursor = 0
	case "end", "G":
		g.Move(len(g.Files))
	}
}

func (m *Model) openMenu(i int) {
	n := len(m.menus)
	m.menuOpen = ((i % n) + n) % n
	m.menuCur = m.menus[m.menuOpen].NextItem(-1, 1)
}

func (m *Model) closeMenu() { m.menuOpen = -1 }

func (m *Model) menuKey(s string) tea.Cmd {
	mn := m.menus[m.menuOpen]
	switch s {
	case "esc", "f10":
		m.closeMenu()
	case "left", "h":
		m.openMenu(m.menuOpen - 1)
	case "right", "l":
		m.openMenu(m.menuOpen + 1)
	case "up", "k", "shift+tab":
		m.menuCur = mn.NextItem(m.menuCur, -1)
	case "down", "j", "tab":
		m.menuCur = mn.NextItem(m.menuCur, 1)
	case "enter", "space", " ":
		a := mn.Items[m.menuCur].Act
		m.closeMenu()
		return m.do(a)
	}
	return nil
}
