package service_test

import (
	"errors"
	"testing"

	"github.com/ToniBlankenburg/boxclub/service"
)

// Die letzten sechs Spalten der Excel-Tabelle: IBAN, Geschlecht,
// Google-Bewertung und Digital am Mitglied, Anmeldedatum und Anmeldegebühr an
// der Mitgliedschaft (Ticket 16). Alle sechs sind freiwillig — keine davon darf
// das Anlegen aufhalten.

func TestCreate_UebernimmtDieRestlichenStammdaten(t *testing.T) {
	svc := neuerService(t)

	anmeldedatum := datum(t, "2025-12-28")

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:         "Anna",
		Nachname:        "Berger",
		Eintritt:        datum(t, "2026-01-01"),
		BeitragCents:    beitragImTest,
		IBAN:            "DE02120300000000202051",
		Geschlecht:      "Frau",
		GoogleBewertung: true,
		Digital:         "Digital",
		Anmeldung: service.Anmeldung{
			Datum:        &anmeldedatum,
			GebuehrCents: 6000,
		},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if m.IBAN != "DE02120300000000202051" {
		t.Errorf("IBAN = %q", m.IBAN)
	}
	if m.Geschlecht != "Frau" {
		t.Errorf("Geschlecht = %q", m.Geschlecht)
	}
	if !m.GoogleBewertung {
		t.Errorf("GoogleBewertung = false, erwartet true")
	}
	if m.Digital != "Digital" {
		t.Errorf("Digital = %q", m.Digital)
	}

	ms := m.Mitgliedschaften[0]
	if ms.Anmeldung.Datum == nil || !ms.Anmeldung.Datum.Equal(anmeldedatum) {
		t.Errorf("Anmeldung.Datum = %v, erwartet %v", ms.Anmeldung.Datum, anmeldedatum)
	}
	if ms.Anmeldung.GebuehrCents != 6000 {
		t.Errorf("Anmeldung.GebuehrCents = %d, erwartet 6000", ms.Anmeldung.GebuehrCents)
	}
}

// Alle sechs Felder sind freiwillig: ein Mitglied ohne sie ist ein gültiges
// Mitglied. Das Anmeldedatum fehlt dann ganz (nil) und ist nicht etwa der
// Eintritt — die beiden sind verschiedene Tage (CONTEXT.md → Anmeldedatum).
func TestCreate_OhneDieRestlichenStammdatenBleibtMoeglich(t *testing.T) {
	svc := neuerService(t)

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:      "Bruno",
		Nachname:     "Costa",
		Eintritt:     datum(t, "2026-01-01"),
		BeitragCents: beitragImTest,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if m.IBAN != "" || m.Geschlecht != "" || m.Digital != "" {
		t.Errorf("erwarte leere Texte, bekam IBAN=%q Geschlecht=%q Digital=%q", m.IBAN, m.Geschlecht, m.Digital)
	}
	if m.GoogleBewertung {
		t.Errorf("GoogleBewertung = true, erwartet false bei Neuanlage")
	}

	ms := m.Mitgliedschaften[0]
	if ms.Anmeldung.Datum != nil {
		t.Errorf("Anmeldung.Datum = %v, erwartet nil", ms.Anmeldung.Datum)
	}
	if ms.Anmeldung.GebuehrCents != 0 {
		t.Errorf("Anmeldung.GebuehrCents = %d, erwartet 0", ms.Anmeldung.GebuehrCents)
	}
}

// Die IBAN ist reiner Text: keine Prüfung, keine Formatierung (ADR-0006). Was
// eingetippt wurde, steht hinterher da.
func TestCreate_NimmtDieIBANUngeprueftEntgegen(t *testing.T) {
	svc := neuerService(t)

	for _, iban := range []string{"DE02 1203 0000 0000 2020 51", "keine iban", "DE02120300000000202051"} {
		t.Run(iban, func(t *testing.T) {
			id, err := svc.Create(service.NeuesMitglied{
				Vorname:      "Clara",
				Nachname:     "Diaz",
				Eintritt:     datum(t, "2026-01-01"),
				BeitragCents: beitragImTest,
				IBAN:         iban,
			})
			if err != nil {
				t.Fatalf("Create: %v", err)
			}

			m, err := svc.Get(id)
			if err != nil {
				t.Fatalf("Get: %v", err)
			}
			if m.IBAN != iban {
				t.Errorf("IBAN = %q, erwartet %q", m.IBAN, iban)
			}
		})
	}
}

func TestUpdate_SchreibtDieRestlichenStammdaten(t *testing.T) {
	svc := neuerService(t)

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:      "Anna",
		Nachname:     "Berger",
		Eintritt:     datum(t, "2026-01-01"),
		BeitragCents: beitragImTest,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	var (
		iban            = "DE02120300000000202051"
		geschlecht      = "Frau"
		googleBewertung = service.GoogleBewertung(true)
		digital         = "Digital"
		anmeldedatum    = datum(t, "2025-12-28")
	)
	if err := svc.Update(id, service.MitgliedPatch{
		IBAN:            &iban,
		Geschlecht:      &geschlecht,
		GoogleBewertung: &googleBewertung,
		Digital:         &digital,
		Anmeldung: &service.Anmeldung{
			Datum:        &anmeldedatum,
			GebuehrCents: 3000,
		},
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if m.IBAN != iban || m.Geschlecht != geschlecht || m.Digital != digital || !m.GoogleBewertung {
		t.Errorf("Stammdaten nach Update: IBAN=%q Geschlecht=%q Digital=%q GoogleBewertung=%v",
			m.IBAN, m.Geschlecht, m.Digital, m.GoogleBewertung)
	}

	ms := m.Mitgliedschaften[0]
	if ms.Anmeldung.Datum == nil || !ms.Anmeldung.Datum.Equal(anmeldedatum) {
		t.Errorf("Anmeldung.Datum = %v, erwartet %v", ms.Anmeldung.Datum, anmeldedatum)
	}
	if ms.Anmeldung.GebuehrCents != 3000 {
		t.Errorf("Anmeldung.GebuehrCents = %d, erwartet 3000", ms.Anmeldung.GebuehrCents)
	}
}

// Leeren muss genauso gehen wie Setzen: ein Feld, das versehentlich gefüllt
// wurde, lässt sich wieder loswerden — auch das Anmeldedatum, dessen leere Form
// nil ist.
func TestUpdate_LeertDieRestlichenStammdaten(t *testing.T) {
	svc := neuerService(t)

	anmeldedatum := datum(t, "2025-12-28")
	id, err := svc.Create(service.NeuesMitglied{
		Vorname:         "Anna",
		Nachname:        "Berger",
		Eintritt:        datum(t, "2026-01-01"),
		BeitragCents:    beitragImTest,
		IBAN:            "DE02120300000000202051",
		Geschlecht:      "Frau",
		GoogleBewertung: true,
		Digital:         "Digital",
		Anmeldung:       service.Anmeldung{Datum: &anmeldedatum, GebuehrCents: 6000},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	var (
		leer  = ""
		keine = service.GoogleBewertung(false)
	)
	if err := svc.Update(id, service.MitgliedPatch{
		IBAN:            &leer,
		Geschlecht:      &leer,
		GoogleBewertung: &keine,
		Digital:         &leer,
		Anmeldung:       &service.Anmeldung{},
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if m.IBAN != "" || m.Geschlecht != "" || m.Digital != "" || m.GoogleBewertung {
		t.Errorf("erwarte geleerte Stammdaten, bekam IBAN=%q Geschlecht=%q Digital=%q GoogleBewertung=%v",
			m.IBAN, m.Geschlecht, m.Digital, m.GoogleBewertung)
	}

	ms := m.Mitgliedschaften[0]
	if ms.Anmeldung.Datum != nil || ms.Anmeldung.GebuehrCents != 0 {
		t.Errorf("Anmeldung = %+v, erwartet leer", ms.Anmeldung)
	}
}

// Ein Patch ohne diese Felder rührt sie nicht an — dieselbe Regel wie bei jedem
// anderen Feld.
func TestUpdate_LaesstDieRestlichenStammdatenOhneAngabeUnberuehrt(t *testing.T) {
	svc := neuerService(t)

	anmeldedatum := datum(t, "2025-12-28")
	id, err := svc.Create(service.NeuesMitglied{
		Vorname:         "Anna",
		Nachname:        "Berger",
		Eintritt:        datum(t, "2026-01-01"),
		BeitragCents:    beitragImTest,
		IBAN:            "DE02120300000000202051",
		Geschlecht:      "Frau",
		GoogleBewertung: true,
		Digital:         "Digital",
		Anmeldung:       service.Anmeldung{Datum: &anmeldedatum, GebuehrCents: 6000},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	nachname := "Berger-Costa"
	if err := svc.Update(id, service.MitgliedPatch{Nachname: &nachname}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if m.IBAN != "DE02120300000000202051" || m.Geschlecht != "Frau" || m.Digital != "Digital" || !m.GoogleBewertung {
		t.Errorf("Stammdaten wurden angetastet: IBAN=%q Geschlecht=%q Digital=%q GoogleBewertung=%v",
			m.IBAN, m.Geschlecht, m.Digital, m.GoogleBewertung)
	}

	ms := m.Mitgliedschaften[0]
	if ms.Anmeldung.Datum == nil || !ms.Anmeldung.Datum.Equal(anmeldedatum) || ms.Anmeldung.GebuehrCents != 6000 {
		t.Errorf("Anmeldung = %+v, erwartet unverändert", ms.Anmeldung)
	}
}

// Die Anmeldegebühr ist ein historischer Wert: sie gehört zu dem Vorgang, mit
// dem ein Zeitraum begann, und wird bei einem Wiedereintritt nicht fortgeführt.
// Genauso das Anmeldedatum — wie die Trainingstermine beginnt der neue Zeitraum
// hier leer.
func TestRejoin_FuehrtAnmeldungNichtFort(t *testing.T) {
	svc := neuerService(t)

	anmeldedatum := datum(t, "2025-12-28")
	id, err := svc.Create(service.NeuesMitglied{
		Vorname:      "Anna",
		Nachname:     "Berger",
		Eintritt:     datum(t, "2026-01-01"),
		BeitragCents: beitragImTest,
		Anmeldung:    service.Anmeldung{Datum: &anmeldedatum, GebuehrCents: 6000},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := svc.SetKuendigung(id, austrittZum(datum(t, "2026-03-31"))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}
	if err := svc.Rejoin(id, datum(t, "2026-09-01")); err != nil {
		t.Fatalf("Rejoin: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(m.Mitgliedschaften) != 2 {
		t.Fatalf("erwarte 2 Mitgliedschaften, bekam %d", len(m.Mitgliedschaften))
	}

	if alt := m.Mitgliedschaften[0].Anmeldung; alt.Datum == nil || !alt.Datum.Equal(anmeldedatum) || alt.GebuehrCents != 6000 {
		t.Errorf("alte Anmeldung = %+v, erwartet unverändert", alt)
	}
	if neu := m.Mitgliedschaften[1].Anmeldung; neu.Datum != nil || neu.GebuehrCents != 0 {
		t.Errorf("neue Anmeldung = %+v, erwartet leer", neu)
	}
}

// Update ändert den Zeitraum, den der Nutzer vor sich hat — bei einem
// Ehemaligen also dessen letzten, genau wie beim Beitrag.
func TestUpdate_SchreibtDieAnmeldungAnDieMassgeblicheMitgliedschaft(t *testing.T) {
	svc := neuerService(t)

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:      "Anna",
		Nachname:     "Berger",
		Eintritt:     datum(t, "2026-01-01"),
		BeitragCents: beitragImTest,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := svc.SetKuendigung(id, austrittZum(datum(t, "2026-03-31"))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	anmeldedatum := datum(t, "2025-12-28")
	if err := svc.Update(id, service.MitgliedPatch{
		Anmeldung: &service.Anmeldung{Datum: &anmeldedatum, GebuehrCents: 6000},
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if ms := m.Mitgliedschaften[0]; ms.Anmeldung.GebuehrCents != 6000 {
		t.Errorf("Anmeldung = %+v, erwartet an der letzten Mitgliedschaft geschrieben", ms.Anmeldung)
	}
}

// Eine negative Anmeldegebühr gibt es nicht. 0 ist dagegen gültig und heißt
// schlicht: es wurde keine erhoben.
func TestCreate_LehntNegativeAnmeldegebuehrAb(t *testing.T) {
	svc := neuerService(t)

	_, err := svc.Create(service.NeuesMitglied{
		Vorname:      "Anna",
		Nachname:     "Berger",
		Eintritt:     datum(t, "2026-01-01"),
		BeitragCents: beitragImTest,
		Anmeldung:    service.Anmeldung{GebuehrCents: -1},
	})

	var validierung *service.ValidierungsFehler
	if !errors.As(err, &validierung) {
		t.Fatalf("Create mit negativer Anmeldegebühr = %v, erwartet *service.ValidierungsFehler", err)
	}
}

// Die Google-Bewertung ist zweiwertig und steht in der Listenzeile — der Verein
// will sehen, wen er noch fragen kann.
func TestEintrag_TraegtDieGoogleBewertung(t *testing.T) {
	svc := neuerService(t)

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:         "Anna",
		Nachname:        "Berger",
		Eintritt:        datum(t, "2026-01-01"),
		BeitragCents:    beitragImTest,
		GoogleBewertung: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	eintrag, err := svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag: %v", err)
	}
	if !eintrag.GoogleBewertung {
		t.Errorf("GoogleBewertung = false, erwartet true")
	}
}

func TestGoogleBewertung_BezeichnungIstZweiwertig(t *testing.T) {
	if got := service.GoogleBewertung(true).Bezeichnung(); got != "hat bewertet" {
		t.Errorf("GoogleBewertung(true).Bezeichnung() = %q", got)
	}
	if got := service.GoogleBewertung(false).Bezeichnung(); got != "hat nicht bewertet" {
		t.Errorf("GoogleBewertung(false).Bezeichnung() = %q", got)
	}
}
