// Package gallery gère la grille de miniatures du répertoire courant.
package gallery

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/laurentpoirierfr/imgview/internal/mathx"
	"github.com/laurentpoirierfr/imgview/pkg/imgrender"
)

// ThumbSizes liste les tailles de miniature en cellules (une cellule fait
// environ 1×2, donc largeur = 2 × hauteur donne un cadre à peu près carré).
var ThumbSizes = []struct{ W, H int }{
	{14, 7}, {20, 10}, {28, 14}, {40, 20},
}

type Gallery struct {
	Dir       string
	Files     []string
	Cursor    int
	RowOffset int
	SizeIdx   int
	Err       error
}

func (g *Gallery) Load(dir string) {
	if dir == g.Dir && g.Err == nil {
		return
	}
	g.Dir, g.Files, g.Cursor, g.RowOffset, g.Err = dir, nil, 0, 0, nil
	entries, err := os.ReadDir(dir)
	if err != nil {
		g.Err = err
		return
	}
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") || !imgrender.IsImage(e.Name()) {
			continue
		}
		g.Files = append(g.Files, filepath.Join(dir, e.Name()))
	}
	sort.Slice(g.Files, func(i, j int) bool {
		return strings.ToLower(filepath.Base(g.Files[i])) < strings.ToLower(filepath.Base(g.Files[j]))
	})
}

func (g *Gallery) Reload() {
	dir, cur := g.Dir, g.Current()
	g.Dir = ""
	g.Load(dir)
	for i, f := range g.Files {
		if f == cur {
			g.Cursor = i
		}
	}
}

func (g *Gallery) Current() string {
	if g.Cursor < 0 || g.Cursor >= len(g.Files) {
		return ""
	}
	return g.Files[g.Cursor]
}

func (g *Gallery) Thumb() (int, int) {
	s := ThumbSizes[g.SizeIdx]
	return s.W, s.H
}

// Layout décrit la grille de miniatures pour une zone donnée.
type Layout struct {
	Cols, Rows   int // colonnes, rangées visibles
	CellW, CellH int // taille extérieure d'une case (bordure comprise)
	TW, TH       int // taille de l'image dans la case
}

func (g *Gallery) Layout(width, height int) Layout {
	tw, th := g.Thumb()
	// Si la zone est trop petite, réduire la miniature.
	for tw > 6 && tw+2 > width {
		tw -= 2
		th = tw / 2
	}
	for th > 3 && th+3 > height {
		th--
		tw = th * 2
	}
	l := Layout{TW: tw, TH: th, CellW: tw + 2, CellH: th + 3}
	l.Cols = max(1, (width+1)/(l.CellW+1))
	l.Rows = max(1, height/l.CellH)
	return l
}

func (g *Gallery) Move(delta int) {
	if len(g.Files) > 0 {
		g.Cursor = mathx.Clamp(g.Cursor+delta, 0, len(g.Files)-1)
	}
}

// ScrollTo garde la rangée sélectionnée visible.
func (g *Gallery) ScrollTo(l Layout) {
	row := g.Cursor / l.Cols
	if row < g.RowOffset {
		g.RowOffset = row
	}
	if row >= g.RowOffset+l.Rows {
		g.RowOffset = row - l.Rows + 1
	}
	total := (len(g.Files) + l.Cols - 1) / l.Cols
	g.RowOffset = mathx.Clamp(g.RowOffset, 0, max(0, total-l.Rows))
}

// Visible renvoie les indices des fichiers affichés.
func (g *Gallery) Visible(l Layout) []int {
	var idx []int
	start := g.RowOffset * l.Cols
	for i := start; i < len(g.Files) && i < start+l.Rows*l.Cols; i++ {
		idx = append(idx, i)
	}
	return idx
}
