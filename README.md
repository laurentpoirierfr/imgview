# imgview

Explorateur d'images pour le terminal, écrit en Go avec
[Bubble Tea v2](https://github.com/charmbracelet/bubbletea) et Lip Gloss v2.

Une arborescence des répertoires à gauche, une grille de miniatures à droite,
une barre de menus, un aperçu et un affichage plein écran.

## Pré-requis

### Obligatoires

| Outil | Version | Rôle |
|---|---|---|
| **Linux** | — | Seul système testé (appels `ioctl` Unix, `xdg-open`, `bash`). |
| **Go** | 1.27.3 ou plus récent (voir `go.mod`) | Compilation. |
| **make** | — | Facultatif, mais toutes les commandes ci-dessous l'utilisent. |
| **Un vrai terminal** | — | Bubble Tea a besoin d'un TTY. Le programme ne démarre pas dans une console sans TTY (par exemple la *Debug Console* de VS Code). |
| **Terminal en couleurs 24 bits** | — | Le rendu intégré utilise des séquences *truecolor*. |
| **Police Nerd Font** | — | Les icônes de dossiers et d'images sont des glyphes [Nerd Fonts](https://www.nerdfonts.com/). Sans cette police, elles apparaissent sous forme de carrés. |

### Optionnels (fonctionnalités en plus)

| Outil | Sans lui | Avec lui |
|---|---|---|
| **[chafa](https://hpjansson.org/chafa/)** | Miniatures dessinées par le rendu Go intégré (demi-blocs `▀`). Le plein écran (`f`) ne fonctionne pas. | Miniatures de meilleure qualité (touche `r` pour basculer) et plein écran haute définition. |
| **Terminal Kitty ou Ghostty** | Pas d'images HD. | Images en pleine résolution grâce au protocole graphique Kitty (touche `d`). |
| **xdg-open** (paquet `xdg-utils`) | « Ouvrir avec le bureau » (`o`) échoue. | Ouvre l'image dans la visionneuse par défaut. |
| **bash** et **tput** (paquet `ncurses`) | Le plein écran ne fonctionne pas. | Utilisés par le script du plein écran. |

Installation des outils optionnels :

```sh
# Debian / Ubuntu
sudo apt install chafa xdg-utils ncurses-bin

# Fedora
sudo dnf install chafa xdg-utils ncurses

# Arch
sudo pacman -S chafa xdg-utils ncurses
```

La fenêtre « Aide → À propos » indique si chafa a été détecté.

#### Détection du protocole Kitty

Les images HD sont activées automatiquement si l'une de ces conditions est vraie :

- `KITTY_WINDOW_ID` est défini ;
- `TERM_PROGRAM=ghostty` ;
- `TERM` vaut `xterm-kitty` ou `xterm-ghostty`.

Pour les désactiver de force : `IMGVIEW_NO_KITTY=1 imgview`.

Les images HD ont besoin que le terminal communique la taille de ses
cellules en pixels. Sinon, imgview suppose 10×20 pixels par cellule.

### Formats d'image

Le rendu intégré décode les formats PNG, JPEG, GIF, BMP, TIFF et WebP.
Les fichiers SVG et AVIF apparaissent dans la galerie, mais seul chafa peut
les afficher, et seulement si votre version de chafa les prend en charge.

## Compilation et lancement

Installation directe depuis GitHub :

```sh
go install github.com/laurentpoirierfr/imgview/cmd/imgview@latest
```

Depuis les sources :

```sh
git clone https://github.com/laurentpoirierfr/imgview.git
cd imgview
make build              # produit bin/imgview
./bin/imgview ~/Images  # répertoire par défaut : ~/Images, sinon .

make run ARGS=~/Images  # compile et lance en une étape
make install            # installe dans ~/.local/bin (PREFIX=… pour changer)
make help               # liste toutes les cibles
```

## Raccourcis principaux

| Touche | Action |
|---|---|
| `F10` / `Alt+F` | Menu |
| `Tab` | Changer de panneau |
| Flèches, `h j k l` | Se déplacer |
| `Entrée` | Déplier un répertoire / aperçu de l'image |
| `f` | Plein écran HD (chafa) |
| `o` | Ouvrir avec le bureau |
| `+` / `-` | Taille des miniatures |
| `.` | Afficher les répertoires cachés |
| `r` | Rendu chafa ou interne |
| `d` | Images HD (Kitty) on/off |
| `F5` | Actualiser |
| `?` | Aide |
| `q` / `Ctrl+C` | Quitter |

La souris est prise en charge : un clic sélectionne, un deuxième clic ouvre
l'aperçu, et la molette fait défiler.

## Organisation du code

```
cmd/imgview/        point d'entrée
internal/ui/        modèle Bubble Tea (vue, touches, souris, actions, HD)
internal/tree/      arborescence des répertoires
internal/gallery/   grille de miniatures
internal/menu/      menus et actions
internal/mathx/     utilitaires numériques
pkg/imgrender/      rendu image → texte (chafa ou demi-blocs), réutilisable
pkg/kittygfx/       protocole graphique Kitty, réutilisable
```
