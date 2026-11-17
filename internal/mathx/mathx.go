// Package mathx regroupe de petites fonctions numériques utilitaires.
package mathx

// Clamp borne v dans l'intervalle [lo, hi]. Si hi < lo, renvoie lo.
func Clamp(v, lo, hi int) int {
	if hi < lo {
		return lo
	}
	return min(max(v, lo), hi)
}
