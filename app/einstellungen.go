package app

import (
	"fmt"
	"net/http"

	"github.com/ToniBlankenburg/boxclub/i18n"
)

// spracheAendern schaltet die Anzeigesprache der gesamten Oberfläche um und
// schreibt sie in die Einstellungsdatei, damit sie den nächsten Start
// übersteht (ADR-0017). Die Sprache ist eine Einstellung fürs ganze
// Programm, nicht für einen einzelnen Bereich — die Rückkehr-Ansicht ist
// deshalb dieselbe wie nach jeder anderen Aktion: die Mitgliederliste
// (App.listeRendern). Welcher Bereich vor dem Umschalten offen war, hält die
// App nicht serverseitig fest.
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

	a.listeRendern(w, meldung{})
}
