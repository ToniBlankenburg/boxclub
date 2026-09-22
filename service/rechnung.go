package service

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// Die Rechnung ist ein Dokument über eine Leistung neben dem Beitrag — typisch
// ein Einzeltraining (CONTEXT.md → Rechnung, ADR-0009). Sie ist kein offener
// Posten: es gibt weder Rechnungsstatus noch Zahlungseingang noch Mahnung, und
// keine eigene Tabelle — nach dem Erzeugen bleibt nur das PDF, als Dokument der
// Art Rechnung am Mitglied abgelegt (siehe dokument.go).
//
// Die Nummer wird eingetippt, nicht vergeben: der Verein führt seine eigene
// Nummernfolge in seiner Buchhaltung, und eine fortlaufende Nummer aus der App
// wäre ein Versprechen, das sie neben einer zweiten Rechnungsquelle nicht halten
// kann (ADR-0009). Beträge sind netto — der Admin trägt den Preis ohne Steuer
// ein, und die App schlägt die Steuer für die Anzeige auf, statt sie aus einem
// Bruttopreis herauszurechnen.

// StandardSteuersatz ist die Vorgabe im Formular: 19 %, änderbar bis 0. Welcher
// Satz gilt, entscheidet der steuerliche Status des Vereins — die App kennt dazu
// keine Regel (CONTEXT.md → Rechnung).
const StandardSteuersatz = 19.0

// ZahlungsfristTage ist der Vorschlag für das Zahlungsziel: 14 Tage nach dem
// Rechnungsdatum. Wie beim regulären Austritt (RegulaererAustritt) ist das ein
// Vorschlag und bindet nichts — das Feld bleibt überschreibbar.
const ZahlungsfristTage = 14

// Zahlungsziel schlägt das Zahlungsziel zu einem Rechnungsdatum vor.
func Zahlungsziel(rechnungsdatum time.Time) time.Time {
	return rechnungsdatum.AddDate(0, 0, ZahlungsfristTage)
}

// Empfaenger ist Name und Anschrift, an die eine Rechnung geht. Er wird aus dem
// Mitglied vorbelegt und ist frei überschreibbar: ein Einzeltraining nimmt auch,
// wer nie eintritt (CONTEXT.md → Rechnung).
type Empfaenger struct {
	Name      string
	Anschrift Anschrift
}

// Rechnungsposition ist eine Zeile der Rechnung: Bezeichnung, Menge und ein
// Einzelpreis in Cent, netto.
//
// Die Menge ist eine ganze Zahl — Einzeltrainings werden gezählt, nicht
// gewogen. Die Zeilensumme ist damit eine Ganzzahlmultiplikation und braucht
// keine Rundung; die rundet erst der Steueranteil des Gesamtbetrags (siehe
// RechnungEingabe.Betraege).
type Rechnungsposition struct {
	Bezeichnung      string
	Menge            int64
	EinzelpreisCents int64
}

// SummeCents ist die Zeilensumme, netto.
func (p Rechnungsposition) SummeCents() int64 {
	return p.Menge * p.EinzelpreisCents
}

// RechnungEingabe sind alle Angaben, aus denen eine Rechnung als PDF entsteht.
type RechnungEingabe struct {
	Empfaenger Empfaenger

	// Nummer wird eingetippt und nicht vergeben (ADR-0009): kein Vorschlag,
	// keine Eindeutigkeitsprüfung — die Nummernfolge führt der Verein in seiner
	// Buchhaltung.
	Nummer string

	Rechnungsdatum time.Time
	Zahlungsziel   time.Time

	// SteuersatzProzent gilt für die ganze Rechnung. Vorgabe StandardSteuersatz,
	// änderbar bis 0; geprüft wird nur, dass die Zahl nicht negativ ist — welche
	// steuerliche Regel gilt, weiß die App nicht.
	SteuersatzProzent float64

	Positionen []Rechnungsposition
}

// Rechnungsbetraege sind Netto, Steuer und Brutto, wie sie auf dem PDF stehen.
type Rechnungsbetraege struct {
	NettoCents  int64
	SteuerCents int64
	BruttoCents int64
}

// Betraege schlägt die Steuer auf den Nettobetrag der Positionen auf — der
// Admin trägt den Preis ohne Steuer ein, und die App rechnet ihn hoch, statt
// ihn vorher aus einem Bruttopreis herausrechnen zu lassen. Bei 0 % ist die
// Steuer 0 und Brutto gleich Netto, ohne den Umweg über eine Multiplikation,
// die dort nichts zu runden hätte.
func (e RechnungEingabe) Betraege() Rechnungsbetraege {
	var netto int64
	for _, p := range e.Positionen {
		netto += p.SummeCents()
	}

	if e.SteuersatzProzent == 0 {
		return Rechnungsbetraege{NettoCents: netto, BruttoCents: netto}
	}

	steuer := int64(math.Round(float64(netto) * e.SteuersatzProzent / 100))

	return Rechnungsbetraege{
		NettoCents:  netto,
		SteuerCents: steuer,
		BruttoCents: netto + steuer,
	}
}

// validieren sammelt alle Regelverstöße einer Rechnung auf einmal ein, wie
// NeuesMitglied.validieren es für die Stammdaten tut.
func (e RechnungEingabe) validieren() error {
	var fehler []string

	if strings.TrimSpace(e.Empfaenger.Name) == "" {
		fehler = append(fehler, "Der Empfänger darf nicht leer sein.")
	}
	if strings.TrimSpace(e.Nummer) == "" {
		fehler = append(fehler, "Die Rechnungsnummer darf nicht leer sein.")
	}
	if e.Rechnungsdatum.IsZero() {
		fehler = append(fehler, "Das Rechnungsdatum darf nicht leer sein.")
	}
	if e.Zahlungsziel.IsZero() {
		fehler = append(fehler, "Das Zahlungsziel darf nicht leer sein.")
	}
	if e.SteuersatzProzent < 0 {
		fehler = append(fehler, "Der Steuersatz darf nicht negativ sein.")
	}
	if len(e.Positionen) == 0 {
		fehler = append(fehler, "Die Rechnung braucht mindestens eine Position.")
	}
	for i, p := range e.Positionen {
		nr := i + 1
		if strings.TrimSpace(p.Bezeichnung) == "" {
			fehler = append(fehler, fmt.Sprintf("Position %d: die Bezeichnung darf nicht leer sein.", nr))
		}
		if p.Menge <= 0 {
			fehler = append(fehler, fmt.Sprintf("Position %d: die Menge muss größer als 0 sein.", nr))
		}
		if p.EinzelpreisCents < 0 {
			fehler = append(fehler, fmt.Sprintf("Position %d: der Einzelpreis darf nicht negativ sein.", nr))
		}
	}

	if len(fehler) > 0 {
		return &ValidierungsFehler{Meldungen: fehler}
	}

	return nil
}

// EinzelpreisAusEuro liest den Einzelpreis einer Position wie den Beitrag: in
// Euro eingetippt, netto, als Cent-Betrag zurück (siehe beitrag.go, dessen
// euroText und centsAusEuroText hier mitbenutzt werden — dieselben
// Schreibweisen, dasselbe Formular-Gefühl). 0 € ist erlaubt: eine Freikarte ist
// eine Position wie jede andere und keine fehlende Angabe.
func EinzelpreisAusEuro(eingabe string) (int64, error) {
	text := euroText(eingabe)
	if text == "" {
		return 0, &ValidierungsFehler{Meldungen: []string{"Bitte einen Einzelpreis angeben (0 für kostenlos)."}}
	}

	cents, ok := centsAusEuroText(text)
	if !ok {
		return 0, &ValidierungsFehler{Meldungen: []string{
			"Der Einzelpreis ist kein gültiger Betrag. Beispiel: 60 oder 60,50.",
		}}
	}

	return cents, nil
}

// SteuersatzAusText liest den eingetippten Steuersatz. Erlaubt sind Komma und
// Punkt als Dezimaltrennzeichen und ein angehängtes Prozentzeichen. Negative
// Sätze gibt es nicht; mehr prüft die App nicht — welcher Satz gilt, entscheidet
// der steuerliche Status des Vereins (CONTEXT.md → Rechnung).
func SteuersatzAusText(eingabe string) (float64, error) {
	text := strings.TrimSpace(eingabe)
	text = strings.TrimSuffix(text, "%")
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, ",", ".")

	if text == "" {
		return 0, &ValidierungsFehler{Meldungen: []string{"Bitte einen Steuersatz angeben (0 für steuerfrei)."}}
	}

	wert, err := strconv.ParseFloat(text, 64)
	if err != nil || wert < 0 {
		return 0, &ValidierungsFehler{Meldungen: []string{
			"Der Steuersatz ist keine gültige Zahl. Beispiel: 19 oder 7.",
		}}
	}

	return wert, nil
}

// SteuersatzAlsText schreibt einen Steuersatz so, wie ihn das Formular zur
// Bearbeitung anbietet — ohne Nachkommastellen, wo keine anfallen ("19" statt
// "19,00"), mit Komma sonst.
func SteuersatzAlsText(prozent float64) string {
	if prozent == math.Trunc(prozent) {
		return strconv.FormatFloat(prozent, 'f', 0, 64)
	}

	return strings.Replace(strconv.FormatFloat(prozent, 'f', -1, 64), ".", ",", 1)
}

// RechnungErstellen erzeugt das PDF einer Rechnung. Ist mitgliedID gesetzt,
// wird es zugleich als Dokument der Art Rechnung an diesem Mitglied abgelegt;
// ohne Mitglied — ein Externer, der nie eintritt — wird nur das PDF geliefert,
// denn es gibt niemanden, an dem es hängen könnte (CONTEXT.md → Rechnung).
//
// beschriftungen sind die Textbausteine, die das PDF selbst trägt (Ticket 06)
// — der Aufrufer (app/rechnung.go) löst sie über die zum Erstellzeitpunkt
// aktive Anzeigesprache auf, damit service/ selbst keine Sprache kennt
// (ADR-0002). Wer keine eigene Übersetzung hat, etwa die Tests dieses
// Pakets, übergibt RechnungBeschriftungenDeutsch.
func (s *MemberService) RechnungErstellen(mitgliedID *int64, eingabe RechnungEingabe, beschriftungen RechnungBeschriftungen) ([]byte, error) {
	if err := eingabe.validieren(); err != nil {
		return nil, err
	}

	// Erst prüfen, dann setzen: ein unbekanntes Mitglied soll nicht erst ein
	// komplettes PDF kosten, bevor rechnungAblegen es ohnehin ablehnt — dieselbe
	// Reihenfolge wie bei VertragAblegen (mitgliedschaftPruefen vor dem Schreiben).
	if mitgliedID != nil {
		if err := mitgliedPruefen(s.db, *mitgliedID); err != nil {
			return nil, err
		}
	}

	verein, err := s.GetVereinsdaten()
	if err != nil {
		return nil, err
	}

	pdf, err := rechnungPDF(verein, eingabe, beschriftungen)
	if err != nil {
		return nil, fmt.Errorf("rechnung als pdf erzeugen: %w", err)
	}

	if mitgliedID == nil {
		return pdf, nil
	}

	if _, err := s.rechnungAblegen(*mitgliedID, eingabe.Nummer, pdf); err != nil {
		return nil, err
	}

	return pdf, nil
}
