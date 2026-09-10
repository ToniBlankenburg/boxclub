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

// zeiger liefert einen Zeiger auf einen Wert — im Patch bedeutet "gesetzt"
// genau das: ein Feld, das nicht nil ist, wird geschrieben.
func zeiger[T any](v T) *T {
	return &v
}

func TestUpdate_SchreibtNurDieGesetztenFelder(t *testing.T) {
	svc := neuerService(t)

	klassen := beitragsklassen(t, svc)
	geburtsdatum := datum(t, "1988-11-02")
	eintritt := datum(t, "2026-01-15")

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:          "Jonas",
		Nachname:         "Krüger",
		Geburtsdatum:     &geburtsdatum,
		Adresse:          "Hauptstraße 1, 10115 Berlin",
		Email:            "jonas@example.org",
		Telefon:          "030 111111",
		BeitragsklasseID: klassen[0].ID,
		Eintritt:         eintritt,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Nur zwei der sechs bearbeitbaren Felder werden gesetzt.
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
	if m.Adresse != "Hauptstraße 1, 10115 Berlin" {
		t.Errorf("Adresse = %q, erwartet unverändert", m.Adresse)
	}
	if m.Telefon != "030 111111" {
		t.Errorf("Telefon = %q, erwartet unverändert %q", m.Telefon, "030 111111")
	}
	if m.BeitragsklasseID != klassen[0].ID {
		t.Errorf("BeitragsklasseID = %d, erwartet unverändert %d", m.BeitragsklasseID, klassen[0].ID)
	}
	if m.Geburtsdatum == nil || !m.Geburtsdatum.Equal(geburtsdatum) {
		t.Errorf("Geburtsdatum = %v, erwartet unverändert %v", m.Geburtsdatum, geburtsdatum)
	}
	if m.BezahltBis != nil {
		t.Errorf("BezahltBis = %v, erwartet unverändert nil", m.BezahltBis)
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

func TestUpdate_WechseltBeitragsklasseUndListeZeigtSie(t *testing.T) {
	svc := neuerService(t)

	klassen := beitragsklassen(t, svc)
	einMalWoche, zweiMalWoche := klassen[0], klassen[1]

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:          "Lena",
		Nachname:         "Hoffmann",
		BeitragsklasseID: einMalWoche.ID,
		Eintritt:         datum(t, "2026-02-01"),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := svc.Update(id, service.MitgliedPatch{BeitragsklasseID: &zweiMalWoche.ID}); err != nil {
		t.Fatalf("Update (Klassenwechsel): %v", err)
	}

	liste, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(liste) != 1 {
		t.Fatalf("List = %d Einträge, erwartet 1", len(liste))
	}
	if liste[0].Beitragsklasse != zweiMalWoche {
		t.Errorf("Beitragsklasse in der Liste = %+v, erwartet %+v", liste[0].Beitragsklasse, zweiMalWoche)
	}
}

func TestUpdate_LehntUnbekannteBeitragsklasseAbUndLaesstDieAlteStehen(t *testing.T) {
	svc := neuerService(t)

	klassen := beitragsklassen(t, svc)

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:          "Timo",
		Nachname:         "Ludwig",
		BeitragsklasseID: klassen[0].ID,
		Eintritt:         datum(t, "2026-02-01"),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := svc.Update(id, service.MitgliedPatch{BeitragsklasseID: zeiger(int64(999))}); err == nil {
		t.Fatal("Update mit unbekannter BeitragsklasseID muss fehlschlagen")
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if m.BeitragsklasseID != klassen[0].ID {
		t.Errorf("BeitragsklasseID = %d, erwartet unverändert %d", m.BeitragsklasseID, klassen[0].ID)
	}
}

func TestUpdate_MeldetLeeregemachtePflichtfelderAufEinmal(t *testing.T) {
	svc := neuerService(t)

	klassen := beitragsklassen(t, svc)

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:          "Sara",
		Nachname:         "Neumann",
		BeitragsklasseID: klassen[0].ID,
		Eintritt:         datum(t, "2026-02-01"),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	err = svc.Update(id, service.MitgliedPatch{
		Vorname:          zeiger(""),
		Nachname:         zeiger(""),
		BeitragsklasseID: zeiger(int64(0)),
	})

	var validierung *service.ValidierungsFehler
	if !errors.As(err, &validierung) {
		t.Fatalf("Update mit leeren Pflichtfeldern = %v, erwartet *service.ValidierungsFehler", err)
	}

	erwartet := []string{
		"Vorname darf nicht leer sein.",
		"Nachname darf nicht leer sein.",
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

	klassen := beitragsklassen(t, svc)

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:          "Pia",
		Nachname:         "Roth",
		BeitragsklasseID: klassen[0].ID,
		Eintritt:         datum(t, "2026-02-01"),
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
		nachher.BeitragsklasseID != vorher.BeitragsklasseID {
		t.Errorf("Mitglied = %+v, erwartet unverändert %+v", nachher, vorher)
	}

	// Auch ohne zu schreibende Felder darf eine unbekannte ID nicht durchgehen.
	if err := svc.Update(4711, service.MitgliedPatch{}); !errors.Is(err, service.ErrNichtGefunden) {
		t.Fatalf("Update(4711, {}) = %v, erwartet ErrNichtGefunden", err)
	}
}

func TestLaufendeMitgliedschaft_LiefertDenOffenenZeitraumUndNachAustrittNil(t *testing.T) {
	svc := neuerService(t)

	klassen := beitragsklassen(t, svc)
	eintritt := datum(t, "2026-01-15")

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:          "Ida",
		Nachname:         "Sommer",
		BeitragsklasseID: klassen[0].ID,
		Eintritt:         eintritt,
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

	if err := svc.AustrittFuerTest(id, datum(t, "2026-06-30")); err != nil {
		t.Fatalf("AustrittFuerTest: %v", err)
	}

	m, err = svc.Get(id)
	if err != nil {
		t.Fatalf("Get nach Austritt: %v", err)
	}
	if laufend := m.LaufendeMitgliedschaft(); laufend != nil {
		t.Errorf("LaufendeMitgliedschaft = %+v, erwartet nil nach dem Austritt", laufend)
	}
}

// heuteVersetzt liefert ein Datum relativ zum heutigen Tag — die Status-Ableitung
// vergleicht gegen "heute", also müssen die Fixtures mitwandern statt feste
// Kalendertage zu setzen.
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
		Vorname:          vorname,
		Nachname:         nachname,
		BeitragsklasseID: beitragsklassen(t, svc)[0].ID,
		Eintritt:         datum(t, "2026-01-05"),
	})
	if err != nil {
		t.Fatalf("Create(%s %s): %v", vorname, nachname, err)
	}

	return id
}

func TestZahlungsstatus_LeitetSichAusBezahltBisUndHeuteAb(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Ravi", "Kumar")

	// Ohne jede Zahlungsangabe ist der Status weder bezahlt noch nicht bezahlt.
	eintrag, err := svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag: %v", err)
	}
	if eintrag.Zahlungsstatus != service.ZahlungsstatusNichtGesetzt {
		t.Errorf("Zahlungsstatus ohne bezahlt_bis = %v, erwartet %v",
			eintrag.Zahlungsstatus, service.ZahlungsstatusNichtGesetzt)
	}

	faelle := []struct {
		name     string
		tage     int
		erwartet service.Zahlungsstatus
	}{
		{"gestern", -1, service.ZahlungsstatusNichtBezahlt},
		{"heute", 0, service.ZahlungsstatusBezahlt},
		{"morgen", 1, service.ZahlungsstatusBezahlt},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			bezahltBis := heuteVersetzt(f.tage)
			if err := svc.SetBezahltBis(id, &bezahltBis); err != nil {
				t.Fatalf("SetBezahltBis: %v", err)
			}

			eintrag, err := svc.Eintrag(id)
			if err != nil {
				t.Fatalf("Eintrag: %v", err)
			}
			if eintrag.Zahlungsstatus != f.erwartet {
				t.Errorf("Zahlungsstatus bei bezahlt_bis = %s = %v, erwartet %v",
					f.name, eintrag.Zahlungsstatus, f.erwartet)
			}
			if eintrag.BezahltBis == nil || !eintrag.BezahltBis.Equal(bezahltBis) {
				t.Errorf("BezahltBis = %v, erwartet %v", eintrag.BezahltBis, bezahltBis)
			}
		})
	}

	// Zurücksetzen führt in den neutralen Zustand, nicht in "nicht bezahlt".
	if err := svc.SetBezahltBis(id, nil); err != nil {
		t.Fatalf("SetBezahltBis(nil): %v", err)
	}

	eintrag, err = svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag nach dem Zurücksetzen: %v", err)
	}
	if eintrag.BezahltBis != nil {
		t.Errorf("BezahltBis = %v, erwartet nil nach dem Zurücksetzen", eintrag.BezahltBis)
	}
	if eintrag.Zahlungsstatus != service.ZahlungsstatusNichtGesetzt {
		t.Errorf("Zahlungsstatus nach dem Zurücksetzen = %v, erwartet %v",
			eintrag.Zahlungsstatus, service.ZahlungsstatusNichtGesetzt)
	}
}

func TestSetBezahltBis_IstBeiWiederholungIdempotent(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Sina", "Petrov")
	bezahltBis := heuteVersetzt(30)

	for versuch := 1; versuch <= 3; versuch++ {
		if err := svc.SetBezahltBis(id, &bezahltBis); err != nil {
			t.Fatalf("SetBezahltBis (Versuch %d): %v", versuch, err)
		}

		eintrag, err := svc.Eintrag(id)
		if err != nil {
			t.Fatalf("Eintrag (Versuch %d): %v", versuch, err)
		}
		if eintrag.BezahltBis == nil || !eintrag.BezahltBis.Equal(bezahltBis) {
			t.Fatalf("BezahltBis nach Versuch %d = %v, erwartet %v",
				versuch, eintrag.BezahltBis, bezahltBis)
		}
		if eintrag.Zahlungsstatus != service.ZahlungsstatusBezahlt {
			t.Fatalf("Zahlungsstatus nach Versuch %d = %v, erwartet %v",
				versuch, eintrag.Zahlungsstatus, service.ZahlungsstatusBezahlt)
		}
	}

	// Auch die Liste kennt das Mitglied danach genau einmal.
	liste, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(liste) != 1 {
		t.Errorf("List = %d Einträge, erwartet 1: %+v", len(liste), liste)
	}
}

func TestSetBezahltBis_LaesstAlleAnderenFelderUnberuehrt(t *testing.T) {
	svc := neuerService(t)

	klassen := beitragsklassen(t, svc)
	geburtsdatum := datum(t, "1994-07-19")

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:          "Lina",
		Nachname:         "Fischer",
		Geburtsdatum:     &geburtsdatum,
		Adresse:          "Hauptstraße 3, 10115 Berlin",
		Email:            "lina@example.org",
		Telefon:          "030 123456",
		BeitragsklasseID: klassen[1].ID,
		Eintritt:         datum(t, "2026-02-01"),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	vorher, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get (vorher): %v", err)
	}

	bezahltBis := heuteVersetzt(14)
	if err := svc.SetBezahltBis(id, &bezahltBis); err != nil {
		t.Fatalf("SetBezahltBis: %v", err)
	}

	nachher, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get (nachher): %v", err)
	}

	if nachher.BezahltBis == nil || !nachher.BezahltBis.Equal(bezahltBis) {
		t.Errorf("BezahltBis = %v, erwartet %v", nachher.BezahltBis, bezahltBis)
	}

	// Alles außer bezahlt_bis muss identisch geblieben sein — inklusive der
	// Mitgliedschaft, die von einer Zahlungsangabe nichts wissen darf.
	erwartet := vorher
	erwartet.BezahltBis = nachher.BezahltBis
	if nachher.ID != erwartet.ID || nachher.Vorname != erwartet.Vorname ||
		nachher.Nachname != erwartet.Nachname || nachher.Adresse != erwartet.Adresse ||
		nachher.Email != erwartet.Email || nachher.Telefon != erwartet.Telefon ||
		nachher.BeitragsklasseID != erwartet.BeitragsklasseID {
		t.Errorf("Stammdaten = %+v, erwartet unverändert %+v", nachher, erwartet)
	}
	if nachher.Geburtsdatum == nil || !nachher.Geburtsdatum.Equal(geburtsdatum) {
		t.Errorf("Geburtsdatum = %v, erwartet %v", nachher.Geburtsdatum, geburtsdatum)
	}
	if len(nachher.Mitgliedschaften) != len(vorher.Mitgliedschaften) {
		t.Fatalf("Mitgliedschaften = %+v, erwartet unverändert %+v",
			nachher.Mitgliedschaften, vorher.Mitgliedschaften)
	}
	for i, ms := range nachher.Mitgliedschaften {
		if ms != vorher.Mitgliedschaften[i] {
			t.Errorf("Mitgliedschaft %d = %+v, erwartet unverändert %+v",
				i, ms, vorher.Mitgliedschaften[i])
		}
	}
}

func TestSetBezahltBis_UnbekannteIDMeldetNichtGefunden(t *testing.T) {
	svc := neuerService(t)

	bezahltBis := heuteVersetzt(7)
	if err := svc.SetBezahltBis(4711, &bezahltBis); !errors.Is(err, service.ErrNichtGefunden) {
		t.Fatalf("SetBezahltBis(4711, …) = %v, erwartet ErrNichtGefunden", err)
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
	bezahltBis := heuteVersetzt(3)
	if err := svc.SetBezahltBis(id, &bezahltBis); err != nil {
		t.Fatalf("SetBezahltBis: %v", err)
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
		eintrag.Nachname != liste[0].Nachname || eintrag.Beitragsklasse != liste[0].Beitragsklasse ||
		eintrag.Zahlungsstatus != liste[0].Zahlungsstatus ||
		!eintrag.Eintritt.Equal(liste[0].Eintritt) ||
		eintrag.BezahltBis == nil || !eintrag.BezahltBis.Equal(*liste[0].BezahltBis) {
		t.Errorf("Eintrag = %+v, erwartet dieselbe Zeile wie List: %+v", eintrag, liste[0])
	}
}

func TestEintrag_OhneLaufendeMitgliedschaftUndUnbekannteIDMeldenNichtGefunden(t *testing.T) {
	svc := neuerService(t)

	if _, err := svc.Eintrag(4711); !errors.Is(err, service.ErrNichtGefunden) {
		t.Fatalf("Eintrag(4711) = %v, erwartet ErrNichtGefunden", err)
	}

	id := mitgliedAnlegen(t, svc, "Timo", "Vogel")
	if err := svc.AustrittFuerTest(id, heuteVersetzt(-10)); err != nil {
		t.Fatalf("AustrittFuerTest: %v", err)
	}

	// Die Liste führt ausgetretene Mitglieder nicht — dann gibt es für sie auch
	// keine Zeile.
	if _, err := svc.Eintrag(id); !errors.Is(err, service.ErrNichtGefunden) {
		t.Fatalf("Eintrag nach Austritt = %v, erwartet ErrNichtGefunden", err)
	}
}
