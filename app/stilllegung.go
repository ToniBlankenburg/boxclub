package app

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/ToniBlankenburg/boxclub/service"
)

// stilllegungDaten speist den Block „Stilllegungen“ im Mitgliedsformular
// (templates/mitglieder_liste.html → mitglied-stilllegungen). Er zeigt die
// Stilllegungen der maßgeblichen Mitgliedschaft als Historie, jede mit eigenen
// Feldern und Schaltflächen, und darunter das Erfassen einer neuen.
type stilllegungDaten struct {
	MitgliedID int64

	// Eintraege sind die Zeilen der Historie, aufsteigend nach Beginn.
	Eintraege []stilllegungZeile

	// Neu ist die Eingabe der neuen Stilllegung, damit eine abgelehnte nicht
	// verloren geht.
	Neu stilllegungEingabe

	// NeuAnbieten sagt, ob eine neue erfasst werden kann: nur an einem laufenden
	// Zeitraum, und nur solange keine Stilllegung läuft.
	NeuAnbieten bool

	// Fehler gelten der Eingabe in NeuFehler bzw. der Zeile, die FehlerID nennt.
	Fehler   []string
	FehlerID int64
}

type stilllegungZeile struct {
	ID      int64
	Eingabe stilllegungEingabe

	// Laeuft und Bevorstehend steuern die Schaltfläche „Wieder aktiv“: sie gibt
	// es nur für eine laufende oder noch nicht begonnene Stilllegung.
	Laeuft       bool
	Bevorstehend bool
}

type stilllegungEingabe struct {
	Von, Bis, Grund string
}

func stilllegungEingabeLesen(r *http.Request) stilllegungEingabe {
	return stilllegungEingabe{
		Von:   r.FormValue("von"),
		Bis:   r.FormValue("bis"),
		Grund: r.FormValue("grund"),
	}
}

// alsAngabe liest die Datumsfelder. Ein leeres Feld bleibt der Nullwert — die
// Pflicht prüft der Service. Nur ein gefülltes, unlesbares Feld ist hier ein Fehler.
func (e stilllegungEingabe) alsAngabe() (service.Stilllegungsangabe, bool) {
	angabe := service.Stilllegungsangabe{Grund: e.Grund}
	lesbar := true

	for _, f := range []struct {
		roh  string
		ziel *time.Time
	}{{e.Von, &angabe.Von}, {e.Bis, &angabe.Bis}} {
		if f.roh == "" {
			continue
		}
		d, err := time.Parse(isoDatum, f.roh)
		if err != nil {
			lesbar = false

			continue
		}
		*f.ziel = d
	}

	return angabe, lesbar
}

// stilllegungDatenZu baut den Block aus dem Stand des Mitglieds.
func stilllegungDatenZu(m service.Mitglied, neu stilllegungEingabe, fehlerID int64, fehler []string) stilllegungDaten {
	daten := stilllegungDaten{MitgliedID: m.ID, Neu: neu, Fehler: fehler, FehlerID: fehlerID}
	heute := time.Now().Format(isoDatum)

	if ms := m.LetzteMitgliedschaft(); ms != nil {
		daten.NeuAnbieten = !ms.Status().Ausgetreten() && !ms.Stillgelegt()
		for _, st := range ms.Stilllegungen {
			von, bis := st.Von.Format(isoDatum), st.Bis.Format(isoDatum)
			daten.Eintraege = append(daten.Eintraege, stilllegungZeile{
				ID:           st.ID,
				Eingabe:      stilllegungEingabe{Von: von, Bis: bis, Grund: st.Grund},
				Laeuft:       von <= heute && heute <= bis,
				Bevorstehend: von > heute,
			})
		}
	}

	return daten
}

// stilllegungenRendern liest den Stand neu und antwortet mit dem Block.
func (a *App) stilllegungenRendern(w http.ResponseWriter, mitgliedID int64, neu stilllegungEingabe, fehlerID int64, fehler []string) {
	m, err := a.svc.Get(mitgliedID)
	if err != nil {
		a.nichtGefundenOderFehler(w, err)

		return
	}

	a.rendern(w, "mitglied-stilllegungen", stilllegungDatenZu(m, neu, fehlerID, fehler))
}

func (a *App) stilllegungenBlock(w http.ResponseWriter, r *http.Request) {
	id, ok := mitgliedID(w, r)
	if !ok {
		return
	}

	a.stilllegungenRendern(w, id, stilllegungEingabe{}, 0, nil)
}

func (a *App) stilllegungAnlegen(w http.ResponseWriter, r *http.Request) {
	id, ok := mitgliedID(w, r)
	if !ok {
		return
	}

	eingabe := stilllegungEingabeLesen(r)
	angabe, lesbar := eingabe.alsAngabe()
	if !lesbar {
		a.stilllegungenRendern(w, id, eingabe, 0, []string{"Datum ist kein gültiges Datum."})

		return
	}

	if _, err := a.svc.CreateStilllegung(id, angabe); err != nil {
		a.stilllegungFehler(w, id, 0, eingabe, err)

		return
	}

	a.stilllegungenRendern(w, id, stilllegungEingabe{}, 0, nil)
}

func (a *App) stilllegungAendern(w http.ResponseWriter, r *http.Request) {
	id, ok := pfadID(w, r, "Stilllegung-ID")
	if !ok {
		return
	}

	eingabe := stilllegungEingabeLesen(r)
	mitgliedID, ok := a.stilllegungMitglied(w, r)
	if !ok {
		return
	}

	angabe, lesbar := eingabe.alsAngabe()
	if !lesbar {
		a.stilllegungenRendern(w, mitgliedID, stilllegungEingabe{}, id, []string{"Datum ist kein gültiges Datum."})

		return
	}

	if err := a.svc.UpdateStilllegung(id, angabe); err != nil {
		a.stilllegungFehler(w, mitgliedID, id, stilllegungEingabe{}, err)

		return
	}

	a.stilllegungenRendern(w, mitgliedID, stilllegungEingabe{}, 0, nil)
}

func (a *App) stilllegungBeenden(w http.ResponseWriter, r *http.Request) {
	a.stilllegungAktion(w, r, a.svc.BeendeStilllegung)
}

func (a *App) stilllegungLoeschen(w http.ResponseWriter, r *http.Request) {
	a.stilllegungAktion(w, r, a.svc.DeleteStilllegung)
}

func (a *App) stilllegungAktion(w http.ResponseWriter, r *http.Request, aktion func(int64) error) {
	id, ok := pfadID(w, r, "Stilllegung-ID")
	if !ok {
		return
	}
	mitgliedID, ok := a.stilllegungMitglied(w, r)
	if !ok {
		return
	}

	if err := aktion(id); err != nil {
		a.stilllegungFehler(w, mitgliedID, id, stilllegungEingabe{}, err)

		return
	}

	a.stilllegungenRendern(w, mitgliedID, stilllegungEingabe{}, 0, nil)
}

// stilllegungMitglied liest die Mitglied-ID, die der Block mitschickt: die
// Antwort ist wieder der Block dieses Mitglieds.
func (a *App) stilllegungMitglied(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.FormValue("mitglied"), 10, 64)
	if err != nil {
		http.Error(w, "ungültige Mitglied-ID", http.StatusBadRequest)

		return 0, false
	}

	return id, true
}

func (a *App) stilllegungFehler(w http.ResponseWriter, mitgliedID, fehlerID int64, neu stilllegungEingabe, err error) {
	var validierung *service.ValidierungsFehler
	if errors.As(err, &validierung) {
		a.stilllegungenRendern(w, mitgliedID, neu, fehlerID, a.uebersetzeMeldungen(validierung.Meldungen))

		return
	}

	a.veralteteAnsichtOderFehler(w, err)
}
