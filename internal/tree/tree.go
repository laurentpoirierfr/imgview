// Package tree gère l'arborescence des répertoires affichée à gauche.
package tree

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/laurentpoirierfr/imgview/internal/mathx"
)

// Node est un répertoire de l'arborescence. Les enfants sont chargés à la
// demande, la première fois que le nœud est déplié.
type Node struct {
	Name     string
	Path     string
	Parent   *Node
	Children []*Node
	Loaded   bool
	Expanded bool
	Err      error
}

// Row est une ligne visible de l'arbre, avec ses connecteurs (├─ └─ │).
type Row struct {
	N      *Node
	Prefix string
}

type Tree struct {
	Root       *Node
	Rows       []Row
	Cursor     int
	Offset     int
	ShowHidden bool
}

func New(path string) *Tree {
	t := &Tree{Root: &Node{Name: filepath.Base(path), Path: path}}
	t.Expand(t.Root)
	return t
}

func (t *Tree) load(n *Node) {
	if n.Loaded {
		return
	}
	n.Loaded = true
	entries, err := os.ReadDir(n.Path)
	if err != nil {
		n.Err = err
		return
	}
	for _, e := range entries {
		name := e.Name()
		if !t.ShowHidden && strings.HasPrefix(name, ".") {
			continue
		}
		p := filepath.Join(n.Path, name)
		if !e.IsDir() {
			// Suivre les liens symboliques vers des répertoires.
			if e.Type()&os.ModeSymlink == 0 {
				continue
			}
			if st, err := os.Stat(p); err != nil || !st.IsDir() {
				continue
			}
		}
		n.Children = append(n.Children, &Node{Name: name, Path: p, Parent: n})
	}
	sort.Slice(n.Children, func(i, j int) bool {
		return strings.ToLower(n.Children[i].Name) < strings.ToLower(n.Children[j].Name)
	})
}

func (t *Tree) Expand(n *Node) {
	t.load(n)
	n.Expanded = true
	t.rebuild()
}

func (t *Tree) Collapse(n *Node) {
	n.Expanded = false
	t.rebuild()
}

// rebuild recalcule la liste des lignes visibles.
func (t *Tree) rebuild() {
	t.Rows = t.Rows[:0]
	t.Rows = append(t.Rows, Row{N: t.Root})
	if t.Root.Expanded {
		t.walk(t.Root, "")
	}
	t.Cursor = mathx.Clamp(t.Cursor, 0, len(t.Rows)-1)
}

func (t *Tree) walk(n *Node, prefix string) {
	for i, c := range n.Children {
		conn, next := "├─ ", "│  "
		if i == len(n.Children)-1 {
			conn, next = "└─ ", "   "
		}
		t.Rows = append(t.Rows, Row{N: c, Prefix: prefix + conn})
		if c.Expanded {
			t.walk(c, prefix+next)
		}
	}
}

func (t *Tree) Selected() *Node {
	if len(t.Rows) == 0 {
		return t.Root
	}
	return t.Rows[t.Cursor].N
}

func (t *Tree) Move(delta int) {
	t.Cursor = mathx.Clamp(t.Cursor+delta, 0, len(t.Rows)-1)
}

// Right déplie le nœud, ou descend sur son premier enfant s'il est déjà déplié.
func (t *Tree) Right() {
	n := t.Selected()
	if !n.Expanded {
		t.Expand(n)
		return
	}
	if len(n.Children) > 0 {
		t.Move(1)
	}
}

// Left replie le nœud, ou remonte sur son parent.
func (t *Tree) Left() {
	n := t.Selected()
	if n.Expanded && n != t.Root {
		t.Collapse(n)
		return
	}
	if n.Parent != nil {
		t.SelectPath(n.Parent.Path)
	}
}

func (t *Tree) Toggle() {
	n := t.Selected()
	if n.Expanded && n != t.Root {
		t.Collapse(n)
	} else {
		t.Expand(n)
	}
}

func (t *Tree) SelectPath(p string) bool {
	for i, r := range t.Rows {
		if r.N.Path == p {
			t.Cursor = i
			return true
		}
	}
	return false
}

// Reload relit le disque en conservant les répertoires dépliés et la sélection.
func (t *Tree) Reload() {
	open := map[string]bool{}
	var collect func(n *Node)
	collect = func(n *Node) {
		if n.Expanded {
			open[n.Path] = true
		}
		for _, c := range n.Children {
			collect(c)
		}
	}
	collect(t.Root)
	sel := t.Selected().Path

	t.Root = &Node{Name: t.Root.Name, Path: t.Root.Path}
	var reopen func(n *Node)
	reopen = func(n *Node) {
		if !open[n.Path] {
			return
		}
		t.load(n)
		n.Expanded = true
		for _, c := range n.Children {
			reopen(c)
		}
	}
	reopen(t.Root)
	t.rebuild()
	if !t.SelectPath(sel) {
		// Le répertoire a disparu : se placer sur l'ancêtre le plus proche.
		for p := filepath.Dir(sel); p != "." && p != "/"; p = filepath.Dir(p) {
			if t.SelectPath(p) {
				break
			}
		}
	}
}

// ScrollTo ajuste le décalage pour que le curseur reste visible.
func (t *Tree) ScrollTo(height int) {
	if height <= 0 {
		return
	}
	if t.Cursor < t.Offset {
		t.Offset = t.Cursor
	}
	if t.Cursor >= t.Offset+height {
		t.Offset = t.Cursor - height + 1
	}
	t.Offset = mathx.Clamp(t.Offset, 0, max(0, len(t.Rows)-height))
}
