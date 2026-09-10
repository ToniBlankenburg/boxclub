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

// beitragsklassen liefert die geseedeten Klassen, aufsteigend nach Preis —
// die Tests brauchen deren IDs, um Mitglieder anlegen zu können.
func beitragsklassen(t *testing.T, svc *service.MemberService) []service.Beitragsklasse {
	t.Helper()

	k, err := svc.AktiveBeitragsklassen()
	if err != nil {
		t.Fatalf("AktiveBeitragsklassen: %v", err)
	}

	return k
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

	klassen := beitragsklassen(t, svc)
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

	klassen := beitragsklassen(t, svc)

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

	klassen := beitragsklassen(t, ersteSitzung)

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

	klassen := beitragsklassen(t, svc)

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

func TestList_LiefertAktiveMitgliederNachNamenSortiert(t *testing.T) {
	svc := neuerService(t)

	klassen := beitragsklassen(t, svc)
	einMalWoche, zweiMalWoche := klassen[0], klassen[1]

	// Bewusst in einer Reihenfolge angelegt, die weder der erwarteten Sortierung
	// noch der ID-Reihenfolge entspricht.
	anlegen := func(vorname, nachname string, klasse service.Beitragsklasse, eintritt string) int64 {
		t.Helper()

		id, err := svc.Create(service.NeuesMitglied{
			Vorname:          vorname,
			Nachname:         nachname,
			BeitragsklasseID: klasse.ID,
			Eintritt:         datum(t, eintritt),
		})
		if err != nil {
			t.Fatalf("Create(%s %s): %v", vorname, nachname, err)
		}

		return id
	}

	schmidtID := anlegen("Bea", "Schmidt", zweiMalWoche, "2026-02-10")
	bergerAnnaID := anlegen("Anna", "Berger", einMalWoche, "2026-01-15")
	bergerZoeID := anlegen("Zoe", "Berger", zweiMalWoche, "2026-03-01")

	liste, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	erwartet := []struct {
		id       int64
		vorname  string
		nachname string
		klasse   service.Beitragsklasse
		eintritt string
	}{
		{bergerAnnaID, "Anna", "Berger", einMalWoche, "2026-01-15"},
		{bergerZoeID, "Zoe", "Berger", zweiMalWoche, "2026-03-01"},
		{schmidtID, "Bea", "Schmidt", zweiMalWoche, "2026-02-10"},
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
		if eintrag.Beitragsklasse != e.klasse {
			t.Errorf("Eintrag %d (%s): Beitragsklasse = %+v, erwartet %+v", i, e.nachname, eintrag.Beitragsklasse, e.klasse)
		}
		if !eintrag.Eintritt.Equal(datum(t, e.eintritt)) {
			t.Errorf("Eintrag %d (%s): Eintritt = %v, erwartet %s", i, e.nachname, eintrag.Eintritt, e.eintritt)
		}
		if eintrag.BezahltBis != nil {
			t.Errorf("Eintrag %d (%s): BezahltBis = %v, erwartet nil bei Neuanlage", i, e.nachname, eintrag.BezahltBis)
		}
	}
}

func TestList_LaesstMitgliederOhneLaufendeMitgliedschaftAus(t *testing.T) {
	svc := neuerService(t)

	klassen := beitragsklassen(t, svc)

	geblieben, err := svc.Create(service.NeuesMitglied{
		Vorname:          "Nora",
		Nachname:         "Wagner",
		BeitragsklasseID: klassen[0].ID,
		Eintritt:         datum(t, "2026-01-05"),
	})
	if err != nil {
		t.Fatalf("Create (bleibendes Mitglied): %v", err)
	}

	ausgetreten, err := svc.Create(service.NeuesMitglied{
		Vorname:          "Olaf",
		Nachname:         "Vogel",
		BeitragsklasseID: klassen[1].ID,
		Eintritt:         datum(t, "2026-01-05"),
	})
	if err != nil {
		t.Fatalf("Create (austretendes Mitglied): %v", err)
	}
	if err := svc.AustrittFuerTest(ausgetreten, datum(t, "2026-06-30")); err != nil {
		t.Fatalf("AustrittFuerTest: %v", err)
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

	klasse := beitragsklassen(t, svc)[0]
	for _, nachname := range []string{"Zimmermann", "Öztürk", "Ärmel", "Adler"} {
		if _, err := svc.Create(service.NeuesMitglied{
			Vorname:          "Kim",
			Nachname:         nachname,
			BeitragsklasseID: klasse.ID,
			Eintritt:         datum(t, "2026-01-05"),
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
