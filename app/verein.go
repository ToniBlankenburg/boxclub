package app

import (
	"net/http"

	"github.com/ToniBlankenburg/boxclub/service"
)

// Die Handler der Vereinsdaten (CONTEXT.md → Vereinsdaten). Es ist die
// schmalste Ansicht der App: ein Formular, ein Speichern, keine Liste und keine
// ID — es gibt einen Verein.
//
// Einen Rohwert-Typ für das Formular gibt es hier nicht, anders als beim
// Mitglied und beim Termin. Alle Angaben sind Text, keine wird umgerechnet und
// keine geprüft: die Service-Angabe *ist* die Eingabe, und ein zweiter Typ
// daneben wäre eine Übersetzung zwischen zwei Gleichen.

// vereinDaten speisen die Ansicht: die gespeicherten Angaben und optional eine
// Rückmeldung.
type vereinDaten struct {
	service.Vereinsdaten
	Meldung    meldung
	Navigation []navigationseintrag
}

// vereinsdatenLesen sammelt die Felder des abgeschickten Formulars ein. Die
// Feldnamen sind dieselben wie am Mitglied, wo es dieselbe Angabe gibt — es ist
// dieselbe Anschrift und dieselbe Art von Bankverbindung.
func vereinsdatenLesen(r *http.Request) service.Vereinsdaten {
	return service.Vereinsdaten{
		Name: r.FormValue("name"),
		Anschrift: service.Anschrift{
			Adresse:      r.FormValue("adresse"),
			Postleitzahl: r.FormValue("postleitzahl"),
			Ort:          r.FormValue("ort"),
		},
		Email:          r.FormValue("email"),
		Telefon:        r.FormValue("telefon"),
		IBAN:           r.FormValue("iban"),
		BIC:            r.FormValue("bic"),
		Kreditinstitut: r.FormValue("kreditinstitut"),
		Fusszeile:      r.FormValue("fusszeile"),
	}
}

// vereinFormular zeigt die gespeicherten Vereinsdaten zum Bearbeiten.
func (a *App) vereinFormular(w http.ResponseWriter, r *http.Request) {
	a.vereinRendern(w, meldung{})
}

// vereinSpeichern schreibt die Angaben und zeigt das Formular erneut.
//
// Abgewiesen werden kann hier nichts: es gibt keine Pflichtangabe und keine
// Regel, gegen die zu prüfen wäre (Ticket 23). Deshalb gibt es auch keinen
// Zweig, der die getippten Werte zurückgeben müsste — was gespeichert wurde,
// liest die Ansicht gleich wieder aus der Datenbank.
func (a *App) vereinSpeichern(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fehlerAntwort(w, err)
		return
	}

	if err := a.svc.SetVereinsdaten(vereinsdatenLesen(r)); err != nil {
		fehlerAntwort(w, err)
		return
	}

	a.vereinRendern(w, meldung{Text: "Die Vereinsdaten wurden gespeichert."})
}

// vereinRendern zeigt die Ansicht mit dem Stand aus der Datenbank.
//
// Gelesen wird auch direkt nach dem Speichern, statt die eingesammelten Werte
// wieder auszugeben: der Service schneidet umschließenden Leerraum ab, und das
// Formular soll zeigen, was tatsächlich gespeichert ist, und nicht, was getippt
// wurde.
func (a *App) vereinRendern(w http.ResponseWriter, m meldung) {
	daten, err := a.svc.GetVereinsdaten()
	if err != nil {
		fehlerAntwort(w, err)
		return
	}

	a.rendern(w, "verein", vereinDaten{
		Vereinsdaten: daten,
		Meldung:      m,
		Navigation:   navigation(bereichVerein),
	})
}
