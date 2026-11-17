package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/laurentpoirierfr/imgview/internal/gallery"
	"github.com/laurentpoirierfr/imgview/internal/menu"
	"github.com/laurentpoirierfr/imgview/pkg/imgrender"
)

// ─── Vue ─────────────────────────────────────────────────────────────────────

func (m *Model) View() tea.View {
	var v tea.View
	if m.w == 0 || m.h == 0 {
		v = tea.NewView("Chargement…")
	} else {
		v = tea.NewView(m.render())
	}
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	v.WindowTitle = "imgview — " + m.gal.Dir
	return v
}

func (m *Model) render() string {
	l := m.layout()
	body := lipgloss.JoinHorizontal(lipgloss.Top, m.viewTree(l), m.viewGallery(l))
	base := lipgloss.JoinVertical(lipgloss.Left, m.viewMenuBar(), body, m.viewStatus(), m.viewHints())

	layers := []*lipgloss.Layer{lipgloss.NewLayer(base)}
	if m.menuOpen >= 0 {
		x, y, _, _ := m.dropdownRect()
		layers = append(layers, lipgloss.NewLayer(m.viewDropdown()).X(x).Y(y).Z(1))
	}
	if m.modal != modalNone {
		box := m.viewModal()
		x := max(0, (m.w-lipgloss.Width(box))/2)
		y := max(0, (m.h-lipgloss.Height(box))/2)
		layers = append(layers, lipgloss.NewLayer(box).X(x).Y(y).Z(2))
	}
	if len(layers) == 1 {
		return base
	}
	return lipgloss.NewCompositor(layers...).Render()
}

// fill complète une ligne avec des espaces stylés jusqu'à la largeur w.
func fill(st lipgloss.Style, s string, w int) string {
	if gap := w - lipgloss.Width(s); gap > 0 {
		return s + st.Render(strings.Repeat(" ", gap))
	}
	return ansi.Truncate(s, w, "…")
}

func (m *Model) viewMenuBar() string {
	var b strings.Builder
	b.WriteString(m.st.brand.Render(brandText))
	b.WriteString(m.st.menuBar.Render(" "))
	for i, mn := range m.menus {
		if i == m.menuOpen {
			b.WriteString(m.st.menuTitleOpen.Render(mn.Title))
		} else {
			b.WriteString(m.st.menuTitle.Render(mn.Title))
		}
	}
	left := b.String()
	right := m.st.menuHint.Render("F10 menu ")
	return fill(m.st.menuBar, left, m.w-lipgloss.Width(right)) + right
}

func (m *Model) checked(a menu.Action) (bool, bool) {
	switch a {
	case menu.ActToggleHidden:
		return m.tree.ShowHidden, true
	case menu.ActToggleHD:
		return m.hd, true
	case menu.ActToggleRenderer:
		return m.kind == imgrender.Chafa, true
	}
	return false, false
}

func (m *Model) viewDropdown() string {
	mn := m.menus[m.menuOpen]
	lw, kw := 0, 0
	for _, it := range mn.Items {
		lw = max(lw, lipgloss.Width(it.Label))
		kw = max(kw, lipgloss.Width(it.Key))
	}
	inner := 1 + 2 + lw + 3 + kw + 1

	lines := make([]string, len(mn.Items))
	for i, it := range mn.Items {
		if it.Sep {
			lines[i] = m.st.dropSep.Render(strings.Repeat("─", inner))
			continue
		}
		mark := "  "
		if on, toggle := m.checked(it.Act); toggle && on {
			mark = "✓ "
		}
		left := " " + mark + it.Label
		pad := strings.Repeat(" ", max(1, inner-lipgloss.Width(left)-lipgloss.Width(it.Key)-1))
		if i == m.menuCur {
			lines[i] = m.st.dropItemSel.Render(left + pad + it.Key + " ")
		} else {
			lines[i] = m.st.dropItem.Render(left+pad) + m.st.dropKey.Render(it.Key+" ")
		}
	}
	return m.st.dropdown.Render(strings.Join(lines, "\n"))
}

func (m *Model) viewTree(l layout) string {
	iw := l.treeW - 2
	focused := m.focus == focusTree && !m.preview
	var lines []string
	for i := m.tree.Offset; i < len(m.tree.Rows) && len(lines) < l.treeIH; i++ {
		r := m.tree.Rows[i]
		icon := iconFolder
		switch {
		case r.N.Err != nil:
			icon = iconWarn
		case r.N.Expanded:
			icon = iconFolderOpen
		}
		if i == m.tree.Cursor {
			sel := m.st.treeSelBlur
			if focused {
				sel = m.st.treeSel
			}
			plain := ansi.Truncate(r.Prefix+icon+" "+r.N.Name, iw, "…")
			lines = append(lines, sel.Width(iw).Render(plain))
			continue
		}
		line := m.st.dim.Render(r.Prefix) + m.st.treeIcon.Render(icon) + " " + m.st.text.Render(r.N.Name)
		lines = append(lines, ansi.Truncate(line, iw, "…"))
	}

	panel := m.st.panel
	if focused {
		panel = m.st.panelFocus
	}
	return panel.Width(l.treeW).Height(l.bodyH).MaxHeight(l.bodyH).Render(strings.Join(lines, "\n"))
}

// image renvoie le rendu en cache, ou un texte d'attente / d'erreur.
func (m *Model) image(p string, w, h int) string {
	key := renderKey(p, w, h, m.kind)
	if out, ok := m.cache[key]; ok {
		return out
	}
	if _, ok := m.failed[key]; ok {
		return m.st.errText.Render(ansi.Truncate("✗ illisible", w, "…"))
	}
	return m.st.dim.Render(ansi.Truncate("chargement…", w, "…"))
}

func (m *Model) viewGallery(l layout) string {
	g := m.gal
	iw, ih := l.galIW, l.galIH
	var header, body string

	switch {
	case g.Err != nil:
		header = iconWarn + " " + filepath.Base(g.Dir)
		body = lipgloss.Place(iw, ih, lipgloss.Center, lipgloss.Center,
			m.st.errText.Render("Impossible de lire ce répertoire"))
	case len(g.Files) == 0:
		header = iconFolderOpen + " " + filepath.Base(g.Dir)
		body = lipgloss.Place(iw, ih, lipgloss.Center, lipgloss.Center,
			m.st.dim.Render("Aucune image dans ce répertoire"))
	case m.preview:
		p := g.Current()
		header = fmt.Sprintf("%s %s  (%d/%d)", iconImage, filepath.Base(p), g.Cursor+1, len(g.Files))
		body = lipgloss.Place(iw, ih, lipgloss.Center, lipgloss.Center, m.imageHD(p, iw, ih))
	default:
		n := len(g.Files)
		unit := "images"
		if n == 1 {
			unit = "image"
		}
		header = fmt.Sprintf("%s %s · %d %s", iconFolderOpen, filepath.Base(g.Dir), n, unit)
		body = m.viewGrid(iw, ih)
	}

	content := ansi.Truncate(m.st.header.Render(header), iw, "…") + "\n" + body
	panel := m.st.panel
	if m.focus == focusGallery || m.preview {
		panel = m.st.panelFocus
	}
	return panel.Width(l.galW).Height(l.bodyH).MaxHeight(l.bodyH).Padding(0, 1).Render(content)
}

func (m *Model) viewGrid(iw, ih int) string {
	gl := m.gal.Layout(iw, ih)
	idx := m.gal.Visible(gl)
	var rows []string
	for r := 0; r < gl.Rows; r++ {
		var cells []string
		for c := 0; c < gl.Cols; c++ {
			k := r*gl.Cols + c
			if k >= len(idx) {
				break
			}
			if c > 0 {
				cells = append(cells, " ")
			}
			cells = append(cells, m.viewCell(idx[k], gl))
		}
		if len(cells) == 0 {
			break
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cells...))
	}
	return strings.Join(rows, "\n")
}

func (m *Model) viewCell(i int, gl gallery.Layout) string {
	p := m.gal.Files[i]
	sel := i == m.gal.Cursor

	img := lipgloss.Place(gl.TW, gl.TH, lipgloss.Center, lipgloss.Center, m.imageHD(p, gl.TW, gl.TH))
	nameSt, box := m.st.name, m.st.cell
	if sel {
		nameSt, box = m.st.nameSel, m.st.cellSelBlur
		if m.focus == focusGallery {
			box = m.st.cellSel
		}
	}
	name := lipgloss.PlaceHorizontal(gl.TW, lipgloss.Center,
		nameSt.Render(ansi.Truncate(filepath.Base(p), gl.TW, "…")))
	return box.Width(gl.CellW).Render(img + "\n" + name)
}

// homeRel remplace le répertoire personnel par « ~ ».
func homeRel(p string) string {
	if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(p, home) {
		return "~" + strings.TrimPrefix(p, home)
	}
	return p
}

func (m *Model) viewStatus() string {
	var mode string
	switch {
	case m.menuOpen >= 0:
		mode = "MENU"
	case m.preview:
		mode = "APERÇU"
	case m.focus == focusTree:
		mode = "ARBRE"
	default:
		mode = "GALERIE"
	}
	left := m.st.badge.Render(mode)
	switch {
	case m.flash != "" && m.flashErr:
		left += m.st.statusErr.Render(m.flash)
	case m.flash != "":
		left += m.st.statusOK.Render(m.flash)
	default:
		left += m.st.statusBar.Render(" " + homeRel(m.gal.Dir))
	}

	var right []string
	if p := m.gal.Current(); p != "" {
		parts := []string{filepath.Base(p)}
		if info, ok := m.info[p]; ok {
			if info.OK {
				parts = append(parts, fmt.Sprintf("%d×%d", info.W, info.H))
			}
			parts = append(parts, info.Format, imgrender.HumanSize(info.Size))
		}
		right = append(right, m.st.statusInfo.Render(strings.Join(parts, " · ")))
	}
	engine := m.kind.String()
	if m.hd {
		engine = "kitty HD"
	}
	right = append(right, m.st.badgeAlt.Render(engine))
	r := strings.Join(right, "")

	if lipgloss.Width(left)+lipgloss.Width(r) > m.w {
		left = ansi.Truncate(left, max(0, m.w-lipgloss.Width(r)), "…")
	}
	return fill(m.st.statusBar, left, m.w-lipgloss.Width(r)) + r
}

func (m *Model) viewHints() string {
	var h [][2]string
	switch {
	case m.modal != modalNone:
		h = [][2]string{{"Échap", "fermer"}}
	case m.menuOpen >= 0:
		h = [][2]string{{"←→", "menus"}, {"↑↓", "choisir"}, {"Entrée", "valider"}, {"Échap", "fermer"}}
	case m.preview:
		h = [][2]string{{"←→", "précédente/suivante"}, {"f", "plein écran"}, {"d", "HD on/off"}, {"o", "ouvrir"}, {"Échap", "retour"}, {"q", "quitter"}}
	case m.focus == focusTree:
		h = [][2]string{{"↑↓", "naviguer"}, {"→←", "déplier/replier"}, {"Tab", "galerie"}, {"F10", "menu"}, {"?", "aide"}, {"q", "quitter"}}
	default:
		h = [][2]string{{"flèches", "choisir"}, {"Entrée", "aperçu"}, {"f", "plein écran"}, {"+/-", "taille"}, {"d", "HD"}, {"Tab", "arbre"}, {"F10", "menu"}, {"q", "quitter"}}
	}
	parts := make([]string, len(h))
	for i, kv := range h {
		parts[i] = m.st.hintKey.Render(kv[0]) + m.st.hints.Render(" "+kv[1])
	}
	line := " " + strings.Join(parts, m.st.hints.Render("  ·  "))
	return fill(lipgloss.NewStyle(), line, m.w)
}

func (m *Model) viewModal() string {
	t, k, d := m.st.modalTitle, m.st.modalKey, m.st.modalText
	row := func(key, desc string) string {
		return k.Render(fmt.Sprintf("  %-14s", key)) + d.Render(desc)
	}

	var lines []string
	switch m.modal {
	case modalHelp:
		lines = []string{
			t.Render("Raccourcis clavier"), "",
			t.Render("Général"),
			row("F10 / Alt+F", "ouvrir le menu"),
			row("Tab", "changer de panneau"),
			row("?", "cette aide"),
			row("q / Ctrl+C", "quitter"), "",
			t.Render("Arbre"),
			row("↑ ↓", "se déplacer"),
			row("→ / ←", "déplier / replier"),
			row("Entrée", "déplier ou replier"), "",
			t.Render("Galerie"),
			row("flèches", "choisir une image"),
			row("Entrée", "aperçu"),
			row("f", "plein écran HD (chafa)"),
			row("o", "ouvrir avec le bureau"),
			row("+ / -", "taille des miniatures"), "",
			t.Render("Divers"),
			row(".", "répertoires cachés"),
			row("r", "rendu chafa / interne"),
			row("d", "images HD (Kitty) on/off"),
			row("F5", "actualiser"),
			row("souris", "clic, re-clic = aperçu, molette"),
		}
	case modalAbout:
		chafa := "non installé"
		if m.chafaOK {
			chafa = "installé"
		}
		lines = []string{
			t.Render(iconImage + "  imgview"), "",
			d.Render("Explorateur d'images pour le terminal"),
			d.Render("Bubble Tea v2 · Lip Gloss v2"), "",
			d.Render("chafa : " + chafa),
			d.Render("Rendu actif : " + m.kind.String()),
		}
	}
	lines = append(lines, "", m.st.dim.Background(m.st.p.mantle).Render("Échap pour fermer"))
	return m.st.modal.Render(strings.Join(lines, "\n"))
}
