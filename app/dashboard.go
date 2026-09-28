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
	// aus weiterarbeiten kann (Ticket 26).
	//
	// RueckstandAnzahl zählt Ausgetretene mit (ein Austritt erlässt keine
	// Schulden, ADR-0006), aber seit Ticket 02 lädt der Statusfilter sie nur
	// noch exklusiv: die aufgerufene Liste zeigt deshalb nur die aktiven
	// Mitglieder im Rückstand und bleibt unter der Zahl, wenn zusätzlich
	// Ausgetretene betroffen sind — dieselbe Grenze wie beim alten
	// "ehemalige"-Parameter, der mit AuchEhemalige entfiel und hier nie
	// nachgezogen wurde.
	RueckstandLink string

	// MoneyMoneyEigenerText ist gesetzt, wenn die Vereinsdaten einen eigenen
	// Verwendungszweck für den MoneyMoney-Export hinterlegt haben (ADR-0016)
	// — der Export-Button zeigt dann einen Warnhinweis, weil dieser Text
	// sich nicht selbst nach Monat und Jahr fortführt.
	MoneyMoneyEigenerText bool
}

// rueckstandLink baut dieselbe Rückstandsfilter-Adresse, die auch die
// Filterleiste der Mitgliederliste erzeugt — über dieselbe Konstante wie
// alsSuchfilter, damit ein geänderter Filterwert nicht an einer Stelle
// aktualisiert wird und an der anderen stehen bleibt.
func rueckstandLink() string {
	werte := url.Values{
		"rueckstand": {rueckstandImRueckstand},
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

	verein, err := a.svc.GetVereinsdaten()
	if err != nil {
		fehlerAntwort(w, err)
		return
	}

	a.rendern(w, "dashboard", dashboardDaten{
		Monatsuebersicht:      uebersicht,
		Navigation:            a.navigation(bereichDashboard),
		Meldung:               m,
		RueckstandLink:        rueckstandLink(),
		MoneyMoneyEigenerText: verein.MoneyMoneyVerwendungszweck != "",
	})
}
