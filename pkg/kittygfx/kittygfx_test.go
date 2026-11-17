package kittygfx

import (
	"encoding/base64"
	"regexp"
	"strings"
	"testing"
)

func TestFit(t *testing.T) {
	tests := []struct {
		name                           string
		imgW, imgH, maxC, maxR, cw, ch int
		wantC, wantR                   int
	}{
		{"carré remplissant la zone", 1000, 1000, 80, 40, 10, 20, 80, 40},
		{"carré limité en largeur", 1000, 1000, 50, 40, 10, 20, 50, 25},
		{"paysage limité en largeur", 2000, 1000, 50, 40, 10, 20, 50, 13},
		{"petite image non agrandie", 100, 100, 80, 40, 10, 20, 10, 5},
		{"taille de cellule inconnue", 1000, 1000, 50, 40, 0, 0, 50, 25},
		{"zone nulle", 100, 100, 0, 10, 10, 20, 0, 0},
	}
	for _, tt := range tests {
		c, r := Fit(tt.imgW, tt.imgH, tt.maxC, tt.maxR, tt.cw, tt.ch)
		if c != tt.wantC || r != tt.wantR {
			t.Errorf("%s : Fit = %d×%d, attendu %d×%d", tt.name, c, r, tt.wantC, tt.wantR)
		}
	}
}

func TestPlaceholder(t *testing.T) {
	const id = 0x123456
	s := Placeholder(id, 3, 2)
	lines := strings.Split(s, "\n")
	if len(lines) != 2 {
		t.Fatalf("%d lignes, attendu 2", len(lines))
	}
	if !strings.HasPrefix(lines[0], "\x1b[38;2;18;52;86m") {
		t.Errorf("couleur d'identifiant incorrecte : %q", lines[0][:20])
	}
	// Ligne 1, colonne 2 : U+10EEEE + diacritique[1] + diacritique[2].
	want := string([]rune{placeholderRune, diacritics[1], diacritics[2]})
	if !strings.Contains(lines[1], want) {
		t.Errorf("cellule (1,2) absente de la ligne 2")
	}
	if n := strings.Count(s, string(placeholderRune)); n != 6 {
		t.Errorf("%d cellules, attendu 6", n)
	}
}

func TestTransmitChunks(t *testing.T) {
	data := make([]byte, 10000) // ≈ 13 336 caractères en base64 → 4 morceaux
	for i := range data {
		data[i] = byte(i)
	}
	s := Transmit(7, data, 4, 3)
	re := regexp.MustCompile(`\x1b_G([^;]*);([^\x1b]*)\x1b\\`)
	parts := re.FindAllStringSubmatch(s, -1)
	if len(parts) != 4 {
		t.Fatalf("%d morceaux, attendu 4", len(parts))
	}
	if !strings.Contains(parts[0][1], "a=T,U=1,f=100") || !strings.Contains(parts[0][1], "i=7,c=4,r=3") {
		t.Errorf("en-tête du premier morceau incorrect : %s", parts[0][1])
	}
	var enc strings.Builder
	for i, p := range parts {
		more := strings.Contains(p[1], "m=1")
		if last := i == len(parts)-1; more == last {
			t.Errorf("morceau %d : m=1 attendu %v", i, !last)
		}
		if len(p[2]) > 4096 {
			t.Errorf("morceau %d trop long : %d", i, len(p[2]))
		}
		enc.WriteString(p[2])
	}
	got, err := base64.StdEncoding.DecodeString(enc.String())
	if err != nil || len(got) != len(data) {
		t.Fatalf("données reconstituées invalides (%d octets, err=%v)", len(got), err)
	}
}
