package imgrender

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"

	"golang.org/x/image/draw"
)

// PNG décode l'image et la renvoie encodée en PNG, réduite si besoin pour
// tenir dans maxW×maxH pixels (proportions conservées). Un fichier PNG déjà
// assez petit est renvoyé tel quel. Renvoie aussi les dimensions d'origine.
func PNG(path string, maxW, maxH int) (data []byte, w, h int, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, 0, err
	}
	defer f.Close()
	src, format, err := image.Decode(f)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("format non pris en charge : %w", err)
	}
	b := src.Bounds()
	w, h = b.Dx(), b.Dy()
	if w == 0 || h == 0 {
		return nil, 0, 0, fmt.Errorf("image vide")
	}

	scale := 1.0
	if maxW > 0 && maxH > 0 {
		scale = min(1, float64(maxW)/float64(w), float64(maxH)/float64(h))
	}
	if scale == 1 && format == "png" {
		data, err = os.ReadFile(path)
		return data, w, h, err
	}

	var img image.Image = src
	if scale < 1 {
		dst := image.NewRGBA(image.Rect(0, 0, max(1, int(float64(w)*scale)), max(1, int(float64(h)*scale))))
		draw.CatmullRom.Scale(dst, dst.Bounds(), src, b, draw.Src, nil)
		img = dst
	}
	var buf bytes.Buffer
	enc := png.Encoder{CompressionLevel: png.BestSpeed}
	if err := enc.Encode(&buf, img); err != nil {
		return nil, 0, 0, err
	}
	return buf.Bytes(), w, h, nil
}
