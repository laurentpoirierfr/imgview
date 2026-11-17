// Package kittygfx affiche de vraies images (pleine résolution) dans le
// terminal grâce au protocole graphique Kitty, pris en charge par Kitty et
// Ghostty.
//
// Il utilise les « Unicode placeholders » : l'image est d'abord envoyée au
// terminal une seule fois (Transmit), puis on écrit à sa place des caractères
// spéciaux U+10EEEE (Placeholder) que le terminal remplace par les morceaux de
// l'image. Ces caractères sont du texte ordinaire, ce qui permet de les
// intégrer dans une interface Bubble Tea / Lip Gloss comme n'importe quelle
// autre chaîne.
//
// Référence : https://sw.kovidgoyal.net/kitty/graphics-protocol/#unicode-placeholders
package kittygfx

import (
	"encoding/base64"
	"fmt"
	"math"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

// placeholderRune est le caractère remplacé par l'image dans le terminal.
const placeholderRune = '\U0010EEEE'

// MaxCells est la taille maximale (en colonnes et en lignes) d'une image :
// elle est limitée par le nombre de diacritiques disponibles.
var MaxCells = len(diacritics)

// Supported indique si le terminal courant gère le protocole graphique Kitty
// avec les Unicode placeholders. La variable IMGVIEW_NO_KITTY le désactive.
func Supported() bool {
	if os.Getenv("IMGVIEW_NO_KITTY") != "" {
		return false
	}
	if os.Getenv("KITTY_WINDOW_ID") != "" || os.Getenv("TERM_PROGRAM") == "ghostty" {
		return true
	}
	switch os.Getenv("TERM") {
	case "xterm-kitty", "xterm-ghostty":
		return true
	}
	return false
}

// CellSize renvoie la taille d'une cellule du terminal en pixels, ou (0, 0)
// si le terminal ne la communique pas.
func CellSize() (w, h int) {
	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil || ws.Col == 0 || ws.Row == 0 || ws.Xpixel == 0 || ws.Ypixel == 0 {
		return 0, 0
	}
	return int(ws.Xpixel) / int(ws.Col), int(ws.Ypixel) / int(ws.Row)
}

// Fit calcule le nombre de colonnes et de lignes nécessaires pour afficher une
// image imgW×imgH pixels dans au plus maxCols×maxRows cellules de
// cellW×cellH pixels, en gardant ses proportions. Une image plus petite que la
// zone n'est pas agrandie (si la taille des cellules est connue).
func Fit(imgW, imgH, maxCols, maxRows, cellW, cellH int) (cols, rows int) {
	maxCols = min(maxCols, MaxCells)
	maxRows = min(maxRows, MaxCells)
	if imgW <= 0 || imgH <= 0 || maxCols <= 0 || maxRows <= 0 {
		return 0, 0
	}
	known := cellW > 0 && cellH > 0
	if !known {
		cellW, cellH = 1, 2 // proportions habituelles d'une cellule
	}
	scale := math.Min(float64(maxCols*cellW)/float64(imgW), float64(maxRows*cellH)/float64(imgH))
	if known {
		scale = math.Min(scale, 1)
	}
	cols = int(math.Round(float64(imgW) * scale / float64(cellW)))
	rows = int(math.Round(float64(imgH) * scale / float64(cellH)))
	return min(max(cols, 1), maxCols), min(max(rows, 1), maxRows)
}

// Transmit renvoie les séquences qui envoient une image PNG au terminal sous
// l'identifiant id, avec un placement virtuel de cols×rows cellules.
// Les données sont découpées en morceaux de 4 Kio, comme l'exige le protocole.
func Transmit(id uint32, png []byte, cols, rows int) string {
	enc := base64.StdEncoding.EncodeToString(png)
	var b strings.Builder
	for first := true; first || len(enc) > 0; first = false {
		n := min(4096, len(enc))
		chunk := enc[:n]
		enc = enc[n:]
		more := 0
		if len(enc) > 0 {
			more = 1
		}
		if first {
			fmt.Fprintf(&b, "\x1b_Ga=T,U=1,f=100,t=d,i=%d,c=%d,r=%d,q=2,m=%d;%s\x1b\\", id, cols, rows, more, chunk)
		} else {
			fmt.Fprintf(&b, "\x1b_Gm=%d,q=2;%s\x1b\\", more, chunk)
		}
	}
	return b.String()
}

// Delete renvoie la séquence qui libère l'image id dans le terminal.
func Delete(id uint32) string {
	return fmt.Sprintf("\x1b_Ga=d,d=I,i=%d,q=2\x1b\\", id)
}

// Placeholder renvoie le texte (cols×rows cellules) que le terminal remplace
// par l'image id. L'identifiant est codé dans la couleur du texte (24 bits),
// la ligne et la colonne de chaque cellule dans deux diacritiques.
func Placeholder(id uint32, cols, rows int) string {
	cols, rows = min(cols, MaxCells), min(rows, MaxCells)
	color := fmt.Sprintf("\x1b[38;2;%d;%d;%dm", (id>>16)&0xff, (id>>8)&0xff, id&0xff)
	var b strings.Builder
	for y := 0; y < rows; y++ {
		if y > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(color)
		for x := 0; x < cols; x++ {
			b.WriteRune(placeholderRune)
			b.WriteRune(diacritics[y])
			b.WriteRune(diacritics[x])
		}
		b.WriteString("\x1b[39m")
	}
	return b.String()
}
