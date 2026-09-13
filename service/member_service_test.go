package service_test

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/ToniBlankenburg/boxclub/service"
)

// neuerService legt für jeden Test eine frische SQLite-Datei in t.TempDir() an.
// Kein Mocking: die Tests laufen gegen eine echte Datenbank und beobachten den
// Zustand ausschließlich über die MemberService-API.
//
// Dies ist das Setup-Muster, das alle weiteren Service-Tests spiegeln.
func neuerService(t *testing.T) *service.MemberService {
	t.Helper()

	svc, err := service.Open(filepath.Join(t.TempDir(), "boxclub.db"))
	if err != nil {
		t.Fatalf("service.Open: %v", err)
	}
	t.Cleanup(func() {
		if err := svc.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	})

	return svc
}

// beitragImTest ist der Beitrag, den die Fixtures vereinbaren, wo der Betrag
// selbst nichts zur Sache tut — 65,00 €.
const beitragImTest = 6500

// mitgliedschaftText verdichtet eine Mitgliedschaft auf ihre Textform, um zwei
// von ihnen zu vergleichen. Seit sie die Trainingstermine trägt, ist sie kein
// vergleichbarer Wert mehr — und die Textform zeigt ohnehin den Datumswert
// hinter dem Austritts-Zeiger statt dessen Adresse.
func mitgliedschaftText(ms service.Mitgliedschaft) string {
	return fmt.Sprintf("%+v", ms)
}

// datum parst ein ISO-Datum für Testfixtures.
func datum(t *testing.T, iso string) time.Time {
	t.Helper()

	d, err := time.Parse("2006-01-02", iso)
	if err != nil {
		t.Fatalf("Testfixture %q ist kein gültiges Datum: %v", iso, err)
	}

	return d
}

func TestCreate_LegtMitgliedUndAktiveMitgliedschaftAn(t *testing.T) {
	svc := neuerService(t)

	geburtsdatum := datum(t, "1990-04-17")
	eintritt := datum(t, "2026-01-15")

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:      "Anna",
		Nachname:     "Berger",
		Geburtsdatum: &geburtsdatum,
		Anschrift:    service.Anschrift{Adresse: "Ringstraße 5", Postleitzahl: "12043", Ort: "Berlin"},
		Email:        "anna.berger@example.org",
		Telefon:      "030 1234567",
		BeitragCents: 8000,
		Eintritt:     eintritt,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if m.ID != id {
		t.Errorf("ID = %d, erwartet %d", m.ID, id)
	}
	if m.Vorname != "Anna" {
		t.Errorf("Vorname = %q, erwartet %q", m.Vorname, "Anna")
	}
	if m.Nachname != "Berger" {
		t.Errorf("Nachname = %q, erwartet %q", m.Nachname, "Berger")
	}
	if m.Geburtsdatum == nil || !m.Geburtsdatum.Equal(geburtsdatum) {
		t.Errorf("Geburtsdatum = %v, erwartet %v", m.Geburtsdatum, geburtsdatum)
	}
	if erwartet := (service.Anschrift{Adresse: "Ringstraße 5", Postleitzahl: "12043", Ort: "Berlin"}); m.Anschrift != erwartet {
		t.Errorf("Anschrift = %+v, erwartet %+v", m.Anschrift, erwartet)
	}
	if m.Email != "anna.berger@example.org" {
		t.Errorf("Email = %q", m.Email)
	}
	if m.Telefon != "030 1234567" {
		t.Errorf("Telefon = %q", m.Telefon)
	}
	if m.Rueckstand.Offen {
		t.Errorf("Rueckstand.Offen = true, erwartet false bei Neuanlage")
	}

	if len(m.Mitgliedschaften) != 1 {
		t.Fatalf("erwarte genau 1 Mitgliedschaft, bekam %d: %+v", len(m.Mitgliedschaften), m.Mitgliedschaften)
	}
	ms := m.Mitgliedschaften[0]
	if !ms.Eintritt.Equal(eintritt) {
		t.Errorf("Eintritt = %v, erwartet %v", ms.Eintritt, eintritt)
	}
	if ms.Austritt != nil {
		t.Errorf("Austritt = %v, erwartet nil (aktive Mitgliedschaft)", ms.Austritt)
	}
	if ms.MitgliedID != id {
		t.Errorf("MitgliedID = %d, erwartet %d", ms.MitgliedID, id)
	}
	if ms.BeitragCents != 8000 {
		t.Errorf("BeitragCents = %d, erwartet 8000", ms.BeitragCents)
	}
}

func TestCreate_ErlaubtNamensgleichheitBeiGleichemGeburtsdatum(t *testing.T) {
	svc := neuerService(t)

	geburtsdatum := datum(t, "2001-09-03")
	stammdaten := service.NeuesMitglied{
		Vorname:      "Max",
		Nachname:     "Müller",
		Geburtsdatum: &geburtsdatum,
		BeitragCents: beitragImTest,
		Eintritt:     datum(t, "2026-03-01"),
	}

	ersteID, err := svc.Create(stammdaten)
	if err != nil {
		t.Fatalf("Create (erstes Mitglied): %v", err)
	}

	zweiteID, err := svc.Create(stammdaten)
	if err != nil {
		t.Fatalf("Create (namensgleiches zweites Mitglied) muss erlaubt sein, schlug fehl: %v", err)
	}

	if ersteID == zweiteID {
		t.Fatalf("beide Mitglieder haben dieselbe ID %d — es müssen zwei Datensätze sein", ersteID)
	}
}

func TestGet_UnbekannteIDMeldetNichtGefunden(t *testing.T) {
	svc := neuerService(t)

	_, err := svc.Get(4711)
	if !errors.Is(err, service.ErrNichtGefunden) {
		t.Fatalf("Get(4711) = %v, erwartet ErrNichtGefunden", err)
	}
}

func TestOpen_BestehendeDatenbankBehaeltIhreMitglieder(t *testing.T) {
	dbPfad := filepath.Join(t.TempDir(), "boxclub.db")

	ersteSitzung, err := service.Open(dbPfad)
	if err != nil {
		t.Fatalf("service.Open (erste Sitzung): %v", err)
	}

	id, err := ersteSitzung.Create(service.NeuesMitglied{
		Vorname:      "Elif",
		Nachname:     "Yilmaz",
		BeitragCents: beitragImTest,
		Eintritt:     datum(t, "2026-04-20"),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := ersteSitzung.Close(); err != nil {
		t.Fatalf("Close (erste Sitzung): %v", err)
	}

	// Zweiter App-Start auf derselben Datei.
	zweiteSitzung, err := service.Open(dbPfad)
	if err != nil {
		t.Fatalf("service.Open (zweite Sitzung): %v", err)
	}
	t.Cleanup(func() {
		if err := zweiteSitzung.Close(); err != nil {
			t.Errorf("Close (zweite Sitzung): %v", err)
		}
	})

	m, err := zweiteSitzung.Get(id)
	if err != nil {
		t.Fatalf("Get nach Neustart: %v", err)
	}
	if m.Nachname != "Yilmaz" {
		t.Errorf("Nachname nach Neustart = %q, erwartet %q", m.Nachname, "Yilmaz")
	}
	if len(m.Mitgliedschaften) != 1 {
		t.Errorf("Mitgliedschaften nach Neustart = %d, erwartet 1", len(m.Mitgliedschaften))
	}
	if m.Mitgliedschaften[0].BeitragCents != beitragImTest {
		t.Errorf("BeitragCents nach Neustart = %d, erwartet %d",
			m.Mitgliedschaften[0].BeitragCents, beitragImTest)
	}
}

func TestCreate_MeldetAlleFehlendenPflichtangabenAufEinmal(t *testing.T) {
	svc := neuerService(t)

	_, err := svc.Create(service.NeuesMitglied{})

	var validierung *service.ValidierungsFehler
	if !errors.As(err, &validierung) {
		t.Fatalf("Create({}) = %v, erwartet *service.ValidierungsFehler", err)
	}

	erwartet := []string{
		"Vorname darf nicht leer sein.",
		"Nachname darf nicht leer sein.",
		"Eintrittsdatum darf nicht leer sein.",
	}
	if len(validierung.Meldungen) != len(erwartet) {
		t.Fatalf("Meldungen = %q, erwartet %d Stück", validierung.Meldungen, len(erwartet))
	}
	for i, meldung := range erwartet {
		if validierung.Meldungen[i] != meldung {
			t.Errorf("Meldung %d = %q, erwartet %q", i, validierung.Meldungen[i], meldung)
		}
	}
}

func TestList_LiefertAktiveMitgliederNachNamenSortiert(t *testing.T) {
	svc := neuerService(t)

	// Bewusst in einer Reihenfolge angelegt, die weder der erwarteten Sortierung
	// noch der ID-Reihenfolge entspricht.
	anlegen := func(vorname, nachname string, beitragCents int64, eintritt string) int64 {
		t.Helper()

		id, err := svc.Create(service.NeuesMitglied{
			Vorname:      vorname,
			Nachname:     nachname,
			BeitragCents: beitragCents,
			Eintritt:     datum(t, eintritt),
		})
		if err != nil {
			t.Fatalf("Create(%s %s): %v", vorname, nachname, err)
		}

		return id
	}

	schmidtID := anlegen("Bea", "Schmidt", 8000, "2026-02-10")
	bergerAnnaID := anlegen("Anna", "Berger", 6000, "2026-01-15")
	bergerZoeID := anlegen("Zoe", "Berger", 0, "2026-03-01")

	liste, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	erwartet := []struct {
		id       int64
		vorname  string
		nachname string
		beitrag  int64
		eintritt string
	}{
		{bergerAnnaID, "Anna", "Berger", 6000, "2026-01-15"},
		{bergerZoeID, "Zoe", "Berger", 0, "2026-03-01"},
		{schmidtID, "Bea", "Schmidt", 8000, "2026-02-10"},
	}
	if len(liste) != len(erwartet) {
		t.Fatalf("List = %d Einträge, erwartet %d: %+v", len(liste), len(erwartet), liste)
	}

	for i, e := range erwartet {
		eintrag := liste[i]
		if eintrag.MitgliedID != e.id {
			t.Errorf("Eintrag %d: MitgliedID = %d, erwartet %d (%s %s)", i, eintrag.MitgliedID, e.id, e.vorname, e.nachname)
		}
		if eintrag.Vorname != e.vorname || eintrag.Nachname != e.nachname {
			t.Errorf("Eintrag %d: Name = %q %q, erwartet %q %q", i, eintrag.Vorname, eintrag.Nachname, e.vorname, e.nachname)
		}
		if eintrag.BeitragCents != e.beitrag {
			t.Errorf("Eintrag %d (%s): BeitragCents = %d, erwartet %d", i, e.nachname, eintrag.BeitragCents, e.beitrag)
		}
		if !eintrag.Eintritt.Equal(datum(t, e.eintritt)) {
			t.Errorf("Eintrag %d (%s): Eintritt = %v, erwartet %s", i, e.nachname, eintrag.Eintritt, e.eintritt)
		}
		if eintrag.Rueckstand.Offen {
			t.Errorf("Eintrag %d (%s): Rueckstand.Offen = true, erwartet false bei Neuanlage", i, e.nachname)
		}
	}
}

func TestList_LaesstMitgliederOhneLaufendeMitgliedschaftAus(t *testing.T) {
	svc := neuerService(t)

	geblieben, err := svc.Create(service.NeuesMitglied{
		Vorname:      "Nora",
		Nachname:     "Wagner",
		BeitragCents: beitragImTest,
		Eintritt:     datum(t, "2026-01-05"),
	})
	if err != nil {
		t.Fatalf("Create (bleibendes Mitglied): %v", err)
	}

	ausgetreten, err := svc.Create(service.NeuesMitglied{
		Vorname:      "Olaf",
		Nachname:     "Vogel",
		BeitragCents: beitragImTest,
		Eintritt:     datum(t, "2026-01-05"),
	})
	if err != nil {
		t.Fatalf("Create (austretendes Mitglied): %v", err)
	}
	if err := svc.SetKuendigung(ausgetreten, austrittZum(datum(t, "2026-06-30"))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	liste, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(liste) != 1 {
		t.Fatalf("List = %d Einträge, erwartet 1 (nur das aktive Mitglied): %+v", len(liste), liste)
	}
	if liste[0].MitgliedID != geblieben {
		t.Errorf("List enthält Mitglied %d, erwartet %d (Ausgetretene gehören nicht in die Liste)",
			liste[0].MitgliedID, geblieben)
	}

	// Der Datensatz selbst bleibt erhalten — nur die Liste zeigt ihn nicht mehr.
	if _, err := svc.Get(ausgetreten); err != nil {
		t.Errorf("Get(%d) nach Austritt: %v — der Mitglied-Datensatz muss erhalten bleiben", ausgetreten, err)
	}
}

func TestList_OhneMitgliederIstLeer(t *testing.T) {
	svc := neuerService(t)

	liste, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(liste) != 0 {
		t.Fatalf("List = %+v, erwartet leer bei frischer Datenbank", liste)
	}
}

func TestList_SortiertUmlauteNachDeutschenRegeln(t *testing.T) {
	svc := neuerService(t)

	for _, nachname := range []string{"Zimmermann", "Öztürk", "Ärmel", "Adler"} {
		if _, err := svc.Create(service.NeuesMitglied{
			Vorname:      "Kim",
			Nachname:     nachname,
			BeitragCents: beitragImTest,
			Eintritt:     datum(t, "2026-01-05"),
		}); err != nil {
			t.Fatalf("Create(%s): %v", nachname, err)
		}
	}

	liste, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	// Deutsche Sortierung: Umlaute zählen wie ihr Grundbuchstabe, "Ärmel" gehört
	// also zwischen "Adler" und "Öztürk" — nicht ans Ende hinter "Zimmermann".
	erwartet := []string{"Adler", "Ärmel", "Öztürk", "Zimmermann"}
	if len(liste) != len(erwartet) {
		t.Fatalf("List = %d Einträge, erwartet %d", len(liste), len(erwartet))
	}
	for i, name := range erwartet {
		if liste[i].Nachname != name {
			var ist []string
			for _, e := range liste {
				ist = append(ist, e.Nachname)
			}
			t.Fatalf("Reihenfolge = %v, erwartet %v", ist, erwartet)
		}
	}
}

// zeiger liefert einen Zeiger auf einen Wert — im Patch bedeutet "gesetzt"
// genau das: ein Feld, das nicht nil ist, wird geschrieben.
func zeiger[T any](v T) *T {
	return &v
}

func TestUpdate_SchreibtNurDieGesetztenFelder(t *testing.T) {
	svc := neuerService(t)

	geburtsdatum := datum(t, "1988-11-02")
	eintritt := datum(t, "2026-01-15")

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:      "Jonas",
		Nachname:     "Krüger",
		Geburtsdatum: &geburtsdatum,
		Anschrift:    service.Anschrift{Adresse: "Hauptstraße 1", Postleitzahl: "10115", Ort: "Berlin"},
		Email:        "jonas@example.org",
		Telefon:      "030 111111",
		BeitragCents: beitragImTest,
		Eintritt:     eintritt,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Nur zwei der sechs bearbeitbaren Angaben werden gesetzt.
	if err := svc.Update(id, service.MitgliedPatch{
		Nachname: zeiger("Krüger-Wolf"),
		Email:    zeiger("jonas.krueger-wolf@example.org"),
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if m.Nachname != "Krüger-Wolf" {
		t.Errorf("Nachname = %q, erwartet %q", m.Nachname, "Krüger-Wolf")
	}
	if m.Email != "jonas.krueger-wolf@example.org" {
		t.Errorf("Email = %q, erwartet %q", m.Email, "jonas.krueger-wolf@example.org")
	}

	// Alles, was nicht im Patch stand, bleibt unangetastet.
	if m.Vorname != "Jonas" {
		t.Errorf("Vorname = %q, erwartet unverändert %q", m.Vorname, "Jonas")
	}
	if erwartet := (service.Anschrift{Adresse: "Hauptstraße 1", Postleitzahl: "10115", Ort: "Berlin"}); m.Anschrift != erwartet {
		t.Errorf("Anschrift = %+v, erwartet unverändert %+v", m.Anschrift, erwartet)
	}
	if m.Telefon != "030 111111" {
		t.Errorf("Telefon = %q, erwartet unverändert %q", m.Telefon, "030 111111")
	}
	if m.Geburtsdatum == nil || !m.Geburtsdatum.Equal(geburtsdatum) {
		t.Errorf("Geburtsdatum = %v, erwartet unverändert %v", m.Geburtsdatum, geburtsdatum)
	}
	if m.Rueckstand.Offen {
		t.Errorf("Rueckstand.Offen = true, erwartet unverändert false")
	}

	// Die Mitgliedschaft ist von einer Stammdaten-Änderung nicht betroffen.
	if len(m.Mitgliedschaften) != 1 {
		t.Fatalf("Mitgliedschaften = %d, erwartet 1", len(m.Mitgliedschaften))
	}
	if !m.Mitgliedschaften[0].Eintritt.Equal(eintritt) {
		t.Errorf("Eintritt = %v, erwartet unverändert %v", m.Mitgliedschaften[0].Eintritt, eintritt)
	}
	if m.Mitgliedschaften[0].Austritt != nil {
		t.Errorf("Austritt = %v, erwartet unverändert nil", m.Mitgliedschaften[0].Austritt)
	}
	if m.Mitgliedschaften[0].BeitragCents != beitragImTest {
		t.Errorf("BeitragCents = %d, erwartet unverändert %d",
			m.Mitgliedschaften[0].BeitragCents, beitragImTest)
	}
}

func TestUpdate_UnbekannteIDMeldetNichtGefundenUndLegtNichtsAn(t *testing.T) {
	svc := neuerService(t)

	err := svc.Update(4711, service.MitgliedPatch{Vorname: zeiger("Niemand")})
	if !errors.Is(err, service.ErrNichtGefunden) {
		t.Fatalf("Update(4711, …) = %v, erwartet ErrNichtGefunden", err)
	}

	// Kein Silent-Fail heißt auch: kein heimliches Neuanlegen.
	liste, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(liste) != 0 {
		t.Fatalf("List = %+v, erwartet leer — Update darf kein Mitglied anlegen", liste)
	}
}

func TestUpdate_MeldetLeeregemachtePflichtfelderAufEinmal(t *testing.T) {
	svc := neuerService(t)

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:      "Sara",
		Nachname:     "Neumann",
		BeitragCents: beitragImTest,
		Eintritt:     datum(t, "2026-02-01"),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	err = svc.Update(id, service.MitgliedPatch{
		Vorname:  zeiger(""),
		Nachname: zeiger(""),
	})

	var validierung *service.ValidierungsFehler
	if !errors.As(err, &validierung) {
		t.Fatalf("Update mit leeren Pflichtfeldern = %v, erwartet *service.ValidierungsFehler", err)
	}

	erwartet := []string{
		"Vorname darf nicht leer sein.",
		"Nachname darf nicht leer sein.",
	}
	if len(validierung.Meldungen) != len(erwartet) {
		t.Fatalf("Meldungen = %q, erwartet %d Stück", validierung.Meldungen, len(erwartet))
	}
	for i, meldung := range erwartet {
		if validierung.Meldungen[i] != meldung {
			t.Errorf("Meldung %d = %q, erwartet %q", i, validierung.Meldungen[i], meldung)
		}
	}

	// Abgelehnt heißt: nichts davon ist in der Datenbank gelandet.
	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if m.Vorname != "Sara" || m.Nachname != "Neumann" {
		t.Errorf("Name = %q %q, erwartet unverändert %q %q", m.Vorname, m.Nachname, "Sara", "Neumann")
	}
}

func TestUpdate_LeererPatchAendertNichtsUndPrueftTrotzdemDieID(t *testing.T) {
	svc := neuerService(t)

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:      "Pia",
		Nachname:     "Roth",
		BeitragCents: beitragImTest,
		Eintritt:     datum(t, "2026-02-01"),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	vorher, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get (vorher): %v", err)
	}

	if err := svc.Update(id, service.MitgliedPatch{}); err != nil {
		t.Fatalf("Update mit leerem Patch: %v", err)
	}

	nachher, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get (nachher): %v", err)
	}
	if nachher.Vorname != vorher.Vorname || nachher.Nachname != vorher.Nachname ||
		mitgliedschaftText(nachher.Mitgliedschaften[0]) != mitgliedschaftText(vorher.Mitgliedschaften[0]) {
		t.Errorf("Mitglied = %+v, erwartet unverändert %+v", nachher, vorher)
	}

	// Auch ohne zu schreibende Felder darf eine unbekannte ID nicht durchgehen.
	if err := svc.Update(4711, service.MitgliedPatch{}); !errors.Is(err, service.ErrNichtGefunden) {
		t.Fatalf("Update(4711, {}) = %v, erwartet ErrNichtGefunden", err)
	}
}

func TestLaufendeMitgliedschaft_LiefertDenOffenenZeitraumUndNachAustrittNil(t *testing.T) {
	svc := neuerService(t)

	eintritt := datum(t, "2026-01-15")

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:      "Ida",
		Nachname:     "Sommer",
		BeitragCents: beitragImTest,
		Eintritt:     eintritt,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	laufend := m.LaufendeMitgliedschaft()
	if laufend == nil {
		t.Fatal("LaufendeMitgliedschaft = nil, erwartet den offenen Zeitraum")
	}
	if !laufend.Eintritt.Equal(eintritt) {
		t.Errorf("Eintritt = %v, erwartet %v", laufend.Eintritt, eintritt)
	}

	if err := svc.SetKuendigung(id, austrittZum(datum(t, "2026-06-30"))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	m, err = svc.Get(id)
	if err != nil {
		t.Fatalf("Get nach Austritt: %v", err)
	}
	if laufend := m.LaufendeMitgliedschaft(); laufend != nil {
		t.Errorf("LaufendeMitgliedschaft = %+v, erwartet nil nach dem Austritt", laufend)
	}
}

// heuteVersetzt liefert ein Datum relativ zum heutigen Tag. Aus- und
// Wiedereintritt werden gegen "heute" geprüft, also müssen die Fixtures
// mitwandern statt feste Kalendertage zu setzen.
func heuteVersetzt(tage int) time.Time {
	// Kalendertag von heute, in UTC wie alle aus der Datenbank gelesenen Daten —
	// so lassen sich die Werte direkt mit Equal vergleichen.
	jetzt := time.Now()
	return time.Date(jetzt.Year(), jetzt.Month(), jetzt.Day(), 0, 0, 0, 0, time.UTC).
		AddDate(0, 0, tage)
}

// mitgliedAnlegen legt ein Mitglied mit Minimalangaben an und liefert dessen ID.
func mitgliedAnlegen(t *testing.T, svc *service.MemberService, vorname, nachname string) int64 {
	t.Helper()

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:      vorname,
		Nachname:     nachname,
		BeitragCents: beitragImTest,
		Eintritt:     datum(t, "2026-01-05"),
	})
	if err != nil {
		t.Fatalf("Create(%s %s): %v", vorname, nachname, err)
	}

	return id
}

// Der Rückstand ist zweiwertig: in Ordnung oder im Rückstand. Einen neutralen
// Zustand gibt es nicht mehr — bei Lastschrift bedeutet die Abwesenheit einer
// Rückgabe tatsächlich, dass gezahlt wurde (ADR-0006).
func TestRueckstand_IstZweiwertigUndStartetInOrdnung(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Ravi", "Kumar")

	eintrag, err := svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag: %v", err)
	}
	if eintrag.Rueckstand.Offen {
		t.Errorf("Rueckstand.Offen = true, erwartet false bei Neuanlage")
	}
	if eintrag.Rueckstand.Notiz != "" {
		t.Errorf("Rueckstand.Notiz = %q, erwartet leer bei Neuanlage", eintrag.Rueckstand.Notiz)
	}

	const notiz = "Rücklastschrift Oktober, angeschrieben am 05.10."
	if err := svc.SetRueckstand(id, service.Rueckstand{Offen: true, Notiz: notiz}); err != nil {
		t.Fatalf("SetRueckstand(true): %v", err)
	}

	eintrag, err = svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag nach dem Setzen: %v", err)
	}
	if !eintrag.Rueckstand.Offen {
		t.Errorf("Rueckstand.Offen = false, erwartet true nach dem Setzen")
	}
	if eintrag.Rueckstand.Notiz != notiz {
		t.Errorf("Rueckstand.Notiz = %q, erwartet %q", eintrag.Rueckstand.Notiz, notiz)
	}
}

// Die Notiz ist änderbar, ohne dass das Kennzeichen wechselt: der Vorgang
// entwickelt sich weiter, während das Geld offen bleibt.
func TestSetRueckstand_AendertDieNotizOhneDasKennzeichenZuWechseln(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Sina", "Petrov")
	if err := svc.SetRueckstand(id, service.Rueckstand{Offen: true, Notiz: "Rücklastschrift Oktober"}); err != nil {
		t.Fatalf("SetRueckstand: %v", err)
	}

	const nachgefasst = "Rücklastschrift Oktober, zweite Mahnung am 19.10."
	if err := svc.SetRueckstand(id, service.Rueckstand{Offen: true, Notiz: nachgefasst}); err != nil {
		t.Fatalf("SetRueckstand (Notiz ändern): %v", err)
	}

	eintrag, err := svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag: %v", err)
	}
	if !eintrag.Rueckstand.Offen || eintrag.Rueckstand.Notiz != nachgefasst {
		t.Errorf("Eintrag = (%v, %q), erwartet (true, %q)",
			eintrag.Rueckstand.Offen, eintrag.Rueckstand.Notiz, nachgefasst)
	}
}

// Beim Aufheben bleibt die Notiz stehen. Sie ist die einzige Spur, die ein
// erledigter Vorgang in v1 hinterlässt — eine Zahlungshistorie gibt es nicht
// (ADR-0006). Wer sie loswerden will, leert sie ausdrücklich.
func TestSetRueckstand_LaesstDieNotizBeimAufhebenStehen(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Ilse", "Brandt")
	const notiz = "Rücklastschrift März, im April nachgezahlt"
	if err := svc.SetRueckstand(id, service.Rueckstand{Offen: true, Notiz: notiz}); err != nil {
		t.Fatalf("SetRueckstand(true): %v", err)
	}

	if err := svc.SetRueckstand(id, service.Rueckstand{Offen: false, Notiz: notiz}); err != nil {
		t.Fatalf("SetRueckstand(false): %v", err)
	}

	eintrag, err := svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag: %v", err)
	}
	if eintrag.Rueckstand.Offen {
		t.Errorf("Rueckstand.Offen = true, erwartet false nach dem Aufheben")
	}
	if eintrag.Rueckstand.Notiz != notiz {
		t.Errorf("Rueckstand.Notiz = %q, erwartet unverändert %q", eintrag.Rueckstand.Notiz, notiz)
	}

	// Leeren ist der ausdrückliche Weg, die Notiz loszuwerden.
	if err := svc.SetRueckstand(id, service.Rueckstand{Offen: false}); err != nil {
		t.Fatalf("SetRueckstand(false, \"\"): %v", err)
	}

	eintrag, err = svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag nach dem Leeren: %v", err)
	}
	if eintrag.Rueckstand.Notiz != "" {
		t.Errorf("Rueckstand.Notiz = %q, erwartet leer", eintrag.Rueckstand.Notiz)
	}
}

func TestSetRueckstand_IstBeiWiederholungIdempotent(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Sina", "Petrov")
	const notiz = "Rücklastschrift November"

	for versuch := 1; versuch <= 2; versuch++ {
		if err := svc.SetRueckstand(id, service.Rueckstand{Offen: true, Notiz: notiz}); err != nil {
			t.Fatalf("SetRueckstand (Versuch %d): %v", versuch, err)
		}

		eintrag, err := svc.Eintrag(id)
		if err != nil {
			t.Fatalf("Eintrag (Versuch %d): %v", versuch, err)
		}
		if !eintrag.Rueckstand.Offen || eintrag.Rueckstand.Notiz != notiz {
			t.Fatalf("Eintrag nach Versuch %d = (%v, %q), erwartet (true, %q)",
				versuch, eintrag.Rueckstand.Offen, eintrag.Rueckstand.Notiz, notiz)
		}
	}

	liste, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(liste) != 1 {
		t.Errorf("List = %d Einträge, erwartet 1 — es darf kein zweites Mitglied entstanden sein", len(liste))
	}
}

// Der Rückstand hängt am Mitglied, nicht an der Mitgliedschaft: ein Austritt
// erlässt keine Schulden. Geprüft am Service-Seam, weil genau hier die
// Zuordnung entschieden wird.
func TestSetRueckstand_BleibtNachDemAustrittBestehen(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Nina", "Klein")
	const notiz = "Rücklastschrift Juni, noch offen"
	if err := svc.SetRueckstand(id, service.Rueckstand{Offen: true, Notiz: notiz}); err != nil {
		t.Fatalf("SetRueckstand: %v", err)
	}

	if err := svc.SetKuendigung(id, austrittZum(heuteVersetzt(-10))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	eintrag, err := svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag: %v", err)
	}
	if !eintrag.Rueckstand.Offen || eintrag.Rueckstand.Notiz != notiz {
		t.Errorf("Eintrag nach dem Austritt = (%v, %q), erwartet (true, %q)",
			eintrag.Rueckstand.Offen, eintrag.Rueckstand.Notiz, notiz)
	}

	// Auch ein Wiedereintritt ändert daran nichts: die neue Mitgliedschaft
	// beginnt, der alte Rückstand bleibt.
	if err := svc.Rejoin(id, heuteVersetzt(-1)); err != nil {
		t.Fatalf("Rejoin: %v", err)
	}

	eintrag, err = svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag nach dem Wiedereintritt: %v", err)
	}
	if !eintrag.Rueckstand.Offen || eintrag.Rueckstand.Notiz != notiz {
		t.Errorf("Eintrag nach dem Wiedereintritt = (%v, %q), erwartet (true, %q)",
			eintrag.Rueckstand.Offen, eintrag.Rueckstand.Notiz, notiz)
	}
}

func TestSetRueckstand_LaesstAlleAnderenFelderUnberuehrt(t *testing.T) {
	svc := neuerService(t)

	geburtsdatum := datum(t, "1994-07-19")

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:      "Lina",
		Nachname:     "Fischer",
		Geburtsdatum: &geburtsdatum,
		Anschrift:    service.Anschrift{Adresse: "Hauptstraße 3", Postleitzahl: "10115", Ort: "Berlin"},
		Email:        "lina@example.org",
		Telefon:      "030 123456",
		BeitragCents: beitragImTest,
		Eintritt:     datum(t, "2026-02-01"),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	vorher, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get (vorher): %v", err)
	}

	if err := svc.SetRueckstand(id, service.Rueckstand{Offen: true, Notiz: "Rücklastschrift Februar"}); err != nil {
		t.Fatalf("SetRueckstand: %v", err)
	}

	nachher, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get (nachher): %v", err)
	}

	if !nachher.Rueckstand.Offen || nachher.Rueckstand.Notiz != "Rücklastschrift Februar" {
		t.Errorf("Rückstand = (%v, %q), erwartet (true, \"Rücklastschrift Februar\")",
			nachher.Rueckstand.Offen, nachher.Rueckstand.Notiz)
	}

	// Alles außer dem Rückstand muss identisch geblieben sein — inklusive der
	// Mitgliedschaft, die von einem Rückstand nichts wissen darf.
	if nachher.ID != vorher.ID || nachher.Vorname != vorher.Vorname ||
		nachher.Nachname != vorher.Nachname || nachher.Anschrift != vorher.Anschrift ||
		nachher.Email != vorher.Email || nachher.Telefon != vorher.Telefon {
		t.Errorf("Stammdaten = %+v, erwartet unverändert %+v", nachher, vorher)
	}
	if nachher.Geburtsdatum == nil || !nachher.Geburtsdatum.Equal(geburtsdatum) {
		t.Errorf("Geburtsdatum = %v, erwartet %v", nachher.Geburtsdatum, geburtsdatum)
	}
	if len(nachher.Mitgliedschaften) != len(vorher.Mitgliedschaften) {
		t.Fatalf("Mitgliedschaften = %+v, erwartet unverändert %+v",
			nachher.Mitgliedschaften, vorher.Mitgliedschaften)
	}
	for i, ms := range nachher.Mitgliedschaften {
		if mitgliedschaftText(ms) != mitgliedschaftText(vorher.Mitgliedschaften[i]) {
			t.Errorf("Mitgliedschaft %d = %+v, erwartet unverändert %+v",
				i, ms, vorher.Mitgliedschaften[i])
		}
	}
}

// Normalisiert wird, wo gespeichert wird: umschließender Leerraum aus einem
// Formularfeld darf nicht als Notiz durchgehen und "leeren" verhindern.
func TestSetRueckstand_SchneidetLeerraumAusDerNotiz(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Jonas", "Weber")
	if err := svc.SetRueckstand(id, service.Rueckstand{
		Offen: true,
		Notiz: "  Rücklastschrift Oktober\t",
	}); err != nil {
		t.Fatalf("SetRueckstand: %v", err)
	}

	eintrag, err := svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag: %v", err)
	}
	if eintrag.Rueckstand.Notiz != "Rücklastschrift Oktober" {
		t.Errorf("Rueckstand.Notiz = %q, erwartet ohne umschließenden Leerraum", eintrag.Rueckstand.Notiz)
	}

	// Eine Notiz aus lauter Leerraum ist keine Notiz.
	if err := svc.SetRueckstand(id, service.Rueckstand{Offen: true, Notiz: "   "}); err != nil {
		t.Fatalf("SetRueckstand (nur Leerraum): %v", err)
	}

	eintrag, err = svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag nach Leerraum-Notiz: %v", err)
	}
	if eintrag.Rueckstand.Notiz != "" {
		t.Errorf("Rueckstand.Notiz = %q, erwartet leer", eintrag.Rueckstand.Notiz)
	}
}

// Die Bezeichnung steht am Typ und nicht im Template, damit Liste und spätere
// Ansichten dieselben Worte benutzen.
func TestRueckstand_BezeichnungIstZweiwertig(t *testing.T) {
	if got := (service.Rueckstand{Offen: true}).Bezeichnung(); got != "im Rückstand" {
		t.Errorf("Bezeichnung bei offen = %q, erwartet \"im Rückstand\"", got)
	}
	// Der Nullwert ist "in Ordnung" — der Normalfall beim Lastschrifteinzug.
	if got := (service.Rueckstand{}).Bezeichnung(); got != "in Ordnung" {
		t.Errorf("Bezeichnung des Nullwerts = %q, erwartet \"in Ordnung\"", got)
	}
}

func TestSetRueckstand_UnbekannteIDMeldetNichtGefunden(t *testing.T) {
	svc := neuerService(t)

	if err := svc.SetRueckstand(4711, service.Rueckstand{Offen: true, Notiz: "egal"}); !errors.Is(err, service.ErrNichtGefunden) {
		t.Fatalf("SetRueckstand(4711, …) = %v, erwartet ErrNichtGefunden", err)
	}

	liste, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(liste) != 0 {
		t.Errorf("List = %+v, erwartet leer — es darf nichts angelegt worden sein", liste)
	}
}

func TestEintrag_LiefertDieselbeZeileWieDieListe(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Mara", "Delgado")
	if err := svc.SetRueckstand(id, service.Rueckstand{Offen: true, Notiz: "Rücklastschrift September"}); err != nil {
		t.Fatalf("SetRueckstand: %v", err)
	}

	liste, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(liste) != 1 {
		t.Fatalf("List = %d Einträge, erwartet 1", len(liste))
	}

	eintrag, err := svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag: %v", err)
	}
	if eintrag.MitgliedID != liste[0].MitgliedID || eintrag.Vorname != liste[0].Vorname ||
		eintrag.Nachname != liste[0].Nachname || eintrag.BeitragCents != liste[0].BeitragCents ||
		eintrag.Rueckstand.Offen != liste[0].Rueckstand.Offen ||
		eintrag.Rueckstand.Notiz != liste[0].Rueckstand.Notiz ||
		!eintrag.Eintritt.Equal(liste[0].Eintritt) {
		t.Errorf("Eintrag = %+v, erwartet dieselbe Zeile wie List: %+v", eintrag, liste[0])
	}
}

func TestEintrag_UnbekannteIDMeldetNichtGefunden(t *testing.T) {
	svc := neuerService(t)

	if _, err := svc.Eintrag(4711); !errors.Is(err, service.ErrNichtGefunden) {
		t.Fatalf("Eintrag(4711) = %v, erwartet ErrNichtGefunden", err)
	}
}

// Seit die Liste auf Wunsch auch Ehemalige zeigt, muss sich deren Zeile auch
// einzeln neu rendern lassen — sonst liefe jede Aktion in einer solchen Zeile
// ins Leere.
func TestEintrag_LiefertAuchDieZeileEinesAusgetretenenMitglieds(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Timo", "Vogel")
	austritt := heuteVersetzt(-10)
	if err := svc.SetKuendigung(id, austrittZum(austritt)); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	eintrag, err := svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag nach Austritt: %v", err)
	}
	if eintrag.MitgliedID != id {
		t.Errorf("MitgliedID = %d, erwartet %d", eintrag.MitgliedID, id)
	}
	if eintrag.Austritt == nil || !eintrag.Austritt.Equal(austritt) {
		t.Errorf("Austritt = %v, erwartet %v", eintrag.Austritt, austritt)
	}

	// In der Standardansicht taucht das Mitglied trotzdem nicht auf.
	liste, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(liste) != 0 {
		t.Errorf("List = %+v, erwartet leer — Ausgetretene gehören nicht in die Standardansicht", liste)
	}
}
