package service_test

import (
	"slices"
	"testing"

	"github.com/ToniBlankenburg/boxclub/service"
)

// frauenMitgliedAnlegen legt ein Mitglied mit der Geschlechtsangabe an, optional
// schon für Termine angemeldet.
func frauenMitgliedAnlegen(t *testing.T, svc *service.MemberService, geschlecht string, terminIDs ...int64) (int64, error) {
	t.Helper()

	return svc.Create(service.NeuesMitglied{
		Vorname:            "Test",
		Nachname:           "Person",
		Geschlecht:         geschlecht,
		BeitragCents:       beitragImTest,
		Eintritt:           datum(t, "2026-01-05"),
		TrainingsterminIDs: terminIDs,
	})
}

func TestFrauentermin_WirdGespeichertUndGeaendert(t *testing.T) {
	svc := neuerService(t)
	termin := frauentermin(t, svc)

	if !termin.NurFrauen {
		t.Fatal("NurFrauen = false nach Anlegen, erwartet true")
	}

	a := terminangabe(service.Dienstag, "19:00", "", "Frauen")
	if err := svc.UpdateTrainingstermin(termin.ID, a); err != nil {
		t.Fatalf("UpdateTrainingstermin: %v", err)
	}

	geaendert, err := svc.GetTrainingstermin(termin.ID)
	if err != nil {
		t.Fatalf("GetTrainingstermin: %v", err)
	}
	if geaendert.NurFrauen {
		t.Error("NurFrauen = true nach Änderung ohne Kennzeichen, erwartet false")
	}
}

func TestFrauentermin_NimmtFrauenAuf(t *testing.T) {
	svc := neuerService(t)
	termin := frauentermin(t, svc)

	for _, geschlecht := range []string{"Frau", " frau "} {
		id, err := frauenMitgliedAnlegen(t, svc, geschlecht, termin.ID)
		if err != nil {
			t.Fatalf("Create(%q): %v", geschlecht, err)
		}
		if gefunden := ids(laufendeTermine(t, svc, id)...); !slices.Equal(gefunden, ids(termin)) {
			t.Errorf("Termine(%q) = %v, erwartet %v", geschlecht, gefunden, ids(termin))
		}
	}
}

func TestFrauentermin_WeistAndereAb(t *testing.T) {
	svc := neuerService(t)
	termin := frauentermin(t, svc)

	for _, geschlecht := range []string{"Mann", "", "divers"} {
		_, err := frauenMitgliedAnlegen(t, svc, geschlecht, termin.ID)
		pruefeValidierungsfehler(t, err, "Create mit Geschlecht "+geschlecht)

		if meldung := einzigeMeldung(t, err); meldung.Schluessel != "validierung.termin.nur_frauen" {
			t.Errorf("Meldung(%q) = %+v, erwartet validierung.termin.nur_frauen", geschlecht, meldung)
		}
	}
}

// Das Geschlecht, das gleich mitgespeichert wird, zählt — nicht das alte.
func TestFrauentermin_PruftDasGeschlechtDesPatches(t *testing.T) {
	svc := neuerService(t)
	termin := frauentermin(t, svc)

	id, err := frauenMitgliedAnlegen(t, svc, "Mann")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	dazu := ids(termin)
	err = svc.Update(id, service.MitgliedPatch{TrainingsterminIDs: &dazu})
	pruefeValidierungsfehler(t, err, "Update eines Mannes in den Frauentermin")

	frau := "Frau"
	if err := svc.Update(id, service.MitgliedPatch{Geschlecht: &frau, TrainingsterminIDs: &dazu}); err != nil {
		t.Fatalf("Update mit Geschlecht Frau und Termin: %v", err)
	}
}

// Wer schon angemeldet ist, bleibt es, wenn sich die Angabe später ändert.
func TestFrauentermin_BestehendeAnmeldungBleibt(t *testing.T) {
	svc := neuerService(t)
	termin := frauentermin(t, svc)

	id, err := frauenMitgliedAnlegen(t, svc, "Frau", termin.ID)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	anders := "divers"
	gleich := ids(termin)
	if err := svc.Update(id, service.MitgliedPatch{Geschlecht: &anders, TrainingsterminIDs: &gleich}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if gefunden := ids(laufendeTermine(t, svc, id)...); !slices.Equal(gefunden, ids(termin)) {
		t.Errorf("Termine = %v, erwartet unverändert %v", gefunden, ids(termin))
	}
}

func TestIstFrau(t *testing.T) {
	for geschlecht, erwartet := range map[string]bool{"Frau": true, " frau ": true, "Mann": false, "": false} {
		if got := service.IstFrau(geschlecht); got != erwartet {
			t.Errorf("IstFrau(%q) = %v, erwartet %v", geschlecht, got, erwartet)
		}
	}
}
