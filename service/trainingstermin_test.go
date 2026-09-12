package service_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/ToniBlankenburg/boxclub/service"
)

// Die Tests des Stundenplans laufen wie alle übrigen gegen eine echte Datenbank
// (neuerService in member_service_test.go) und beobachten den Zustand
// ausschließlich über die MemberService-API.

// terminangabe ist die Fixture-Kurzform für einen vollständigen Termin.
func terminangabe(tag service.Wochentag, beginn, ende, bezeichnung string) service.Trainingsterminangabe {
	return service.Trainingsterminangabe{
		Wochentag:   tag,
		Beginn:      beginn,
		Ende:        ende,
		Bezeichnung: bezeichnung,
	}
}

// angelegterTermin legt einen Termin an und liefert ihn zurückgelesen — der
// Weg, den fast jeder Test hier zuerst geht.
func angelegterTermin(t *testing.T, svc *service.MemberService, a service.Trainingsterminangabe) service.Trainingstermin {
	t.Helper()

	id, err := svc.CreateTrainingstermin(a)
	if err != nil {
		t.Fatalf("CreateTrainingstermin: %v", err)
	}

	termin, err := svc.GetTrainingstermin(id)
	if err != nil {
		t.Fatalf("GetTrainingstermin: %v", err)
	}

	return termin
}

// anzeigen verdichtet eine Terminliste auf ihre Anzeigetexte — die Form, in der
// sich Reihenfolge und Inhalt einer Liste in einem Zug vergleichen lassen.
func anzeigen(termine []service.Trainingstermin) []string {
	texte := make([]string, 0, len(termine))
	for _, t := range termine {
		texte = append(texte, t.Anzeige())
	}

	return texte
}

func TestCreateTrainingstermin_HaeltAlleVierAngaben(t *testing.T) {
	svc := neuerService(t)

	termin := angelegterTermin(t, svc, terminangabe(service.Samstag, "10:30", "12:00", "Anfänger"))

	if termin.Wochentag != service.Samstag {
		t.Errorf("Wochentag = %d, erwartet Samstag", termin.Wochentag)
	}
	if termin.Beginn != "10:30" || termin.Ende != "12:00" {
		t.Errorf("Beginn/Ende = %q/%q, erwartet 10:30/12:00", termin.Beginn, termin.Ende)
	}
	if termin.Bezeichnung != "Anfänger" {
		t.Errorf("Bezeichnung = %q, erwartet Anfänger", termin.Bezeichnung)
	}
	if termin.Archiviert {
		t.Error("ein neuer Termin ist archiviert, erwartet: nicht archiviert")
	}
	if termin.Anzeige() != "Samstag 10:30 – 12:00 · Anfänger" {
		t.Errorf("Anzeige = %q", termin.Anzeige())
	}
}

// Ende und Bezeichnung sind freiwillig: nicht jeder Verein weiß, wann das
// Training endet, und "Dienstag 18:00" ist auch ohne Namen ein Termin.
func TestCreateTrainingstermin_OhneEndeUndBezeichnung(t *testing.T) {
	svc := neuerService(t)

	termin := angelegterTermin(t, svc, terminangabe(service.Dienstag, "18:00", "", ""))

	if !termin.Ende.Leer() {
		t.Errorf("Ende = %q, erwartet leer", termin.Ende)
	}
	if termin.Anzeige() != "Dienstag 18:00" {
		t.Errorf("Anzeige = %q, erwartet »Dienstag 18:00«", termin.Anzeige())
	}
}

// Uhrzeiten werden beim Speichern vereinheitlicht. Daran hängt mehr als die
// Optik: sortiert und verglichen wird die Textform, und "9:05" stünde sonst
// hinter "10:30".
func TestCreateTrainingstermin_VereinheitlichtUhrzeitUndSchneidetLeerraum(t *testing.T) {
	svc := neuerService(t)

	termin := angelegterTermin(t, svc, terminangabe(service.Montag, " 9:05 ", " 10:05 ", "  Wettkampf  "))

	if termin.Beginn != "09:05" {
		t.Errorf("Beginn = %q, erwartet 09:05", termin.Beginn)
	}
	if termin.Ende != "10:05" {
		t.Errorf("Ende = %q, erwartet 10:05", termin.Ende)
	}
	if termin.Bezeichnung != "Wettkampf" {
		t.Errorf("Bezeichnung = %q, erwartet »Wettkampf«", termin.Bezeichnung)
	}
}

// Die Liste steht in Wochenreihenfolge und innerhalb des Tages nach Beginn —
// deshalb ist der Wochentag eine Zahl und kein Text (ADR-0008): nach dem
// Alphabet käme Dienstag vor Donnerstag vor Freitag vor Mittwoch.
func TestListTrainingstermine_StehtInWochenreihenfolge(t *testing.T) {
	svc := neuerService(t)

	for _, a := range []service.Trainingsterminangabe{
		terminangabe(service.Mittwoch, "18:00", "", ""),
		terminangabe(service.Samstag, "10:30", "12:00", "Anfänger"),
		terminangabe(service.Montag, "20:00", "", "Sparring"),
		terminangabe(service.Montag, "18:00", "19:30", ""),
		terminangabe(service.Donnerstag, "18:00", "", ""),
	} {
		if _, err := svc.CreateTrainingstermin(a); err != nil {
			t.Fatalf("CreateTrainingstermin: %v", err)
		}
	}

	termine, err := svc.ListTrainingstermine(false)
	if err != nil {
		t.Fatalf("ListTrainingstermine: %v", err)
	}

	erwartet := []string{
		"Montag 18:00 – 19:30",
		"Montag 20:00 · Sparring",
		"Mittwoch 18:00",
		"Donnerstag 18:00",
		"Samstag 10:30 – 12:00 · Anfänger",
	}
	if got := anzeigen(termine); !slices.Equal(got, erwartet) {
		t.Errorf("Reihenfolge = %v, erwartet %v", got, erwartet)
	}
}

func TestUpdateTrainingstermin_ErsetztAlleAngaben(t *testing.T) {
	svc := neuerService(t)

	termin := angelegterTermin(t, svc, terminangabe(service.Samstag, "10:30", "12:00", "Anfänger"))

	if err := svc.UpdateTrainingstermin(termin.ID, terminangabe(service.Freitag, "17:00", "", "")); err != nil {
		t.Fatalf("UpdateTrainingstermin: %v", err)
	}

	geaendert, err := svc.GetTrainingstermin(termin.ID)
	if err != nil {
		t.Fatalf("GetTrainingstermin: %v", err)
	}

	if geaendert.Anzeige() != "Freitag 17:00" {
		t.Errorf("Anzeige = %q, erwartet »Freitag 17:00« — Ende und Bezeichnung sollen geleert sein", geaendert.Anzeige())
	}
}

// Archivieren statt Löschen (ADR-0008): der Termin verschwindet aus der Liste
// und damit aus jeder Auswahl, bleibt aber lesbar — sonst nähme das Aufräumen
// des Stundenplans jeder daran angemeldeten Mitgliedschaft still ihre Frequenz.
func TestSetTrainingsterminArchiviert_NimmtDenTerminAusDerListe(t *testing.T) {
	svc := neuerService(t)

	termin := angelegterTermin(t, svc, terminangabe(service.Samstag, "10:30", "", "Anfänger"))
	bleibt := angelegterTermin(t, svc, terminangabe(service.Montag, "18:00", "", ""))

	if err := svc.SetTrainingsterminArchiviert(termin.ID, true); err != nil {
		t.Fatalf("SetTrainingsterminArchiviert: %v", err)
	}

	offen, err := svc.ListTrainingstermine(false)
	if err != nil {
		t.Fatalf("ListTrainingstermine: %v", err)
	}
	if len(offen) != 1 || offen[0].ID != bleibt.ID {
		t.Errorf("Liste = %v, erwartet nur den nicht archivierten Termin", anzeigen(offen))
	}

	alle, err := svc.ListTrainingstermine(true)
	if err != nil {
		t.Fatalf("ListTrainingstermine(true): %v", err)
	}
	if len(alle) != 2 {
		t.Fatalf("Liste mit Archivierten = %v, erwartet beide Termine", anzeigen(alle))
	}

	archiviert, err := svc.GetTrainingstermin(termin.ID)
	if err != nil {
		t.Fatalf("GetTrainingstermin: %v", err)
	}
	if !archiviert.Archiviert {
		t.Error("Archiviert = false, erwartet true")
	}
}

// Ein Fehlgriff darf nicht endgültig sein: was archiviert wurde, lässt sich
// wieder aufnehmen.
func TestSetTrainingsterminArchiviert_ReaktiviertWieder(t *testing.T) {
	svc := neuerService(t)

	termin := angelegterTermin(t, svc, terminangabe(service.Samstag, "10:30", "", "Anfänger"))

	if err := svc.SetTrainingsterminArchiviert(termin.ID, true); err != nil {
		t.Fatalf("archivieren: %v", err)
	}
	if err := svc.SetTrainingsterminArchiviert(termin.ID, false); err != nil {
		t.Fatalf("reaktivieren: %v", err)
	}

	offen, err := svc.ListTrainingstermine(false)
	if err != nil {
		t.Fatalf("ListTrainingstermine: %v", err)
	}
	if len(offen) != 1 || offen[0].ID != termin.ID {
		t.Errorf("Liste = %v, erwartet den reaktivierten Termin", anzeigen(offen))
	}
}

// Ein Ende vor dem Beginn ist ein Tippfehler und keine Nachtschicht.
func TestCreateTrainingstermin_EndeVorBeginnIstEinFehler(t *testing.T) {
	svc := neuerService(t)

	_, err := svc.CreateTrainingstermin(terminangabe(service.Samstag, "12:00", "10:30", ""))

	meldung := einzigeMeldung(t, err)
	if !strings.Contains(meldung, "Ende") {
		t.Errorf("Meldung = %q, erwartet einen Hinweis auf das Ende", meldung)
	}
}

func TestUpdateTrainingstermin_EndeVorBeginnIstEinFehler(t *testing.T) {
	svc := neuerService(t)

	termin := angelegterTermin(t, svc, terminangabe(service.Samstag, "10:30", "12:00", "Anfänger"))

	err := svc.UpdateTrainingstermin(termin.ID, terminangabe(service.Samstag, "12:00", "10:30", "Anfänger"))
	einzigeMeldung(t, err)

	unveraendert, err := svc.GetTrainingstermin(termin.ID)
	if err != nil {
		t.Fatalf("GetTrainingstermin: %v", err)
	}
	if unveraendert.Anzeige() != "Samstag 10:30 – 12:00 · Anfänger" {
		t.Errorf("Anzeige = %q — die abgewiesene Änderung darf nichts geschrieben haben", unveraendert.Anzeige())
	}
}

// Gleicher Zeitpunkt ist kein Ende vor dem Beginn; abgewiesen wird nur, was
// rückwärts läuft (Ticket 20).
func TestCreateTrainingstermin_EndeGleichBeginnIstErlaubt(t *testing.T) {
	svc := neuerService(t)

	if _, err := svc.CreateTrainingstermin(terminangabe(service.Samstag, "10:30", "10:30", "")); err != nil {
		t.Fatalf("CreateTrainingstermin: %v", err)
	}
}

func TestCreateTrainingstermin_PflichtangabenFehlen(t *testing.T) {
	svc := neuerService(t)

	faelle := []struct {
		name   string
		angabe service.Trainingsterminangabe
	}{
		{"ohne Wochentag", terminangabe(0, "18:00", "", "")},
		{"unbekannter Wochentag", terminangabe(service.Wochentag(8), "18:00", "", "")},
		{"ohne Beginn", terminangabe(service.Montag, "", "", "Anfänger")},
		{"Beginn ist keine Uhrzeit", terminangabe(service.Montag, "abends", "", "")},
		{"Ende ist keine Uhrzeit", terminangabe(service.Montag, "18:00", "25:00", "")},
	}

	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			if _, err := svc.CreateTrainingstermin(f.angabe); err == nil {
				t.Fatal("CreateTrainingstermin nahm die Angabe an, erwartet ValidierungsFehler")
			} else {
				var validierung *service.ValidierungsFehler
				if !errors.As(err, &validierung) {
					t.Fatalf("Fehler = %v, erwartet ValidierungsFehler", err)
				}
			}
		})
	}
}

// Eine leere Datenbank hat keinen Stundenplan: geseedet wird nichts (CLAUDE.md).
func TestListTrainingstermine_LeereDatenbankHatKeineTermine(t *testing.T) {
	svc := neuerService(t)

	termine, err := svc.ListTrainingstermine(true)
	if err != nil {
		t.Fatalf("ListTrainingstermine: %v", err)
	}
	if len(termine) != 0 {
		t.Errorf("Liste = %v, erwartet leer", anzeigen(termine))
	}
}

func TestTrainingstermin_UnbekannteIDMeldetNichtGefunden(t *testing.T) {
	svc := neuerService(t)

	if _, err := svc.GetTrainingstermin(999); !errors.Is(err, service.ErrNichtGefunden) {
		t.Errorf("GetTrainingstermin = %v, erwartet ErrNichtGefunden", err)
	}
	if err := svc.UpdateTrainingstermin(999, terminangabe(service.Montag, "18:00", "", "")); !errors.Is(err, service.ErrNichtGefunden) {
		t.Errorf("UpdateTrainingstermin = %v, erwartet ErrNichtGefunden", err)
	}
	if err := svc.SetTrainingsterminArchiviert(999, true); !errors.Is(err, service.ErrNichtGefunden) {
		t.Errorf("SetTrainingsterminArchiviert = %v, erwartet ErrNichtGefunden", err)
	}
}

// Die Wochentage sind die Auswahlliste des Formulars und stehen deshalb in
// Wochenreihenfolge.
func TestWochentage_StehenInWochenreihenfolge(t *testing.T) {
	erwartet := []string{"Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag", "Sonntag"}

	tage := service.Wochentage()
	if len(tage) != len(erwartet) {
		t.Fatalf("Wochentage() hat %d Einträge, erwartet %d", len(tage), len(erwartet))
	}

	for i, tag := range tage {
		if tag.Bezeichnung() != erwartet[i] {
			t.Errorf("Wochentage()[%d] = %q, erwartet %q", i, tag.Bezeichnung(), erwartet[i])
		}
	}
}

// einzigeMeldung besteht darauf, dass der Fehler ein ValidierungsFehler mit
// genau einer Meldung ist, und liefert sie.
func einzigeMeldung(t *testing.T, err error) string {
	t.Helper()

	var validierung *service.ValidierungsFehler
	if !errors.As(err, &validierung) {
		t.Fatalf("Fehler = %v, erwartet ValidierungsFehler", err)
	}
	if len(validierung.Meldungen) != 1 {
		t.Fatalf("Meldungen = %v, erwartet genau eine", validierung.Meldungen)
	}

	return validierung.Meldungen[0]
}
