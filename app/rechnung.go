package app

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ToniBlankenburg/boxclub/i18n"
	"github.com/ToniBlankenburg/boxclub/service"
)

// Die Handler der Rechnung (CONTEXT.md → Rechnung, ADR-0009). Eine Rechnung
// ist ein erzeugtes PDF und kein offener Posten — nach dem Erstellen bleibt
// nur das PDF, und dieses Package bietet es genau einmal zum Speichern an. Es
// gibt weder eine Liste erzeugter Rechnungen noch einen zweiten Export dafür:
// wer den Datei-Dialog abbricht, muss die Rechnung noch einmal erstellen.
//
// Zwei Wege führen zum selben Formular: "am Mitglied" — POST
// /api/mitglied/{id}/rechnung, eingebettet unter dessen Stammdaten, Empfänger
// aus ihnen vorbelegt, das PDF wird dort abgelegt — und, für einen Externen,
// der nie eintritt, ohne Mitglied über den eigenen Bereich "Rechnung": der
// Empfänger ist dort leer, und das PDF wird nur ausgegeben (CONTEXT.md →
// Rechnung).

// rechnungPositionenImFormular ist die Zahl der Positionszeilen, die das
// Formular anbietet. Eine feste Zahl statt eines dynamischen "Zeile
// hinzufügen": ein Einzeltraining braucht selten mehr als zwei, drei Zeilen,
// und eine leere Zeile zählt beim Erstellen nicht mit (siehe
// rechnungEingabe.alsRechnungEingabe) — wer mehr braucht, tippt einfach weiter,
// bis die letzte Zeile auch voll ist, und erstellt danach eine zweite Rechnung.
const rechnungPositionenImFormular = 5

// rechnungPositionEingabe hält die Rohwerte einer Positionszeile.
type rechnungPositionEingabe struct {
	Bezeichnung string
	Menge       string
	Einzelpreis string
}

// rechnungEingabe hält die Rohwerte des Rechnungsformulars.
type rechnungEingabe struct {
	Name         string
	Adresse      string
	Postleitzahl string
	Ort          string

	Nummer         string
	Rechnungsdatum string
	Zahlungsziel   string
	Steuersatz     string

	Positionen []rechnungPositionEingabe
}

// rechnungEingabeLesen sammelt die Rohwerte des abgeschickten Formulars ein.
// Alle Positionszeilen heißen gleich und kommen als Liste an — wie die
// angekreuzten Trainingstermine im Mitgliedsformular —, ausgerichtet über den
// Index und nicht über eindeutige Feldnamen: das erspart dem Formular, beim
// Anzeigen weiterer Zeilen Namen hochzuzählen.
func rechnungEingabeLesen(r *http.Request) rechnungEingabe {
	bezeichnungen := r.Form["position-bezeichnung"]
	mengen := r.Form["position-menge"]
	einzelpreise := r.Form["position-einzelpreis"]

	positionen := make([]rechnungPositionEingabe, len(bezeichnungen))
	for i, bezeichnung := range bezeichnungen {
		p := rechnungPositionEingabe{Bezeichnung: bezeichnung}
		if i < len(mengen) {
			p.Menge = mengen[i]
		}
		if i < len(einzelpreise) {
			p.Einzelpreis = einzelpreise[i]
		}
		positionen[i] = p
	}

	return rechnungEingabe{
		Name:           r.FormValue("empfaenger_name"),
		Adresse:        r.FormValue("empfaenger_adresse"),
		Postleitzahl:   r.FormValue("empfaenger_postleitzahl"),
		Ort:            r.FormValue("empfaenger_ort"),
		Nummer:         r.FormValue("nummer"),
		Rechnungsdatum: r.FormValue("rechnungsdatum"),
		Zahlungsziel:   r.FormValue("zahlungsziel"),
		Steuersatz:     r.FormValue("steuersatz"),
		Positionen:     positionen,
	}
}

// alsRechnungEingabe übersetzt die Rohwerte in die Service-Eingabe und sammelt
// dabei alle Parse-Fehler, statt beim ersten abzubrechen — wie
// formularEingabe.alsNeuesMitglied es für das Mitgliedsformular tut. Leere
// Pflichtfelder meldet nicht diese Funktion, sondern der Service; hier wird nur
// geprüft, was *gefüllt* und trotzdem unlesbar ist.
func (e rechnungEingabe) alsRechnungEingabe(sprache i18n.Sprache) (service.RechnungEingabe, []string) {
	var fehler []string

	rechnung := service.RechnungEingabe{
		Empfaenger: service.Empfaenger{
			Name:      e.Name,
			Anschrift: service.Anschrift{Adresse: e.Adresse, Postleitzahl: e.Postleitzahl, Ort: e.Ort},
		},
		Nummer: e.Nummer,
	}

	if e.Rechnungsdatum != "" {
		d, err := time.Parse(isoDatum, e.Rechnungsdatum)
		if err != nil {
			fehler = append(fehler, i18n.Text(sprache, "rechnung.fehler_rechnungsdatum"))
		} else {
			rechnung.Rechnungsdatum = d
		}
	}

	if e.Zahlungsziel != "" {
		d, err := time.Parse(isoDatum, e.Zahlungsziel)
		if err != nil {
			fehler = append(fehler, i18n.Text(sprache, "rechnung.fehler_zahlungsziel"))
		} else {
			rechnung.Zahlungsziel = d
		}
	}

	steuersatz, err := service.SteuersatzAusText(e.Steuersatz)
	fehler = append(fehler, validierungsMeldungen(sprache, err)...)
	rechnung.SteuersatzProzent = steuersatz

	for i, p := range e.Positionen {
		if position, meldungen := p.alsRechnungsposition(i+1, sprache); meldungen != nil {
			fehler = append(fehler, meldungen...)
		} else if position != nil {
			rechnung.Positionen = append(rechnung.Positionen, *position)
		}
	}

	return rechnung, fehler
}

// alsRechnungsposition übersetzt eine Zeile. Eine ganz leere Zeile liefert
// weder eine Position noch eine Meldung — sie kommt von den ungenutzten
// Feldern, die das Formular vorsorglich anbietet (rechnungPositionenImFormular),
// und zählt beim Erstellen nicht mit.
func (p rechnungPositionEingabe) alsRechnungsposition(nr int, sprache i18n.Sprache) (*service.Rechnungsposition, []string) {
	if strings.TrimSpace(p.Bezeichnung) == "" && strings.TrimSpace(p.Menge) == "" && strings.TrimSpace(p.Einzelpreis) == "" {
		return nil, nil
	}

	var fehler []string
	position := service.Rechnungsposition{Bezeichnung: p.Bezeichnung}

	if menge, err := strconv.ParseInt(strings.TrimSpace(p.Menge), 10, 64); err != nil {
		fehler = append(fehler, i18n.Text(sprache, "rechnung.fehler_position_menge", nr))
	} else {
		position.Menge = menge
	}

	cents, err := service.EinzelpreisAusEuro(p.Einzelpreis)
	for _, meldung := range validierungsMeldungen(sprache, err) {
		fehler = append(fehler, i18n.Text(sprache, "rechnung.fehler_position_praefix", nr, meldung))
	}
	position.EinzelpreisCents = cents

	if len(fehler) > 0 {
		return nil, fehler
	}

	return &position, nil
}

// validierungsMeldungen liest die Meldungen aus einem Service-Fehler, sofern es
// einer ist — der Rückweg zu strconv.ParseInt & Co., die App-Handler sonst nie
// mit *service.ValidierungsFehler sehen, weil sie ihn nur weiterreichen. Ein
// nil-Fehler liefert nichts. Die Schlüssel des Service werden hier übersetzt —
// der einzige Weg, auf dem sie in diese Datei gelangen.
func validierungsMeldungen(sprache i18n.Sprache, err error) []string {
	if err == nil {
		return nil
	}

	var validierung *service.ValidierungsFehler
	if errors.As(err, &validierung) {
		fehler := make([]string, len(validierung.Meldungen))
		for i, m := range validierung.Meldungen {
			fehler[i] = i18n.Text(sprache, m.Schluessel, m.Args...)
		}

		return fehler
	}

	return []string{err.Error()}
}

// rechnungEingabeLeer ist ein frisches Formular: Rechnungsdatum heute,
// Zahlungsziel 14 Tage später (service.Zahlungsziel), Steuersatz die Vorgabe —
// alles überschreibbar (Ticket 25, AC).
func rechnungEingabeLeer() rechnungEingabe {
	heute := time.Now()

	return rechnungEingabe{
		Rechnungsdatum: heute.Format(isoDatum),
		Zahlungsziel:   service.Zahlungsziel(heute).Format(isoDatum),
		Steuersatz:     service.SteuersatzAlsText(service.StandardSteuersatz),
		Positionen:     make([]rechnungPositionEingabe, rechnungPositionenImFormular),
	}
}

// rechnungEingabeVorbelegt ist dasselbe mit dem Empfänger aus den Stammdaten
// eines Mitglieds — frei überschreibbar (CONTEXT.md → Rechnung).
func rechnungEingabeVorbelegt(m service.Mitglied) rechnungEingabe {
	eingabe := rechnungEingabeLeer()
	eingabe.Name = strings.TrimSpace(m.Vorname + " " + m.Nachname)
	eingabe.Adresse = m.Anschrift.Adresse
	eingabe.Postleitzahl = m.Anschrift.Postleitzahl
	eingabe.Ort = m.Anschrift.Ort

	return eingabe
}

// rechnungDaten speist das Formular-Template.
type rechnungDaten struct {
	// MitgliedID ist gesetzt, wenn die Rechnung am Mitglied entsteht — dann ist
	// der Empfänger vorbelegt und das PDF wird dort abgelegt. Ohne Mitglied
	// (nil) ist der Empfänger frei getippt, und das PDF wird nur ausgegeben: es
	// gibt niemanden, an dem es hängen könnte (CONTEXT.md → Rechnung).
	MitgliedID *int64
	AktionsURL string

	Eingabe rechnungEingabe
	Meldung meldung
	Fehler  []string

	// Navigation ist nur beim eigenständigen Bereich gesetzt (ohne Mitglied);
	// eingebettet unter den Stammdaten eines Mitglieds bleibt die Kopfzeile
	// stehen, wie sie ist (siehe navigation).
	Navigation []navigationseintrag
}

// rechnungBeschriftungen löst die Textbausteine des PDFs (ADR-0009,
// service.RechnungBeschriftungen) in sprache auf — die zum Erstellzeitpunkt
// aktive Anzeigesprache, damit eine erzeugte Rechnung sie trägt (Ticket 06).
// SpalteBezeichnung/-Menge/-Einzelpreis greifen bewusst auf dieselben
// feld.*-Schlüssel zurück wie das Formular selbst, statt sie zu verdoppeln.
func rechnungBeschriftungen(sprache i18n.Sprache) service.RechnungBeschriftungen {
	return service.RechnungBeschriftungen{
		TelefonPraefix:         i18n.Text(sprache, "rechnung.pdf.telefon_praefix"),
		EMailPraefix:           i18n.Text(sprache, "rechnung.pdf.email_praefix"),
		RechnungsnummerPraefix: i18n.Text(sprache, "rechnung.pdf.rechnungsnummer_praefix"),
		RechnungsdatumPraefix:  i18n.Text(sprache, "rechnung.pdf.rechnungsdatum_praefix"),
		ZahlungszielPraefix:    i18n.Text(sprache, "rechnung.pdf.zahlungsziel_praefix"),
		TitelPraefix:           i18n.Text(sprache, "rechnung.pdf.titel_praefix"),
		SpalteBezeichnung:      i18n.Text(sprache, "feld.position_bezeichnung"),
		SpalteMenge:            i18n.Text(sprache, "feld.menge"),
		SpalteEinzelpreis:      i18n.Text(sprache, "feld.einzelpreis_netto"),
		SpalteSumme:            i18n.Text(sprache, "rechnung.pdf.spalte_summe"),
		NettoPraefix:           i18n.Text(sprache, "rechnung.pdf.netto_praefix"),
		SteuerVorlage:          i18n.Text(sprache, "rechnung.pdf.steuer_vorlage"),
		GesamtbetragPraefix:    i18n.Text(sprache, "rechnung.pdf.gesamtbetrag_praefix"),
		IBANPraefix:            i18n.Text(sprache, "rechnung.pdf.iban_praefix"),
		BICPraefix:             i18n.Text(sprache, "rechnung.pdf.bic_praefix"),
	}
}

// rechnungAktionsURL ist die Adresse, an die das Formular abschickt.
func rechnungAktionsURL(mitgliedID *int64) string {
	if mitgliedID == nil {
		return "/api/rechnung"
	}

	return fmt.Sprintf("/api/mitglied/%d/rechnung", *mitgliedID)
}

// rechnungBereichAmMitglied baut den Block, der unter den Stammdaten eines
// Mitglieds steht — wie vertragsbloecke für die Verträge.
func rechnungBereichAmMitglied(m service.Mitglied) rechnungDaten {
	return rechnungDaten{
		MitgliedID: &m.ID,
		AktionsURL: rechnungAktionsURL(&m.ID),
		Eingabe:    rechnungEingabeVorbelegt(m),
	}
}

// rechnungFormular zeigt den eigenständigen Bereich für einen Empfänger ohne
// Mitglied.
func (a *App) rechnungFormular(w http.ResponseWriter, r *http.Request) {
	a.rechnungFormularRendern(w, nil, rechnungEingabeLeer(), nil, meldung{})
}

// rechnungErstellenAmMitglied erstellt die Rechnung am Mitglied, dessen ID im
// Pfad steht.
func (a *App) rechnungErstellenAmMitglied(w http.ResponseWriter, r *http.Request) {
	id, ok := mitgliedID(w, r)
	if !ok {
		return
	}

	a.rechnungVerarbeiten(w, r, &id)
}

// rechnungErstellen erstellt die Rechnung für einen Empfänger ohne Mitglied.
func (a *App) rechnungErstellen(w http.ResponseWriter, r *http.Request) {
	a.rechnungVerarbeiten(w, r, nil)
}

// rechnungVerarbeiten ist beiden Wegen gemeinsam: Formular lesen, Rechnung
// erstellen, das PDF anbieten. mitgliedID entscheidet, ob dabei ein Dokument
// entsteht — der Service kennt dieselbe Unterscheidung (RechnungErstellen).
func (a *App) rechnungVerarbeiten(w http.ResponseWriter, r *http.Request, mitgliedID *int64) {
	if err := r.ParseForm(); err != nil {
		fehlerAntwort(w, err)
		return
	}

	eingabe := rechnungEingabeLesen(r)
	rechnung, fehler := eingabe.alsRechnungEingabe(a.Sprache())

	if len(fehler) == 0 {
		pdf, err := a.svc.RechnungErstellen(mitgliedID, rechnung, rechnungBeschriftungen(a.Sprache()))
		if err == nil {
			a.rechnungAnbieten(w, mitgliedID, rechnung.Nummer, pdf)
			return
		}

		var validierung *service.ValidierungsFehler
		switch {
		case errors.As(err, &validierung):
			fehler = a.uebersetzeMeldungen(validierung.Meldungen)
		case mitgliedID != nil:
			a.nichtGefundenOderFehler(w, err)
			return
		default:
			fehlerAntwort(w, err)
			return
		}
	}

	a.rechnungFormularRendern(w, mitgliedID, eingabe, fehler, meldung{})
}

// rechnungAnbieten bietet ein frisch erstelltes PDF über den Datei-Dialog zum
// Speichern an — denselben Weg, den der Vertragsexport nutzt (app/dokument.go),
// und aus demselben Grund: das WebView von Wails kennt keine Downloads.
//
// Anders als beim Vertrag gibt es hier keinen zweiten Versuch: eine Rechnung
// wird nicht noch einmal exportiert, sie wird noch einmal erstellt. Wer den
// Dialog abbricht, bekommt das gesagt — für einen Externen ist das PDF dann
// weg, für ein Mitglied bleibt es zwar am Datensatz liegen, aber ohne einen
// erneuten Export von hier aus nicht mehr zu erreichen.
func (a *App) rechnungAnbieten(w http.ResponseWriter, mitgliedID *int64, nummer string, pdf []byte) {
	sprache := a.Sprache()
	vorschlag := i18n.Text(sprache, "rechnung.dateiname_praefix") + strings.TrimSpace(nummer) + ".pdf"

	if a.speicherziel == nil {
		a.rechnungNeuRendern(w, mitgliedID, meldung{
			Text:    i18n.Text(sprache, "rechnung.dialog_nicht_verfuegbar"),
			Warnung: true,
		})

		return
	}

	ziel, err := a.speicherziel(vorschlag, pdfFilterBeschriftung, pdfFilterMuster)
	if err != nil {
		fehlerAntwort(w, fmt.Errorf("speicherort erfragen: %w", err))
		return
	}

	if ziel == "" {
		text := i18n.Text(sprache, "rechnung.kein_speicherort_ohne_mitglied")
		if mitgliedID != nil {
			text = i18n.Text(sprache, "rechnung.kein_speicherort_mit_mitglied")
		}

		a.rechnungNeuRendern(w, mitgliedID, meldung{Text: text, Warnung: true})

		return
	}

	// 0600: eine Rechnung nennt Namen und Anschrift des Empfängers und geht
	// niemanden sonst an — dieselbe Überlegung wie beim Vertrag.
	if err := os.WriteFile(ziel, pdf, 0o600); err != nil {
		a.rechnungNeuRendern(w, mitgliedID, meldung{
			Text:    i18n.Text(sprache, "rechnung.speichern_fehlgeschlagen", err),
			Warnung: true,
		})

		return
	}

	a.rechnungNeuRendern(w, mitgliedID, meldung{Text: i18n.Text(sprache, "rechnung.gespeichert_nach", ziel)})
}

// rechnungNeuRendern zeigt ein frisches Formular mit einer Rückmeldung — nach
// dem Erstellen ist die alte Eingabe abgearbeitet, eine zweite Rechnung fängt
// bei den Vorgaben wieder an.
func (a *App) rechnungNeuRendern(w http.ResponseWriter, mitgliedID *int64, m meldung) {
	if mitgliedID == nil {
		a.rechnungFormularRendern(w, nil, rechnungEingabeLeer(), nil, m)
		return
	}

	mitglied, err := a.svc.Get(*mitgliedID)
	if err != nil {
		a.nichtGefundenOderFehler(w, err)
		return
	}

	a.rechnungFormularRendern(w, mitgliedID, rechnungEingabeVorbelegt(mitglied), nil, m)
}

// rechnungFormularRendern zeigt den Rechnungsbereich mit den übergebenen
// Werten — bei einer abgelehnten Eingabe die getippten, sonst frische.
func (a *App) rechnungFormularRendern(w http.ResponseWriter, mitgliedID *int64, eingabe rechnungEingabe, fehler []string, m meldung) {
	var nav []navigationseintrag
	if mitgliedID == nil {
		nav = a.navigation(bereichRechnung)
	}

	a.rendern(w, "rechnung-bereich", rechnungDaten{
		MitgliedID: mitgliedID,
		AktionsURL: rechnungAktionsURL(mitgliedID),
		Eingabe:    eingabe,
		Meldung:    m,
		Fehler:     fehler,
		Navigation: nav,
	})
}
