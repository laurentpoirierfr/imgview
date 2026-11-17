package ui

import (
	"fmt"
	"path/filepath"

	tea "charm.land/bubbletea/v2"

	"github.com/laurentpoirierfr/imgview/internal/gallery"
	"github.com/laurentpoirierfr/imgview/internal/mathx"
	"github.com/laurentpoirierfr/imgview/internal/menu"
	"github.com/laurentpoirierfr/imgview/pkg/imgrender"
)

// ─── Actions ─────────────────────────────────────────────────────────────────

func (m *Model) do(a menu.Action) tea.Cmd {
	cur := m.gal.Current()
	switch a {
	case menu.ActPreview:
		if cur == "" {
			return m.setFlash("Aucune image sélectionnée", true)
		}
		m.preview, m.focus = true, focusGallery

	case menu.ActFullscreen:
		if cur == "" {
			return m.setFlash("Aucune image sélectionnée", true)
		}
		if !m.chafaOK {
			return m.setFlash("chafa n'est pas installé", true)
		}
		return fullscreenCmd(cur)

	case menu.ActOpenExternal:
		if cur == "" {
			return m.setFlash("Aucune image sélectionnée", true)
		}
		if err := openExternal(cur); err != nil {
			return m.setFlash("Impossible d'ouvrir : "+err.Error(), true)
		}
		return m.setFlash("Ouverture de "+filepath.Base(cur), false)

	case menu.ActRefresh:
		m.tree.Reload()
		m.gal.Reload()
		m.cache, m.failed, m.info = map[string]string{}, map[string]string{}, map[string]imgrender.Info{}
		return m.setFlash("Actualisé", false)

	case menu.ActQuit:
		// Libérer les images HD avant de rendre la main au terminal.
		if free := m.hdFree(); free != nil {
			return tea.Sequence(free, tea.Quit)
		}
		return tea.Quit

	case menu.ActZoomIn, menu.ActZoomOut:
		d := 1
		if a == menu.ActZoomOut {
			d = -1
		}
		m.gal.SizeIdx = mathx.Clamp(m.gal.SizeIdx+d, 0, len(gallery.ThumbSizes)-1)
		w, h := m.gal.Thumb()
		return m.setFlash(fmt.Sprintf("Miniatures : %d×%d", w, h), false)

	case menu.ActToggleHidden:
		m.tree.ShowHidden = !m.tree.ShowHidden
		m.tree.Reload()
		if m.tree.ShowHidden {
			return m.setFlash("Répertoires cachés affichés", false)
		}
		return m.setFlash("Répertoires cachés masqués", false)

	case menu.ActToggleRenderer:
		if !m.chafaOK {
			return m.setFlash("chafa introuvable : rendu interne uniquement", true)
		}
		if m.kind == imgrender.Chafa {
			m.kind = imgrender.Internal
		} else {
			m.kind = imgrender.Chafa
		}
		return m.setFlash("Rendu : "+m.kind.String(), false)

	case menu.ActHelp:
		m.modal = modalHelp
	case menu.ActAbout:
		m.modal = modalAbout

	case menu.ActToggleHD:
		if !m.hdOK {
			return m.setFlash("Images HD : terminal non compatible (Ghostty ou Kitty requis)", true)
		}
		m.hd = !m.hd
		if m.hd {
			return m.setFlash("Images HD activées", false)
		}
		return tea.Batch(m.hdFree(), m.setFlash("Images HD désactivées", false))
	}
	return nil
}
