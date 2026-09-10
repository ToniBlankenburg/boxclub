package service_test

import (
	"errors"
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

func TestOpen_SeedetDieBeidenBeitragsklassen(t *testing.T) {
	svc := neuerService(t)

	klassen, err := svc.AktiveBeitragsklassen()
	if err != nil {
		t.Fatalf("AktiveBeitragsklassen: %v", err)
	}

	if len(klassen) != 2 {
		t.Fatalf("erwarte 2 geseedete Beitragsklassen, bekam %d: %+v", len(klassen), klassen)
	}

	erwartet := []struct {
		name  string
		cents int64
	}{
		{"Erwachsen 1×/Woche", 6000},
		{"Erwachsen 2×/Woche", 8000},
	}
	for i, e := range erwartet {
		if klassen[i].Name != e.name {
			t.Errorf("Klasse %d: Name = %q, erwartet %q", i, klassen[i].Name, e.name)
		}
		if klassen[i].PreisMonatlichCents != e.cents {
			t.Errorf("Klasse %d (%s): PreisMonatlichCents = %d, erwartet %d",
				i, e.name, klassen[i].PreisMonatlichCents, e.cents)
		}
		if !klassen[i].Aktiv {
			t.Errorf("Klasse %d (%s): Aktiv = false, erwartet true", i, e.name)
		}
	}
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

	klassen, err := svc.AktiveBeitragsklassen()
	if err != nil {
		t.Fatalf("AktiveBeitragsklassen: %v", err)
	}
	zweiMalWoche := klassen[1]

	geburtsdatum := datum(t, "1990-04-17")
	eintritt := datum(t, "2026-01-15")

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:          "Anna",
		Nachname:         "Berger",
		Geburtsdatum:     &geburtsdatum,
		Adresse:          "Ringstraße 5, 12043 Berlin",
		Email:            "anna.berger@example.org",
		Telefon:          "030 1234567",
		BeitragsklasseID: zweiMalWoche.ID,
		Eintritt:         eintritt,
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
	if m.Adresse != "Ringstraße 5, 12043 Berlin" {
		t.Errorf("Adresse = %q", m.Adresse)
	}
	if m.Email != "anna.berger@example.org" {
		t.Errorf("Email = %q", m.Email)
	}
	if m.Telefon != "030 1234567" {
		t.Errorf("Telefon = %q", m.Telefon)
	}
	if m.BeitragsklasseID != zweiMalWoche.ID {
		t.Errorf("BeitragsklasseID = %d, erwartet %d", m.BeitragsklasseID, zweiMalWoche.ID)
	}
	if m.BezahltBis != nil {
		t.Errorf("BezahltBis = %v, erwartet nil bei Neuanlage", m.BezahltBis)
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
}

func TestCreate_LehntUnbekannteBeitragsklasseAb(t *testing.T) {
	svc := neuerService(t)

	_, err := svc.Create(service.NeuesMitglied{
		Vorname:          "Carl",
		Nachname:         "Dietrich",
		BeitragsklasseID: 999,
		Eintritt:         datum(t, "2026-02-01"),
	})
	if err == nil {
		t.Fatal("Create mit unbekannter BeitragsklasseID muss fehlschlagen")
	}
}

func TestCreate_ErlaubtNamensgleichheitBeiGleichemGeburtsdatum(t *testing.T) {
	svc := neuerService(t)

	klassen, err := svc.AktiveBeitragsklassen()
	if err != nil {
		t.Fatalf("AktiveBeitragsklassen: %v", err)
	}

	geburtsdatum := datum(t, "2001-09-03")
	stammdaten := service.NeuesMitglied{
		Vorname:          "Max",
		Nachname:         "Müller",
		Geburtsdatum:     &geburtsdatum,
		BeitragsklasseID: klassen[0].ID,
		Eintritt:         datum(t, "2026-03-01"),
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

func TestOpen_BestehendeDatenbankBehaeltMitgliederUndSeedetNichtDoppelt(t *testing.T) {
	dbPfad := filepath.Join(t.TempDir(), "boxclub.db")

	ersteSitzung, err := service.Open(dbPfad)
	if err != nil {
		t.Fatalf("service.Open (erste Sitzung): %v", err)
	}

	klassen, err := ersteSitzung.AktiveBeitragsklassen()
	if err != nil {
		t.Fatalf("AktiveBeitragsklassen: %v", err)
	}

	id, err := ersteSitzung.Create(service.NeuesMitglied{
		Vorname:          "Elif",
		Nachname:         "Yilmaz",
		BeitragsklasseID: klassen[0].ID,
		Eintritt:         datum(t, "2026-04-20"),
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

	nachNeustart, err := zweiteSitzung.AktiveBeitragsklassen()
	if err != nil {
		t.Fatalf("AktiveBeitragsklassen nach Neustart: %v", err)
	}
	if len(nachNeustart) != 2 {
		t.Errorf("Beitragsklassen nach Neustart = %d, erwartet 2 (Seed darf nicht doppeln)", len(nachNeustart))
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
		"Bitte eine Beitragsklasse wählen.",
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

func TestBeitragsklasse_LiefertEinzelneKlasseUndMeldetUnbekannte(t *testing.T) {
	svc := neuerService(t)

	klassen, err := svc.AktiveBeitragsklassen()
	if err != nil {
		t.Fatalf("AktiveBeitragsklassen: %v", err)
	}

	klasse, err := svc.Beitragsklasse(klassen[1].ID)
	if err != nil {
		t.Fatalf("Beitragsklasse: %v", err)
	}
	if klasse != klassen[1] {
		t.Errorf("Beitragsklasse(%d) = %+v, erwartet %+v", klassen[1].ID, klasse, klassen[1])
	}

	if _, err := svc.Beitragsklasse(999); !errors.Is(err, service.ErrNichtGefunden) {
		t.Fatalf("Beitragsklasse(999) = %v, erwartet ErrNichtGefunden", err)
	}
}
