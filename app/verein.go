package app

import (
	"errors"
	"io"
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

// maxVereinUpload begrenzt, was das Formular beim Logo-Upload überhaupt
// entgegennimmt — derselbe Gedanke wie maxDokumentUpload: etwas Luft über der
// Grenze des Service, damit eine Datei knapp über MaxLogoBytes die genaue
// Meldung des Service bekommt und nicht die grobe von hier.
const maxVereinUpload = service.MaxLogoBytes + (1 << 20)

// vereinDaten speisen die Ansicht: die gespeicherten Angaben und optional eine
// Rückmeldung.
type vereinDaten struct {
	service.Vereinsdaten
	Meldung    meldung
	Navigation []navigationseintrag

	// Fehler sind die Gründe, aus denen ein hochgeladenes Logo nicht
	// angenommen wurde — dieselbe Art Auskunft wie beim Vertrag (app/dokument.go):
	// zu groß und falsches Format können beide zugleich zutreffen.
	Fehler []string
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
		Email:                      r.FormValue("email"),
		Telefon:                    r.FormValue("telefon"),
		IBAN:                       r.FormValue("iban"),
		BIC:                        r.FormValue("bic"),
		Kreditinstitut:             r.FormValue("kreditinstitut"),
		Fusszeile:                  r.FormValue("fusszeile"),
		MoneyMoneyVerwendungszweck: r.FormValue("moneymoney_verwendungszweck"),
	}
}

// vereinFormular zeigt die gespeicherten Vereinsdaten zum Bearbeiten.
func (a *App) vereinFormular(w http.ResponseWriter, r *http.Request) {
	a.vereinRendern(w, meldung{}, nil)
}

// vereinSpeichern schreibt die Angaben und zeigt das Formular erneut.
//
// Für die Textfelder kann hier nichts abgewiesen werden: es gibt keine
// Pflichtangabe und keine Regel, gegen die zu prüfen wäre (Ticket 23). Das
// Logo dagegen hat eine Grenze und ein festes Format (LogoAusUpload,
// ADR-0012) — deshalb der einzige Fehlerzweig, den dieses Formular kennt.
//
// Ersetzen und Entfernen des Logos wirken erst hier, zusammen mit den
// Textfeldern: das Formular ist eines, kein Feld ist Pflicht, und eine
// Sonderregel nur fürs Logo (etwa ein sofortiges Schreiben per Checkbox)
// bräche mit dem Rest des Formulars.
func (a *App) vereinSpeichern(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxVereinUpload)

	if err := r.ParseMultipartForm(maxVereinUpload); err != nil {
		a.vereinRendern(w, meldung{}, []string{
			"Die Formulardaten ließen sich nicht entgegennehmen — das Logo ist größer als " +
				service.Logogrenze() + "."})

		return
	}

	// Das gespeicherte Logo ist der Ausgangspunkt, den LogoAktualisieren
	// braucht, um "unverändert" von "entfernt" und "ersetzt" zu unterscheiden
	// — SetVereinsdaten selbst kennt kein "unverändert lassen", es ersetzt die
	// ganze Zeile.
	aktuell, err := a.svc.GetVereinsdaten()
	if err != nil {
		fehlerAntwort(w, err)
		return
	}

	aenderung := service.LogoAenderung{Entfernen: r.FormValue("logo_entfernen") != ""}
	if datei, kopf, err := r.FormFile("logo"); err == nil {
		defer datei.Close()

		inhalt, err := io.ReadAll(datei)
		if err != nil {
			fehlerAntwort(w, err)
			return
		}

		aenderung.Hochgeladen, aenderung.Dateiname, aenderung.Inhalt = true, kopf.Filename, inhalt
	} else if !errors.Is(err, http.ErrMissingFile) {
		fehlerAntwort(w, err)
		return
	}

	logo, mime, err := service.LogoAktualisieren(aktuell.Logo, aktuell.LogoMime, aenderung)

	var validierung *service.ValidierungsFehler
	if errors.As(err, &validierung) {
		a.vereinRendern(w, meldung{}, validierung.Meldungen)
		return
	}
	if err != nil {
		fehlerAntwort(w, err)
		return
	}

	daten := vereinsdatenLesen(r)
	daten.Logo, daten.LogoMime = logo, mime

	if err := a.svc.SetVereinsdaten(daten); err != nil {
		fehlerAntwort(w, err)
		return
	}

	a.vereinRendern(w, meldung{Text: "Die Vereinsdaten wurden gespeichert."}, nil)
}

// vereinLogo liefert das hinterlegte Vereinslogo aus — der einzige Weg, auf
// dem das Bild außerhalb eines erzeugten PDFs zu sehen ist (CONTEXT.md →
// Vereinslogo). Ohne Logo antwortet die Route mit 404: die Ansicht bindet den
// Vorschau-<img> nur ein, wenn eines hinterlegt ist (siehe verein.html), ein
// Aufruf von anderswo soll trotzdem nicht ins Leere laufen.
func (a *App) vereinLogo(w http.ResponseWriter, r *http.Request) {
	daten, err := a.svc.GetVereinsdaten()
	if err != nil {
		fehlerAntwort(w, err)
		return
	}
	if len(daten.Logo) == 0 {
		http.NotFound(w, r)
		return
	}

	// Kein Zwischenspeichern im WebView: ein ersetztes oder entferntes Logo
	// soll sofort wirken und nicht erst nach einem harten Neuladen — derselbe
	// Grund wie bei jedem gerenderten Fragment (siehe App.rendern).
	w.Header().Set("Content-Type", daten.LogoMime)
	w.Header().Set("Cache-Control", "no-store")
	w.Write(daten.Logo)
}

// vereinRendern zeigt die Ansicht mit dem Stand aus der Datenbank.
//
// Gelesen wird auch direkt nach dem Speichern, statt die eingesammelten Werte
// wieder auszugeben: der Service schneidet umschließenden Leerraum ab, und das
// Formular soll zeigen, was tatsächlich gespeichert ist, und nicht, was getippt
// wurde.
func (a *App) vereinRendern(w http.ResponseWriter, m meldung, fehler []string) {
	daten, err := a.svc.GetVereinsdaten()
	if err != nil {
		fehlerAntwort(w, err)
		return
	}

	a.rendern(w, "verein", vereinDaten{
		Vereinsdaten: daten,
		Meldung:      m,
		Fehler:       fehler,
		Navigation:   a.navigation(bereichVerein),
	})
}
