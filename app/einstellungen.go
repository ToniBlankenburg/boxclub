package app

import (
	"fmt"
	"net/http"

	"github.com/ToniBlankenburg/boxclub/i18n"
)

// spracheAendern schaltet die Anzeigesprache der gesamten Oberfläche um und
// schreibt sie in die Einstellungsdatei, damit sie den nächsten Start
// übersteht (ADR-0017). Das Auswahlfeld dafür steht auf der Verein-Seite
// (templates/verein.html) — die Rückkehr-Ansicht ist deshalb dieselbe Seite,
// jetzt in der neuen Sprache, und nicht die Mitgliederliste.
func (a *App) spracheAendern(w http.ResponseWriter, r *http.Request) {
	sprache := i18n.Sprache(r.FormValue("sprache"))
	if !sprache.Gueltig() {
		fehlerAntwort(w, fmt.Errorf("unbekannte Sprache %q", sprache))
		return
	}

	if err := a.spracheSetzen(sprache); err != nil {
		fehlerAntwort(w, err)
		return
	}

	a.vereinRendern(w, meldung{}, nil)
}
