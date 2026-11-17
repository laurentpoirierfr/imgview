package ui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/laurentpoirierfr/imgview/internal/menu"
)

// ─── Souris ──────────────────────────────────────────────────────────────────

type span struct{ x, w int }

const brandText = iconImage + " imgview"

// menuSpans renvoie la position horizontale de chaque titre de menu.
func (m *Model) menuSpans() []span {
	x := lipgloss.Width(m.st.brand.Render(brandText)) + 1
	spans := make([]span, len(m.menus))
	for i, mn := range m.menus {
		w := lipgloss.Width(mn.Title) + 2
		spans[i] = span{x, w}
		x += w
	}
	return spans
}

// dropdownRect renvoie la position et la taille du menu déroulant ouvert.
func (m *Model) dropdownRect() (x, y, w, h int) {
	box := m.viewDropdown()
	return m.menuSpans()[m.menuOpen].x, menuH, lipgloss.Width(box), lipgloss.Height(box)
}

func (m *Model) handleClick(x, y int) tea.Cmd {
	if m.modal != modalNone {
		m.modal = modalNone
		return nil
	}

	// Clic dans le menu déroulant.
	if m.menuOpen >= 0 {
		dx, dy, dw, dh := m.dropdownRect()
		if x >= dx && x < dx+dw && y >= dy && y < dy+dh {
			items := m.menus[m.menuOpen].Items
			if i := y - dy - 1; i >= 0 && i < len(items) && !items[i].Sep {
				m.closeMenu()
				return m.do(items[i].Act)
			}
			return nil
		}
	}

	// Clic dans la barre de menus.
	if y < menuH {
		for i, sp := range m.menuSpans() {
			if x >= sp.x && x < sp.x+sp.w {
				if m.menuOpen == i {
					m.closeMenu()
				} else {
					m.openMenu(i)
				}
				return nil
			}
		}
		m.closeMenu()
		return nil
	}
	m.closeMenu()

	l := m.layout()
	if y < l.bodyY || y >= l.bodyY+l.bodyH {
		return nil
	}

	// Clic dans l'arbre : sélectionner, ou déplier/replier si déjà sélectionné.
	if x < l.treeW {
		m.focus, m.preview = focusTree, false
		row := y - l.bodyY - 1
		if row < 0 || row >= l.treeIH {
			return nil
		}
		if idx := m.tree.Offset + row; idx < len(m.tree.Rows) {
			if idx == m.tree.Cursor {
				m.tree.Toggle()
			} else {
				m.tree.Cursor = idx
			}
		}
		return nil
	}

	// Clic dans la galerie : sélectionner, ou ouvrir l'aperçu si déjà sélectionnée.
	m.focus = focusGallery
	if m.preview {
		return nil
	}
	gl := m.gridLayout()
	rx, ry := x-l.galX-2, y-l.bodyY-2
	if rx < 0 || ry < 0 || rx%(gl.CellW+1) == gl.CellW {
		return nil
	}
	col, row := rx/(gl.CellW+1), ry/gl.CellH
	if col >= gl.Cols || row >= gl.Rows {
		return nil
	}
	if idx := (m.gal.RowOffset+row)*gl.Cols + col; idx < len(m.gal.Files) {
		if idx == m.gal.Cursor {
			return m.do(menu.ActPreview)
		}
		m.gal.Cursor = idx
	}
	return nil
}

func (m *Model) handleWheel(x, d int) {
	if m.modal != modalNone || m.menuOpen >= 0 {
		return
	}
	switch l := m.layout(); {
	case x < l.treeW:
		m.tree.Move(3 * d)
	case m.preview:
		m.gal.Move(d)
	default:
		m.gal.Move(d * m.gridLayout().Cols)
	}
}
