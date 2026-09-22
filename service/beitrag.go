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
// Die Anmeldegebühr steht mit im selben Formular und teilt sich deshalb die
// Umrechnung: zwei Betragsfelder nebeneinander, die unterschiedliche
// Schreibweisen verstünden, wären eine Falle.
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
	text := euroText(eingabe)

	if text == "" {
		return 0, &ValidierungsFehler{Meldungen: []Meldung{meldung("validierung.beitrag.leer")}}
	}

	cents, ok := centsAusEuroText(text)
	if !ok {
		return 0, beitragUnlesbar()
	}

	return cents, nil
}

// AnmeldegebuehrAusEuro liest die einmalige Gebühr beim Eintritt als
// Cent-Betrag. Sie liest dieselben Schreibweisen wie der Beitrag — beide
// kommen aus demselben Formular und sollen sich nicht unterschiedlich
// verhalten.
//
// Der eine Unterschied: die Gebühr ist freiwillig. Ein leeres Feld heißt
// "keine Gebühr erhoben" und ergibt 0 Cent, wo ein leerer Beitrag ein Fehler
// ist. Das ist keine Ausnahme von der Regel, sondern ihr Gegenstück: beim
// Beitrag ist die Null eine Vereinbarung, die hingeschrieben gehört; bei der
// Gebühr ist sie schlicht die Abwesenheit einer Zahlung, und bei den meisten
// Altbeständen steht dazu nichts.
//
// Die Gebühr ist ein historischer Wert: sie hält fest, was beim Eintritt
// tatsächlich gezahlt wurde, und wird nie neu berechnet.
func AnmeldegebuehrAusEuro(eingabe string) (int64, error) {
	text := euroText(eingabe)

	if text == "" {
		return 0, nil
	}

	cents, ok := centsAusEuroText(text)
	if !ok {
		return 0, &ValidierungsFehler{Meldungen: []Meldung{
			meldung("validierung.beitrag.gebuehr_ungueltig"),
		}}
	}

	return cents, nil
}

// euroText schneidet Leerraum und ein angehängtes Eurozeichen ab. Ob der Rest
// leer sein darf, entscheidet der Aufrufer — beim Beitrag nicht, bei der
// Anmeldegebühr schon.
func euroText(eingabe string) string {
	text := strings.TrimSpace(eingabe)
	text = strings.TrimSuffix(text, "€")

	return strings.TrimSpace(text)
}

// centsAusEuroText rechnet einen aufbereiteten, nicht leeren Euro-Text in Cent
// um. Es meldet nur, ob das gelungen ist; die Meldung dazu formuliert der
// Aufrufer, weil sie den Namen des Feldes trägt.
func centsAusEuroText(text string) (int64, bool) {
	euro, cent, _ := strings.Cut(strings.ReplaceAll(text, ",", "."), ".")

	// Ein Vereinsbetrag hat höchstens vier Stellen vor dem Komma. Die Grenze
	// steht hier nicht, um Übermut zu bestrafen, sondern damit euroWert*100
	// unten nicht überläuft und aus einer Zahlenwurst stillschweigend ein
	// gültiger Betrag wird.
	if len(euro) > 4 {
		return 0, false
	}

	euroWert, err := nurZiffern(euro)
	if err != nil {
		return 0, false
	}

	// Fehlt der Nachkommateil ganz, sind es null Cent; ein einzelner steht für
	// Zehntel-Euro ("35,5" sind 35,50 €).
	centWert := int64(0)
	switch len(cent) {
	case 0:
		if strings.ContainsAny(text, ",.") {
			// "60," oder "60." — der Nutzer war mitten im Tippen.
			return 0, false
		}
	case 1, 2:
		if centWert, err = nurZiffern(cent + strings.Repeat("0", 2-len(cent))); err != nil {
			return 0, false
		}
	default:
		return 0, false
	}

	return euroWert*100 + centWert, true
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
	return &ValidierungsFehler{Meldungen: []Meldung{
		meldung("validierung.beitrag.ungueltig"),
	}}
}

// BeitragAlsEuro schreibt einen Cent-Betrag als Euro-Betrag ohne Währungszeichen
// — die Form, in der das Formular ihn zur Bearbeitung anbietet.
func BeitragAlsEuro(cents int64) string {
	return fmt.Sprintf("%d,%02d", cents/100, cents%100)
}

// AnmeldegebuehrAlsEuro schreibt die Gebühr so, wie das Formular sie zur
// Bearbeitung anbietet. Die nicht erhobene Gebühr wird dabei zum leeren Feld
// und nicht zu "0,00": ein Feld, in dem nichts steht, ist die ehrliche Anzeige
// für "dazu ist nichts erfasst" — und genau das liest
// AnmeldegebuehrAusEuro daraus wieder zurück.
func AnmeldegebuehrAlsEuro(cents int64) string {
	if cents == 0 {
		return ""
	}

	return BeitragAlsEuro(cents)
}
