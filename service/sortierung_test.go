package service_test

import (
	"testing"

	"github.com/ToniBlankenburg/boxclub/service"
)

// eintraegeSortieren am MemberService-Seam (Ticket 27). Jede Sortierordnung
// bekommt ihren eigenen Test, weil jede ihre eigene Regel hat: alphabetisch,
// numerisch, chronologisch oder nach dem Lebenszyklus — und weil zwei von
// ihnen (Anschrift, Training) leere Werte kennen, die immer ans Ende gehören,
// gleich in welcher Richtung.
//
// Beobachtet wird ausschließlich über Search: Sortierung ist ein Feld von
// Suchfilter und reist deshalb im selben Aufruf wie Suchbegriff und die
// übrigen Filter.

func sortiert(t *testing.T, svc *service.MemberService, sortierung service.Sortierung) []string {
	t.Helper()

	liste, err := svc.Search("", service.Suchfilter{AuchEhemalige: true, Sortierung: sortierung})
	if err != nil {
		t.Fatalf("Search mit Sortierung %+v: %v", sortierung, err)
	}

	return namen(liste)
}

// TestSearch_SortiertNachNameAbsteigend ergänzt die aufsteigende Sortierung,
// die List schon als Standard testet (TestList_LiefertAktiveMitgliederNachNamenSortiert):
// ein erneuter Klick auf die aktive Spalte dreht die Richtung um.
func TestSearch_SortiertNachNameAbsteigend(t *testing.T) {
	svc := neuerService(t)

	mitgliedAnlegen(t, svc, "Anna", "Adler")
	mitgliedAnlegen(t, svc, "Bea", "Berger")
	mitgliedAnlegen(t, svc, "Cem", "Celik")

	got := sortiert(t, svc, service.Sortierung{
		Spalte: service.SortierspalteName, Richtung: service.SortierrichtungAbsteigend,
	})
	erwartet := []string{"Celik, Cem", "Berger, Bea", "Adler, Anna"}
	pruefeReihenfolge(t, got, erwartet)
}

// TestSearch_SortiertNachMitgliedIDNumerisch: die Nummer ist die
// Vereinsnummer aus der Excel (CONTEXT.md → Mitglieds-ID) und wird numerisch
// sortiert, nicht als Text — sonst stünde "10" vor "2".
func TestSearch_SortiertNachMitgliedIDNumerisch(t *testing.T) {
	svc := neuerService(t)

	// Angelegt in umgekehrter Reihenfolge zur Namensfolge — steigende IDs
	// vergibt Create ohnehin in Anlage-Reihenfolge (siehe schema): käme die
	// Namenssortierung versehentlich noch zum Zug, fiele das hier auf.
	mitgliedAnlegen(t, svc, "Zoe", "Adler")
	mitgliedAnlegen(t, svc, "Anna", "Berger")
	mitgliedAnlegen(t, svc, "Mona", "Celik")

	auf := sortiert(t, svc, service.Sortierung{Spalte: service.SortierspalteMitgliedID})
	pruefeReihenfolge(t, auf, []string{"Adler, Zoe", "Berger, Anna", "Celik, Mona"})

	ab := sortiert(t, svc, service.Sortierung{
		Spalte: service.SortierspalteMitgliedID, Richtung: service.SortierrichtungAbsteigend,
	})
	pruefeReihenfolge(t, ab, []string{"Celik, Mona", "Berger, Anna", "Adler, Zoe"})
}

// TestSearch_SortiertNachStatusLebenszyklusNichtAlphabetisch: Neu, Aktiv, In
// Kündigungsfrist, Ausgetreten — in dieser Reihenfolge, weil es die des
// Lebenszyklus ist (status.go). Die Namen sind bewusst so gewählt, dass die
// alphabetische Reihenfolge eine andere wäre — käme sie stattdessen zum
// Zug, fiele das hier auf.
func TestSearch_SortiertNachStatusLebenszyklusNichtAlphabetisch(t *testing.T) {
	svc := neuerService(t)

	// "Neu" hieße alphabetisch vor "Zumbach" und "Anders" — die erwartete
	// Statusreihenfolge sagt aber, dass "Zumbach" (Neu) zuerst kommt.
	neuID := mitgliedAnlegenZum(t, svc, "Zumbach", heuteVersetzt(10))
	aktivID := mitgliedAnlegenZum(t, svc, "Meier", heuteVersetzt(-10))

	kuendigungID := mitgliedAnlegenZum(t, svc, "Doerr", heuteVersetzt(-30))
	if err := svc.SetKuendigung(kuendigungID, austrittZum(heuteVersetzt(30))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	ausgetretenID := mitgliedAnlegenZum(t, svc, "Anders", heuteVersetzt(-100))
	if err := svc.SetKuendigung(ausgetretenID, austrittZum(heuteVersetzt(-1))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	namenVon := func(id int64) string {
		e, err := svc.Eintrag(id)
		if err != nil {
			t.Fatalf("Eintrag(%d): %v", id, err)
		}
		return e.Nachname + ", " + e.Vorname
	}

	auf := sortiert(t, svc, service.Sortierung{Spalte: service.SortierspalteStatus})
	pruefeReihenfolge(t, auf, []string{
		namenVon(neuID), namenVon(aktivID), namenVon(kuendigungID), namenVon(ausgetretenID),
	})

	ab := sortiert(t, svc, service.Sortierung{
		Spalte: service.SortierspalteStatus, Richtung: service.SortierrichtungAbsteigend,
	})
	pruefeReihenfolge(t, ab, []string{
		namenVon(ausgetretenID), namenVon(kuendigungID), namenVon(aktivID), namenVon(neuID),
	})
}

// TestSearch_SortiertNachAnschriftLeereAmEnde: alphabetisch nach Ort, und wer
// keine Anschrift hinterlegt hat, steht am Ende — in beiden Richtungen. Ein
// Mitglied ohne Anschrift darf nicht die halbe erste Seite füllen.
func TestSearch_SortiertNachAnschriftLeereAmEnde(t *testing.T) {
	svc := neuerService(t)

	mitAnschrift := func(vorname, nachname, ort string) int64 {
		t.Helper()

		id, err := svc.Create(service.NeuesMitglied{
			Vorname: vorname, Nachname: nachname, BeitragCents: beitragImTest,
			Eintritt:  datum(t, "2026-01-05"),
			Anschrift: service.Anschrift{Adresse: "Hauptstraße 1", Postleitzahl: "00000", Ort: ort},
		})
		if err != nil {
			t.Fatalf("Create(%s %s): %v", vorname, nachname, err)
		}
		return id
	}

	mitAnschrift("Tim", "Zeller", "Hamburg")
	mitgliedAnlegen(t, svc, "Ohne", "Anschrift") // Nullwert: keine Anschrift erfasst
	mitAnschrift("Ida", "Adler", "Berlin")

	auf := sortiert(t, svc, service.Sortierung{Spalte: service.SortierspalteAnschrift})
	pruefeReihenfolge(t, auf, []string{"Adler, Ida", "Zeller, Tim", "Anschrift, Ohne"})

	ab := sortiert(t, svc, service.Sortierung{
		Spalte: service.SortierspalteAnschrift, Richtung: service.SortierrichtungAbsteigend,
	})
	// Absteigend dreht Hamburg und Berlin um — "Anschrift, Ohne" bleibt aber
	// am Ende stehen und rutscht nicht etwa nach vorn.
	pruefeReihenfolge(t, ab, []string{"Zeller, Tim", "Adler, Ida", "Anschrift, Ohne"})
}

// TestSearch_SortiertNachTrainingLeereAmEnde: numerisch nach der abgeleiteten
// Trainingsfrequenz (CONTEXT.md → Trainingsfrequenz), und wer für keinen
// Termin angemeldet ist, steht am Ende — auch hier in beiden Richtungen.
func TestSearch_SortiertNachTrainingLeereAmEnde(t *testing.T) {
	svc := neuerService(t)

	montag, mittwoch, samstag := stundenplan(t, svc)

	if _, err := svc.Create(service.NeuesMitglied{
		Vorname: "Drei", Nachname: "Mal", BeitragCents: beitragImTest,
		Eintritt: datum(t, "2026-01-05"), TrainingsterminIDs: ids(montag, mittwoch, samstag),
	}); err != nil {
		t.Fatalf("Create (dreimal): %v", err)
	}

	if _, err := svc.Create(service.NeuesMitglied{
		Vorname: "Ein", Nachname: "Mal", BeitragCents: beitragImTest,
		Eintritt: datum(t, "2026-01-05"), TrainingsterminIDs: ids(montag),
	}); err != nil {
		t.Fatalf("Create (einmal): %v", err)
	}

	mitgliedAnlegen(t, svc, "Kein", "Mal")

	auf := sortiert(t, svc, service.Sortierung{Spalte: service.SortierspalteTraining})
	pruefeReihenfolge(t, auf, []string{"Mal, Ein", "Mal, Drei", "Mal, Kein"})

	ab := sortiert(t, svc, service.Sortierung{
		Spalte: service.SortierspalteTraining, Richtung: service.SortierrichtungAbsteigend,
	})
	pruefeReihenfolge(t, ab, []string{"Mal, Drei", "Mal, Ein", "Mal, Kein"})
}

// TestSearch_SortiertNachBeitragNumerischOhneNullAmEnde: numerisch nach dem
// Beitrag — und 0 € ist ein gültiger Beitrag (ADR-0005), keine fehlende
// Angabe. Anders als Anschrift und Training landet er deshalb an seiner
// numerischen Stelle und nicht am Ende.
func TestSearch_SortiertNachBeitragNumerischOhneNullAmEnde(t *testing.T) {
	svc := neuerService(t)

	beitrag := func(vorname string, cents int64) {
		t.Helper()
		if _, err := svc.Create(service.NeuesMitglied{
			Vorname: vorname, Nachname: "Zahler", BeitragCents: cents, Eintritt: datum(t, "2026-01-05"),
		}); err != nil {
			t.Fatalf("Create(%s): %v", vorname, err)
		}
	}

	beitrag("Fuenfzig", 5000)
	beitrag("Null", 0)
	beitrag("Dreissig", 3000)

	auf := sortiert(t, svc, service.Sortierung{Spalte: service.SortierspalteBeitrag})
	pruefeReihenfolge(t, auf, []string{"Zahler, Null", "Zahler, Dreissig", "Zahler, Fuenfzig"})

	ab := sortiert(t, svc, service.Sortierung{
		Spalte: service.SortierspalteBeitrag, Richtung: service.SortierrichtungAbsteigend,
	})
	pruefeReihenfolge(t, ab, []string{"Zahler, Fuenfzig", "Zahler, Dreissig", "Zahler, Null"})
}

// TestSearch_SortiertNachRueckstand: "in Ordnung" vor "im Rückstand"
// aufsteigend — der Normalfall zuerst, die Ausnahmeliste danach (ADR-0006).
func TestSearch_SortiertNachRueckstand(t *testing.T) {
	svc := neuerService(t)

	mitgliedAnlegen(t, svc, "Ida", "Ordentlich")

	imRueckstandID := mitgliedAnlegen(t, svc, "Rico", "Rueckstand")
	if err := svc.SetRueckstand(imRueckstandID, service.Rueckstand{Offen: true, Notiz: "Testfall"}); err != nil {
		t.Fatalf("SetRueckstand: %v", err)
	}

	auf := sortiert(t, svc, service.Sortierung{Spalte: service.SortierspalteRueckstand})
	pruefeReihenfolge(t, auf, []string{"Ordentlich, Ida", "Rueckstand, Rico"})

	ab := sortiert(t, svc, service.Sortierung{
		Spalte: service.SortierspalteRueckstand, Richtung: service.SortierrichtungAbsteigend,
	})
	pruefeReihenfolge(t, ab, []string{"Rueckstand, Rico", "Ordentlich, Ida"})
}

// TestSearch_SortiertNachEintrittChronologisch: nach dem Kalendertag, nicht
// nach dem Text des Datums — sonst stünde "2026-02-01" vor "2026-01-15", weil
// hier zufällig auch die Textform übereinstimmt; entscheidend ist trotzdem
// die chronologische Regel und nicht die alphabetische.
func TestSearch_SortiertNachEintrittChronologisch(t *testing.T) {
	svc := neuerService(t)

	eintritt := func(vorname, iso string) {
		t.Helper()
		if _, err := svc.Create(service.NeuesMitglied{
			Vorname: vorname, Nachname: "Beitritt", BeitragCents: beitragImTest, Eintritt: datum(t, iso),
		}); err != nil {
			t.Fatalf("Create(%s): %v", vorname, err)
		}
	}

	eintritt("Maerz", "2026-03-01")
	eintritt("Januar", "2026-01-15")
	eintritt("Februar", "2026-02-01")

	auf := sortiert(t, svc, service.Sortierung{Spalte: service.SortierspalteEintritt})
	pruefeReihenfolge(t, auf, []string{"Beitritt, Januar", "Beitritt, Februar", "Beitritt, Maerz"})

	ab := sortiert(t, svc, service.Sortierung{
		Spalte: service.SortierspalteEintritt, Richtung: service.SortierrichtungAbsteigend,
	})
	pruefeReihenfolge(t, ab, []string{"Beitritt, Maerz", "Beitritt, Februar", "Beitritt, Januar"})
}

// TestSearch_SortierungBleibtBeiFilterErhalten: Sortierung ist ein Feld von
// Suchfilter und wirkt deshalb im selben Aufruf wie Suchbegriff und die
// übrigen Filter — genau das verlangt das Ticket ("Sortierung bleibt bei
// Suche und Filter erhalten und umgekehrt").
func TestSearch_SortierungBleibtBeiFilterErhalten(t *testing.T) {
	svc := neuerService(t)

	teuerID, err := svc.Create(service.NeuesMitglied{
		Vorname: "Teuer", Nachname: "Rueckstand", BeitragCents: 9000, Eintritt: datum(t, "2026-01-05"),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := svc.SetRueckstand(teuerID, service.Rueckstand{Offen: true, Notiz: "Testfall"}); err != nil {
		t.Fatalf("SetRueckstand: %v", err)
	}

	guenstigID, err := svc.Create(service.NeuesMitglied{
		Vorname: "Guenstig", Nachname: "Rueckstand", BeitragCents: 3000, Eintritt: datum(t, "2026-01-05"),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := svc.SetRueckstand(guenstigID, service.Rueckstand{Offen: true, Notiz: "Testfall"}); err != nil {
		t.Fatalf("SetRueckstand: %v", err)
	}

	// Ohne Rückstand, gehört wegen des Filters unten nicht ins Ergebnis —
	// stünde er trotzdem drin, hätte die Sortierung den Filter verdrängt statt
	// mit ihm zusammenzuwirken.
	mitgliedAnlegen(t, svc, "Ausserhalb", "Filter")

	liste, err := svc.Search("", service.Suchfilter{
		Rueckstand: service.RueckstandsfilterImRueckstand,
		Sortierung: service.Sortierung{Spalte: service.SortierspalteBeitrag, Richtung: service.SortierrichtungAbsteigend},
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	pruefeReihenfolge(t, namen(liste), []string{"Rueckstand, Teuer", "Rueckstand, Guenstig"})
}

// pruefeReihenfolge vergleicht ein Suchergebnis auf Namen und Reihenfolge —
// beides gehört zusammen, denn genau die Reihenfolge ist es, was diese Tests
// beobachten.
func pruefeReihenfolge(t *testing.T, got, erwartet []string) {
	t.Helper()

	if len(got) != len(erwartet) {
		t.Fatalf("Ergebnis = %v, erwartet %v", got, erwartet)
	}
	for i := range erwartet {
		if got[i] != erwartet[i] {
			t.Errorf("Position %d = %q, erwartet %q — ganzes Ergebnis: %v", i, got[i], erwartet[i], got)
		}
	}
}
