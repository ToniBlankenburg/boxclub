package service_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/ToniBlankenburg/boxclub/service"
)

// importsatz baut einen vollständigen Satz, wie ihn der ExcelImporter aus einer
// Zeile liefert. Die Tests ändern davon nur, worum es ihnen geht.
func importsatz(t *testing.T, id int64, vorname, nachname string) service.Importsatz {
	t.Helper()

	geburtsdatum := datum(t, "2000-03-16")

	return service.Importsatz{
		ID: id,
		NeuesMitglied: service.NeuesMitglied{
			Vorname:      vorname,
			Nachname:     nachname,
			Geburtsdatum: &geburtsdatum,
			Eintritt:     datum(t, "2026-10-01"),
			BeitragCents: beitragImTest,
		},
	}
}

func TestUebernehmen_LegtMitgliedUnterDerExcelNummerAn(t *testing.T) {
	svc := neuerService(t)

	wirkung, err := svc.Uebernehmen(importsatz(t, 47, "Erika", "Musterfrau"))
	if err != nil {
		t.Fatalf("Uebernehmen: %v", err)
	}
	if wirkung != service.ImportNeu {
		t.Errorf("Wirkung = %v, erwartet ImportNeu", wirkung)
	}

	m, err := svc.Get(47)
	if err != nil {
		t.Fatalf("Get(47): %v", err)
	}
	if m.Vorname != "Erika" || m.Nachname != "Musterfrau" {
		t.Errorf("Mitglied 47 = %s %s, erwartet Erika Musterfrau", m.Vorname, m.Nachname)
	}
}

func TestUebernehmen_ZweiterLaufAktualisiertStattAnzulegen(t *testing.T) {
	svc := neuerService(t)

	if _, err := svc.Uebernehmen(importsatz(t, 47, "Erika", "Musterfrau")); err != nil {
		t.Fatalf("erster Lauf: %v", err)
	}

	geaendert := importsatz(t, 47, "Erika", "Musterfrau")
	geaendert.Email = "erika@example.org"
	geaendert.BeitragCents = 8000

	wirkung, err := svc.Uebernehmen(geaendert)
	if err != nil {
		t.Fatalf("zweiter Lauf: %v", err)
	}
	if wirkung != service.ImportAktualisiert {
		t.Errorf("Wirkung = %v, erwartet ImportAktualisiert", wirkung)
	}

	liste, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(liste) != 1 {
		t.Fatalf("nach zwei Läufen %d Mitglieder, erwartet 1", len(liste))
	}

	m, err := svc.Get(47)
	if err != nil {
		t.Fatalf("Get(47): %v", err)
	}
	if m.Email != "erika@example.org" {
		t.Errorf("E-Mail = %q, erwartet die geänderte Adresse", m.Email)
	}
	if len(m.Mitgliedschaften) != 1 {
		t.Fatalf("%d Mitgliedschaften, erwartet 1", len(m.Mitgliedschaften))
	}
	if m.Mitgliedschaften[0].BeitragCents != 8000 {
		t.Errorf("Beitrag = %d, erwartet den geänderten Betrag 8000", m.Mitgliedschaften[0].BeitragCents)
	}
}

func TestUebernehmen_AutomatischeVergabeZaehltOberhalbDerImportiertenNummern(t *testing.T) {
	svc := neuerService(t)

	if _, err := svc.Uebernehmen(importsatz(t, 47, "Erika", "Musterfrau")); err != nil {
		t.Fatalf("Uebernehmen: %v", err)
	}

	id, err := svc.Create(service.NeuesMitglied{
		Vorname: "Anna", Nachname: "Berger",
		Eintritt: datum(t, "2026-11-01"), BeitragCents: beitragImTest,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 47 {
		t.Errorf("nach dem Import vergebene ID = %d, erwartet oberhalb von 47", id)
	}
}

func TestUebernehmen_SchreibtKuendigungRuhendUndSlots(t *testing.T) {
	svc := neuerService(t)

	kuendigungsdatum := datum(t, "2026-11-03")
	austritt := datum(t, "2027-02-28")

	satz := importsatz(t, 47, "Erika", "Musterfrau")
	satz.Kuendigung = service.Kuendigung{Datum: &kuendigungsdatum, Austritt: &austritt}
	satz.Ruhend = true
	satz.Trainingsslots = []string{"Samstag 10:30 Uhr", "Dienstag 19:30 Uhr"}

	if _, err := svc.Uebernehmen(satz); err != nil {
		t.Fatalf("Uebernehmen: %v", err)
	}

	m, err := svc.Get(47)
	if err != nil {
		t.Fatalf("Get(47): %v", err)
	}

	ms := m.Mitgliedschaften[0]
	if ms.Kuendigungsdatum == nil || !ms.Kuendigungsdatum.Equal(kuendigungsdatum) {
		t.Errorf("Kündigungsdatum = %v, erwartet %v", ms.Kuendigungsdatum, kuendigungsdatum)
	}
	if ms.Austritt == nil || !ms.Austritt.Equal(austritt) {
		t.Errorf("Austritt = %v, erwartet %v", ms.Austritt, austritt)
	}
	if !ms.Ruhend {
		t.Error("Ruhend = false, erwartet true")
	}
	if ms.Trainingsfrequenz() != 2 {
		t.Errorf("Trainingsfrequenz = %d, erwartet 2", ms.Trainingsfrequenz())
	}
}

func TestUebernehmen_LaesstDenRueckstandUnangetastet(t *testing.T) {
	svc := neuerService(t)

	if _, err := svc.Uebernehmen(importsatz(t, 47, "Erika", "Musterfrau")); err != nil {
		t.Fatalf("erster Lauf: %v", err)
	}

	m, err := svc.Get(47)
	if err != nil {
		t.Fatalf("Get(47): %v", err)
	}
	if m.Rueckstand.Offen {
		t.Error("frisch importiertes Mitglied ist im Rückstand, erwartet in Ordnung")
	}

	notiz := service.Rueckstand{Offen: true, Notiz: "Rücklastschrift Oktober"}
	if err := svc.SetRueckstand(47, notiz); err != nil {
		t.Fatalf("SetRueckstand: %v", err)
	}

	if _, err := svc.Uebernehmen(importsatz(t, 47, "Erika", "Musterfrau")); err != nil {
		t.Fatalf("zweiter Lauf: %v", err)
	}

	m, err = svc.Get(47)
	if err != nil {
		t.Fatalf("Get(47): %v", err)
	}
	if m.Rueckstand != notiz {
		t.Errorf("Rückstand = %+v, erwartet unverändert %+v", m.Rueckstand, notiz)
	}
}

func TestUebernehmen_FindetVonHandAngelegtesMitgliedUeberDenNamen(t *testing.T) {
	svc := neuerService(t)

	geburtsdatum := datum(t, "2000-03-16")
	id, err := svc.Create(service.NeuesMitglied{
		Vorname: "Erika", Nachname: "Musterfrau", Geburtsdatum: &geburtsdatum,
		Eintritt: datum(t, "2026-10-01"), BeitragCents: beitragImTest,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	wirkung, err := svc.Uebernehmen(importsatz(t, 47, "Erika", "Musterfrau"))
	if err != nil {
		t.Fatalf("Uebernehmen: %v", err)
	}
	if wirkung != service.ImportAktualisiert {
		t.Errorf("Wirkung = %v, erwartet ImportAktualisiert", wirkung)
	}

	liste, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(liste) != 1 {
		t.Fatalf("%d Mitglieder, erwartet 1 — der Import hat eine Dublette angelegt", len(liste))
	}
	if liste[0].MitgliedID != id {
		t.Errorf("Mitglied hat ID %d, erwartet die bereits vergebene %d", liste[0].MitgliedID, id)
	}
}

func TestUebernehmen_WeistUngueltigeSaetzeAb(t *testing.T) {
	svc := neuerService(t)

	ohneNummer := importsatz(t, 0, "Erika", "Musterfrau")

	var vf *service.ValidierungsFehler
	if _, err := svc.Uebernehmen(ohneNummer); !errors.As(err, &vf) {
		t.Fatalf("Uebernehmen ohne Nummer = %v, erwartet ValidierungsFehler", err)
	}
}

func TestUebernehmen_WeistEineFremdeMitgliedsIDAb(t *testing.T) {
	svc := neuerService(t)

	// Von Hand angelegt: bekommt die automatisch vergebene ID 1 — dieselbe
	// Nummer, unter der in der Excel jemand ganz anderes steht.
	id, err := svc.Create(service.NeuesMitglied{
		Vorname: "Anna", Nachname: "Berger",
		Eintritt: datum(t, "2026-01-15"), BeitragCents: beitragImTest,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	var vf *service.ValidierungsFehler
	if _, err := svc.Uebernehmen(importsatz(t, id, "Erika", "Musterfrau")); !errors.As(err, &vf) {
		t.Fatalf("Uebernehmen = %v, erwartet ValidierungsFehler statt stiller Überschreibung", err)
	}
	if !strings.Contains(vf.Error(), "Anna Berger") {
		t.Errorf("Meldung = %q, erwartet den Namen des betroffenen Mitglieds", vf.Error())
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get(%d): %v", id, err)
	}
	if m.Vorname != "Anna" || m.Nachname != "Berger" {
		t.Errorf("Mitglied %d = %s %s, erwartet unverändert Anna Berger", id, m.Vorname, m.Nachname)
	}
}

func TestUebernehmen_StoertSichNichtAnDerSchreibweiseDesNamens(t *testing.T) {
	svc := neuerService(t)

	if _, err := svc.Uebernehmen(importsatz(t, 47, "Erika", "Musterfrau")); err != nil {
		t.Fatalf("erster Lauf: %v", err)
	}

	// Groß- und Kleinschreibung sind keine andere Person.
	if _, err := svc.Uebernehmen(importsatz(t, 47, "ERIKA", "musterfrau")); err != nil {
		t.Fatalf("zweiter Lauf: %v", err)
	}
}

func TestUebernehmen_RuehrtMehrereMitgliedschaftenNichtAn(t *testing.T) {
	svc := neuerService(t)

	satz := importsatz(t, 47, "Erika", "Musterfrau")
	satz.Eintritt = datum(t, "2020-01-01")
	if _, err := svc.Uebernehmen(satz); err != nil {
		t.Fatalf("erster Lauf: %v", err)
	}

	// Austritt und Wiedereintritt in der App: ab jetzt hat Erika zwei Zeiträume,
	// von denen die Excel nur den ersten kennt.
	austritt := datum(t, "2023-12-31")
	if err := svc.SetKuendigung(47, service.Kuendigung{Austritt: &austritt}); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}
	if err := svc.Rejoin(47, datum(t, "2026-01-01")); err != nil {
		t.Fatalf("Rejoin: %v", err)
	}

	// In der Excel wird der alte Eintritt um einen Tag korrigiert — damit passt
	// er zu keinem der beiden Zeiträume mehr.
	satz.Eintritt = datum(t, "2020-01-02")

	var vf *service.ValidierungsFehler
	if _, err := svc.Uebernehmen(satz); !errors.As(err, &vf) {
		t.Fatalf("Uebernehmen = %v, erwartet ValidierungsFehler statt stillen Überschreibens", err)
	}
	if !strings.Contains(vf.Error(), "mehrere Mitgliedschaften") {
		t.Errorf("Meldung = %q, erwartet einen Hinweis auf die mehreren Zeiträume", vf.Error())
	}

	m, err := svc.Get(47)
	if err != nil {
		t.Fatalf("Get(47): %v", err)
	}
	laufend := m.LaufendeMitgliedschaft()
	if laufend == nil {
		t.Fatal("keine laufende Mitgliedschaft mehr — der Wiedereintritt wurde überschrieben")
	}
	if !laufend.Eintritt.Equal(datum(t, "2026-01-01")) {
		t.Errorf("laufender Eintritt = %s, erwartet unverändert 2026-01-01",
			laufend.Eintritt.Format("2006-01-02"))
	}
}

func TestUebernehmen_KorrigiertDenEintrittBeiNurEinerMitgliedschaft(t *testing.T) {
	svc := neuerService(t)

	satz := importsatz(t, 47, "Erika", "Musterfrau")
	satz.Eintritt = datum(t, "2020-01-01")
	if _, err := svc.Uebernehmen(satz); err != nil {
		t.Fatalf("erster Lauf: %v", err)
	}

	satz.Eintritt = datum(t, "2020-01-02")
	if _, err := svc.Uebernehmen(satz); err != nil {
		t.Fatalf("zweiter Lauf: %v", err)
	}

	m, err := svc.Get(47)
	if err != nil {
		t.Fatalf("Get(47): %v", err)
	}
	if len(m.Mitgliedschaften) != 1 {
		t.Fatalf("%d Mitgliedschaften, erwartet 1", len(m.Mitgliedschaften))
	}
	if !m.Mitgliedschaften[0].Eintritt.Equal(datum(t, "2020-01-02")) {
		t.Errorf("Eintritt = %s, erwartet den korrigierten 2020-01-02",
			m.Mitgliedschaften[0].Eintritt.Format("2006-01-02"))
	}
}

func TestUebernehmen_PruefteineKuendigungWieDasFormular(t *testing.T) {
	svc := neuerService(t)

	austritt := datum(t, "2026-09-30")
	kuendigungsdatum := datum(t, "2026-11-03")

	faelle := map[string]service.Kuendigung{
		"Austritt vor Eintritt":             {Austritt: &austritt},
		"Kündigungsdatum nach dem Austritt": {Datum: &kuendigungsdatum, Austritt: &austritt},
	}

	for name, k := range faelle {
		t.Run(name, func(t *testing.T) {
			satz := importsatz(t, 47, "Erika", "Musterfrau")
			satz.Eintritt = datum(t, "2026-10-01")
			satz.Kuendigung = k

			var vf *service.ValidierungsFehler
			if _, err := svc.Uebernehmen(satz); !errors.As(err, &vf) {
				t.Fatalf("Uebernehmen = %v, erwartet ValidierungsFehler", err)
			}
		})
	}
}

func TestUebernehmen_NimmtEineMitgliedschaftOhneKuendigungAn(t *testing.T) {
	svc := neuerService(t)

	// Der Normalfall des Imports: die Spalte „Gekündigt“ ist leer. Anders als im
	// Formular ist das keine leere Eingabe, sondern schlicht keine Kündigung.
	if _, err := svc.Uebernehmen(importsatz(t, 47, "Erika", "Musterfrau")); err != nil {
		t.Fatalf("Uebernehmen ohne Kündigung: %v", err)
	}
}
