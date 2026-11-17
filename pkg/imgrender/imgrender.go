// Package imgrender convertit des images en texte coloré pour le terminal,
// soit avec chafa (si installé), soit avec un rendu Go intégré en demi-blocs.
//
// Ce paquet ne dépend pas de Bubble Tea et peut être réutilisé ailleurs.
package imgrender

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// Kind choisit comment une image est convertie en texte.
type Kind int

const (
	Chafa    Kind = iota // chafa (meilleure qualité, si installé)
	Internal             // rendu Go intégré en demi-blocs ▀
)

func (k Kind) String() string {
	if k == Chafa {
		return "chafa"
	}
	return "interne"
}

var imageExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true,
	".bmp": true, ".tif": true, ".tiff": true, ".svg": true, ".avif": true,
}

// IsImage indique si le nom de fichier a une extension d'image connue.
func IsImage(name string) bool {
	return imageExts[strings.ToLower(filepath.Ext(name))]
}

// Info décrit une image sur disque.
type Info struct {
	W, H   int
	Size   int64
	Format string
	OK     bool // dimensions lues avec succès
}

// ReadInfo lit la taille du fichier et, si possible, les dimensions de l'image.
func ReadInfo(path string) Info {
	var info Info
	st, err := os.Stat(path)
	if err != nil {
		return info
	}
	info.Size = st.Size()
	info.Format = strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	f, err := os.Open(path)
	if err != nil {
		return info
	}
	defer f.Close()
	if cfg, format, err := image.DecodeConfig(f); err == nil {
		info.W, info.H, info.Format, info.OK = cfg.Width, cfg.Height, format, true
	}
	return info
}

// Render convertit l'image avec le moteur demandé, en se rabattant sur le
// rendu interne si chafa échoue.
func Render(path string, w, h int, k Kind) (string, error) {
	if k == Chafa {
		if out, err := RenderChafa(path, w, h); err == nil {
			return out, nil
		}
	}
	return RenderInternal(path, w, h)
}

// ChafaAvailable indique si l'exécutable chafa est dans le PATH.
func ChafaAvailable() bool {
	_, err := exec.LookPath("chafa")
	return err == nil
}

// Séquences que chafa peut émettre et qui perturberaient l'affichage
// (curseur masqué, retour à la ligne automatique…).
var privateModeRe = regexp.MustCompile(`\x1b\[\?[0-9;]*[hl]`)

// RenderChafa convertit l'image avec chafa en w×h cellules.
func RenderChafa(path string, w, h int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "chafa",
		"--format", "symbols", "--colors", "full", "--animate", "off", "--polite", "on",
		"--size", fmt.Sprintf("%dx%d", w, h), "--", path)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("chafa: %w", err)
	}
	s := privateModeRe.ReplaceAllString(string(out), "")
	s = strings.ReplaceAll(s, "\r", "")
	return strings.TrimRight(s, "\n"), nil
}

// RenderInternal dessine l'image avec des demi-blocs « ▀ » : chaque cellule
// affiche deux pixels (haut = couleur du texte, bas = couleur du fond).
func RenderInternal(path string, w, h int) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	src, _, err := image.Decode(f)
	if err != nil {
		return "", fmt.Errorf("format non pris en charge : %w", err)
	}

	b := src.Bounds()
	if b.Dx() == 0 || b.Dy() == 0 || w <= 0 || h <= 0 {
		return "", fmt.Errorf("image vide")
	}
	ratio := float64(b.Dx()) / float64(b.Dy())
	pw, ph := w, int(float64(w)/ratio+0.5)
	if ph > 2*h {
		ph = 2 * h
		pw = int(float64(ph)*ratio + 0.5)
	}
	pw, ph = max(pw, 1), max(ph, 1)

	dst := image.NewRGBA(image.Rect(0, 0, pw, ph))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, b, draw.Over, nil)

	var sb strings.Builder
	for y := 0; y < ph; y += 2 {
		for x := 0; x < pw; x++ {
			top := dst.RGBAAt(x, y)
			bot := color.RGBA{}
			if y+1 < ph {
				bot = dst.RGBAAt(x, y+1)
			}
			tOK, bOK := top.A >= 128, bot.A >= 128
			switch {
			case tOK && bOK:
				fmt.Fprintf(&sb, "\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm▀", top.R, top.G, top.B, bot.R, bot.G, bot.B)
			case tOK:
				fmt.Fprintf(&sb, "\x1b[49m\x1b[38;2;%d;%d;%dm▀", top.R, top.G, top.B)
			case bOK:
				fmt.Fprintf(&sb, "\x1b[49m\x1b[38;2;%d;%d;%dm▄", bot.R, bot.G, bot.B)
			default:
				sb.WriteString("\x1b[0m ")
			}
		}
		sb.WriteString("\x1b[0m")
		if y+2 < ph {
			sb.WriteByte('\n')
		}
	}
	return sb.String(), nil
}

// HumanSize formate une taille en octets (o, Ko, Mo, Go).
func HumanSize(n int64) string {
	switch {
	case n >= 1<<30:
		return fmt.Sprintf("%.1f Go", float64(n)/(1<<30))
	case n >= 1<<20:
		return fmt.Sprintf("%.1f Mo", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.0f Ko", float64(n)/(1<<10))
	}
	return fmt.Sprintf("%d o", n)
}
