package service_test

import (
	"errors"
	"testing"

	"github.com/ToniBlankenburg/boxclub/service"
)

// Der Beitrag wird als Euro-Betrag eingetippt und in Cent abgelegt. Die
// Umrechnung muss verlustfrei sein — deshalb rechnet sie ganzzahlig und nicht
// über float64, wo 0,29 € als 28,999… ankäme.
func TestBeitragAusEuro_RechnetVerlustfreiInCentUm(t *testing.T) {
	faelle := []struct {
		eingabe string
		cents   int64
	}{
		{"60", 6000},
		{"60,00", 6000},
		{"60.00", 6000},
		{"144,50", 14450},
		{"0", 0},
		{"0,00", 0},
		{"0,29", 29},
		{"35,5", 3550},
		{"  80 € ", 8000},
	}

	for _, f := range faelle {
		t.Run(f.eingabe, func(t *testing.T) {
			cents, err := service.BeitragAusEuro(f.eingabe)
			if err != nil {
				t.Fatalf("BeitragAusEuro(%q): %v", f.eingabe, err)
			}
			if cents != f.cents {
				t.Errorf("BeitragAusEuro(%q) = %d Cent, erwartet %d", f.eingabe, cents, f.cents)
			}
		})
	}
}

// Was kein Betrag ist, wird abgewiesen statt stillschweigend zurechtgebogen:
// ein Tippfehler im Beitrag fällt seit dem Wegfall der Beitragsklassen nicht
// mehr durch einen Fremdschlüssel auf (ADR-0005).
func TestBeitragAusEuro_WeistUngueltigeEingabenAb(t *testing.T) {
	for _, eingabe := range []string{
		"", "   ", "abc", "-5", "60,005", "1.234,56", "60,", ",50", "6 0",
		// Mehr Stellen, als ein Beitrag haben kann: stumm überlaufen darf das
		// keinesfalls.
		"12345", "200000000000000000",
	} {
		t.Run(eingabe, func(t *testing.T) {
			cents, err := service.BeitragAusEuro(eingabe)

			var validierung *service.ValidierungsFehler
			if !errors.As(err, &validierung) {
				t.Fatalf("BeitragAusEuro(%q) = %d, %v — erwartet *service.ValidierungsFehler", eingabe, cents, err)
			}
			if cents != 0 {
				t.Errorf("BeitragAusEuro(%q) = %d Cent, erwartet 0 neben dem Fehler", eingabe, cents)
			}
		})
	}
}

// Anzeige und Eingabe müssen zueinander passen: was das Formular vorbelegt,
// muss es auch wieder einlesen können.
func TestBeitragAlsEuro_IstDieUmkehrungDerEingabe(t *testing.T) {
	faelle := []struct {
		cents int64
		text  string
	}{
		{0, "0,00"},
		{29, "0,29"},
		{3550, "35,50"},
		{6000, "60,00"},
		{14450, "144,50"},
	}

	for _, f := range faelle {
		t.Run(f.text, func(t *testing.T) {
			if text := service.BeitragAlsEuro(f.cents); text != f.text {
				t.Errorf("BeitragAlsEuro(%d) = %q, erwartet %q", f.cents, text, f.text)
			}

			zurueck, err := service.BeitragAusEuro(f.text)
			if err != nil {
				t.Fatalf("BeitragAusEuro(%q): %v", f.text, err)
			}
			if zurueck != f.cents {
				t.Errorf("Rundlauf über %q = %d Cent, erwartet %d", f.text, zurueck, f.cents)
			}
		})
	}
}

// Der Beitrag hängt an der Mitgliedschaft, nicht am Mitglied — er ist Teil der
// Vereinbarung, die mit dem Eintritt zustande kam.
func TestCreate_LegtDenBeitragAnDerMitgliedschaftAb(t *testing.T) {
	svc := neuerService(t)

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:      "Anna",
		Nachname:     "Berger",
		BeitragCents: 6500,
		Eintritt:     datum(t, "2026-01-15"),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(m.Mitgliedschaften) != 1 {
		t.Fatalf("Mitgliedschaften = %d, erwartet 1", len(m.Mitgliedschaften))
	}
	if m.Mitgliedschaften[0].BeitragCents != 6500 {
		t.Errorf("BeitragCents an der Mitgliedschaft = %d, erwartet 6500", m.Mitgliedschaften[0].BeitragCents)
	}

	// Und die Liste zeigt ihn, ohne dass der Aufrufer nachschlagen muss.
	eintrag, err := svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag: %v", err)
	}
	if eintrag.BeitragCents != 6500 {
		t.Errorf("BeitragCents in der Liste = %d, erwartet 6500", eintrag.BeitragCents)
	}
}

// 0 € ist eine gültige Vereinbarung (Trainer zahlen nichts) und keine fehlende
// Angabe: die Anlage muss durchgehen und die Liste den Betrag zeigen.
func TestCreate_ErlaubtBeitragVonNullEuro(t *testing.T) {
	svc := neuerService(t)

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:      "Tomas",
		Nachname:     "Trainer",
		BeitragCents: 0,
		Eintritt:     datum(t, "2026-01-15"),
	})
	if err != nil {
		t.Fatalf("Create mit 0 € muss erlaubt sein, schlug fehl: %v", err)
	}

	eintrag, err := svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag: %v", err)
	}
	if eintrag.BeitragCents != 0 {
		t.Errorf("BeitragCents = %d, erwartet 0", eintrag.BeitragCents)
	}
}

// Ein Beitrag lässt sich neu verhandeln; die Änderung trifft genau dieses
// Mitglied (ADR-0005).
func TestUpdate_AendertDenBeitragUndDieListeZeigtIhn(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Lena", "Hoffmann")
	unberuehrt := mitgliedAnlegen(t, svc, "Paul", "Wagner")

	if err := svc.Update(id, service.MitgliedPatch{BeitragCents: zeiger(int64(7500))}); err != nil {
		t.Fatalf("Update (Beitragsänderung): %v", err)
	}

	geaendert, err := svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag: %v", err)
	}
	if geaendert.BeitragCents != 7500 {
		t.Errorf("BeitragCents nach Update = %d, erwartet 7500", geaendert.BeitragCents)
	}

	// Auf 0 € herabsetzen ist eine Änderung wie jede andere.
	if err := svc.Update(id, service.MitgliedPatch{BeitragCents: zeiger(int64(0))}); err != nil {
		t.Fatalf("Update auf 0 €: %v", err)
	}
	if geaendert, err = svc.Eintrag(id); err != nil {
		t.Fatalf("Eintrag: %v", err)
	}
	if geaendert.BeitragCents != 0 {
		t.Errorf("BeitragCents nach Update auf 0 € = %d, erwartet 0", geaendert.BeitragCents)
	}

	// Kein anderes Mitglied wird davon berührt — genau das war der Grund, die
	// Beitragsklassen aufzugeben.
	anderer, err := svc.Eintrag(unberuehrt)
	if err != nil {
		t.Fatalf("Eintrag (unbeteiligtes Mitglied): %v", err)
	}
	if anderer.BeitragCents != beitragImTest {
		t.Errorf("BeitragCents des unbeteiligten Mitglieds = %d, erwartet unverändert %d",
			anderer.BeitragCents, beitragImTest)
	}
}

// Beim Wiedereintritt entsteht eine neue Mitgliedschaft mit eigenem Beitrag.
// Der alte bleibt an der alten Mitgliedschaft stehen — dafür hängt der Beitrag
// dort und nicht am Mitglied.
func TestRejoin_NeueMitgliedschaftBekommtEigenenBeitrag(t *testing.T) {
	svc := neuerService(t)

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:      "Nina",
		Nachname:     "Klein",
		BeitragCents: 6000,
		Eintritt:     datum(t, "2020-01-01"),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := svc.MarkExit(id, datum(t, "2021-12-31")); err != nil {
		t.Fatalf("MarkExit: %v", err)
	}
	if err := svc.Rejoin(id, datum(t, "2026-02-01")); err != nil {
		t.Fatalf("Rejoin: %v", err)
	}

	// Die neue Mitgliedschaft startet mit dem zuletzt vereinbarten Beitrag.
	eintrag, err := svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag: %v", err)
	}
	if eintrag.BeitragCents != 6000 {
		t.Errorf("BeitragCents nach Wiedereintritt = %d, erwartet 6000", eintrag.BeitragCents)
	}

	// Eine Neuverhandlung trifft nur die laufende Mitgliedschaft.
	if err := svc.Update(id, service.MitgliedPatch{BeitragCents: zeiger(int64(8500))}); err != nil {
		t.Fatalf("Update nach Wiedereintritt: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(m.Mitgliedschaften) != 2 {
		t.Fatalf("Mitgliedschaften = %d, erwartet 2", len(m.Mitgliedschaften))
	}
	if m.Mitgliedschaften[0].BeitragCents != 6000 {
		t.Errorf("Beitrag der beendeten Mitgliedschaft = %d, erwartet unverändert 6000",
			m.Mitgliedschaften[0].BeitragCents)
	}
	if m.Mitgliedschaften[1].BeitragCents != 8500 {
		t.Errorf("Beitrag der laufenden Mitgliedschaft = %d, erwartet 8500",
			m.Mitgliedschaften[1].BeitragCents)
	}
}

// Ein ausgetretenes Mitglied hat keine laufende Mitgliedschaft mehr. Geändert
// wird dann die zuletzt begonnene — dieselbe, die auch die Liste zeigt.
func TestUpdate_AendertDenBeitragAuchBeiEinemEhemaligenMitglied(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Olaf", "Vogel")
	if err := svc.MarkExit(id, datum(t, "2026-06-30")); err != nil {
		t.Fatalf("MarkExit: %v", err)
	}

	if err := svc.Update(id, service.MitgliedPatch{BeitragCents: zeiger(int64(4000))}); err != nil {
		t.Fatalf("Update nach Austritt: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(m.Mitgliedschaften) != 1 {
		t.Fatalf("Mitgliedschaften = %d, erwartet 1", len(m.Mitgliedschaften))
	}
	if m.Mitgliedschaften[0].BeitragCents != 4000 {
		t.Errorf("BeitragCents = %d, erwartet 4000", m.Mitgliedschaften[0].BeitragCents)
	}
}

// Ein negativer Beitrag kann nur aus einem programmatischen Aufruf stammen —
// die Eingabe fängt ihn schon ab. Abgewiesen wird er trotzdem hier, weil die
// Regeln im Service liegen und nicht in der Oberfläche.
func TestCreate_LehntNegativenBeitragAb(t *testing.T) {
	svc := neuerService(t)

	_, err := svc.Create(service.NeuesMitglied{
		Vorname:      "Carl",
		Nachname:     "Dietrich",
		BeitragCents: -100,
		Eintritt:     datum(t, "2026-02-01"),
	})

	var validierung *service.ValidierungsFehler
	if !errors.As(err, &validierung) {
		t.Fatalf("Create mit negativem Beitrag = %v, erwartet *service.ValidierungsFehler", err)
	}

	id := mitgliedAnlegen(t, svc, "Carl", "Dietrich")
	if err := svc.Update(id, service.MitgliedPatch{BeitragCents: zeiger(int64(-1))}); !errors.As(err, &validierung) {
		t.Fatalf("Update mit negativem Beitrag = %v, erwartet *service.ValidierungsFehler", err)
	}
}

// Die Anmeldegebühr teilt sich die Umrechnung mit dem Beitrag und liest
// deshalb dieselben Schreibweisen. Der eine Unterschied: sie ist freiwillig —
// ein leeres Feld heißt "keine Gebühr" und ist kein Fehler.
func TestAnmeldegebuehrAusEuro_RechnetWieDerBeitragUndErlaubtLeer(t *testing.T) {
	faelle := []struct {
		eingabe string
		cents   int64
	}{
		{"", 0},
		{"   ", 0},
		{"0", 0},
		{"60", 6000},
		{"60,50", 6050},
		{"60.50", 6050},
		{"  30 € ", 3000},
	}

	for _, f := range faelle {
		t.Run(f.eingabe, func(t *testing.T) {
			cents, err := service.AnmeldegebuehrAusEuro(f.eingabe)
			if err != nil {
				t.Fatalf("AnmeldegebuehrAusEuro(%q): %v", f.eingabe, err)
			}
			if cents != f.cents {
				t.Errorf("AnmeldegebuehrAusEuro(%q) = %d Cent, erwartet %d", f.eingabe, cents, f.cents)
			}
		})
	}
}

// Was kein Betrag ist, wird auch hier abgewiesen statt zurechtgebogen — die
// Gebühr ist ein historischer Wert, den niemand nachrechnet.
func TestAnmeldegebuehrAusEuro_WeistUngueltigeEingabenAb(t *testing.T) {
	for _, eingabe := range []string{"abc", "-5", "60,005", "1.234,56", "60,", ",50", "12345"} {
		t.Run(eingabe, func(t *testing.T) {
			cents, err := service.AnmeldegebuehrAusEuro(eingabe)

			var validierung *service.ValidierungsFehler
			if !errors.As(err, &validierung) {
				t.Fatalf("AnmeldegebuehrAusEuro(%q) = %d, %v — erwartet *service.ValidierungsFehler", eingabe, cents, err)
			}
			if cents != 0 {
				t.Errorf("AnmeldegebuehrAusEuro(%q) = %d Cent, erwartet 0 neben dem Fehler", eingabe, cents)
			}
		})
	}
}

// Anzeige und Eingabe müssen auch bei der Gebühr zusammenpassen — und die nicht
// erhobene Gebühr muss als leeres Feld hin- und zurückkommen.
func TestAnmeldegebuehrAlsEuro_IstDieUmkehrungDerEingabe(t *testing.T) {
	faelle := []struct {
		cents int64
		text  string
	}{
		{0, ""},
		{3000, "30,00"},
		{6050, "60,50"},
	}

	for _, f := range faelle {
		t.Run(f.text, func(t *testing.T) {
			if got := service.AnmeldegebuehrAlsEuro(f.cents); got != f.text {
				t.Errorf("AnmeldegebuehrAlsEuro(%d) = %q, erwartet %q", f.cents, got, f.text)
			}

			zurueck, err := service.AnmeldegebuehrAusEuro(f.text)
			if err != nil {
				t.Fatalf("AnmeldegebuehrAusEuro(%q): %v", f.text, err)
			}
			if zurueck != f.cents {
				t.Errorf("AnmeldegebuehrAusEuro(%q) = %d Cent, erwartet %d", f.text, zurueck, f.cents)
			}
		})
	}
}
