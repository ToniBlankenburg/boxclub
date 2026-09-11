package service_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/ToniBlankenburg/boxclub/service"
)

// slotsImTest sind drei Trainingstermine in der Schreibweise, in der der Verein
// sie tippt — Wochentag und Uhrzeit als freier Text (CONTEXT.md → Trainingsslot).
var slotsImTest = []string{"Montag 18:00 Uhr", "Mittwoch 19:30 Uhr", "Samstag 10:30 Uhr"}

// mitSlotsAnlegen legt ein Mitglied mit den angegebenen Trainingsslots an.
func mitSlotsAnlegen(t *testing.T, svc *service.MemberService, nachname string, slots ...string) int64 {
	t.Helper()

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:        "Test",
		Nachname:       nachname,
		BeitragCents:   beitragImTest,
		Eintritt:       datum(t, "2026-01-05"),
		Trainingsslots: slots,
	})
	if err != nil {
		t.Fatalf("Create(%s, %v): %v", nachname, slots, err)
	}

	return id
}

// laufendeSlots liest die Slots der laufenden Mitgliedschaft über die
// Service-API — direkt in die Tabelle sieht kein Test.
func laufendeSlots(t *testing.T, svc *service.MemberService, id int64) []string {
	t.Helper()

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get(%d): %v", id, err)
	}

	laufend := m.LaufendeMitgliedschaft()
	if laufend == nil {
		t.Fatalf("Mitglied %d hat keine laufende Mitgliedschaft", id)
	}

	return laufend.Trainingsslots
}

// Die Trainingsfrequenz wird nirgends gespeichert: sie ist die Anzahl der
// Slots und kann deshalb gar nicht von ihr abweichen (CONTEXT.md →
// Trainingsfrequenz). Null Slots heißen "keine Frequenz" und nicht "1×".
func TestTrainingsfrequenz_ErgibtSichAusDerAnzahlDerSlots(t *testing.T) {
	svc := neuerService(t)

	for anzahl := 0; anzahl <= service.MaxTrainingsslots; anzahl++ {
		slots := slotsImTest[:anzahl]
		id := mitSlotsAnlegen(t, svc, "Frequenz"+strings.Repeat("x", anzahl+1), slots...)

		m, err := svc.Get(id)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}

		laufend := m.LaufendeMitgliedschaft()
		if laufend == nil {
			t.Fatalf("keine laufende Mitgliedschaft bei %d Slots", anzahl)
		}
		if !slices.Equal(laufend.Trainingsslots, slots) {
			t.Errorf("Trainingsslots = %v, erwartet %v", laufend.Trainingsslots, slots)
		}
		if got := laufend.Trainingsfrequenz(); int(got) != anzahl {
			t.Errorf("Trainingsfrequenz bei %d Slots = %d, erwartet %d", anzahl, int(got), anzahl)
		}

		// Dieselbe Ableitung muss die Listenzeile zeigen — sonst zeigte die
		// Liste etwas anderes als das Formular.
		eintrag, err := svc.Eintrag(id)
		if err != nil {
			t.Fatalf("Eintrag: %v", err)
		}
		if !slices.Equal(eintrag.Trainingsslots, slots) {
			t.Errorf("Listeneintrag.Trainingsslots = %v, erwartet %v", eintrag.Trainingsslots, slots)
		}
		if int(eintrag.Trainingsfrequenz()) != anzahl {
			t.Errorf("Listeneintrag.Trainingsfrequenz = %d, erwartet %d", int(eintrag.Trainingsfrequenz()), anzahl)
		}
	}
}

// Drei Slots sind die Obergrenze: mehr Trainingstermine gibt es im Verein nicht,
// und ein vierter wäre eine Frequenz, die es nicht gibt.
func TestTrainingsslots_WeisenDenViertenSlotAb(t *testing.T) {
	svc := neuerService(t)

	vier := []string{"Montag 18:00 Uhr", "Dienstag 19:30 Uhr", "Mittwoch 19:30 Uhr", "Samstag 10:30 Uhr"}

	_, err := svc.Create(service.NeuesMitglied{
		Vorname:        "Test",
		Nachname:       "Zuviel",
		BeitragCents:   beitragImTest,
		Eintritt:       datum(t, "2026-01-05"),
		Trainingsslots: vier,
	})
	pruefeValidierungsfehler(t, err, "Create mit vier Slots")

	// Und ebenso wenig lässt sich ein vierter nachträglich anhängen.
	id := mitSlotsAnlegen(t, svc, "Wagner", slotsImTest...)

	err = svc.Update(id, service.MitgliedPatch{Trainingsslots: &vier})
	pruefeValidierungsfehler(t, err, "Update mit vier Slots")

	// Der abgewiesene Versuch darf den Bestand nicht angerührt haben.
	if gefunden := laufendeSlots(t, svc, id); !slices.Equal(gefunden, slotsImTest) {
		t.Errorf("Slots nach abgewiesenem Update = %v, erwartet unverändert %v", gefunden, slotsImTest)
	}
}

// Ein leer gelassenes Feld im Formular ist kein Slot: es zählt weder gegen die
// Obergrenze noch für die Frequenz.
func TestTrainingsslots_LeereAngabenSindKeineSlots(t *testing.T) {
	svc := neuerService(t)

	id := mitSlotsAnlegen(t, svc, "Leerfeld", " Montag 18:00 Uhr ", "", "   ")

	gefunden := laufendeSlots(t, svc, id)
	if !slices.Equal(gefunden, []string{"Montag 18:00 Uhr"}) {
		t.Errorf("Slots = %v, erwartet [Montag 18:00 Uhr] — Leerraum abgeschnitten, leere Felder weg", gefunden)
	}

	// Vier Felder, von denen eines leer bleibt, sind drei Slots und damit erlaubt.
	vierFelder := []string{"Montag 18:00 Uhr", "", "Mittwoch 19:30 Uhr", "Samstag 10:30 Uhr"}
	if err := svc.Update(id, service.MitgliedPatch{Trainingsslots: &vierFelder}); err != nil {
		t.Fatalf("Update mit drei gefüllten von vier Feldern: %v", err)
	}
	if gefunden := laufendeSlots(t, svc, id); !slices.Equal(gefunden, slotsImTest) {
		t.Errorf("Slots = %v, erwartet %v", gefunden, slotsImTest)
	}
}

// Hinzufügen und Entfernen sind derselbe Vorgang: das Formular schickt die
// Slots, die danach gelten sollen.
func TestTrainingsslots_AenderungErsetztDenBestand(t *testing.T) {
	svc := neuerService(t)

	id := mitSlotsAnlegen(t, svc, "Berger", "Montag 18:00 Uhr")

	dazu := []string{"Montag 18:00 Uhr", "Samstag 10:30 Uhr"}
	if err := svc.Update(id, service.MitgliedPatch{Trainingsslots: &dazu}); err != nil {
		t.Fatalf("Update (hinzufügen): %v", err)
	}
	if gefunden := laufendeSlots(t, svc, id); !slices.Equal(gefunden, dazu) {
		t.Errorf("Slots nach dem Hinzufügen = %v, erwartet %v", gefunden, dazu)
	}

	weniger := []string{"Samstag 10:30 Uhr"}
	if err := svc.Update(id, service.MitgliedPatch{Trainingsslots: &weniger}); err != nil {
		t.Fatalf("Update (entfernen): %v", err)
	}
	if gefunden := laufendeSlots(t, svc, id); !slices.Equal(gefunden, weniger) {
		t.Errorf("Slots nach dem Entfernen = %v, erwartet %v", gefunden, weniger)
	}

	// Alle entfernen ist erlaubt und heißt: keine Frequenz.
	keine := []string{}
	if err := svc.Update(id, service.MitgliedPatch{Trainingsslots: &keine}); err != nil {
		t.Fatalf("Update (alle entfernen): %v", err)
	}
	if gefunden := laufendeSlots(t, svc, id); len(gefunden) != 0 {
		t.Errorf("Slots nach dem Leeren = %v, erwartet keine", gefunden)
	}
}

// Ein Patch ohne Slots ist keine Aussage über sie: wer nur die Adresse
// korrigiert, darf damit nicht das Training löschen.
func TestTrainingsslots_BleibenOhneAngabeImPatchUnberuehrt(t *testing.T) {
	svc := neuerService(t)

	id := mitSlotsAnlegen(t, svc, "Öztürk", slotsImTest...)

	ort := "Berlin"
	if err := svc.Update(id, service.MitgliedPatch{Anschrift: &service.Anschrift{Ort: ort}}); err != nil {
		t.Fatalf("Update ohne Slots: %v", err)
	}

	if gefunden := laufendeSlots(t, svc, id); !slices.Equal(gefunden, slotsImTest) {
		t.Errorf("Slots = %v, erwartet unverändert %v", gefunden, slotsImTest)
	}
}

// Die Slots hängen an der Mitgliedschaft, nicht am Mitglied: ein Wiedereintritt
// ist eine neue Vereinbarung und fängt ohne Trainingszeiten an, während die
// alte Mitgliedschaft ihre behält.
func TestTrainingsslots_WiedereintrittLaesstDieAltenSlotsStehen(t *testing.T) {
	svc := neuerService(t)

	id := mitSlotsAnlegen(t, svc, "Klein", "Montag 18:00 Uhr", "Samstag 10:30 Uhr")

	if err := svc.MarkExit(id, datum(t, "2026-06-30")); err != nil {
		t.Fatalf("MarkExit: %v", err)
	}
	if err := svc.Rejoin(id, datum(t, "2027-01-01")); err != nil {
		t.Fatalf("Rejoin: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(m.Mitgliedschaften) != 2 {
		t.Fatalf("erwarte 2 Mitgliedschaften, bekam %d", len(m.Mitgliedschaften))
	}

	alt, neu := m.Mitgliedschaften[0], m.Mitgliedschaften[1]
	if erwartet := []string{"Montag 18:00 Uhr", "Samstag 10:30 Uhr"}; !slices.Equal(alt.Trainingsslots, erwartet) {
		t.Errorf("Slots der alten Mitgliedschaft = %v, erwartet unverändert %v", alt.Trainingsslots, erwartet)
	}
	if len(neu.Trainingsslots) != 0 {
		t.Errorf("Slots der neuen Mitgliedschaft = %v, erwartet keine", neu.Trainingsslots)
	}
	if neu.Trainingsfrequenz() != 0 {
		t.Errorf("Frequenz der neuen Mitgliedschaft = %d, erwartet 0", int(neu.Trainingsfrequenz()))
	}

	// Und die neuen Slots landen an der neuen Mitgliedschaft, nicht an der alten.
	nurSamstag := []string{"Samstag 10:30 Uhr"}
	if err := svc.Update(id, service.MitgliedPatch{Trainingsslots: &nurSamstag}); err != nil {
		t.Fatalf("Update nach Wiedereintritt: %v", err)
	}

	m, err = svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if erwartet := []string{"Montag 18:00 Uhr", "Samstag 10:30 Uhr"}; !slices.Equal(m.Mitgliedschaften[0].Trainingsslots, erwartet) {
		t.Errorf("Slots der alten Mitgliedschaft = %v, erwartet unverändert %v", m.Mitgliedschaften[0].Trainingsslots, erwartet)
	}
	if !slices.Equal(m.Mitgliedschaften[1].Trainingsslots, nurSamstag) {
		t.Errorf("Slots der neuen Mitgliedschaft = %v, erwartet %v", m.Mitgliedschaften[1].Trainingsslots, nurSamstag)
	}
}

// Ein ausgetretenes Mitglied zeigt in der Liste die Slots seines letzten
// Zeitraums — dieselbe Mitgliedschaft, aus der auch Beitrag und Eintritt kommen.
func TestTrainingsslots_ListeZeigtDieSlotsDerMassgeblichenMitgliedschaft(t *testing.T) {
	svc := neuerService(t)

	id := mitSlotsAnlegen(t, svc, "Klein", "Montag 18:00 Uhr", "Samstag 10:30 Uhr")
	if err := svc.MarkExit(id, datum(t, "2026-06-30")); err != nil {
		t.Fatalf("MarkExit: %v", err)
	}

	eintrag, err := svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag: %v", err)
	}
	if erwartet := []string{"Montag 18:00 Uhr", "Samstag 10:30 Uhr"}; !slices.Equal(eintrag.Trainingsslots, erwartet) {
		t.Errorf("Slots des ausgetretenen Mitglieds = %v, erwartet %v", eintrag.Trainingsslots, erwartet)
	}
}

// pruefeValidierungsfehler verlangt einen ValidierungsFehler mit mindestens
// einer Meldung — die Meldungen sind für den Nutzer bestimmt und dürfen nicht
// leer sein.
func pruefeValidierungsfehler(t *testing.T, err error, was string) {
	t.Helper()

	var validierung *service.ValidierungsFehler
	if !errors.As(err, &validierung) {
		t.Fatalf("%s = %v, erwartet ValidierungsFehler", was, err)
	}
	if len(validierung.Meldungen) == 0 {
		t.Errorf("%s: ValidierungsFehler ohne Meldung", was)
	}
}
