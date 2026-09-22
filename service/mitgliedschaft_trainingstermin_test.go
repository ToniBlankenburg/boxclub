package service_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/ToniBlankenburg/boxclub/service"
)

// stundenplan sind die drei Termine, an denen die Zuordnungstests arbeiten —
// über die Woche verteilt, damit die Wochenreihenfolge einer Liste beobachtbar
// ist. Zurück kommen sie in eben dieser Reihenfolge.
func stundenplan(t *testing.T, svc *service.MemberService) (montag, mittwoch, samstag service.Trainingstermin) {
	t.Helper()

	// Angelegt in verdrehter Reihenfolge: käme die Auswahl in der Reihenfolge
	// der Anlage statt in der der Woche, fiele es hier auf.
	samstag = angelegterTermin(t, svc, terminangabe(service.Samstag, "10:30", "12:00", ""))
	montag = angelegterTermin(t, svc, terminangabe(service.Montag, "18:00", "", ""))
	mittwoch = angelegterTermin(t, svc, terminangabe(service.Mittwoch, "19:30", "", ""))

	return montag, mittwoch, samstag
}

// ids verdichtet Termine auf ihre IDs — die Währung, in der Formular und
// Service über die Zuordnung sprechen.
func ids(termine ...service.Trainingstermin) []int64 {
	gesammelt := make([]int64, 0, len(termine))
	for _, termin := range termine {
		gesammelt = append(gesammelt, termin.ID)
	}

	return gesammelt
}

// mitTerminenAnlegen legt ein Mitglied an, das für die angegebenen Termine
// angemeldet ist.
func mitTerminenAnlegen(t *testing.T, svc *service.MemberService, nachname string, terminIDs ...int64) int64 {
	t.Helper()

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:            "Test",
		Nachname:           nachname,
		BeitragCents:       beitragImTest,
		Eintritt:           datum(t, "2026-01-05"),
		TrainingsterminIDs: terminIDs,
	})
	if err != nil {
		t.Fatalf("Create(%s, %v): %v", nachname, terminIDs, err)
	}

	return id
}

// laufendeTermine liest die Termine der laufenden Mitgliedschaft über die
// Service-API — direkt in die Tabelle sieht kein Test.
func laufendeTermine(t *testing.T, svc *service.MemberService, id int64) []service.Trainingstermin {
	t.Helper()

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get(%d): %v", id, err)
	}

	laufend := m.LaufendeMitgliedschaft()
	if laufend == nil {
		t.Fatalf("Mitglied %d hat keine laufende Mitgliedschaft", id)
	}

	return laufend.Trainingstermine
}

// Die Trainingsfrequenz wird nirgends gespeichert: sie ist die Anzahl der
// Termine und kann deshalb gar nicht von ihr abweichen (CONTEXT.md →
// Trainingsfrequenz). Null Termine heißen "keine Frequenz" und nicht "1×".
func TestTrainingsfrequenz_ErgibtSichAusDerAnzahlDerTermine(t *testing.T) {
	svc := neuerService(t)
	montag, mittwoch, samstag := stundenplan(t, svc)

	alle := ids(montag, mittwoch, samstag)

	for anzahl := 0; anzahl <= service.MaxTrainingstermine; anzahl++ {
		gewaehlt := alle[:anzahl]
		id := mitTerminenAnlegen(t, svc, "Frequenz"+strings.Repeat("x", anzahl+1), gewaehlt...)

		m, err := svc.Get(id)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}

		laufend := m.LaufendeMitgliedschaft()
		if laufend == nil {
			t.Fatalf("keine laufende Mitgliedschaft bei %d Terminen", anzahl)
		}
		if gefunden := ids(laufend.Trainingstermine...); !slices.Equal(gefunden, gewaehlt) {
			t.Errorf("Trainingstermine = %v, erwartet %v", gefunden, gewaehlt)
		}
		if got := laufend.Trainingsfrequenz(); int(got) != anzahl {
			t.Errorf("Trainingsfrequenz bei %d Terminen = %d, erwartet %d", anzahl, int(got), anzahl)
		}

		// Dieselbe Ableitung muss die Listenzeile zeigen — sonst zeigte die
		// Liste etwas anderes als das Formular.
		eintrag, err := svc.Eintrag(id)
		if err != nil {
			t.Fatalf("Eintrag: %v", err)
		}
		if gefunden := ids(eintrag.Trainingstermine...); !slices.Equal(gefunden, gewaehlt) {
			t.Errorf("Listeneintrag.Trainingstermine = %v, erwartet %v", gefunden, gewaehlt)
		}
		if int(eintrag.Trainingsfrequenz()) != anzahl {
			t.Errorf("Listeneintrag.Trainingsfrequenz = %d, erwartet %d", int(eintrag.Trainingsfrequenz()), anzahl)
		}
	}
}

// Die Termine kommen in Wochenreihenfolge zurück, ganz gleich, in welcher
// Reihenfolge sie angekreuzt wurden: gezeigt wird ein Stundenplanauszug und
// keine Anklick-Historie.
func TestTrainingstermine_StehenInWochenreihenfolge(t *testing.T) {
	svc := neuerService(t)
	montag, mittwoch, samstag := stundenplan(t, svc)

	id := mitTerminenAnlegen(t, svc, "Wochenfolge", samstag.ID, montag.ID, mittwoch.ID)

	erwartet := []string{"Montag 18:00", "Mittwoch 19:30", "Samstag 10:30 – 12:00"}
	if gefunden := anzeigen(laufendeTermine(t, svc, id)); !slices.Equal(gefunden, erwartet) {
		t.Errorf("Trainingstermine = %v, erwartet %v", gefunden, erwartet)
	}
}

// Drei Termine sind die Obergrenze: mehr trainiert im Verein niemand, und ein
// vierter wäre eine Frequenz, die es nicht gibt.
func TestTrainingstermine_WeisenDenViertenAb(t *testing.T) {
	svc := neuerService(t)
	montag, mittwoch, samstag := stundenplan(t, svc)
	dienstag := angelegterTermin(t, svc, terminangabe(service.Dienstag, "19:30", "", ""))

	drei := ids(montag, mittwoch, samstag)
	vier := ids(montag, dienstag, mittwoch, samstag)

	_, err := svc.Create(service.NeuesMitglied{
		Vorname:            "Test",
		Nachname:           "Zuviel",
		BeitragCents:       beitragImTest,
		Eintritt:           datum(t, "2026-01-05"),
		TrainingsterminIDs: vier,
	})
	pruefeValidierungsfehler(t, err, "Create mit vier Terminen")

	// Und ebenso wenig lässt sich ein vierter nachträglich anhängen.
	id := mitTerminenAnlegen(t, svc, "Wagner", drei...)

	err = svc.Update(id, service.MitgliedPatch{TrainingsterminIDs: &vier})
	pruefeValidierungsfehler(t, err, "Update mit vier Terminen")

	// Die Meldung nennt die Obergrenze: „vier geht nicht" allein sagte dem
	// Nutzer nicht, wie viele denn gehen.
	meldung := einzigeMeldung(t, err)
	if meldung.Schluessel != "validierung.termin.zu_viele" || len(meldung.Args) == 0 ||
		meldung.Args[0] != service.MaxTrainingstermine {
		t.Errorf("Meldung = %+v, erwartet einen Hinweis auf die Obergrenze %d", meldung, service.MaxTrainingstermine)
	}

	// Der abgewiesene Versuch darf den Bestand nicht angerührt haben.
	if gefunden := ids(laufendeTermine(t, svc, id)...); !slices.Equal(gefunden, drei) {
		t.Errorf("Termine nach abgewiesenem Update = %v, erwartet unverändert %v", gefunden, drei)
	}
}

// Derselbe Termin doppelt angekreuzt ist derselbe Termin und nicht zwei: er
// zählt einmal für die Frequenz und einmal gegen die Obergrenze. Der Nullwert
// ist gar keine Auswahl und zählt ebenso wenig — er kann nur aus einem
// selbstgebauten Request stammen.
func TestTrainingstermine_DoppelteUndLeereAuswahlZaehlenNicht(t *testing.T) {
	svc := neuerService(t)
	montag, mittwoch, _ := stundenplan(t, svc)

	// Fünf Angaben, aber nur zwei Termine — damit unterhalb der Obergrenze.
	id := mitTerminenAnlegen(t, svc, "Doppelt", montag.ID, 0, mittwoch.ID, montag.ID, mittwoch.ID)

	erwartet := ids(montag, mittwoch)
	if gefunden := ids(laufendeTermine(t, svc, id)...); !slices.Equal(gefunden, erwartet) {
		t.Errorf("Termine = %v, erwartet %v", gefunden, erwartet)
	}

	// Gefragt wird die Ableitung selbst und nicht deren Zutat: eine hier
	// nachgerechnete Länge könnte gar nicht auffallen, wenn sie bricht.
	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if frequenz := m.LaufendeMitgliedschaft().Trainingsfrequenz(); frequenz != 2 {
		t.Errorf("Frequenz = %d, erwartet 2", int(frequenz))
	}
}

// Einen Termin, den es nicht gibt, kann niemand auswählen — das kann nur aus
// einer veralteten Ansicht kommen, und stillschweigend wegzulassen hieße, eine
// Trainingszeit zu verschlucken.
func TestTrainingstermine_UnbekannteIDWirdAbgewiesen(t *testing.T) {
	svc := neuerService(t)
	montag, _, _ := stundenplan(t, svc)

	_, err := svc.Create(service.NeuesMitglied{
		Vorname:            "Test",
		Nachname:           "Unbekannt",
		BeitragCents:       beitragImTest,
		Eintritt:           datum(t, "2026-01-05"),
		TrainingsterminIDs: []int64{montag.ID, 9999},
	})
	pruefeValidierungsfehler(t, err, "Create mit unbekanntem Termin")

	// Auch nachträglich nicht — und der Bestand bleibt, wie er war.
	id := mitTerminenAnlegen(t, svc, "Bestand", montag.ID)

	unbekannt := []int64{9999}
	pruefeValidierungsfehler(t,
		svc.Update(id, service.MitgliedPatch{TrainingsterminIDs: &unbekannt}),
		"Update mit unbekanntem Termin")

	if gefunden := ids(laufendeTermine(t, svc, id)...); !slices.Equal(gefunden, ids(montag)) {
		t.Errorf("Termine nach abgewiesenem Update = %v, erwartet unverändert %v", gefunden, ids(montag))
	}
}

// Ein archivierter Termin ist aus dem Stundenplan genommen: neu vergeben wird
// er nicht mehr (ADR-0008).
func TestTrainingstermine_ArchivierterLaesstSichNichtNeuVergeben(t *testing.T) {
	svc := neuerService(t)
	montag, mittwoch, _ := stundenplan(t, svc)

	if err := svc.SetTrainingsterminArchiviert(mittwoch.ID, true); err != nil {
		t.Fatalf("SetTrainingsterminArchiviert: %v", err)
	}

	id := mitTerminenAnlegen(t, svc, "Neuling", montag.ID)

	dazu := ids(montag, mittwoch)
	err := svc.Update(id, service.MitgliedPatch{TrainingsterminIDs: &dazu})
	pruefeValidierungsfehler(t, err, "Update mit archiviertem Termin")

	// Die Meldung nennt den Termin beim Namen — in einer Auswahl von zehn
	// Zeilen soll niemand raten müssen, welche gemeint ist.
	if meldung := einzigeMeldung(t, err); meldung.Schluessel != "validierung.termin.archiviert" ||
		len(meldung.Args) == 0 || meldung.Args[0] != mittwoch.Anzeige() {
		t.Errorf("Meldung = %+v, erwartet den Termin %q darin", meldung, mittwoch.Anzeige())
	}

	if gefunden := ids(laufendeTermine(t, svc, id)...); !slices.Equal(gefunden, ids(montag)) {
		t.Errorf("Termine = %v, erwartet unverändert %v", gefunden, ids(montag))
	}
}

// Wird ein bereits zugeordneter Termin archiviert, bleibt die Anmeldung stehen
// und zählt weiter zur Frequenz: den Stundenplan aufzuräumen darf nicht
// stillschweigend ändern, was Mitglieder vereinbart haben.
func TestTrainingstermine_ArchivierterBleibtZugeordnetUndZaehltWeiter(t *testing.T) {
	svc := neuerService(t)
	montag, mittwoch, _ := stundenplan(t, svc)

	id := mitTerminenAnlegen(t, svc, "Bestandsschutz", montag.ID, mittwoch.ID)

	if err := svc.SetTrainingsterminArchiviert(mittwoch.ID, true); err != nil {
		t.Fatalf("SetTrainingsterminArchiviert: %v", err)
	}

	gefunden := laufendeTermine(t, svc, id)
	if !slices.Equal(ids(gefunden...), ids(montag, mittwoch)) {
		t.Errorf("Termine = %v, erwartet beide weiterhin", ids(gefunden...))
	}
	if !gefunden[1].Archiviert {
		t.Error("der zugeordnete Termin ist nicht als archiviert erkennbar")
	}

	eintrag, err := svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag: %v", err)
	}
	if eintrag.Trainingsfrequenz() != 2 {
		t.Errorf("Frequenz = %d, erwartet 2 — die archivierte Zuordnung zählt mit",
			int(eintrag.Trainingsfrequenz()))
	}

	// Speichern mit unveränderter Auswahl muss durchgehen: sonst ließe sich an
	// diesem Mitglied gar nichts mehr ändern, solange der alte Termin dranhängt.
	unveraendert := ids(montag, mittwoch)
	if err := svc.Update(id, service.MitgliedPatch{TrainingsterminIDs: &unveraendert}); err != nil {
		t.Fatalf("Update mit unverändert zugeordnetem archiviertem Termin: %v", err)
	}

	// Abwählen lässt er sich dagegen sehr wohl.
	nurMontag := ids(montag)
	if err := svc.Update(id, service.MitgliedPatch{TrainingsterminIDs: &nurMontag}); err != nil {
		t.Fatalf("Update (archivierten abwählen): %v", err)
	}
	if gefunden := ids(laufendeTermine(t, svc, id)...); !slices.Equal(gefunden, nurMontag) {
		t.Errorf("Termine nach dem Abwählen = %v, erwartet %v", gefunden, nurMontag)
	}
}

// Terminauswahl ist, was das Formular anbietet: der gepflegte Stundenplan, dazu
// die bereits zugeordneten archivierten — und zwar an ihrem Platz in der Woche.
func TestTerminauswahl_ZeigtDenStundenplanUndZugeordneteArchivierte(t *testing.T) {
	svc := neuerService(t)
	montag, mittwoch, samstag := stundenplan(t, svc)

	if err := svc.SetTrainingsterminArchiviert(mittwoch.ID, true); err != nil {
		t.Fatalf("SetTrainingsterminArchiviert: %v", err)
	}

	// Ohne Zuordnung fehlt der archivierte: neu vergeben wird er nicht.
	auswahl, err := svc.Terminauswahl(nil)
	if err != nil {
		t.Fatalf("Terminauswahl(nil): %v", err)
	}
	if erwartet := ids(montag, samstag); !slices.Equal(ids(auswahl...), erwartet) {
		t.Errorf("Terminauswahl ohne Zuordnung = %v, erwartet %v", ids(auswahl...), erwartet)
	}

	// Mit Zuordnung steht er mit da, an seinem Platz in der Woche.
	auswahl, err = svc.Terminauswahl(ids(mittwoch))
	if err != nil {
		t.Fatalf("Terminauswahl: %v", err)
	}
	if erwartet := ids(montag, mittwoch, samstag); !slices.Equal(ids(auswahl...), erwartet) {
		t.Errorf("Terminauswahl mit Zuordnung = %v, erwartet %v", ids(auswahl...), erwartet)
	}

	// Eine ID, zu der es nichts gibt, erweitert die Auswahl nicht und ist kein
	// Fehler: welche Auswahl gültig ist, entscheidet das Speichern.
	auswahl, err = svc.Terminauswahl([]int64{9999})
	if err != nil {
		t.Fatalf("Terminauswahl mit unbekannter ID: %v", err)
	}
	if erwartet := ids(montag, samstag); !slices.Equal(ids(auswahl...), erwartet) {
		t.Errorf("Terminauswahl mit unbekannter ID = %v, erwartet %v", ids(auswahl...), erwartet)
	}
}

// Ankreuzen und Abwählen sind derselbe Vorgang: das Formular schickt die
// Termine, die danach gelten sollen.
func TestTrainingstermine_AenderungErsetztDenBestand(t *testing.T) {
	svc := neuerService(t)
	montag, _, samstag := stundenplan(t, svc)

	id := mitTerminenAnlegen(t, svc, "Berger", montag.ID)

	dazu := ids(montag, samstag)
	if err := svc.Update(id, service.MitgliedPatch{TrainingsterminIDs: &dazu}); err != nil {
		t.Fatalf("Update (ankreuzen): %v", err)
	}
	if gefunden := ids(laufendeTermine(t, svc, id)...); !slices.Equal(gefunden, dazu) {
		t.Errorf("Termine nach dem Ankreuzen = %v, erwartet %v", gefunden, dazu)
	}

	weniger := ids(samstag)
	if err := svc.Update(id, service.MitgliedPatch{TrainingsterminIDs: &weniger}); err != nil {
		t.Fatalf("Update (abwählen): %v", err)
	}
	if gefunden := ids(laufendeTermine(t, svc, id)...); !slices.Equal(gefunden, weniger) {
		t.Errorf("Termine nach dem Abwählen = %v, erwartet %v", gefunden, weniger)
	}

	// Alle abwählen ist erlaubt und heißt: keine Frequenz.
	keine := []int64{}
	if err := svc.Update(id, service.MitgliedPatch{TrainingsterminIDs: &keine}); err != nil {
		t.Fatalf("Update (alle abwählen): %v", err)
	}
	if gefunden := laufendeTermine(t, svc, id); len(gefunden) != 0 {
		t.Errorf("Termine nach dem Leeren = %v, erwartet keine", anzeigen(gefunden))
	}
}

// Ein Patch ohne Termine ist keine Aussage über sie: wer nur die Adresse
// korrigiert, darf damit nicht das Training löschen.
func TestTrainingstermine_BleibenOhneAngabeImPatchUnberuehrt(t *testing.T) {
	svc := neuerService(t)
	montag, mittwoch, samstag := stundenplan(t, svc)

	alle := ids(montag, mittwoch, samstag)
	id := mitTerminenAnlegen(t, svc, "Öztürk", alle...)

	ort := "Berlin"
	if err := svc.Update(id, service.MitgliedPatch{Anschrift: &service.Anschrift{Ort: ort}}); err != nil {
		t.Fatalf("Update ohne Termine: %v", err)
	}

	if gefunden := ids(laufendeTermine(t, svc, id)...); !slices.Equal(gefunden, alle) {
		t.Errorf("Termine = %v, erwartet unverändert %v", gefunden, alle)
	}
}

// Die Anmeldung hängt an der Mitgliedschaft, nicht am Mitglied: ein
// Wiedereintritt ist eine neue Vereinbarung und fängt ohne Trainingszeiten an,
// während die alte Mitgliedschaft ihre behält.
func TestTrainingstermine_WiedereintrittLaesstDieAltenStehen(t *testing.T) {
	svc := neuerService(t)
	montag, _, samstag := stundenplan(t, svc)

	alt := ids(montag, samstag)
	id := mitTerminenAnlegen(t, svc, "Klein", alt...)

	if err := svc.SetKuendigung(id, austrittZum(datum(t, "2026-06-30"))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
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

	alte, neu := m.Mitgliedschaften[0], m.Mitgliedschaften[1]
	if gefunden := ids(alte.Trainingstermine...); !slices.Equal(gefunden, alt) {
		t.Errorf("Termine der alten Mitgliedschaft = %v, erwartet unverändert %v", gefunden, alt)
	}
	if len(neu.Trainingstermine) != 0 {
		t.Errorf("Termine der neuen Mitgliedschaft = %v, erwartet keine", anzeigen(neu.Trainingstermine))
	}
	if neu.Trainingsfrequenz() != 0 {
		t.Errorf("Frequenz der neuen Mitgliedschaft = %d, erwartet 0", int(neu.Trainingsfrequenz()))
	}

	// Und die neue Auswahl landet an der neuen Mitgliedschaft, nicht an der alten.
	nurSamstag := ids(samstag)
	if err := svc.Update(id, service.MitgliedPatch{TrainingsterminIDs: &nurSamstag}); err != nil {
		t.Fatalf("Update nach Wiedereintritt: %v", err)
	}

	m, err = svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if gefunden := ids(m.Mitgliedschaften[0].Trainingstermine...); !slices.Equal(gefunden, alt) {
		t.Errorf("Termine der alten Mitgliedschaft = %v, erwartet unverändert %v", gefunden, alt)
	}
	if gefunden := ids(m.Mitgliedschaften[1].Trainingstermine...); !slices.Equal(gefunden, nurSamstag) {
		t.Errorf("Termine der neuen Mitgliedschaft = %v, erwartet %v", gefunden, nurSamstag)
	}
}

// Ein ausgetretenes Mitglied zeigt in der Liste die Termine seines letzten
// Zeitraums — dieselbe Mitgliedschaft, aus der auch Beitrag und Eintritt kommen.
func TestTrainingstermine_ListeZeigtDieDerMassgeblichenMitgliedschaft(t *testing.T) {
	svc := neuerService(t)
	montag, _, samstag := stundenplan(t, svc)

	erwartet := ids(montag, samstag)
	id := mitTerminenAnlegen(t, svc, "Klein", erwartet...)
	if err := svc.SetKuendigung(id, austrittZum(datum(t, "2026-06-30"))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	eintrag, err := svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag: %v", err)
	}
	if gefunden := ids(eintrag.Trainingstermine...); !slices.Equal(gefunden, erwartet) {
		t.Errorf("Termine des ausgetretenen Mitglieds = %v, erwartet %v", gefunden, erwartet)
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
