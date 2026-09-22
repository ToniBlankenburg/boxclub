package service_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ToniBlankenburg/boxclub/service"
)

// rechnungsdatumImTest ist ein festes Rechnungsdatum für alle Tests dieser
// Datei — welcher Tag es ist, spielt für keinen von ihnen eine Rolle.
var rechnungsdatumImTest = time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC)

// rechnungEingabeImTest ist eine vollständig ausgefüllte Rechnung über ein
// Einzeltraining — die Tests ändern davon nur, worum es ihnen geht.
func rechnungEingabeImTest() service.RechnungEingabe {
	return service.RechnungEingabe{
		Empfaenger: service.Empfaenger{
			Name: "Max Mustermann",
			Anschrift: service.Anschrift{
				Adresse: "Musterweg 1", Postleitzahl: "12345", Ort: "Musterstadt",
			},
		},
		Nummer:            "2026-014",
		Rechnungsdatum:    rechnungsdatumImTest,
		Zahlungsziel:      service.Zahlungsziel(rechnungsdatumImTest),
		SteuersatzProzent: service.StandardSteuersatz,
		Positionen: []service.Rechnungsposition{
			{Bezeichnung: "Einzeltraining Boxen", Menge: 2, EinzelpreisCents: 3000},
		},
	}
}

// Die Zeilensumme ist eine Ganzzahlmultiplikation: Menge zählt Einzeltrainings,
// keine Bruchteile davon, und braucht deshalb keine Rundung.
func TestRechnungsposition_SummeCents(t *testing.T) {
	p := service.Rechnungsposition{Bezeichnung: "Einzeltraining", Menge: 3, EinzelpreisCents: 3333}

	if summe := p.SummeCents(); summe != 9999 {
		t.Errorf("SummeCents = %d, erwartet 9999", summe)
	}
}

// Betraege addiert die Positionssumme über mehrere Zeilen, bevor die Steuer
// darauf aufgeschlagen wird.
func TestRechnungEingabe_Betraege_AddiertDiePositionen(t *testing.T) {
	e := service.RechnungEingabe{
		SteuersatzProzent: 0,
		Positionen: []service.Rechnungsposition{
			{Bezeichnung: "Einzeltraining", Menge: 2, EinzelpreisCents: 3000},
			{Bezeichnung: "Handschuhe geliehen", Menge: 1, EinzelpreisCents: 500},
		},
	}

	betraege := e.Betraege()
	if betraege.BruttoCents != 6500 {
		t.Errorf("BruttoCents = %d, erwartet 6500", betraege.BruttoCents)
	}
}

// Bei 0 % ist die Steuer 0 und Netto gleich Brutto — ohne den Umweg über eine
// Multiplikation, die hier nichts zu runden hätte.
func TestRechnungEingabe_Betraege_NullProzent(t *testing.T) {
	e := service.RechnungEingabe{
		SteuersatzProzent: 0,
		Positionen:        []service.Rechnungsposition{{Menge: 1, EinzelpreisCents: 6500}},
	}

	betraege := e.Betraege()
	if betraege != (service.Rechnungsbetraege{NettoCents: 6500, SteuerCents: 0, BruttoCents: 6500}) {
		t.Errorf("Betraege = %+v, erwartet Netto=Brutto=6500, Steuer=0", betraege)
	}
}

// Bei einer glatten Multiplikation geht die Rechnung ohne Rest auf: 100,00 €
// netto bei 19 % sind exakt 19,00 € Steuer und 119,00 € brutto.
func TestRechnungEingabe_Betraege_SchlaegtSteuerAufDasNettoAuf(t *testing.T) {
	e := service.RechnungEingabe{
		SteuersatzProzent: 19,
		Positionen:        []service.Rechnungsposition{{Menge: 1, EinzelpreisCents: 10000}},
	}

	betraege := e.Betraege()
	if betraege != (service.Rechnungsbetraege{NettoCents: 10000, SteuerCents: 1900, BruttoCents: 11900}) {
		t.Errorf("Betraege = %+v, erwartet Netto=10000, Steuer=1900, Brutto=11900", betraege)
	}
}

// 84,03 € netto bei 19 % gehen nicht glatt auf (84,03 * 1,19 = 99,9957) — das
// prüft die Rundung auf den Cent.
func TestRechnungEingabe_Betraege_RundetAufDenCent(t *testing.T) {
	e := service.RechnungEingabe{
		SteuersatzProzent: 19,
		Positionen:        []service.Rechnungsposition{{Menge: 1, EinzelpreisCents: 8403}},
	}

	betraege := e.Betraege()
	if betraege != (service.Rechnungsbetraege{NettoCents: 8403, SteuerCents: 1597, BruttoCents: 10000}) {
		t.Errorf("Betraege = %+v, erwartet Netto=8403, Steuer=1597, Brutto=10000", betraege)
	}
}

// RechnungErstellen legt das PDF als Dokument der Art Rechnung am Mitglied ab,
// wenn eine ID mitgegeben wird.
func TestRechnungErstellen_LegtSieAmMitgliedAb(t *testing.T) {
	svc := neuerService(t)
	if err := svc.SetVereinsdaten(vereinsdatenImTest()); err != nil {
		t.Fatalf("SetVereinsdaten: %v", err)
	}

	mitgliedID := mitgliedAnlegen(t, svc, "Anna", "Berger")

	pdf, err := svc.RechnungErstellen(&mitgliedID, rechnungEingabeImTest(), service.RechnungBeschriftungenDeutsch)
	if err != nil {
		t.Fatalf("RechnungErstellen: %v", err)
	}
	if !bytes.HasPrefix(pdf, []byte("%PDF-")) {
		t.Errorf("PDF beginnt mit %q, erwartet die PDF-Kennung", pdf[:min(20, len(pdf))])
	}

	rechnungen, err := svc.RechnungenDesMitglieds(mitgliedID)
	if err != nil {
		t.Fatalf("RechnungenDesMitglieds: %v", err)
	}
	if len(rechnungen) != 1 {
		t.Fatalf("%d Rechnungen abgelegt, erwartet 1", len(rechnungen))
	}
	if rechnungen[0].Art != service.ArtRechnung {
		t.Errorf("Art = %q, erwartet %q", rechnungen[0].Art, service.ArtRechnung)
	}
	if !strings.Contains(rechnungen[0].Name, "2026-014") {
		t.Errorf("Name = %q, erwartet die Rechnungsnummer darin", rechnungen[0].Name)
	}

	geholt, err := svc.RechnungInhalt(rechnungen[0].ID)
	if err != nil {
		t.Fatalf("RechnungInhalt: %v", err)
	}
	if !bytes.Equal(geholt.Inhalt, pdf) {
		t.Error("RechnungInhalt liefert nicht dasselbe PDF, das RechnungErstellen zurückgegeben hat")
	}
}

// Von Rechnungen gibt es je Mitglied beliebig viele — anders als beim Vertrag
// ersetzt eine neue keine alte.
func TestRechnungErstellen_LegtMehrereNebeneinanderAb(t *testing.T) {
	svc := neuerService(t)
	if err := svc.SetVereinsdaten(vereinsdatenImTest()); err != nil {
		t.Fatalf("SetVereinsdaten: %v", err)
	}

	mitgliedID := mitgliedAnlegen(t, svc, "Jonas", "Hartmann")

	erste := rechnungEingabeImTest()
	erste.Nummer = "2026-001"
	if _, err := svc.RechnungErstellen(&mitgliedID, erste, service.RechnungBeschriftungenDeutsch); err != nil {
		t.Fatalf("erstes RechnungErstellen: %v", err)
	}

	zweite := rechnungEingabeImTest()
	zweite.Nummer = "2026-002"
	if _, err := svc.RechnungErstellen(&mitgliedID, zweite, service.RechnungBeschriftungenDeutsch); err != nil {
		t.Fatalf("zweites RechnungErstellen: %v", err)
	}

	rechnungen, err := svc.RechnungenDesMitglieds(mitgliedID)
	if err != nil {
		t.Fatalf("RechnungenDesMitglieds: %v", err)
	}
	if len(rechnungen) != 2 {
		t.Fatalf("%d Rechnungen abgelegt, erwartet 2 — nebeneinander, nicht ersetzt", len(rechnungen))
	}
}

// Für einen Empfänger ohne Mitglied wird das PDF ausgegeben, aber nicht
// abgelegt — es gibt niemanden, an dem es hängen könnte (CONTEXT.md → Rechnung).
func TestRechnungErstellen_OhneMitgliedWirdNichtsAbgelegt(t *testing.T) {
	svc := neuerService(t)
	if err := svc.SetVereinsdaten(vereinsdatenImTest()); err != nil {
		t.Fatalf("SetVereinsdaten: %v", err)
	}

	// Ein Mitglied steht daneben, an dem nichts landen darf — ohne dieses
	// Zeugnis ließe sich "nirgends abgelegt" nicht von "es gibt niemanden, bei
	// dem ich nachsehen könnte" unterscheiden.
	mitgliedID := mitgliedAnlegen(t, svc, "Externer", "Test")

	eingabe := rechnungEingabeImTest()
	eingabe.Empfaenger.Name = "Läuft mal vorbei"

	pdf, err := svc.RechnungErstellen(nil, eingabe, service.RechnungBeschriftungenDeutsch)
	if err != nil {
		t.Fatalf("RechnungErstellen: %v", err)
	}
	if !bytes.HasPrefix(pdf, []byte("%PDF-")) {
		t.Error("kein PDF geliefert, obwohl kein Fehler zurückkam")
	}

	rechnungen, err := svc.RechnungenDesMitglieds(mitgliedID)
	if err != nil {
		t.Fatalf("RechnungenDesMitglieds: %v", err)
	}
	if len(rechnungen) != 0 {
		t.Errorf("%d Rechnungen abgelegt, erwartet 0", len(rechnungen))
	}
}

func TestRechnungErstellen_MeldetUnbekanntesMitglied(t *testing.T) {
	svc := neuerService(t)
	if err := svc.SetVereinsdaten(vereinsdatenImTest()); err != nil {
		t.Fatalf("SetVereinsdaten: %v", err)
	}

	unbekannt := int64(4711)
	if _, err := svc.RechnungErstellen(&unbekannt, rechnungEingabeImTest(), service.RechnungBeschriftungenDeutsch); !errors.Is(err, service.ErrNichtGefunden) {
		t.Errorf("RechnungErstellen = %v, erwartet ErrNichtGefunden", err)
	}
}

// Alle Regelverstöße auf einmal — wie bei NeuesMitglied.validieren.
func TestRechnungErstellen_MeldetAlleFehlerAufEinmal(t *testing.T) {
	svc := neuerService(t)

	eingabe := service.RechnungEingabe{SteuersatzProzent: -1}

	_, err := svc.RechnungErstellen(nil, eingabe, service.RechnungBeschriftungenDeutsch)

	var validierung *service.ValidierungsFehler
	if !errors.As(err, &validierung) {
		t.Fatalf("RechnungErstellen = %v, erwartet einen ValidierungsFehler", err)
	}
	// Empfänger leer, Nummer leer, Rechnungsdatum leer, Zahlungsziel leer,
	// Steuersatz negativ, keine Position: sechs voneinander unabhängige Gründe.
	if len(validierung.Meldungen) != 6 {
		t.Errorf("Meldungen = %q, erwartet 6 Gründe auf einmal", validierung.Meldungen)
	}
}

func TestRechnungErstellen_MeldetFehlerhaftePosition(t *testing.T) {
	svc := neuerService(t)

	eingabe := rechnungEingabeImTest()
	eingabe.Positionen = []service.Rechnungsposition{
		{Bezeichnung: "", Menge: 0, EinzelpreisCents: -100},
	}

	_, err := svc.RechnungErstellen(nil, eingabe, service.RechnungBeschriftungenDeutsch)

	var validierung *service.ValidierungsFehler
	if !errors.As(err, &validierung) {
		t.Fatalf("RechnungErstellen = %v, erwartet einen ValidierungsFehler", err)
	}
	if len(validierung.Meldungen) != 3 {
		t.Errorf("Meldungen = %q, erwartet je einen Grund für Bezeichnung, Menge und Einzelpreis", validierung.Meldungen)
	}
}

func TestEinzelpreisAusEuro(t *testing.T) {
	faelle := []struct {
		eingabe  string
		erwartet int64
	}{
		{"60", 6000},
		{"60,50", 6050},
		{"0", 0},
		{" 12,00 € ", 1200},
	}

	for _, fall := range faelle {
		cents, err := service.EinzelpreisAusEuro(fall.eingabe)
		if err != nil {
			t.Errorf("EinzelpreisAusEuro(%q): %v", fall.eingabe, err)
			continue
		}
		if cents != fall.erwartet {
			t.Errorf("EinzelpreisAusEuro(%q) = %d, erwartet %d", fall.eingabe, cents, fall.erwartet)
		}
	}
}

func TestEinzelpreisAusEuro_WeistLeereEingabeAb(t *testing.T) {
	if _, err := service.EinzelpreisAusEuro(""); err == nil {
		t.Error("EinzelpreisAusEuro(\"\") = nil, erwartet einen Fehler — 0 muss hingeschrieben werden")
	}
}

func TestSteuersatzAusText(t *testing.T) {
	faelle := []struct {
		eingabe  string
		erwartet float64
	}{
		{"19", 19},
		{"7,5", 7.5},
		{"0", 0},
		{" 19 % ", 19},
	}

	for _, fall := range faelle {
		wert, err := service.SteuersatzAusText(fall.eingabe)
		if err != nil {
			t.Errorf("SteuersatzAusText(%q): %v", fall.eingabe, err)
			continue
		}
		if wert != fall.erwartet {
			t.Errorf("SteuersatzAusText(%q) = %v, erwartet %v", fall.eingabe, wert, fall.erwartet)
		}
	}
}

func TestSteuersatzAusText_WeistNegativenSatzAb(t *testing.T) {
	if _, err := service.SteuersatzAusText("-1"); err == nil {
		t.Error("SteuersatzAusText(\"-1\") = nil, erwartet einen Fehler")
	}
}

// SteuersatzAlsText ist die Umkehrung von SteuersatzAusText — was hier steht,
// muss dort wieder ankommen (dieselbe Rundtrip-Regel wie bei BeitragAlsEuro).
func TestSteuersatzAlsText(t *testing.T) {
	faelle := []struct {
		prozent  float64
		erwartet string
	}{
		{19, "19"},
		{0, "0"},
		{7.5, "7,5"},
	}

	for _, fall := range faelle {
		if text := service.SteuersatzAlsText(fall.prozent); text != fall.erwartet {
			t.Errorf("SteuersatzAlsText(%v) = %q, erwartet %q", fall.prozent, text, fall.erwartet)
		}
	}
}
