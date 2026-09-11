package service_test

import (
	"slices"
	"testing"

	"github.com/ToniBlankenburg/boxclub/service"
)

// alleBeitragsklassen ruft ListBeitragsklassen auf und bricht bei einem Fehler
// ab. Gegenstück zum Helper beitragsklassen, der nur die aktiven liefert.
func alleBeitragsklassen(t *testing.T, svc *service.MemberService) []service.Beitragsklasse {
	t.Helper()

	klassen, err := svc.ListBeitragsklassen()
	if err != nil {
		t.Fatalf("ListBeitragsklassen: %v", err)
	}

	return klassen
}

// Die Beitragsklassen-Ansicht ist die Referenz des Admins: nach dem Seed muss
// sie genau die beiden definierten Klassen mit ihren Preisen zeigen,
// aufsteigend nach Preis.
func TestListBeitragsklassen_ZeigtDieGeseedetenKlassenMitPreis(t *testing.T) {
	svc := neuerService(t)

	klassen := alleBeitragsklassen(t, svc)

	erwartet := []struct {
		name  string
		cents int64
	}{
		{"Erwachsen 1×/Woche", 6000},
		{"Erwachsen 2×/Woche", 8000},
	}

	if len(klassen) != len(erwartet) {
		t.Fatalf("erwarte %d Beitragsklassen, bekam %d: %+v", len(erwartet), len(klassen), klassen)
	}

	for i, e := range erwartet {
		if klassen[i].Name != e.name {
			t.Errorf("Klasse %d: Name = %q, erwartet %q", i, klassen[i].Name, e.name)
		}
		if klassen[i].PreisMonatlichCents != e.cents {
			t.Errorf("Klasse %d (%s): PreisMonatlichCents = %d, erwartet %d",
				i, e.name, klassen[i].PreisMonatlichCents, e.cents)
		}
		if klassen[i].ID == 0 {
			t.Errorf("Klasse %d (%s): ID = 0, erwartet eine vergebene ID", i, e.name)
		}
	}
}

// Die Liste beschreibt das Angebot des Vereins, nicht die Belegung: eine Klasse
// ohne ein einziges zugeordnetes Mitglied muss genauso erscheinen wie eine
// belegte. Geprüft wird deshalb die unbelegte Klasse selbst und nicht bloß die
// Anzahl — sonst hinge der Test an derselben Quelle, die er absichern soll.
func TestListBeitragsklassen_ZeigtAuchKlassenOhneMitglieder(t *testing.T) {
	svc := neuerService(t)

	klassen := beitragsklassen(t, svc)
	belegt, unbelegt := klassen[0], klassen[1]

	if _, err := svc.Create(service.NeuesMitglied{
		Vorname:          "Nadia",
		Nachname:         "Öztürk",
		BeitragsklasseID: belegt.ID,
		Eintritt:         datum(t, "2024-03-01"),
	}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Nur belegt hat jetzt ein Mitglied — unbelegt muss trotzdem dabei sein,
	// unverändert in Name und Preis.
	alle := alleBeitragsklassen(t, svc)
	if !slices.Contains(alle, unbelegt) {
		t.Errorf("Beitragsklasse %+v fehlt, weil ihr kein Mitglied zugeordnet ist; geliefert wurde:\n%+v",
			unbelegt, alle)
	}
}
