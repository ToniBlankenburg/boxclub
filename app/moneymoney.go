package app

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ToniBlankenburg/boxclub/service"
)

// csvFilterBeschriftung und csvFilterMuster sind der Dialog-Filter für den
// MoneyMoney-Export (ADR-0013) — die einzige Datei, die diese App als CSV und
// nicht als PDF anbietet.
const csvFilterBeschriftung, csvFilterMuster = "CSV-Dateien (*.csv)", "*.csv"

// moneyMoneyExportieren erzeugt den Lastschrift-Export und bietet ihn über den
// Datei-Dialog zum Speichern an — denselben Weg wie Vertrag und Rechnung
// (app/dokument.go, app/rechnung.go), aus demselben Grund: das WebView von
// Wails kennt keine Downloads.
//
// Eine enthaltene Anmeldegebühr gilt erst als eingezogen, nachdem die Datei
// tatsächlich geschrieben wurde — ein abgebrochener Dialog oder ein
// Schreibfehler dürfen sie nicht als erledigt markieren, sonst fehlte sie im
// nächsten Lauf ersatzlos (ADR-0013).
func (a *App) moneyMoneyExportieren(w http.ResponseWriter, r *http.Request) {
	export, err := a.svc.MoneyMoneyExportieren()
	if err != nil {
		fehlerAntwort(w, err)
		return
	}

	if a.speicherziel == nil {
		a.dashboardRendern(w, meldung{
			Text:    "Der Datei-Dialog steht nicht zur Verfügung — der Export wurde nicht angeboten.",
			Warnung: true,
		})

		return
	}

	daten, err := service.MoneyMoneyCSV(export)
	if err != nil {
		fehlerAntwort(w, err)
		return
	}

	vorschlag := fmt.Sprintf("MoneyMoney-Einzug-%s.csv", time.Now().Format("2006-01"))

	ziel, err := a.speicherziel(vorschlag, csvFilterBeschriftung, csvFilterMuster)
	if err != nil {
		fehlerAntwort(w, fmt.Errorf("speicherort erfragen: %w", err))
		return
	}
	if ziel == "" {
		a.dashboardRendern(w, meldung{})
		return
	}

	// 0600: eine Liste von IBANs und Namen ist personenbezogen wie jede andere
	// Unterlage dieser App (siehe auch das Verzeichnis der Datenbank).
	if err := os.WriteFile(ziel, daten, 0o600); err != nil {
		a.dashboardRendern(w, meldung{
			Text:    fmt.Sprintf("Der Export ließ sich nicht speichern: %v", err),
			Warnung: true,
		})

		return
	}

	// Die Datei liegt jetzt unwiderruflich beim Benutzer — ein Fehler ab hier
	// darf deshalb nicht als 500er enden, der ihm verschweigt, dass der Export
	// trotzdem stattgefunden hat. Bleibt das Markieren aus, taucht die
	// enthaltene Anmeldegebühr im nächsten Lauf einfach noch einmal auf; das
	// muss der Admin dann von Hand bemerken, kann es aber nur, wenn er hier
	// überhaupt davon erfährt.
	if err := a.svc.AnmeldegebuehrenAlsEingezogenMarkieren(export.MitgliedschaftIDs); err != nil {
		a.dashboardRendern(w, meldung{
			Text: fmt.Sprintf(
				"Der Export wurde nach %s gespeichert, aber enthaltene Anmeldegebühren ließen sich nicht als "+
					"eingezogen vormerken (%v). Vor dem nächsten Export prüfen, ob sie darin doppelt auftauchen.",
				ziel, err),
			Warnung: true,
		})

		return
	}

	a.dashboardRendern(w, meldung{
		Text: fmt.Sprintf("Der Export wurde nach %s gespeichert (%d Zeilen).", ziel, len(export.Zeilen)),
	})
}
