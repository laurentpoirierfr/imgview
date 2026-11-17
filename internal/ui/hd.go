package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/laurentpoirierfr/imgview/pkg/imgrender"
	"github.com/laurentpoirierfr/imgview/pkg/kittygfx"
)

// Images haute définition (aperçu et miniatures) : chaque image est envoyée
// au terminal avec le protocole graphique Kitty, puis affichée via des
// caractères « placeholder » que Bubble Tea dessine comme du texte
// (voir pkg/kittygfx).
//
// Mémoire du terminal : les grandes images (aperçu) sont libérées dès
// qu'elles ne sont plus affichées ; les miniatures restent en mémoire pour un
// retour rapide, jusqu'à hdMaxImages, au-delà de quoi celles qui ne sont pas
// visibles sont libérées.
const hdMaxImages = 150

// hdImage est une image déjà envoyée au terminal.
type hdImage struct {
	id         uint32
	cols, rows int
	big        bool // image d'aperçu (libérée dès qu'elle n'est plus visible)
}

// hdReadyMsg arrive quand l'image a été préparée (PNG réduit à la bonne taille).
type hdReadyMsg struct {
	key        string
	path       string
	png        []byte
	cols, rows int
	big        bool
	info       imgrender.Info
	err        error
}

func hdKey(path string, w, h int) string {
	return fmt.Sprintf("%s|%dx%d|hd", path, w, h)
}

// hdCmd prépare l'image en arrière-plan pour une zone de w×h cellules.
func hdCmd(key, path string, w, h, cellW, cellH int, big bool) tea.Cmd {
	return func() tea.Msg {
		renderSem <- struct{}{}
		defer func() { <-renderSem }()

		msg := hdReadyMsg{key: key, path: path, big: big, info: imgrender.ReadInfo(path)}
		if !msg.info.OK {
			msg.err = fmt.Errorf("dimensions illisibles")
			return msg
		}
		msg.cols, msg.rows = kittygfx.Fit(msg.info.W, msg.info.H, w, h, cellW, cellH)
		if msg.cols == 0 {
			msg.err = fmt.Errorf("zone trop petite")
			return msg
		}
		// Taille utile en pixels (10×20 par cellule si le terminal ne la donne pas).
		pw, ph := cellW, cellH
		if pw <= 0 || ph <= 0 {
			pw, ph = 10, 20
		}
		msg.png, _, _, msg.err = imgrender.PNG(path, msg.cols*pw, msg.rows*ph)
		return msg
	}
}

// hdWanted indique si l'image doit être affichée en HD pour cette zone
// (HD activé et pas d'échec précédent), et renvoie sa clé.
func (m *Model) hdWanted(p string, w, h int) (key string, use bool) {
	if !m.hd || p == "" {
		return "", false
	}
	key = hdKey(p, w, h)
	_, failed := m.hdFailed[key]
	return key, !failed
}

// hdRequest marque l'image comme affichée et lance sa préparation si elle
// n'a pas encore été envoyée. Renvoie false si elle doit être affichée en
// texte (HD désactivé ou en échec).
func (m *Model) hdRequest(p string, w, h int, big bool, cmds *[]tea.Cmd) bool {
	key, use := m.hdWanted(p, w, h)
	if !use {
		return false
	}
	m.hdVisible[key] = true
	if _, ok := m.hdImgs[key]; !ok && !m.hdPending[key] {
		m.hdPending[key] = true
		*cmds = append(*cmds, hdCmd(key, p, w, h, m.cellW, m.cellH, big))
	}
	return true
}

// hdTransmit envoie au terminal une image préparée, si elle est toujours
// affichée (sinon elle sera redemandée quand elle redeviendra visible).
func (m *Model) hdTransmit(msg hdReadyMsg) tea.Cmd {
	delete(m.hdPending, msg.key)
	if msg.info.Size > 0 {
		m.info[msg.path] = msg.info
	}
	if msg.err != nil {
		m.hdFailed[msg.key] = msg.err.Error()
		return nil
	}
	if !m.hd || !m.hdVisible[msg.key] {
		return nil
	}
	if _, ok := m.hdImgs[msg.key]; ok {
		return nil
	}
	m.hdNextID++
	img := hdImage{id: m.hdNextID, cols: msg.cols, rows: msg.rows, big: msg.big}
	m.hdImgs[msg.key] = img
	return tea.Raw(kittygfx.Transmit(img.id, msg.png, img.cols, img.rows))
}

// hdEvict libère les images qui ne sont plus affichées : toujours pour les
// aperçus, et pour les miniatures quand il y en a trop.
func (m *Model) hdEvict() tea.Cmd {
	over := len(m.hdImgs) > hdMaxImages
	var b strings.Builder
	for k, img := range m.hdImgs {
		if m.hdVisible[k] || !(img.big || over) {
			continue
		}
		b.WriteString(kittygfx.Delete(img.id))
		delete(m.hdImgs, k)
	}
	if b.Len() == 0 {
		return nil
	}
	return tea.Raw(b.String())
}

// hdFreeSeq oublie toutes les images HD et renvoie les séquences qui les
// libèrent dans le terminal.
func (m *Model) hdFreeSeq() string {
	var b strings.Builder
	for k, img := range m.hdImgs {
		b.WriteString(kittygfx.Delete(img.id))
		delete(m.hdImgs, k)
	}
	return b.String()
}

// hdFree libère toutes les images HD (changement de mode, sortie…).
func (m *Model) hdFree() tea.Cmd {
	if s := m.hdFreeSeq(); s != "" {
		return tea.Raw(s)
	}
	return nil
}

// imageHD renvoie l'image HD si elle est prête, un texte d'attente si elle
// est en préparation, ou le rendu texte si le HD n'est pas utilisable.
func (m *Model) imageHD(p string, w, h int) string {
	if key, use := m.hdWanted(p, w, h); use {
		if img, ok := m.hdImgs[key]; ok {
			return kittygfx.Placeholder(img.id, img.cols, img.rows)
		}
		return m.st.dim.Render("chargement…")
	}
	return m.image(p, w, h)
}
