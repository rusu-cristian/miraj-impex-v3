// Package textkey builds the comparison key of a name or a column header
// (SPEC §10.2): the same key for "Hîrbovăț", "Hirbovat" and "HÎRBOVAŢ".
package textkey

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// unitWords are dropped when they stand alone ("Hrană g" → "hrana",
// "Ouat %" → "ouat").
var unitWords = map[string]bool{"g": true, "%": true}

// Key returns s decomposed (NFD) without its combining marks, so ș/ş and ț/ţ
// both fold to s/t, in lower case, with the spaces and the unit words "g" and
// "%" removed. It fills name_key for farms, guides, feed types and series
// (D111, D114, D116) and matches guide headers.
func Key(s string) string {
	var plain strings.Builder
	for _, r := range norm.NFD.String(s) {
		if !unicode.Is(unicode.Mn, r) {
			plain.WriteRune(unicode.ToLower(r))
		}
	}
	var b strings.Builder
	for _, word := range strings.Fields(plain.String()) {
		if !unitWords[word] {
			b.WriteString(word)
		}
	}
	return b.String()
}
