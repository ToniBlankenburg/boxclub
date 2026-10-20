package service_test

import (
	"slices"
	"testing"
	"time"

	"github.com/ToniBlankenburg/boxclub/service"
)

// kindertermin ist ein Termin nur für Kinder.
func kindertermin(t *testing.T, svc *service.MemberService) service.Trainingstermin {
	t.Helper()

	a := terminangabe(service.Mittwoch, "16:00", "", "Kinder")
	a.NurKinder = true

	return angelegterTermin(t, svc, a)
}

func TestKindertermin_WirdGespeichertUndGeaendert(t *testing.T) {
	svc := neuerService(t)
	termin := kindertermin(t, svc)

	if !termin.NurKinder {
		t.Fatal("NurKinder = false nach Anlegen, erwartet true")
	}

	a := terminangabe(service.Mittwoch, "16:00", "", "Kinder")
	if err := svc.UpdateTrainingstermin(termin.ID, a); err != nil {
		t.Fatalf("UpdateTrainingstermin: %v", err)
	}

	geaendert, err := svc.GetTrainingstermin(termin.ID)
	if err != nil {
		t.Fatalf("GetTrainingstermin: %v", err)
	}
	if geaendert.NurKinder {
		t.Error("NurKinder = true nach Änderung ohne Kennzeichen, erwartet false")
	}
}

// Ein Termin hat eine Zielgruppe, nie zwei (CONTEXT.md → Trainingstermin).
func TestKindertermin_KannNichtZugleichNurFrauenSein(t *testing.T) {
	svc := neuerService(t)

	a := terminangabe(service.Mittwoch, "16:00", "", "Mädchen")
	a.NurKinder = true
	a.NurFrauen = true

	_, err := svc.CreateTrainingstermin(a)
	pruefeValidierungsfehler(t, err, "CreateTrainingstermin mit beiden Kennzeichen")
	if meldung := einzigeMeldung(t, err); meldung.Schluessel != "validierung.trainingstermin.zielgruppe" {
		t.Errorf("Meldung = %+v, erwartet validierung.trainingstermin.zielgruppe", meldung)
	}

	termin := kindertermin(t, svc)
	pruefeValidierungsfehler(t, svc.UpdateTrainingstermin(termin.ID, a), "UpdateTrainingstermin mit beiden Kennzeichen")
}

// kindMitgliedAnlegen legt ein Mitglied mit dem Geburtsdatum an (nil: keines),
// optional schon für Termine angemeldet.
func kindMitgliedAnlegen(t *testing.T, svc *service.MemberService, geburtsdatum *time.Time, kindManuell bool, terminIDs ...int64) (int64, error) {
	t.Helper()

	return svc.Create(service.NeuesMitglied{
		Vorname:            "Test",
		Nachname:           "Person",
		Geburtsdatum:       geburtsdatum,
		KindManuell:        kindManuell,
		BeitragCents:       beitragImTest,
		Eintritt:           datum(t, "2026-01-05"),
		TrainingsterminIDs: terminIDs,
	})
}

func vorJahren(jahre int) *time.Time {
	d := time.Now().AddDate(-jahre, 0, -1)

	return &d
}

func TestKindertermin_NimmtKinderAuf(t *testing.T) {
	svc := neuerService(t)
	termin := kindertermin(t, svc)

	for name, geburtsdatum := range map[string]*time.Time{"mit Geburtsdatum": vorJahren(10), "ohne, von Hand": nil} {
		id, err := kindMitgliedAnlegen(t, svc, geburtsdatum, true, termin.ID)
		if err != nil {
			t.Fatalf("Create (%s): %v", name, err)
		}
		if gefunden := ids(laufendeTermine(t, svc, id)...); !slices.Equal(gefunden, ids(termin)) {
			t.Errorf("Termine (%s) = %v, erwartet %v", name, gefunden, ids(termin))
		}
	}
}

func TestKindertermin_WeistErwachseneAb(t *testing.T) {
	svc := neuerService(t)
	termin := kindertermin(t, svc)

	faelle := map[string]struct {
		geburtsdatum *time.Time
		kindManuell  bool
	}{
		"volljährig":                        {vorJahren(30), false},
		"volljährig, Kennzeichen ignoriert": {vorJahren(30), true},
		"ohne Geburtsdatum und Kennzeichen": {nil, false},
	}
	for name, f := range faelle {
		_, err := kindMitgliedAnlegen(t, svc, f.geburtsdatum, f.kindManuell, termin.ID)
		pruefeValidierungsfehler(t, err, "Create "+name)

		if meldung := einzigeMeldung(t, err); meldung.Schluessel != "validierung.termin.nur_kinder" {
			t.Errorf("Meldung (%s) = %+v, erwartet validierung.termin.nur_kinder", name, meldung)
		}
	}
}

// Das Kennzeichen, das gleich mitgespeichert wird, zählt — nicht das alte.
func TestKindertermin_PruftDenStandDesPatches(t *testing.T) {
	svc := neuerService(t)
	termin := kindertermin(t, svc)

	id, err := kindMitgliedAnlegen(t, svc, nil, false)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	dazu := ids(termin)
	err = svc.Update(id, service.MitgliedPatch{TrainingsterminIDs: &dazu})
	pruefeValidierungsfehler(t, err, "Update eines Erwachsenen in den Kindertermin")

	kind := true
	if err := svc.Update(id, service.MitgliedPatch{KindManuell: &kind, TrainingsterminIDs: &dazu}); err != nil {
		t.Fatalf("Update mit Kennzeichen und Termin: %v", err)
	}
}

// Wer schon angemeldet ist, bleibt es, wenn er kein Kind mehr ist.
func TestKindertermin_BestehendeAnmeldungBleibt(t *testing.T) {
	svc := neuerService(t)
	termin := kindertermin(t, svc)

	id, err := kindMitgliedAnlegen(t, svc, nil, true, termin.ID)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	erwachsen := false
	gleich := ids(termin)
	if err := svc.Update(id, service.MitgliedPatch{KindManuell: &erwachsen, TrainingsterminIDs: &gleich}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if gefunden := ids(laufendeTermine(t, svc, id)...); !slices.Equal(gefunden, ids(termin)) {
		t.Errorf("Termine = %v, erwartet unverändert %v", gefunden, ids(termin))
	}
}

// Kinder dürfen weiter in jeden Termin ohne Einschränkung.
func TestKindertermin_KinderDuerfenInOffeneTermine(t *testing.T) {
	svc := neuerService(t)
	offen := angelegterTermin(t, svc, terminangabe(service.Montag, "18:00", "", "Alle"))

	if _, err := kindMitgliedAnlegen(t, svc, vorJahren(10), false, offen.ID); err != nil {
		t.Fatalf("Create: %v", err)
	}
}

// Wird ein Termin erst nach der Anmeldung auf „nur Kinder“ gestellt, bleibt die
// Anmeldung eines Erwachsenen bestehen.
func TestKindertermin_NachtraeglichGesetztLaesstAnmeldungenStehen(t *testing.T) {
	svc := neuerService(t)
	offen := angelegterTermin(t, svc, terminangabe(service.Mittwoch, "16:00", "", "Alle"))

	id, err := kindMitgliedAnlegen(t, svc, vorJahren(30), false, offen.ID)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	a := terminangabe(service.Mittwoch, "16:00", "", "Alle")
	a.NurKinder = true
	if err := svc.UpdateTrainingstermin(offen.ID, a); err != nil {
		t.Fatalf("UpdateTrainingstermin: %v", err)
	}

	gleich := ids(offen)
	if err := svc.Update(id, service.MitgliedPatch{TrainingsterminIDs: &gleich}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if gefunden := ids(laufendeTermine(t, svc, id)...); !slices.Equal(gefunden, ids(offen)) {
		t.Errorf("Termine = %v, erwartet unverändert %v", gefunden, ids(offen))
	}
}
