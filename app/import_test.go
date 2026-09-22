package app

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ToniBlankenburg/boxclub/i18n"
	"github.com/ToniBlankenburg/boxclub/importer"
	"github.com/ToniBlankenburg/boxclub/service"
)

// Der Orchestrator ist die zweite Ausnahme von "app/ wird nicht unit-getestet"
// (CLAUDE.md): getestet wird kein Handler, sondern die Zählung, die der Bericht
// zeigt. Sie ist die einzige Stelle, an der die beiden Hälften des Imports
// zusammenkommen — und die vier Zahlen sind das, woran der Verein einen zweiten
// Lauf von einem ersten unterscheidet.

func importsatz(id int64, vorname string) service.Importsatz {
	return service.Importsatz{
		ID: id,
		NeuesMitglied: service.NeuesMitglied{
			Vorname: vorname, Nachname: "Musterfrau",
			Eintritt:     time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
			BeitragCents: 6000,
		},
	}
}

func TestUebernehmen_ZaehltNeuAktualisiertUndGescheitert(t *testing.T) {
	svc, err := service.Open(filepath.Join(t.TempDir(), "boxclub.db"))
	if err != nil {
		t.Fatalf("service.Open: %v", err)
	}
	t.Cleanup(func() { svc.Close() })

	gelesen := importer.Ergebnis{
		Saetze: []importer.Zeilensatz{
			{Zeile: 2, Satz: importsatz(1, "Erika")},
			{Zeile: 3, Satz: importsatz(2, "Hans")},
		},
		Fehler: []importer.Zeilenmeldung{{Zeile: 5, Meldungen: []string{"unbekannter Status \"Halbtot\""}}},
	}

	erster, err := uebernehmen(svc, gelesen, i18n.Deutsch)
	if err != nil {
		t.Fatalf("erster Lauf: %v", err)
	}
	if erster.Uebernommen != 2 || erster.Neu != 2 || erster.Aktualisiert != 0 || erster.Gescheitert != 1 {
		t.Errorf("erster Lauf = %+v, erwartet 2 übernommen / 2 neu / 0 aktualisiert / 1 gescheitert", erster)
	}
	if erster.Erfolgreich() {
		t.Error("Erfolgreich = true, erwartet false bei einer gescheiterten Zeile")
	}

	zweiter, err := uebernehmen(svc, gelesen, i18n.Deutsch)
	if err != nil {
		t.Fatalf("zweiter Lauf: %v", err)
	}
	if zweiter.Uebernommen != 2 || zweiter.Neu != 0 || zweiter.Aktualisiert != 2 {
		t.Errorf("zweiter Lauf = %+v, erwartet 2 übernommen / 0 neu / 2 aktualisiert", zweiter)
	}
}

func TestUebernehmen_MeldetEineAbgewieseneZeileMitIhrerZeilennummer(t *testing.T) {
	svc, err := service.Open(filepath.Join(t.TempDir(), "boxclub.db"))
	if err != nil {
		t.Fatalf("service.Open: %v", err)
	}
	t.Cleanup(func() { svc.Close() })

	ohneNachname := importsatz(1, "Erika")
	ohneNachname.Nachname = ""

	bericht, err := uebernehmen(svc, importer.Ergebnis{
		Saetze: []importer.Zeilensatz{
			{Zeile: 2, Satz: ohneNachname},
			{Zeile: 3, Satz: importsatz(2, "Hans")},
		},
	}, i18n.Deutsch)
	if err != nil {
		t.Fatalf("uebernehmen: %v", err)
	}

	if bericht.Uebernommen != 1 || bericht.Gescheitert != 1 {
		t.Fatalf("Bericht = %+v, erwartet 1 übernommen / 1 gescheitert", bericht)
	}
	if len(bericht.Fehler) != 1 || bericht.Fehler[0].Zeile != 2 {
		t.Errorf("Fehler = %+v, erwartet einen Eintrag zu Zeile 2", bericht.Fehler)
	}
}

func TestUebernehmen_HaeltDenLaufBeiEinerAbgewiesenenZeileNichtAn(t *testing.T) {
	svc, err := service.Open(filepath.Join(t.TempDir(), "boxclub.db"))
	if err != nil {
		t.Fatalf("service.Open: %v", err)
	}
	t.Cleanup(func() { svc.Close() })

	// Von Hand angelegt: belegt die automatisch vergebene ID 1, unter der in der
	// Excel jemand anderes steht.
	if _, err := svc.Create(service.NeuesMitglied{
		Vorname: "Anna", Nachname: "Berger",
		Eintritt: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC), BeitragCents: 6500,
	}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	bericht, err := uebernehmen(svc, importer.Ergebnis{
		Saetze: []importer.Zeilensatz{
			{Zeile: 2, Satz: importsatz(1, "Erika")},
			{Zeile: 3, Satz: importsatz(2, "Hans")},
		},
	}, i18n.Deutsch)
	if err != nil {
		t.Fatalf("uebernehmen: %v", err)
	}

	if bericht.Uebernommen != 1 || bericht.Gescheitert != 1 {
		t.Fatalf("Bericht = %+v, erwartet 1 übernommen / 1 gescheitert", bericht)
	}
	if len(bericht.Fehler) != 1 || bericht.Fehler[0].Zeile != 2 {
		t.Fatalf("Fehler = %+v, erwartet einen Eintrag zu Zeile 2", bericht.Fehler)
	}
	if !strings.Contains(strings.Join(bericht.Fehler[0].Meldungen, " "), "Anna Berger") {
		t.Errorf("Grund = %q, erwartet den Namen des betroffenen Mitglieds", bericht.Fehler[0].Meldungen)
	}
}

// Ein Hinweis ist keine gescheiterte Zeile: das Mitglied ist da, sein
// Trainingstermin nicht. Der Bericht muss beides auseinanderhalten, sonst
// passen die Zahlen darüber nicht mehr zu den Listen darunter.
func TestUebernehmen_ZaehltHinweiseNichtAlsGescheitert(t *testing.T) {
	svc, err := service.Open(filepath.Join(t.TempDir(), "boxclub.db"))
	if err != nil {
		t.Fatalf("service.Open: %v", err)
	}
	t.Cleanup(func() { svc.Close() })

	bericht, err := uebernehmen(svc, importer.Ergebnis{
		Saetze: []importer.Zeilensatz{{Zeile: 2, Satz: importsatz(1, "Erika")}},
		Hinweise: []importer.Zeilenmeldung{
			{Zeile: 2, Meldungen: []string{"in der Spalte „Training - 1“ steht „Sa 10:30“ nicht im Stundenplan"}},
		},
	}, i18n.Deutsch)
	if err != nil {
		t.Fatalf("uebernehmen: %v", err)
	}

	if bericht.Uebernommen != 1 || bericht.Gescheitert != 0 {
		t.Errorf("Bericht = %+v, erwartet 1 übernommen / 0 gescheitert", bericht)
	}
	if len(bericht.Hinweise) != 1 || bericht.Hinweise[0].Zeile != 2 {
		t.Errorf("Hinweise = %+v, erwartet einen zu Zeile 2", bericht.Hinweise)
	}
	if len(bericht.Fehler) != 0 {
		t.Errorf("Fehler = %+v, erwartet keine", bericht.Fehler)
	}
	if bericht.Erfolgreich() {
		t.Error("Erfolgreich = true, erwartet false — die Trainingszeit fehlt noch")
	}
}

// Scheitert die Zeile doch noch am Service, gehören ihre Hinweise zu den
// Gründen. In beiden Listen zu stehen hieße, sie zweimal zu zählen.
func TestUebernehmen_NimmtDieHinweiseEinerGescheitertenZeileMit(t *testing.T) {
	svc, err := service.Open(filepath.Join(t.TempDir(), "boxclub.db"))
	if err != nil {
		t.Fatalf("service.Open: %v", err)
	}
	t.Cleanup(func() { svc.Close() })

	ohneNachname := importsatz(1, "Erika")
	ohneNachname.Nachname = ""

	bericht, err := uebernehmen(svc, importer.Ergebnis{
		Saetze:   []importer.Zeilensatz{{Zeile: 2, Satz: ohneNachname}},
		Hinweise: []importer.Zeilenmeldung{{Zeile: 2, Meldungen: []string{"„Sa 10:30“ steht nicht im Stundenplan"}}},
	}, i18n.Deutsch)
	if err != nil {
		t.Fatalf("uebernehmen: %v", err)
	}

	if bericht.Gescheitert != 1 || len(bericht.Hinweise) != 0 {
		t.Fatalf("Bericht = %+v, erwartet 1 gescheitert und keinen offenen Hinweis", bericht)
	}
	if len(bericht.Fehler) != 1 || len(bericht.Fehler[0].Meldungen) != 2 {
		t.Errorf("Gründe = %+v, erwartet den Grund samt Hinweis", bericht.Fehler)
	}
}
