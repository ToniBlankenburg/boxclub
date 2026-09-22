package service_test

import (
	"strings"
	"testing"
	"time"

	"github.com/ToniBlankenburg/boxclub/service"
)

// moneyMoneyMitglied legt ein Mitglied mit IBAN und Beitrag an, wie es der
// Export braucht — Eintritt in der Vergangenheit, damit es als aktiv gilt.
func moneyMoneyMitglied(t *testing.T, svc *service.MemberService, vorname, nachname, iban string) int64 {
	t.Helper()

	id, err := svc.Create(service.NeuesMitglied{
		Vorname: vorname, Nachname: nachname, IBAN: iban,
		BeitragCents: beitragImTest, Eintritt: heuteVersetzt(-100),
	})
	if err != nil {
		t.Fatalf("Create(%s %s): %v", vorname, nachname, err)
	}

	return id
}

// zeileZuMitglied findet die Exportzeile zu einem Mitglied — die Reihenfolge
// des Exports ist nicht Teil der Zusicherung.
func zeileZuMitglied(t *testing.T, export service.MoneyMoneyExport, mitgliedID int64) service.MoneyMoneyZeile {
	t.Helper()

	for _, z := range export.Zeilen {
		if z.MitgliedID == mitgliedID {
			return z
		}
	}

	t.Fatalf("keine Exportzeile für Mitglied %d, erwartet eine unter %d Zeilen", mitgliedID, len(export.Zeilen))
	return service.MoneyMoneyZeile{}
}

func TestMoneyMoneyExportieren_EnthaeltAktiveMitgliedschaft(t *testing.T) {
	svc := neuerService(t)

	id := moneyMoneyMitglied(t, svc, "Anna", "Berger", "DE02120300000000202051")

	export, err := svc.MoneyMoneyExportieren()
	if err != nil {
		t.Fatalf("MoneyMoneyExportieren: %v", err)
	}

	zeile := zeileZuMitglied(t, export, id)
	if zeile.Zahlungspflichtiger != "Anna Berger" {
		t.Errorf("Zahlungspflichtiger = %q, erwartet %q", zeile.Zahlungspflichtiger, "Anna Berger")
	}
	if zeile.IBAN != "DE02120300000000202051" {
		t.Errorf("IBAN = %q, erwartet unverändert übernommen", zeile.IBAN)
	}
	if zeile.BetragCents != beitragImTest {
		t.Errorf("BetragCents = %d, erwartet nur den Beitrag %d (keine Gebühr erfasst)", zeile.BetragCents, beitragImTest)
	}
}

func TestMoneyMoneyExportieren_MandatsreferenzIstDieMitgliedsID(t *testing.T) {
	svc := neuerService(t)

	id := moneyMoneyMitglied(t, svc, "Bruno", "Cordes", "DE02120300000000202051")

	export, err := svc.MoneyMoneyExportieren()
	if err != nil {
		t.Fatalf("MoneyMoneyExportieren: %v", err)
	}

	zeile := zeileZuMitglied(t, export, id)
	if zeile.MitgliedID != id {
		t.Errorf("MitgliedID = %d, erwartet %d — sie ist die Mandatsreferenz (CONTEXT.md → Mitglieds-ID)", zeile.MitgliedID, id)
	}
}

// Ruhende, noch nicht begonnene und ausgetretene Mitgliedschaften ziehen
// nichts ein — dieselbe Menge wie im Monatssoll des Dashboards.
func TestMoneyMoneyExportieren_SchliesstRuhendeNeueUndAusgetreteneAus(t *testing.T) {
	svc := neuerService(t)

	ruhendID := moneyMoneyMitglied(t, svc, "Clara", "Diehl", "")
	if err := svc.SetRuhend(ruhendID, true); err != nil {
		t.Fatalf("SetRuhend: %v", err)
	}

	neuID, err := svc.Create(service.NeuesMitglied{
		Vorname: "David", Nachname: "Ebert",
		BeitragCents: beitragImTest, Eintritt: heuteVersetzt(30),
	})
	if err != nil {
		t.Fatalf("Create (neu): %v", err)
	}

	ausgetretenID := moneyMoneyMitglied(t, svc, "Erika", "Fels", "")
	gekuendigt, ausgetreten := heuteVersetzt(-60), heuteVersetzt(-1)
	if err := svc.SetKuendigung(ausgetretenID, service.Kuendigung{
		Datum: &gekuendigt, Austritt: &ausgetreten,
	}); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	export, err := svc.MoneyMoneyExportieren()
	if err != nil {
		t.Fatalf("MoneyMoneyExportieren: %v", err)
	}

	for _, ausgeschlossen := range []int64{ruhendID, neuID, ausgetretenID} {
		for _, z := range export.Zeilen {
			if z.MitgliedID == ausgeschlossen {
				t.Errorf("Mitglied %d steht im Export, sollte aber ausgeschlossen sein", ausgeschlossen)
			}
		}
	}
}

// In der Kündigungsfrist wird weiter eingezogen — die Mitgliedschaft läuft
// noch (CONTEXT.md → Kündigungsfrist).
func TestMoneyMoneyExportieren_EnthaeltKuendigungsfrist(t *testing.T) {
	svc := neuerService(t)

	id := moneyMoneyMitglied(t, svc, "Frank", "Gerber", "")
	gekuendigt, austritt := heuteVersetzt(-10), heuteVersetzt(60)
	if err := svc.SetKuendigung(id, service.Kuendigung{
		Datum: &gekuendigt, Austritt: &austritt,
	}); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	export, err := svc.MoneyMoneyExportieren()
	if err != nil {
		t.Fatalf("MoneyMoneyExportieren: %v", err)
	}

	zeileZuMitglied(t, export, id)
}

// Eine noch nicht eingezogene Anmeldegebühr reist in derselben Zeile mit dem
// Beitrag zusammen — der Verein zieht nicht zweimal im selben Monat ein.
func TestMoneyMoneyExportieren_AddiertOffeneAnmeldegebuehr(t *testing.T) {
	svc := neuerService(t)

	id, err := svc.Create(service.NeuesMitglied{
		Vorname: "Greta", Nachname: "Hoff", BeitragCents: beitragImTest,
		Eintritt:  heuteVersetzt(-5),
		Anmeldung: service.Anmeldung{GebuehrCents: 3000},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	export, err := svc.MoneyMoneyExportieren()
	if err != nil {
		t.Fatalf("MoneyMoneyExportieren: %v", err)
	}

	zeile := zeileZuMitglied(t, export, id)
	if erwartet := int64(beitragImTest + 3000); zeile.BetragCents != erwartet {
		t.Errorf("BetragCents = %d, erwartet Beitrag+Gebühr = %d", zeile.BetragCents, erwartet)
	}
	if !strings.Contains(zeile.Verwendungszweck, "Anmeldegebühr") {
		t.Errorf("Verwendungszweck = %q, erwartet einen Hinweis auf die Anmeldegebühr", zeile.Verwendungszweck)
	}
	if len(export.MitgliedschaftIDs) != 1 {
		t.Errorf("MitgliedschaftIDs = %v, erwartet genau einen Eintrag für die enthaltene Gebühr", export.MitgliedschaftIDs)
	}
}

// Ein eigener Verwendungszweck in den Vereinsdaten (ADR-0016) ersetzt den
// automatisch erzeugten "Vereinsname Beitrag MM/JJJJ"-Teil.
func TestMoneyMoneyExportieren_EigenerVerwendungszweckErsetztAutomatischenText(t *testing.T) {
	svc := neuerService(t)

	verein, err := svc.GetVereinsdaten()
	if err != nil {
		t.Fatalf("GetVereinsdaten: %v", err)
	}
	verein.MoneyMoneyVerwendungszweck = "Nachzahlung Turnier"
	if err := svc.SetVereinsdaten(verein); err != nil {
		t.Fatalf("SetVereinsdaten: %v", err)
	}

	id := moneyMoneyMitglied(t, svc, "Ella", "Fuchs", "DE02120300000000202051")

	export, err := svc.MoneyMoneyExportieren()
	if err != nil {
		t.Fatalf("MoneyMoneyExportieren: %v", err)
	}

	zeile := zeileZuMitglied(t, export, id)
	if zeile.Verwendungszweck != "Nachzahlung Turnier" {
		t.Errorf("Verwendungszweck = %q, erwartet den eigenen Text unverändert", zeile.Verwendungszweck)
	}
}

// Das "Anmeldegebühr + "-Präfix bleibt automatisch, auch wenn ein eigener
// Verwendungszweck gesetzt ist — es ist Faktenstand der Zeile, kein Stiltext
// (ADR-0016).
func TestMoneyMoneyExportieren_EigenerVerwendungszweckBehaeltAnmeldegebuehrPraefix(t *testing.T) {
	svc := neuerService(t)

	verein, err := svc.GetVereinsdaten()
	if err != nil {
		t.Fatalf("GetVereinsdaten: %v", err)
	}
	verein.MoneyMoneyVerwendungszweck = "Nachzahlung Turnier"
	if err := svc.SetVereinsdaten(verein); err != nil {
		t.Fatalf("SetVereinsdaten: %v", err)
	}

	id, err := svc.Create(service.NeuesMitglied{
		Vorname: "Finn", Nachname: "Graf", BeitragCents: beitragImTest,
		Eintritt:  heuteVersetzt(-5),
		Anmeldung: service.Anmeldung{GebuehrCents: 3000},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	export, err := svc.MoneyMoneyExportieren()
	if err != nil {
		t.Fatalf("MoneyMoneyExportieren: %v", err)
	}

	zeile := zeileZuMitglied(t, export, id)
	if erwartet := "Anmeldegebühr + Nachzahlung Turnier"; zeile.Verwendungszweck != erwartet {
		t.Errorf("Verwendungszweck = %q, erwartet %q", zeile.Verwendungszweck, erwartet)
	}
}

// Ist die Anmeldegebühr schon eingezogen, taucht sie nicht noch einmal auf.
func TestMoneyMoneyExportieren_LaesstBereitsEingezogeneGebuehrWeg(t *testing.T) {
	svc := neuerService(t)

	id, err := svc.Create(service.NeuesMitglied{
		Vorname: "Hans", Nachname: "Imhoff", BeitragCents: beitragImTest,
		Eintritt:  heuteVersetzt(-40),
		Anmeldung: service.Anmeldung{GebuehrCents: 3000},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Erster Lauf: die Gebühr ist noch offen und wird als eingezogen markiert
	// — genau der Ablauf, den der Handler nach erfolgreichem Speichern nimmt.
	erster, err := svc.MoneyMoneyExportieren()
	if err != nil {
		t.Fatalf("erster Export: %v", err)
	}
	if err := svc.AnmeldegebuehrenAlsEingezogenMarkieren(erster.MitgliedschaftIDs); err != nil {
		t.Fatalf("AnmeldegebuehrenAlsEingezogenMarkieren: %v", err)
	}

	zweiter, err := svc.MoneyMoneyExportieren()
	if err != nil {
		t.Fatalf("zweiter Export: %v", err)
	}

	zeile := zeileZuMitglied(t, zweiter, id)
	if zeile.BetragCents != beitragImTest {
		t.Errorf("BetragCents = %d, erwartet nur den Beitrag %d — die Gebühr ist schon eingezogen", zeile.BetragCents, beitragImTest)
	}
	if len(zweiter.MitgliedschaftIDs) != 0 {
		t.Errorf("MitgliedschaftIDs = %v, erwartet leer, da nichts mehr offen ist", zweiter.MitgliedschaftIDs)
	}
}

// Unterschrieben am ist das Anmeldedatum, wenn es erfasst ist — sonst der
// Eintritt (CONTEXT.md → Anmeldedatum).
func TestMoneyMoneyExportieren_UnterschriebenAmAnmeldedatumSonstEintritt(t *testing.T) {
	svc := neuerService(t)

	anmeldedatum := heuteVersetzt(-20)
	mitAnmeldedatum, err := svc.Create(service.NeuesMitglied{
		Vorname: "Ida", Nachname: "Jansen", BeitragCents: beitragImTest,
		Eintritt:  heuteVersetzt(-10),
		Anmeldung: service.Anmeldung{Datum: &anmeldedatum},
	})
	if err != nil {
		t.Fatalf("Create (mit Anmeldedatum): %v", err)
	}

	eintritt := heuteVersetzt(-15)
	ohneAnmeldedatum, err := svc.Create(service.NeuesMitglied{
		Vorname: "Jonas", Nachname: "Klein", BeitragCents: beitragImTest,
		Eintritt: eintritt,
	})
	if err != nil {
		t.Fatalf("Create (ohne Anmeldedatum): %v", err)
	}

	export, err := svc.MoneyMoneyExportieren()
	if err != nil {
		t.Fatalf("MoneyMoneyExportieren: %v", err)
	}

	if z := zeileZuMitglied(t, export, mitAnmeldedatum); !z.UnterschriebenAm.Equal(anmeldedatum) {
		t.Errorf("UnterschriebenAm = %v, erwartet das Anmeldedatum %v", z.UnterschriebenAm, anmeldedatum)
	}
	if z := zeileZuMitglied(t, export, ohneAnmeldedatum); !z.UnterschriebenAm.Equal(eintritt) {
		t.Errorf("UnterschriebenAm = %v, erwartet ersatzweise den Eintritt %v", z.UnterschriebenAm, eintritt)
	}
}

func TestMoneyMoneyCSV_SchreibtKopfzeileUndZeilenSemikolonGetrennt(t *testing.T) {
	export := service.MoneyMoneyExport{
		Zeilen: []service.MoneyMoneyZeile{
			{
				MitgliedID: 42, Zahlungspflichtiger: "Anna Berger",
				IBAN: "DE02120300000000202051", BetragCents: 6000,
				Verwendungszweck: "Boxclub Musterstadt Beitrag 10/2026",
				UnterschriebenAm: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	daten, err := service.MoneyMoneyCSV(export)
	if err != nil {
		t.Fatalf("MoneyMoneyCSV: %v", err)
	}

	zeilen := strings.Split(strings.TrimRight(string(daten), "\r\n"), "\r\n")
	if len(zeilen) != 2 {
		t.Fatalf("CSV hat %d Zeilen, erwartet Kopf- und eine Datenzeile: %q", len(zeilen), string(daten))
	}

	if erwartet := "Mandatsreferenz;Zahlungspflichtiger;IBAN;BIC;Betrag;Verwendungszweck;Unterschrieben am"; zeilen[0] != erwartet {
		t.Errorf("Kopfzeile = %q, erwartet %q", zeilen[0], erwartet)
	}

	erwarteteZeile := strings.Join([]string{
		"42", "Anna Berger", "DE02120300000000202051", "",
		"60,00", "Boxclub Musterstadt Beitrag 10/2026", "01.09.2026",
	}, ";")
	if zeilen[1] != erwarteteZeile {
		t.Errorf("Datenzeile = %q, erwartet %q", zeilen[1], erwarteteZeile)
	}
}

func TestMoneyMoneyCSV_OhneZeilenNurKopfzeile(t *testing.T) {
	daten, err := service.MoneyMoneyCSV(service.MoneyMoneyExport{})
	if err != nil {
		t.Fatalf("MoneyMoneyCSV: %v", err)
	}

	if strings.Count(string(daten), "\r\n") != 1 {
		t.Errorf("CSV = %q, erwartet genau die Kopfzeile und keine Datenzeile", string(daten))
	}
}

func TestAnmeldegebuehrenAlsEingezogenMarkieren_OhneIDsTutNichts(t *testing.T) {
	svc := neuerService(t)

	if err := svc.AnmeldegebuehrenAlsEingezogenMarkieren(nil); err != nil {
		t.Errorf("AnmeldegebuehrenAlsEingezogenMarkieren(nil): %v", err)
	}
}

// Export und Monatssoll teilen sich dieselbe Regel, wer diesen Monat einzieht
// (zaehltInsMonatssoll, service/dashboard.go) — dieser Test bewacht das über
// beide öffentlichen Wege statt über die interne Funktion: ohne
// Anmeldegebühren muss die Summe der Exportzeilen exakt das Monatssoll
// ergeben, sonst sind die beiden Regeln auseinandergelaufen.
func TestMoneyMoneyExportieren_SummeEntsprichtDemMonatssoll(t *testing.T) {
	svc := neuerService(t)

	moneyMoneyMitglied(t, svc, "Anna", "Berger", "")

	ruhendID := moneyMoneyMitglied(t, svc, "Clara", "Diehl", "")
	if err := svc.SetRuhend(ruhendID, true); err != nil {
		t.Fatalf("SetRuhend: %v", err)
	}

	if _, err := svc.Create(service.NeuesMitglied{
		Vorname: "David", Nachname: "Ebert",
		BeitragCents: beitragImTest, Eintritt: heuteVersetzt(30),
	}); err != nil {
		t.Fatalf("Create (neu): %v", err)
	}

	kuendigungID := moneyMoneyMitglied(t, svc, "Frank", "Gerber", "")
	gekuendigt, austritt := heuteVersetzt(-10), heuteVersetzt(60)
	if err := svc.SetKuendigung(kuendigungID, service.Kuendigung{
		Datum: &gekuendigt, Austritt: &austritt,
	}); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	uebersicht, err := svc.Monatsuebersicht()
	if err != nil {
		t.Fatalf("Monatsuebersicht: %v", err)
	}

	export, err := svc.MoneyMoneyExportieren()
	if err != nil {
		t.Fatalf("MoneyMoneyExportieren: %v", err)
	}

	var summe int64
	for _, z := range export.Zeilen {
		summe += z.BetragCents
	}

	if summe != uebersicht.MonatssollCents {
		t.Errorf("Summe der Exportzeilen = %d, erwartet dasselbe Monatssoll %d (%d Zeilen)",
			summe, uebersicht.MonatssollCents, len(export.Zeilen))
	}
}
