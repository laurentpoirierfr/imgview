package ui

import (
	"fmt"
	"os/exec"

	tea "charm.land/bubbletea/v2"

	"github.com/laurentpoirierfr/imgview/pkg/imgrender"
)

// renderMsg arrive quand une image a été convertie en texte.
type renderMsg struct {
	key  string
	path string
	out  string
	info imgrender.Info
	err  error
}

// Limite le nombre de conversions simultanées.
var renderSem = make(chan struct{}, 4)

func renderKey(path string, w, h int, k imgrender.Kind) string {
	return fmt.Sprintf("%s|%dx%d|%d", path, w, h, k)
}

func renderCmd(key, path string, w, h int, k imgrender.Kind) tea.Cmd {
	return func() tea.Msg {
		renderSem <- struct{}{}
		defer func() { <-renderSem }()

		msg := renderMsg{key: key, path: path, info: imgrender.ReadInfo(path)}
		msg.out, msg.err = imgrender.Render(path, w, h, k)
		return msg
	}
}

// execDoneMsg arrive au retour d'un programme externe (plein écran).
type execDoneMsg struct{ err error }

// fullscreenCmd suspend l'interface et affiche l'image en plein écran avec
// chafa, qui choisit le meilleur protocole du terminal (Kitty, Sixel…).
func fullscreenCmd(path string) tea.Cmd {
	const script = `clear
cols=$(tput cols); lines=$(tput lines)
chafa --animate off --align mid,mid --size "${cols}x$((lines - 2))" -- "$1"
printf '\n\033[2m  %s — appuie sur une touche pour revenir\033[0m' "$(basename "$1")"
read -rsn1
printf '\033_Ga=d\033\\'
clear`
	c := exec.Command("bash", "-c", script, "imgview", path)
	return tea.ExecProcess(c, func(err error) tea.Msg { return execDoneMsg{err} })
}

// openExternal ouvre l'image dans la visionneuse par défaut du bureau.
func openExternal(path string) error {
	cmd := exec.Command("xdg-open", path)
	cmd.Stdout, cmd.Stderr = nil, nil
	if err := cmd.Start(); err != nil {
		return err
	}
	go cmd.Wait()
	return nil
}
