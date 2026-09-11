package service_test

import (
	"testing"

	"github.com/ToniBlankenburg/boxclub/service"
)

// berlinKanal ist die Anschrift, mit der die Tests hier arbeiten — dreigeteilt,
// so wie der Verein sie in seiner Excel-Tabelle führt.
var berlinKanal = service.Anschrift{
	Adresse:      "Kanalstraße 12",
	Postleitzahl: "12043",
	Ort:          "Berlin",
}

func TestCreate_UebernimmtAnschriftAlsDreiFelder(t *testing.T) {
	svc := neuerService(t)

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:      "Anna",
		Nachname:     "Berger",
		Anschrift:    berlinKanal,
		BeitragCents: beitragImTest,
		Eintritt:     datum(t, "2026-01-05"),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if m.Anschrift != berlinKanal {
		t.Errorf("Anschrift = %+v, erwartet %+v", m.Anschrift, berlinKanal)
	}
}

func TestCreate_OhneAnschriftBleibtMoeglich(t *testing.T) {
	svc := neuerService(t)

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:      "Nina",
		Nachname:     "Klein",
		BeitragCents: beitragImTest,
		Eintritt:     datum(t, "2026-01-05"),
	})
	if err != nil {
		t.Fatalf("Create ohne Anschrift: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if !m.Anschrift.Leer() {
		t.Errorf("Anschrift = %+v, erwartet leer", m.Anschrift)
	}
}

// Eine halb gefüllte Anschrift muss genauso durch die Datenbank kommen wie eine
// vollständige: die Felder sind einzeln freiwillig, nicht nur gemeinsam.
func TestCreate_HaeltTeilweiseGefuellteAnschrift(t *testing.T) {
	svc := neuerService(t)

	nurOrt := service.Anschrift{Ort: "Berlin"}

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:      "Jörg",
		Nachname:     "Meier-Schmidt",
		Anschrift:    nurOrt,
		BeitragCents: beitragImTest,
		Eintritt:     datum(t, "2026-01-05"),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if m.Anschrift != nurOrt {
		t.Errorf("Anschrift = %+v, erwartet %+v", m.Anschrift, nurOrt)
	}
	if m.Anschrift.Leer() {
		t.Error("Leer() = true, erwartet false bei gesetztem Ort")
	}

	// Und eine Ergänzung füllt die fehlenden Felder nach.
	vollstaendig := berlinKanal
	if err := svc.Update(id, service.MitgliedPatch{Anschrift: &vollstaendig}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if m, err = svc.Get(id); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if m.Anschrift != vollstaendig {
		t.Errorf("Anschrift = %+v, erwartet %+v", m.Anschrift, vollstaendig)
	}
}

func TestUpdate_SchreibtAlleDreiAnschriftsfelder(t *testing.T) {
	svc := neuerService(t)

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:      "Paul",
		Nachname:     "Wagner",
		Anschrift:    berlinKanal,
		BeitragCents: beitragImTest,
		Eintritt:     datum(t, "2026-01-05"),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	umgezogen := service.Anschrift{
		Adresse:      "Hauptstraße 1a",
		Postleitzahl: "10115",
		Ort:          "Berlin",
	}
	if err := svc.Update(id, service.MitgliedPatch{Anschrift: &umgezogen}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if m.Anschrift != umgezogen {
		t.Errorf("Anschrift = %+v, erwartet %+v", m.Anschrift, umgezogen)
	}
}

func TestUpdate_LeertDieAnschrift(t *testing.T) {
	svc := neuerService(t)

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:      "Paul",
		Nachname:     "Wagner",
		Anschrift:    berlinKanal,
		BeitragCents: beitragImTest,
		Eintritt:     datum(t, "2026-01-05"),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := svc.Update(id, service.MitgliedPatch{Anschrift: &service.Anschrift{}}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if !m.Anschrift.Leer() {
		t.Errorf("Anschrift = %+v, erwartet leer", m.Anschrift)
	}
}

func TestEintrag_TraegtDieAnschriftFuerDieListe(t *testing.T) {
	svc := neuerService(t)

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:      "Anna",
		Nachname:     "Berger",
		Anschrift:    berlinKanal,
		BeitragCents: beitragImTest,
		Eintritt:     datum(t, "2026-01-05"),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	e, err := svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag: %v", err)
	}

	if e.Anschrift != berlinKanal {
		t.Errorf("Anschrift = %+v, erwartet %+v", e.Anschrift, berlinKanal)
	}
	if got := e.Anschrift.OrtZeile(); got != "12043 Berlin" {
		t.Errorf("OrtZeile = %q, erwartet %q", got, "12043 Berlin")
	}
}

// Die Suche bleibt bei Name, E-Mail, Telefon und Mitglieds-ID. Die Anschrift
// steht in der Liste, gehört aber nicht zum Suchumfang — sonst brächte ein Ort
// wie "Berlin" den halben Verein zurück und das Suchfeld wäre entwertet.
func TestSearch_TrifftKeineAnschriftsfelder(t *testing.T) {
	svc := neuerService(t)

	if _, err := svc.Create(service.NeuesMitglied{
		Vorname:      "Anna",
		Nachname:     "Berger",
		Anschrift:    berlinKanal,
		Email:        "anna.berger@example.org",
		BeitragCents: beitragImTest,
		Eintritt:     datum(t, "2026-01-05"),
	}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	for _, begriff := range []string{"Kanalstraße", "12043", "Berlin"} {
		treffer, err := svc.Search(begriff, service.Suchfilter{})
		if err != nil {
			t.Fatalf("Search(%q): %v", begriff, err)
		}
		if len(treffer) != 0 {
			t.Errorf("Search(%q) = %d Treffer, erwartet 0", begriff, len(treffer))
		}
	}

	// Gegenprobe: der Name trifft weiterhin.
	treffer, err := svc.Search("berger", service.Suchfilter{})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(treffer) != 1 {
		t.Fatalf("Search(\"berger\") = %d Treffer, erwartet 1", len(treffer))
	}
}

func TestAnschrift_OrtZeile(t *testing.T) {
	faelle := []struct {
		name      string
		anschrift service.Anschrift
		erwartet  string
	}{
		{"vollständig", berlinKanal, "12043 Berlin"},
		{"nur Postleitzahl", service.Anschrift{Postleitzahl: "12043"}, "12043"},
		{"nur Ort", service.Anschrift{Ort: "Berlin"}, "Berlin"},
		{"leer", service.Anschrift{}, ""},
	}

	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			if got := f.anschrift.OrtZeile(); got != f.erwartet {
				t.Errorf("OrtZeile = %q, erwartet %q", got, f.erwartet)
			}
		})
	}
}

func TestAnschrift_Einzeilig(t *testing.T) {
	faelle := []struct {
		name      string
		anschrift service.Anschrift
		erwartet  string
	}{
		{"vollständig", berlinKanal, "Kanalstraße 12, 12043 Berlin"},
		{"ohne Straße", service.Anschrift{Postleitzahl: "12043", Ort: "Berlin"}, "12043 Berlin"},
		{"ohne Ortsangabe", service.Anschrift{Adresse: "Kanalstraße 12"}, "Kanalstraße 12"},
		{"leer", service.Anschrift{}, ""},
	}

	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			if got := f.anschrift.Einzeilig(); got != f.erwartet {
				t.Errorf("Einzeilig = %q, erwartet %q", got, f.erwartet)
			}
		})
	}
}
