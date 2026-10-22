package app

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ToniBlankenburg/boxclub/i18n"
	"github.com/ToniBlankenburg/boxclub/service"
)

const (
	sicherungFilterBeschriftung, sicherungFilterMuster = "Sicherung (*.db)", "*.db"
	excelFilterBeschriftung, excelFilterMuster         = "Excel-Dateien (*.xlsx)", "*.xlsx"
)

// beendenNach ist die Frist zwischen der Meldung "Wiederherstellung fertig"
// und dem Beenden der App — lang genug, dass die Meldung noch zu lesen ist.
const beendenNach = 3 * time.Second

// Oeffnenziel ist der Datei-Dialog zum Auswählen einer vorhandenen Datei,
// das Gegenstück zu Speicherziel. Ein leerer Pfad heißt: abgebrochen.
type Oeffnenziel func(filterBeschriftung, filterMuster string) (string, error)

// OeffnenzielSetzen hinterlegt den Auswahldialog, analog zu SpeicherzielSetzen.
func (a *App) OeffnenzielSetzen(ziel Oeffnenziel) {
	a.oeffnenziel = ziel
}

// BeendenSetzen hinterlegt, wie die App sich beendet. Die Wiederherstellung
// braucht es, weil danach der Service geschlossen ist (service.Wiederherstellen).
func (a *App) BeendenSetzen(beenden func()) {
	a.beenden = beenden
}

// sicherungErstellen schreibt die Sicherung dorthin, wo der Benutzer sie
// ablegen will.
func (a *App) sicherungErstellen(w http.ResponseWriter, r *http.Request) {
	sprache := a.Sprache()
	if a.speicherziel == nil {
		a.datenRendern(w, meldung{Text: i18n.Text(sprache, "verein.dialog_fehlt"), Warnung: true})
		return
	}

	vorschlag := fmt.Sprintf("boxclub-sicherung-%s.db", time.Now().Format("2006-01-02"))
	ziel, err := a.speicherziel(vorschlag, sicherungFilterBeschriftung, sicherungFilterMuster)
	if err != nil {
		fehlerAntwort(w, fmt.Errorf("speicherort erfragen: %w", err))
		return
	}
	if ziel == "" {
		a.datenRendern(w, meldung{})
		return
	}

	if err := a.svc.SicherungErstellen(ziel); err != nil {
		a.datenRendern(w, meldung{
			Text: i18n.Text(sprache, "verein.sicherung_fehler", err), Warnung: true,
		})
		return
	}

	a.datenRendern(w, meldung{Text: i18n.Text(sprache, "verein.sicherung_erstellt", ziel)})
}

// sicherungWiederherstellen ersetzt die Datenbank durch eine gewählte Sicherung
// und beendet danach die App: der Service ist geschlossen, und was die App im
// Speicher hält, gehört zur alten Datei.
func (a *App) sicherungWiederherstellen(w http.ResponseWriter, r *http.Request) {
	sprache := a.Sprache()
	if a.oeffnenziel == nil || a.rueckfrage == nil || a.beenden == nil {
		a.datenRendern(w, meldung{Text: i18n.Text(sprache, "verein.dialog_fehlt"), Warnung: true})
		return
	}

	quelle, err := a.oeffnenziel(sicherungFilterBeschriftung, sicherungFilterMuster)
	if err != nil {
		fehlerAntwort(w, fmt.Errorf("datei erfragen: %w", err))
		return
	}
	if quelle == "" {
		a.datenRendern(w, meldung{})
		return
	}

	knopf := i18n.Text(sprache, "verein.wiederherstellen_knopf")
	ja, err := a.rueckfrage(knopf, i18n.Text(sprache, "verein.wiederherstellen_bestaetigung"),
		knopf, i18n.Text(sprache, "allgemein.abbrechen"))
	if err != nil {
		fehlerAntwort(w, err)
		return
	}
	if !ja {
		a.datenRendern(w, meldung{})
		return
	}

	if err := a.svc.Wiederherstellen(quelle); err != nil {
		schluessel := "verein.wiederherstellen_fehler"
		if errors.Is(err, service.ErrKeineSicherung) {
			schluessel = "verein.wiederherstellen_keine_sicherung"
		}
		a.datenRendern(w, meldung{Text: i18n.Text(sprache, schluessel, err), Warnung: true})
		return
	}

	// Ab hier ist der Service geschlossen: die Antwort darf nichts mehr aus
	// der Datenbank lesen, deshalb nur die Meldung und kein vereinRendern.
	time.AfterFunc(beendenNach, a.beenden)
	a.rendern(w, "meldung", meldung{Text: i18n.Text(sprache, "verein.wiederhergestellt")})
}

// mitgliederExportieren bietet die Mitgliederliste als Excel-Datei an.
func (a *App) mitgliederExportieren(w http.ResponseWriter, r *http.Request) {
	sprache := a.Sprache()
	if a.speicherziel == nil {
		a.datenRendern(w, meldung{Text: i18n.Text(sprache, "verein.dialog_fehlt"), Warnung: true})
		return
	}

	zeilen, err := a.svc.MitgliederExport()
	if err != nil {
		fehlerAntwort(w, err)
		return
	}

	daten, err := service.MitgliederExcel(zeilen)
	if err != nil {
		fehlerAntwort(w, err)
		return
	}

	vorschlag := fmt.Sprintf("Mitglieder-%s.xlsx", time.Now().Format("2006-01-02"))
	ziel, err := a.speicherziel(vorschlag, excelFilterBeschriftung, excelFilterMuster)
	if err != nil {
		fehlerAntwort(w, fmt.Errorf("speicherort erfragen: %w", err))
		return
	}
	if ziel == "" {
		a.datenRendern(w, meldung{})
		return
	}

	// 0600: IBANs und Anschriften sind personenbezogen wie die Datenbank.
	if err := os.WriteFile(ziel, daten, 0o600); err != nil {
		a.datenRendern(w, meldung{
			Text: i18n.Text(sprache, "verein.export_fehler", err), Warnung: true,
		})
		return
	}

	a.datenRendern(w, meldung{Text: i18n.Text(sprache, "verein.export_gespeichert", len(zeilen), ziel)})
}

// datenRendern zeigt die Verein-Seite mit offener Karte "Daten".
func (a *App) datenRendern(w http.ResponseWriter, m meldung) {
	a.vereinRendernMitKarte(w, m, nil, "daten")
}
