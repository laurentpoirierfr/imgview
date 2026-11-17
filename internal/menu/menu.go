// Package menu décrit la barre de menus et les actions qu'elle déclenche.
package menu

type Action int

const (
	ActNone Action = iota
	ActPreview
	ActFullscreen
	ActOpenExternal
	ActRefresh
	ActQuit
	ActZoomIn
	ActZoomOut
	ActToggleHidden
	ActToggleRenderer
	ActHelp
	ActAbout
	ActToggleHD
)

type Item struct {
	Label string
	Key   string
	Act   Action
	Sep   bool
}

type Menu struct {
	Title string
	Items []Item
}

func Default() []Menu {
	return []Menu{
		{Title: "Fichier", Items: []Item{
			{Label: "Aperçu", Key: "Entrée", Act: ActPreview},
			{Label: "Plein écran HD", Key: "f", Act: ActFullscreen},
			{Label: "Ouvrir avec le bureau", Key: "o", Act: ActOpenExternal},
			{Sep: true},
			{Label: "Actualiser", Key: "F5", Act: ActRefresh},
			{Sep: true},
			{Label: "Quitter", Key: "q", Act: ActQuit},
		}},
		{Title: "Affichage", Items: []Item{
			{Label: "Agrandir les miniatures", Key: "+", Act: ActZoomIn},
			{Label: "Réduire les miniatures", Key: "-", Act: ActZoomOut},
			{Sep: true},
			{Label: "Répertoires cachés", Key: ".", Act: ActToggleHidden},
			{Label: "Rendu chafa", Key: "r", Act: ActToggleRenderer},
			{Label: "Images HD (Kitty)", Key: "d", Act: ActToggleHD},
		}},
		{Title: "Aide", Items: []Item{
			{Label: "Raccourcis clavier", Key: "?", Act: ActHelp},
			{Label: "À propos", Act: ActAbout},
		}},
	}
}

// NextItem renvoie l'élément sélectionnable suivant (en sautant les séparateurs).
func (mn Menu) NextItem(cur, delta int) int {
	n := len(mn.Items)
	for i := 1; i <= n; i++ {
		j := ((cur+delta*i)%n + n) % n
		if !mn.Items[j].Sep {
			return j
		}
	}
	return cur
}
