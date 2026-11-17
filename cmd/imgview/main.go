// imgview : explorateur d'images dans le terminal, construit avec Bubble Tea v2.
//
// Usage : imgview [répertoire]   (par défaut ~/Images)
//
//	imgview --version
package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"

	"github.com/laurentpoirierfr/imgview/internal/ui"
)

// Version est renseignée à la compilation (-ldflags "-X main.Version=…").
var Version = "dev"

func main() {
	root := defaultRoot()
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-v", "--version", "version":
			fmt.Println("imgview", Version)
			return
		case "-h", "--help":
			fmt.Println("Usage : imgview [répertoire]   (par défaut ~/Images)")
			fmt.Println("        imgview --version")
			return
		}
		root = os.Args[1]
	}
	abs, err := filepath.Abs(root)
	if err == nil {
		var st os.FileInfo
		st, err = os.Stat(abs)
		if err == nil && !st.IsDir() {
			err = fmt.Errorf("ce n'est pas un répertoire")
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "imgview: %s : %v\n", root, err)
		os.Exit(1)
	}

	if _, err := tea.NewProgram(ui.New(abs)).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "imgview:", err)
		os.Exit(1)
	}
}

func defaultRoot() string {
	if home, err := os.UserHomeDir(); err == nil {
		d := filepath.Join(home, "Images")
		if st, err := os.Stat(d); err == nil && st.IsDir() {
			return d
		}
	}
	return "."
}
