// Package ui assemble l'interface Bubble Tea : barre de menus en haut,
// arborescence à gauche, galerie d'images à droite, barre d'état en bas.
package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/laurentpoirierfr/imgview/internal/gallery"
	"github.com/laurentpoirierfr/imgview/internal/mathx"
	"github.com/laurentpoirierfr/imgview/internal/menu"
	"github.com/laurentpoirierfr/imgview/internal/tree"
	"github.com/laurentpoirierfr/imgview/pkg/imgrender"
	"github.com/laurentpoirierfr/imgview/pkg/kittygfx"
)

// ─── État ────────────────────────────────────────────────────────────────────

type focus int

const (
	focusTree focus = iota
	focusGallery
)

type modal int

const (
	modalNone modal = iota
	modalHelp
	modalAbout
)

// clearFlashMsg efface le message temporaire de la barre d'état.
type clearFlashMsg int

const (
	menuH   = 1 // hauteur de la barre de menus
	footerH = 2 // barre d'état + ligne d'aide
)

const (
	iconFolder     = ""
	iconFolderOpen = ""
	iconWarn       = ""
	iconImage      = ""
)

type Model struct {
	w, h int
	st   styles

	tree  *tree.Tree
	gal   *gallery.Gallery
	focus focus

	menus    []menu.Menu
	menuOpen int // -1 : aucun menu ouvert
	menuCur  int

	modal   modal
	preview bool

	kind    imgrender.Kind
	chafaOK bool

	cache   map[string]string // rendu texte par clé (fichier, taille, moteur)
	failed  map[string]string // erreurs de rendu par clé
	pending map[string]bool   // rendus en cours
	info    map[string]imgrender.Info

	// Aperçu haute définition (protocole graphique Kitty), voir hd.go.
	hd, hdOK     bool
	cellW, cellH int // taille d'une cellule en pixels (0 si inconnue)
	hdImgs       map[string]hdImage
	hdPending    map[string]bool
	hdFailed     map[string]string
	hdVisible    map[string]bool // images HD affichées au dernier rendu
	hdNextID     uint32

	flash    string
	flashErr bool
	flashID  int
}

// New crée le modèle pour explorer le répertoire root.
func New(root string) *Model {
	m := &Model{
		st:        newStyles(true),
		tree:      tree.New(root),
		gal:       &gallery.Gallery{SizeIdx: 1},
		menus:     menu.Default(),
		menuOpen:  -1,
		chafaOK:   imgrender.ChafaAvailable(),
		cache:     map[string]string{},
		failed:    map[string]string{},
		pending:   map[string]bool{},
		info:      map[string]imgrender.Info{},
		hdOK:      kittygfx.Supported(),
		hdImgs:    map[string]hdImage{},
		hdPending: map[string]bool{},
		hdFailed:  map[string]string{},
		hdVisible: map[string]bool{},
		hdNextID:  0x494d00, // identifiants d'images propres à imgview
	}
	m.hd = m.hdOK
	if !m.chafaOK {
		m.kind = imgrender.Internal
	}
	m.gal.Load(root)
	return m
}

func (m *Model) Init() tea.Cmd {
	// La réponse (BackgroundColorMsg) permet d'adapter le thème clair/sombre.
	return tea.RequestBackgroundColor
}

// ─── Mise en page ────────────────────────────────────────────────────────────

type layout struct {
	bodyY, bodyH int
	treeW        int
	galX, galW   int
	treeIH       int // lignes visibles dans l'arbre
	galIW, galIH int // zone utile de la galerie (sous l'en-tête)
}

func (m *Model) layout() layout {
	l := layout{bodyY: menuH, bodyH: max(5, m.h-menuH-footerH)}
	l.treeW = mathx.Clamp(m.w/4, 22, 42)
	if m.w < 70 {
		l.treeW = max(12, m.w/3)
	}
	l.galX = l.treeW
	l.galW = max(10, m.w-l.treeW)
	l.treeIH = l.bodyH - 2 // bordures
	l.galIW = l.galW - 4   // bordures + marge intérieure
	l.galIH = l.bodyH - 3  // bordures + ligne d'en-tête
	return l
}

func (m *Model) gridLayout() gallery.Layout {
	l := m.layout()
	return m.gal.Layout(l.galIW, l.galIH)
}

// ─── Update ──────────────────────────────────────────────────────────────────

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		m.cellW, m.cellH = kittygfx.CellSize()

	case tea.BackgroundColorMsg:
		m.st = newStyles(msg.IsDark())

	case renderMsg:
		delete(m.pending, msg.key)
		if msg.err != nil {
			m.failed[msg.key] = msg.err.Error()
		} else {
			m.cache[msg.key] = msg.out
		}
		if msg.info.Size > 0 {
			m.info[msg.path] = msg.info
		}

	case hdReadyMsg:
		cmds = append(cmds, m.hdTransmit(msg))

	case execDoneMsg:
		// Le plein écran efface les images du terminal : il faudra les renvoyer.
		cmds = append(cmds, m.hdFree())
		if msg.err != nil {
			cmds = append(cmds, m.setFlash("Plein écran : "+msg.err.Error(), true))
		}

	case clearFlashMsg:
		if int(msg) == m.flashID {
			m.flash = ""
		}

	case tea.KeyPressMsg:
		cmds = append(cmds, m.handleKey(msg))

	case tea.MouseClickMsg:
		mo := tea.Mouse(msg)
		if mo.Button == tea.MouseLeft {
			cmds = append(cmds, m.handleClick(mo.X, mo.Y))
		}

	case tea.MouseWheelMsg:
		mo := tea.Mouse(msg)
		switch mo.Button {
		case tea.MouseWheelUp:
			m.handleWheel(mo.X, -1)
		case tea.MouseWheelDown:
			m.handleWheel(mo.X, 1)
		}
	}

	m.sync()
	cmds = append(cmds, m.ensureImages()...)
	cmds = append(cmds, m.hdEvict())
	return m, tea.Batch(cmds...)
}

// sync charge la galerie du répertoire sélectionné et garde les curseurs visibles.
func (m *Model) sync() {
	if sel := m.tree.Selected().Path; sel != m.gal.Dir {
		m.gal.Load(sel)
		m.preview = false
	}
	if m.w == 0 {
		return
	}
	m.tree.ScrollTo(m.layout().treeIH)
	if !m.preview {
		m.gal.ScrollTo(m.gridLayout())
	}
}

// ensureImages lance le rendu des images visibles qui ne sont pas en cache.
func (m *Model) ensureImages() []tea.Cmd {
	if m.w == 0 {
		return nil
	}
	clear(m.hdVisible)
	var cmds []tea.Cmd
	want := func(p string, w, h int) {
		key := renderKey(p, w, h, m.kind)
		if _, ok := m.cache[key]; ok {
			return
		}
		if _, ok := m.failed[key]; ok || m.pending[key] {
			return
		}
		m.pending[key] = true
		cmds = append(cmds, renderCmd(key, p, w, h, m.kind))
	}

	if m.preview {
		l := m.layout()
		p := m.gal.Current()
		if !m.hdRequest(p, l.galIW, l.galIH, true, &cmds) && p != "" {
			want(p, l.galIW, l.galIH)
		}
		return cmds
	}
	gl := m.gridLayout()
	for _, i := range m.gal.Visible(gl) {
		p := m.gal.Files[i]
		if !m.hdRequest(p, gl.TW, gl.TH, false, &cmds) {
			want(p, gl.TW, gl.TH)
		}
	}
	return cmds
}

func (m *Model) setFlash(s string, isErr bool) tea.Cmd {
	m.flashID++
	id := m.flashID
	m.flash, m.flashErr = s, isErr
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg { return clearFlashMsg(id) })
}
