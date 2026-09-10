package service_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/ToniBlankenburg/boxclub/service"
)

// Aus- und Wiedereintritt am MemberService-Seam. Beobachtet wird ausschließlich
// über die Service-API — dass zwei Mitgliedschaften entstanden sind, zeigt Get,
// nicht ein Blick in die Tabelle.

func TestMarkExit_BeendetDieLaufendeMitgliedschaft(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Nina", "Klein")
	austritt := datum(t, "2026-06-30")

	if err := svc.MarkExit(id, austritt); err != nil {
		t.Fatalf("MarkExit: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if laufend := m.LaufendeMitgliedschaft(); laufend != nil {
		t.Errorf("LaufendeMitgliedschaft = %+v, erwartet nil nach dem Austritt", laufend)
	}
	if len(m.Mitgliedschaften) != 1 {
		t.Fatalf("Mitgliedschaften = %d, erwartet 1 (der Austritt legt keine neue an)", len(m.Mitgliedschaften))
	}
	if m.Mitgliedschaften[0].Austritt == nil || !m.Mitgliedschaften[0].Austritt.Equal(austritt) {
		t.Errorf("Austritt = %v, erwartet %v", m.Mitgliedschaften[0].Austritt, austritt)
	}
}

// Der Sinn des Austritts aus Sicht des Nutzers: die Liste wird wieder übersichtlich,
// ohne dass der Datensatz verloren geht.
func TestMarkExit_NimmtDasMitgliedAusDerStandardansicht(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Nina", "Klein")
	if err := svc.MarkExit(id, datum(t, "2026-06-30")); err != nil {
		t.Fatalf("MarkExit: %v", err)
	}

	liste, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(liste) != 0 {
		t.Errorf("List = %+v, erwartet leer — Ausgetretene gehören nicht in die Standardansicht", liste)
	}

	ehemalige, err := svc.Search("", service.Suchfilter{AuchEhemalige: true})
	if err != nil {
		t.Fatalf("Search (auch Ehemalige): %v", err)
	}
	if len(ehemalige) != 1 || ehemalige[0].MitgliedID != id {
		t.Fatalf("Search (auch Ehemalige) = %+v, erwartet Mitglied %d", ehemalige, id)
	}
	if ehemalige[0].Austritt == nil {
		t.Error("Austritt = nil, erwartet das Austrittsdatum — sonst ist die Zeile nicht als ehemalig erkennbar")
	}
}

// Der Kern des Tickets: dieselbe Person, zwei Zeiträume. Wer bei einem
// Wiedereintritt ein zweites Mitglied anlegte, verlöre Stammdaten und Historie.
func TestRejoin_LegtNeueMitgliedschaftAmSelbenMitgliedAn(t *testing.T) {
	svc := neuerService(t)

	// Vollständig gefüllte Stammdaten: die Zusicherung des Tickets lautet
	// "Stammdaten, ID, Historie bleiben" — mit leeren Feldern wäre sie nicht
	// bewiesen, sondern nur nicht widerlegt.
	geburtsdatum := datum(t, "1994-11-02")
	bezahltBis := datum(t, "2026-12-31")

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:          "Nina",
		Nachname:         "Klein",
		Geburtsdatum:     &geburtsdatum,
		Adresse:          "Kanalstraße 12, 12043 Berlin",
		Email:            "nina.klein@example.org",
		Telefon:          "0160 4443322",
		BeitragsklasseID: beitragsklassen(t, svc)[1].ID,
		Eintritt:         datum(t, "2026-01-05"),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := svc.SetBezahltBis(id, &bezahltBis); err != nil {
		t.Fatalf("SetBezahltBis: %v", err)
	}

	vorher, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get vor dem Austritt: %v", err)
	}

	austritt := datum(t, "2026-06-30")
	if err := svc.MarkExit(id, austritt); err != nil {
		t.Fatalf("MarkExit: %v", err)
	}

	wiedereintritt := datum(t, "2026-09-01")
	if err := svc.Rejoin(id, wiedereintritt); err != nil {
		t.Fatalf("Rejoin: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get nach dem Wiedereintritt: %v", err)
	}

	// Stammdaten und ID bleiben dieselben — ein Wiedereintritt ist keine
	// Neuanlage. Verglichen wird das ganze Mitglied ohne seine Zeiträume: so
	// deckt der Test auch ein Feld ab, das erst später dazukommt.
	stammdaten := func(m service.Mitglied) service.Mitglied {
		m.Mitgliedschaften = nil
		return m
	}
	if !reflect.DeepEqual(stammdaten(m), stammdaten(vorher)) {
		t.Errorf("Stammdaten = %+v, erwartet unverändert %+v", stammdaten(m), stammdaten(vorher))
	}

	if len(m.Mitgliedschaften) != 2 {
		t.Fatalf("Mitgliedschaften = %d, erwartet 2 (der beendete und der neue Zeitraum)", len(m.Mitgliedschaften))
	}

	// Der erste Zeitraum bleibt beendet, der zweite läuft.
	erster := m.Mitgliedschaften[0]
	if erster.Austritt == nil || !erster.Austritt.Equal(austritt) {
		t.Errorf("erster Zeitraum: Austritt = %v, erwartet %v", erster.Austritt, austritt)
	}
	if !erster.Eintritt.Equal(vorher.Mitgliedschaften[0].Eintritt) {
		t.Errorf("erster Zeitraum: Eintritt = %v, erwartet unverändert %v",
			erster.Eintritt, vorher.Mitgliedschaften[0].Eintritt)
	}

	laufend := m.LaufendeMitgliedschaft()
	if laufend == nil {
		t.Fatal("LaufendeMitgliedschaft = nil, erwartet den neuen Zeitraum")
	}
	if !laufend.Eintritt.Equal(wiedereintritt) {
		t.Errorf("laufender Eintritt = %v, erwartet %v", laufend.Eintritt, wiedereintritt)
	}
	if laufend.ID == erster.ID {
		t.Error("Rejoin hat den beendeten Zeitraum wiederbelebt, erwartet eine neue Mitgliedschaft")
	}

	// Und damit steht das Mitglied wieder in der Standardansicht.
	liste, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(liste) != 1 || liste[0].MitgliedID != id {
		t.Errorf("List = %+v, erwartet das wieder eingetretene Mitglied %d", liste, id)
	}
}

// Zweimal austreten geht nicht: nach dem ersten Austritt gibt es keinen
// laufenden Zeitraum mehr, den ein zweiter beenden könnte. Das stillschweigend
// durchzuwinken würde das erste Austrittsdatum überschreiben.
func TestMarkExit_ZweimalIstEinFehler(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Nina", "Klein")
	austritt := datum(t, "2026-06-30")

	if err := svc.MarkExit(id, austritt); err != nil {
		t.Fatalf("erster MarkExit: %v", err)
	}

	err := svc.MarkExit(id, datum(t, "2026-07-31"))
	if !errors.Is(err, service.ErrNichtAktiv) {
		t.Fatalf("zweiter MarkExit = %v, erwartet ErrNichtAktiv", err)
	}

	// Und das erste Austrittsdatum steht unverändert.
	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(m.Mitgliedschaften) != 1 {
		t.Fatalf("Mitgliedschaften = %d, erwartet 1", len(m.Mitgliedschaften))
	}
	if m.Mitgliedschaften[0].Austritt == nil || !m.Mitgliedschaften[0].Austritt.Equal(austritt) {
		t.Errorf("Austritt = %v, erwartet unverändert %v", m.Mitgliedschaften[0].Austritt, austritt)
	}
}

// Wiedereintritt setzt einen Austritt voraus. Ohne diese Regel entstünden zwei
// gleichzeitig laufende Zeiträume — und "aktiv seit wann" hätte zwei Antworten.
func TestRejoin_AufAktivesMitgliedIstEinFehler(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Nina", "Klein")

	err := svc.Rejoin(id, datum(t, "2026-09-01"))
	if !errors.Is(err, service.ErrBereitsAktiv) {
		t.Fatalf("Rejoin = %v, erwartet ErrBereitsAktiv", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(m.Mitgliedschaften) != 1 {
		t.Errorf("Mitgliedschaften = %d, erwartet 1 — der abgelehnte Wiedereintritt darf nichts anlegen",
			len(m.Mitgliedschaften))
	}
}

// Ein Austritt vor dem Eintritt ergäbe einen Zeitraum negativer Länge. Das ist
// immer ein Tippfehler und wird abgelehnt, statt still gespeichert zu werden.
func TestMarkExit_VorDemEintrittIstEinFehler(t *testing.T) {
	svc := neuerService(t)

	// mitgliedAnlegen setzt den Eintritt auf den 05.01.2026.
	id := mitgliedAnlegen(t, svc, "Nina", "Klein")

	err := svc.MarkExit(id, datum(t, "2026-01-04"))

	var validierung *service.ValidierungsFehler
	if !errors.As(err, &validierung) {
		t.Fatalf("MarkExit = %v, erwartet einen ValidierungsFehler", err)
	}
	if len(validierung.Meldungen) == 0 {
		t.Error("ValidierungsFehler ohne Meldung — das Formular hätte nichts anzuzeigen")
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if laufend := m.LaufendeMitgliedschaft(); laufend == nil {
		t.Error("die Mitgliedschaft wurde trotz abgelehntem Datum beendet")
	}
}

// Der Eintrittstag selbst ist erlaubt: wer am Tag der Anmeldung wieder abspringt,
// hat einen Zeitraum von einem Tag — ungewöhnlich, aber kein Fehler.
func TestMarkExit_AmEintrittstagIstErlaubt(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Nina", "Klein")
	eintritt := datum(t, "2026-01-05")

	if err := svc.MarkExit(id, eintritt); err != nil {
		t.Fatalf("MarkExit am Eintrittstag: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if m.Mitgliedschaften[0].Austritt == nil || !m.Mitgliedschaften[0].Austritt.Equal(eintritt) {
		t.Errorf("Austritt = %v, erwartet %v", m.Mitgliedschaften[0].Austritt, eintritt)
	}
}

// Ohne Datum gibt es keinen Zeitraum. Die Regel liegt hier und nicht im
// Formular, damit sie nur an einer Stelle steht.
func TestAusUndWiedereintritt_OhneDatumIstEinFehler(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Nina", "Klein")

	var validierung *service.ValidierungsFehler

	if err := svc.MarkExit(id, time.Time{}); !errors.As(err, &validierung) {
		t.Errorf("MarkExit ohne Datum = %v, erwartet einen ValidierungsFehler", err)
	}

	if err := svc.MarkExit(id, datum(t, "2026-06-30")); err != nil {
		t.Fatalf("MarkExit: %v", err)
	}

	if err := svc.Rejoin(id, time.Time{}); !errors.As(err, &validierung) {
		t.Errorf("Rejoin ohne Datum = %v, erwartet einen ValidierungsFehler", err)
	}
}

// Ein Wiedereintritt vor dem letzten Austritt ließe zwei Zeiträume überlappen —
// dieselbe Tippfehler-Klasse wie ein Austritt vor dem Eintritt.
func TestRejoin_VorDemLetztenAustrittIstEinFehler(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Nina", "Klein")
	if err := svc.MarkExit(id, datum(t, "2026-06-30")); err != nil {
		t.Fatalf("MarkExit: %v", err)
	}

	err := svc.Rejoin(id, datum(t, "2026-06-29"))

	var validierung *service.ValidierungsFehler
	if !errors.As(err, &validierung) {
		t.Fatalf("Rejoin = %v, erwartet einen ValidierungsFehler", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(m.Mitgliedschaften) != 1 {
		t.Errorf("Mitgliedschaften = %d, erwartet 1 — der abgelehnte Wiedereintritt darf nichts anlegen",
			len(m.Mitgliedschaften))
	}
}

// Zu einer unbekannten ID gibt es weder Aus- noch Wiedereintritt.
func TestAusUndWiedereintritt_UnbekannteIDMeldetNichtGefunden(t *testing.T) {
	svc := neuerService(t)

	if err := svc.MarkExit(4711, datum(t, "2026-06-30")); !errors.Is(err, service.ErrNichtGefunden) {
		t.Errorf("MarkExit(4711) = %v, erwartet ErrNichtGefunden", err)
	}
	if err := svc.Rejoin(4711, datum(t, "2026-09-01")); !errors.Is(err, service.ErrNichtGefunden) {
		t.Errorf("Rejoin(4711) = %v, erwartet ErrNichtGefunden", err)
	}
}

// LetzteMitgliedschaft ist die Angabe, die eine Ansicht braucht, wenn sie einen
// Eintritt zeigen will: für ein aktives Mitglied den laufenden Zeitraum, für ein
// ehemaliges den zuletzt beendeten. Ohne sie müsste die Oberfläche selbst
// entscheiden, welcher Zeitraum zählt — eine Lebenszyklus-Regel, die laut
// ADR-0002 nicht dorthin gehört.
func TestLetzteMitgliedschaft_FolgtDemLebenszyklus(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Nina", "Klein")
	ersterEintritt := datum(t, "2026-01-05")

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	letzte := m.LetzteMitgliedschaft()
	if letzte == nil || !letzte.Eintritt.Equal(ersterEintritt) {
		t.Fatalf("LetzteMitgliedschaft = %+v, erwartet den laufenden Zeitraum ab %v", letzte, ersterEintritt)
	}

	// Ausgetreten: der zuletzt beendete Zeitraum bleibt der maßgebliche.
	if err := svc.MarkExit(id, datum(t, "2026-06-30")); err != nil {
		t.Fatalf("MarkExit: %v", err)
	}
	if m, err = svc.Get(id); err != nil {
		t.Fatalf("Get nach dem Austritt: %v", err)
	}
	letzte = m.LetzteMitgliedschaft()
	if letzte == nil || !letzte.Eintritt.Equal(ersterEintritt) {
		t.Fatalf("LetzteMitgliedschaft = %+v, erwartet den beendeten Zeitraum ab %v", letzte, ersterEintritt)
	}

	// Wieder eingetreten: jetzt zählt der neue Zeitraum.
	zweiterEintritt := datum(t, "2026-09-01")
	if err := svc.Rejoin(id, zweiterEintritt); err != nil {
		t.Fatalf("Rejoin: %v", err)
	}
	if m, err = svc.Get(id); err != nil {
		t.Fatalf("Get nach dem Wiedereintritt: %v", err)
	}
	letzte = m.LetzteMitgliedschaft()
	if letzte == nil || !letzte.Eintritt.Equal(zweiterEintritt) {
		t.Errorf("LetzteMitgliedschaft = %+v, erwartet den neuen Zeitraum ab %v", letzte, zweiterEintritt)
	}
}
