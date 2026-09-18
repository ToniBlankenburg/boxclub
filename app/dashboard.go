package app

import (
	"net/http"
	"net/url"

	"github.com/ToniBlankenburg/boxclub/service"
)

// dashboardDaten speist die Übersicht (CONTEXT.md → Monatssoll, ADR-0009). Sie
// trägt die Berechnung des Service unverändert weiter — anders als beim
// Mitgliedsformular gibt es hier keine Rohwerte und keine Eingabe, die
// zwischen Formular und Service übersetzt werden müssten: das Dashboard zeigt
// nur an, es nimmt nichts entgegen.
type dashboardDaten struct {
	service.Monatsuebersicht
	Navigation []navigationseintrag
	Meldung    meldung

	// RueckstandLink ist die Adresse der Mitgliederliste mit gesetztem
	// Rückstandsfilter — die Zahl daneben ist nur nützlich, wenn man von ihr
	// aus weiterarbeiten kann (Ticket 26). Sie schließt Ehemalige ein: die
	// Rückstands-Zahl selbst zählt sie mit (ein Austritt erlässt keine
	// Schulden, ADR-0006), und ohne den Filter zeigte die geöffnete Liste
	// weniger Mitglieder, als die Zahl gerade versprochen hat.
	RueckstandLink string
}

// rueckstandLink baut dieselbe Rückstandsfilter-Adresse, die auch die
// Filterleiste der Mitgliederliste erzeugt — über dieselbe Konstante wie
// alsSuchfilter, damit ein geänderter Filterwert nicht an einer Stelle
// aktualisiert wird und an der anderen stehen bleibt.
func rueckstandLink() string {
	werte := url.Values{
		"rueckstand": {rueckstandImRueckstand},
		"ehemalige":  {"1"},
	}

	return "/api/mitglieder?" + werte.Encode()
}

// dashboard zeigt das Monatssoll des laufenden Monats und die Mitgliederzahlen
// für Trainer und Admin. Gerechnet wird bei jedem Aufruf neu; gespeichert wird
// hier nichts.
func (a *App) dashboard(w http.ResponseWriter, r *http.Request) {
	a.dashboardRendern(w, meldung{})
}

// dashboardRendern zeigt das Dashboard mit dem Stand aus der Datenbank, dazu
// eine Rückmeldung — gebraucht vom MoneyMoney-Export (app/moneymoney.go), der
// nach dem Speichern auf dasselbe Dashboard zurückkehrt, auf dem der
// Exportknopf steht.
func (a *App) dashboardRendern(w http.ResponseWriter, m meldung) {
	uebersicht, err := a.svc.Monatsuebersicht()
	if err != nil {
		fehlerAntwort(w, err)
		return
	}

	a.rendern(w, "dashboard", dashboardDaten{
		Monatsuebersicht: uebersicht,
		Navigation:       navigation(bereichDashboard),
		Meldung:          m,
		RueckstandLink:   rueckstandLink(),
	})
}
