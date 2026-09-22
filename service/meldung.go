package service

import "strings"

// Meldung ist eine einzelne Validierungsmeldung als stabiler Schlüssel samt
// Formatargumenten — kein fertiger Satz. service/ kennt keine Sprache
// (ADR-0002, ADR-0018): erst app/ löst Schluessel über i18n.Text in Text auf,
// mit Args wie bei fmt.Sprintf eingesetzt.
type Meldung struct {
	Schluessel string
	Args       []any
}

// meldung baut eine Meldung aus Schlüssel und den Argumenten, die ihre
// Platzhalter füllen. Ohne Argumente bleibt Args nil, wie bei den meisten
// Regelverstößen, die keinen Wert nennen.
func meldung(schluessel string, args ...any) Meldung {
	return Meldung{Schluessel: schluessel, Args: args}
}

// String gibt den Schlüssel aus, notdürftig lesbar für Logs und
// Fehlerausgaben, die keine Sprache kennen — nicht für die Oberfläche.
func (m Meldung) String() string {
	return m.Schluessel
}

// ValidierungsFehler bündelt alle Regelverstöße einer Eingabe. Bewusst als
// Liste: das Formular soll alle fehlenden Pflichtangaben auf einmal anzeigen
// können, statt den Nutzer eine nach der anderen abarbeiten zu lassen. Die
// Meldungen sind Schlüssel, keine fertigen Sätze — die übersetzt erst app/
// (ADR-0018).
type ValidierungsFehler struct {
	Meldungen []Meldung
}

// Error gibt die Schlüssel aus, nicht übersetzten Text. Er ist für Logs und
// den Fallback in app.fehlerAntwort gedacht, nicht für die Oberfläche — dort
// liest jeder Aufrufer, der ValidierungsFehler erwartet, .Meldungen direkt
// und übersetzt sie.
func (f *ValidierungsFehler) Error() string {
	schluessel := make([]string, len(f.Meldungen))
	for i, m := range f.Meldungen {
		schluessel[i] = m.Schluessel
	}

	return strings.Join(schluessel, "; ")
}
