package service_test

import (
	"testing"

	"github.com/ToniBlankenburg/boxclub/service"
)

func kindAm(t *testing.T, geburtsdatum, stichtag string, manuell bool) bool {
	t.Helper()

	svc := neuerService(t)
	neu := service.NeuesMitglied{
		KindManuell: manuell,
		Vorname:     "Max", Nachname: "Muster",
		Eintritt: datum(t, "2020-01-01"), BeitragCents: beitragImTest,
	}
	if geburtsdatum != "" {
		neu.Geburtsdatum = zeiger(datum(t, geburtsdatum))
	}
	id, err := svc.Create(neu)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	return m.IstKindAm(datum(t, stichtag))
}

func TestIstKindAm_MitGeburtsdatumBisZumTagVorDem18Geburtstag(t *testing.T) {
	tests := []struct {
		name, geboren, stichtag string
		erwartet                bool
	}{
		{"Tag vor dem 18. Geburtstag", "2008-05-10", "2026-05-09", true},
		{"am 18. Geburtstag", "2008-05-10", "2026-05-10", false},
		{"Tag nach dem 18. Geburtstag", "2008-05-10", "2026-05-11", false},
		{"Schalttag, Jahr ohne Schalttag: 28.02. noch Kind", "2008-02-29", "2026-02-28", true},
		{"Schalttag, Jahr ohne Schalttag: 01.03. volljährig", "2008-02-29", "2026-03-01", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := kindAm(t, tt.geboren, tt.stichtag, false); got != tt.erwartet {
				t.Errorf("IstKindAm = %v, erwartet %v", got, tt.erwartet)
			}
		})
	}
}

func TestIstKindAm_OhneGeburtsdatumGiltDasHandkennzeichen(t *testing.T) {
	if !kindAm(t, "", "2026-01-01", true) {
		t.Error("ohne Geburtsdatum, Handkennzeichen gesetzt: erwartet Kind")
	}
	if kindAm(t, "", "2026-01-01", false) {
		t.Error("ohne Geburtsdatum, ohne Handkennzeichen: erwartet kein Kind")
	}
}

func TestIstKindAm_MitGeburtsdatumWirdDasHandkennzeichenIgnoriert(t *testing.T) {
	if kindAm(t, "1980-01-01", "2026-01-01", true) {
		t.Error("Erwachsener mit Handkennzeichen: erwartet kein Kind")
	}
	if !kindAm(t, "2015-01-01", "2026-01-01", false) {
		t.Error("Kind ohne Handkennzeichen: erwartet Kind")
	}
}

func TestUpdate_SetztUndLoeschtDasHandkennzeichen(t *testing.T) {
	svc := neuerService(t)
	id, err := svc.Create(service.NeuesMitglied{
		Vorname: "Max", Nachname: "Muster",
		Eintritt: datum(t, "2020-01-01"), BeitragCents: beitragImTest,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	for _, gesetzt := range []bool{true, false} {
		if err := svc.Update(id, service.MitgliedPatch{KindManuell: &gesetzt}); err != nil {
			t.Fatalf("Update: %v", err)
		}
		m, err := svc.Get(id)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if m.IstKind() != gesetzt {
			t.Errorf("IstKind = %v nach Update auf %v", m.IstKind(), gesetzt)
		}
	}
}
