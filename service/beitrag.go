package service

import (
	"fmt"
	"strconv"
	"strings"
)

// Der Beitrag liegt überall als Cent-Betrag vor — eingetippt und angezeigt wird
// er aber in Euro. Beide Richtungen stehen hier beieinander, damit sie nicht
// auseinanderlaufen können: was BeitragAlsEuro schreibt, muss BeitragAusEuro
// wieder einlesen.
//
// Gerechnet wird ganzzahlig. Über float64 käme 0,29 € als 28,999… Cent an, und
// eine Beitragsliste, die sich um einen Cent verzählt, ist wertlos.

// BeitragAusEuro liest einen eingetippten Euro-Betrag als Cent-Betrag. Erlaubt
// sind Komma und Punkt als Dezimaltrennzeichen, ein angehängtes Eurozeichen und
// Leerraum ringsum — also "60", "60,50", "60.50" und " 60 € ".
//
// Alles andere wird abgewiesen statt zurechtgebogen, insbesondere der
// Tausenderpunkt: "1.234,56" ließe sich als 1234,56 € oder als 1,23456 € lesen,
// und raten ist hier die schlechteste aller Antworten. Seit die Beitragsklassen
// weg sind, fällt ein Tippfehler im Beitrag durch nichts anderes mehr auf
// (ADR-0005).
//
// 0 € ist ein gültiger Betrag: Trainer zahlen nichts. Eine leere Eingabe ist
// deshalb kein Weg, "keine Angabe" zu meinen, sondern ein Fehler — die Null
// gehört hingeschrieben.
func BeitragAusEuro(eingabe string) (int64, error) {
	text := strings.TrimSpace(eingabe)
	text = strings.TrimSuffix(text, "€")
	text = strings.TrimSpace(text)

	if text == "" {
		return 0, &ValidierungsFehler{Meldungen: []string{"Bitte einen Beitrag angeben (0 für beitragsfrei)."}}
	}

	euro, cent, _ := strings.Cut(strings.ReplaceAll(text, ",", "."), ".")

	// Ein Vereinsbeitrag hat höchstens vier Stellen vor dem Komma. Die Grenze
	// steht hier nicht, um Übermut zu bestrafen, sondern damit euroWert*100
	// unten nicht überläuft und aus einer Zahlenwurst stillschweigend ein
	// gültiger Betrag wird.
	if len(euro) > 4 {
		return 0, beitragUnlesbar()
	}

	euroWert, err := nurZiffern(euro)
	if err != nil {
		return 0, beitragUnlesbar()
	}

	// Fehlt der Nachkommateil ganz, sind es null Cent; ein einzelner steht für
	// Zehntel-Euro ("35,5" sind 35,50 €).
	centWert := int64(0)
	switch len(cent) {
	case 0:
		if strings.ContainsAny(text, ",.") {
			// "60," oder "60." — der Nutzer war mitten im Tippen.
			return 0, beitragUnlesbar()
		}
	case 1, 2:
		if centWert, err = nurZiffern(cent + strings.Repeat("0", 2-len(cent))); err != nil {
			return 0, beitragUnlesbar()
		}
	default:
		return 0, beitragUnlesbar()
	}

	return euroWert*100 + centWert, nil
}

// nurZiffern parst eine Zahl, die ausschließlich aus Ziffern bestehen darf.
// strconv.ParseInt allein ließe "+60" und "-60" durch; ein Vorzeichen hat in
// einem Beitrag aber nichts zu suchen.
func nurZiffern(text string) (int64, error) {
	if text == "" || strings.ContainsFunc(text, func(r rune) bool { return r < '0' || r > '9' }) {
		return 0, fmt.Errorf("%q besteht nicht nur aus ziffern", text)
	}

	return strconv.ParseInt(text, 10, 64)
}

func beitragUnlesbar() error {
	return &ValidierungsFehler{Meldungen: []string{
		"Der Beitrag ist kein gültiger Betrag. Beispiel: 60 oder 60,50.",
	}}
}

// BeitragAlsEuro schreibt einen Cent-Betrag als Euro-Betrag ohne Währungszeichen
// — die Form, in der das Formular ihn zur Bearbeitung anbietet.
func BeitragAlsEuro(cents int64) string {
	return fmt.Sprintf("%d,%02d", cents/100, cents%100)
}
