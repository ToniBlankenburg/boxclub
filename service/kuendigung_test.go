package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ToniBlankenburg/boxclub/service"
)

// Kündigung und Ruhend am MemberService-Seam. Beide Felder hängen an der
// Mitgliedschaft; beobachtet wird ausschließlich über die Service-API.

// austrittZum ist die Kündigung, die nur den Austritt festhält — der Fall der
// Altbestände und der Lebenszyklus-Tests: jemand tritt aus, ohne dass der Tag
// der Kündigungserklärung erfasst wäre.
func austrittZum(austritt time.Time) service.Kuendigung {
	return service.Kuendigung{Austritt: &austritt}
}

// Die reguläre Frist der Satzung: drei Monate, aufgerundet auf das Monatsende.
// Gerechnet wird sie allein für den Vorschlag im Formular — gespeichert wird
// weiterhin nur, was jemand abschickt (siehe
// TestSetKuendigung_OhneAustrittLeitetKeinenAustrittAb).
func TestRegulaererAustritt_DreiMonateZumMonatsende(t *testing.T) {
	faelle := []struct {
		gekuendigt string
		austritt   string
		warum      string
	}{
		{"2026-09-12", "2026-12-31", "Monatsmitte: drei Monate weiter ist der Dezember"},
		{"2027-01-05", "2027-04-30", "Jahreswechsel spielt keine Rolle"},
		{"2026-06-30", "2026-09-30", "am Monatsletzten erklärt, drei Monate später ebenso"},
		{"2026-12-31", "2027-03-31", "über den Jahreswechsel hinweg"},

		// Der Fall, an dem eine taggenaue Rechnung scheitert: der 30. November
		// plus drei Monate ist der 30. Februar, den es nicht gibt. Weil auf das
		// Monatsende aufgerundet wird, stellt sich die Frage hier gar nicht.
		{"2026-11-30", "2027-02-28", "Februar hat keinen 30."},
		{"2027-11-30", "2028-02-29", "und im Schaltjahr einen 29."},
	}

	for _, f := range faelle {
		got := service.RegulaererAustritt(datum(t, f.gekuendigt))
		if erwartet := datum(t, f.austritt); !got.Equal(erwartet) {
			t.Errorf("RegulaererAustritt(%s) = %s, erwartet %s — %s",
				f.gekuendigt, got.Format("2006-01-02"), f.austritt, f.warum)
		}
	}
}

// Der Vorschlag ist ein Vorschlag: er begrenzt nichts. Eine kürzere Frist
// (Aufhebungsvertrag, Kulanz) bleibt erfassbar — das prüft SetKuendigung nicht
// gegen die Satzung, sondern nur gegen den Eintritt.
func TestRegulaererAustritt_BindetSetKuendigungNicht(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Nina", "Klein")
	gekuendigt := heuteVersetzt(0)
	sofort := heuteVersetzt(1)

	if err := svc.SetKuendigung(id, service.Kuendigung{Datum: &gekuendigt, Austritt: &sofort}); err != nil {
		t.Fatalf("SetKuendigung mit kürzerer Frist als der regulären: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if m.Mitgliedschaften[0].Austritt == nil || !m.Mitgliedschaften[0].Austritt.Equal(sofort) {
		t.Errorf("Austritt = %v, erwartet %v — gespeichert wird der erfasste Tag, nicht der reguläre",
			m.Mitgliedschaften[0].Austritt, sofort)
	}
}

// Der Normalfall des Tickets: gekündigt wird an einem Tag, wirksam wird der
// Austritt an einem späteren. Beide Daten stehen nebeneinander an der
// Mitgliedschaft.
func TestSetKuendigung_HaeltBeideDatenFest(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Nina", "Klein")
	gekuendigt := datum(t, "2026-03-15")
	austritt := datum(t, "2026-06-30")

	if err := svc.SetKuendigung(id, service.Kuendigung{Datum: &gekuendigt, Austritt: &austritt}); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(m.Mitgliedschaften) != 1 {
		t.Fatalf("Mitgliedschaften = %d, erwartet 1 — eine Kündigung legt keine neue an", len(m.Mitgliedschaften))
	}

	ms := m.Mitgliedschaften[0]
	if ms.Kuendigungsdatum == nil || !ms.Kuendigungsdatum.Equal(gekuendigt) {
		t.Errorf("Kuendigungsdatum = %v, erwartet %v", ms.Kuendigungsdatum, gekuendigt)
	}
	if ms.Austritt == nil || !ms.Austritt.Equal(austritt) {
		t.Errorf("Austritt = %v, erwartet %v", ms.Austritt, austritt)
	}
}

// Eine Kündigung liegt vor, der Termin ist noch offen. Daraus darf kein
// Austrittsdatum errechnet werden: Fristen haben Sonderfälle, und ein Datum,
// das man überschreiben muss, ist lästiger als ein leeres Feld.
func TestSetKuendigung_OhneAustrittLeitetKeinenAustrittAb(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Nina", "Klein")
	gekuendigt := datum(t, "2026-03-15")

	if err := svc.SetKuendigung(id, service.Kuendigung{Datum: &gekuendigt}); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	laufend := m.LaufendeMitgliedschaft()
	if laufend == nil {
		t.Fatal("LaufendeMitgliedschaft = nil, erwartet den weiterlaufenden Zeitraum — ohne Austrittsdatum endet nichts")
	}
	if laufend.Kuendigungsdatum == nil || !laufend.Kuendigungsdatum.Equal(gekuendigt) {
		t.Errorf("Kuendigungsdatum = %v, erwartet %v", laufend.Kuendigungsdatum, gekuendigt)
	}
	if laufend.Austritt != nil {
		t.Errorf("Austritt = %v, erwartet nil — das Austrittsdatum wird nicht berechnet", laufend.Austritt)
	}

	// Und das Mitglied bleibt in der Standardansicht: ausgetreten ist es nicht.
	liste, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(liste) != 1 || liste[0].MitgliedID != id {
		t.Fatalf("List = %+v, erwartet das gekündigte, aber noch nicht ausgetretene Mitglied %d", liste, id)
	}
	if liste[0].Kuendigungsdatum == nil || !liste[0].Kuendigungsdatum.Equal(gekuendigt) {
		t.Errorf("Kuendigungsdatum der Zeile = %v, erwartet %v — sonst ist die Kündigung in der Liste unsichtbar",
			liste[0].Kuendigungsdatum, gekuendigt)
	}
}

// Der Termin wird nachgetragen, sobald er feststeht. Das erfasste
// Kündigungsdatum überlebt das: es hält fest, wann gekündigt wurde, und daran
// ändert der später vereinbarte Termin nichts.
func TestSetKuendigung_TraegtDenAustrittSpaeterNach(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Nina", "Klein")
	gekuendigt := datum(t, "2026-03-15")

	if err := svc.SetKuendigung(id, service.Kuendigung{Datum: &gekuendigt}); err != nil {
		t.Fatalf("erstes SetKuendigung: %v", err)
	}

	austritt := datum(t, "2026-06-30")
	if err := svc.SetKuendigung(id, service.Kuendigung{Datum: &gekuendigt, Austritt: &austritt}); err != nil {
		t.Fatalf("zweites SetKuendigung: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(m.Mitgliedschaften) != 1 {
		t.Fatalf("Mitgliedschaften = %d, erwartet 1", len(m.Mitgliedschaften))
	}

	ms := m.Mitgliedschaften[0]
	if ms.Kuendigungsdatum == nil || !ms.Kuendigungsdatum.Equal(gekuendigt) {
		t.Errorf("Kuendigungsdatum = %v, erwartet unverändert %v", ms.Kuendigungsdatum, gekuendigt)
	}
	if ms.Austritt == nil || !ms.Austritt.Equal(austritt) {
		t.Errorf("Austritt = %v, erwartet %v", ms.Austritt, austritt)
	}
}

// Erfasst wird die Kündigung als Ganzes: was der Aufrufer nicht mitschickt,
// steht danach auch nicht mehr da. Das ist der Weg, eine irrtümlich eingetragene
// Kündigungserklärung wieder wegzunehmen — und der Grund, warum das Formular den
// erfassten Wert immer mitschickt.
func TestSetKuendigung_ErsetztBeideAngabenVollstaendig(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Nina", "Klein")
	gekuendigt := datum(t, "2026-03-15")
	if err := svc.SetKuendigung(id, service.Kuendigung{Datum: &gekuendigt}); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	// Nur der Austritt: das Kündigungsdatum ist damit zurückgenommen.
	austritt := datum(t, "2026-06-30")
	if err := svc.SetKuendigung(id, austrittZum(austritt)); err != nil {
		t.Fatalf("SetKuendigung nur mit Austritt: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	ms := m.Mitgliedschaften[0]
	if ms.Kuendigungsdatum != nil {
		t.Errorf("Kuendigungsdatum = %v, erwartet nil — was nicht mitkommt, steht danach nicht mehr da",
			ms.Kuendigungsdatum)
	}
	if ms.Austritt == nil || !ms.Austritt.Equal(austritt) {
		t.Errorf("Austritt = %v, erwartet %v", ms.Austritt, austritt)
	}
}

// Der Normalfall einer laufenden Kündigungsfrist: der Austritt liegt in der
// Zukunft. Ein Datum abzulehnen, nur weil es noch nicht erreicht ist, hieße,
// genau den Fall nicht erfassen zu können, für den die Frist da ist.
func TestSetKuendigung_AustrittInDerZukunftIstErlaubt(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Nina", "Klein")
	gekuendigt := heuteVersetzt(0)
	austritt := heuteVersetzt(90)

	if err := svc.SetKuendigung(id, service.Kuendigung{Datum: &gekuendigt, Austritt: &austritt}); err != nil {
		t.Fatalf("SetKuendigung mit künftigem Austritt: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if m.Mitgliedschaften[0].Austritt == nil || !m.Mitgliedschaften[0].Austritt.Equal(austritt) {
		t.Errorf("Austritt = %v, erwartet %v", m.Mitgliedschaften[0].Austritt, austritt)
	}
}

// Ohne jedes Datum ist nichts erfasst. Das abzuweisen schützt die bereits
// erfasste Kündigung vor einem versehentlich leer abgeschickten Formular; eine
// Kündigung zurückzunehmen ist kein Fall dieses Tickets.
func TestSetKuendigung_OhneDatumIstEinFehler(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Nina", "Klein")
	gekuendigt := datum(t, "2026-03-15")
	if err := svc.SetKuendigung(id, service.Kuendigung{Datum: &gekuendigt}); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	err := svc.SetKuendigung(id, service.Kuendigung{})

	var validierung *service.ValidierungsFehler
	if !errors.As(err, &validierung) {
		t.Fatalf("SetKuendigung ohne Datum = %v, erwartet einen ValidierungsFehler", err)
	}
	if len(validierung.Meldungen) == 0 {
		t.Error("ValidierungsFehler ohne Meldung — das Formular hätte nichts anzuzeigen")
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if m.Mitgliedschaften[0].Kuendigungsdatum == nil {
		t.Error("Kuendigungsdatum = nil, erwartet das erfasste Datum — die abgelehnte Eingabe darf nichts löschen")
	}
}

// Gekündigt wird vor dem Austritt, nie danach: dazwischen liegt die
// Kündigungsfrist. Die umgekehrte Reihenfolge ist ein Tippfehler.
func TestSetKuendigung_KuendigungsdatumNachDemAustrittIstEinFehler(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Nina", "Klein")
	gekuendigt := datum(t, "2026-07-01")
	austritt := datum(t, "2026-06-30")

	err := svc.SetKuendigung(id, service.Kuendigung{Datum: &gekuendigt, Austritt: &austritt})

	var validierung *service.ValidierungsFehler
	if !errors.As(err, &validierung) {
		t.Fatalf("SetKuendigung = %v, erwartet einen ValidierungsFehler", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if m.LaufendeMitgliedschaft() == nil {
		t.Error("die Mitgliedschaft wurde trotz abgelehnter Eingabe beendet")
	}
}

// Am selben Tag zu kündigen und auszutreten ist erlaubt: eine Kündigungsfrist
// von null Tagen gibt es (Aufhebungsvertrag), ein Zeitraum negativer Länge nicht.
func TestSetKuendigung_KuendigungsdatumAmAustrittstagIstErlaubt(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Nina", "Klein")
	tag := datum(t, "2026-06-30")

	if err := svc.SetKuendigung(id, service.Kuendigung{Datum: &tag, Austritt: &tag}); err != nil {
		t.Fatalf("SetKuendigung am Austrittstag: %v", err)
	}
}

// Ruhend ist das Merkmal, das sich aus keinem Datum ergibt: es wird gesetzt und
// wieder zurückgenommen (CONTEXT.md → Ruhend).
func TestSetRuhend_SchaltetHinUndZurueck(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Nina", "Klein")

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if m.Mitgliedschaften[0].Ruhend {
		t.Error("Ruhend = true bei einem frisch angelegten Mitglied, erwartet false")
	}

	if err := svc.SetRuhend(id, true); err != nil {
		t.Fatalf("SetRuhend(true): %v", err)
	}
	if m, err = svc.Get(id); err != nil {
		t.Fatalf("Get nach SetRuhend(true): %v", err)
	}
	if !m.Mitgliedschaften[0].Ruhend {
		t.Error("Ruhend = false, erwartet true")
	}

	// Ruhend ist kein Ende: die Mitgliedschaft läuft weiter, das Mitglied bleibt
	// in der Standardansicht.
	if m.LaufendeMitgliedschaft() == nil {
		t.Error("LaufendeMitgliedschaft = nil, erwartet den weiterlaufenden Zeitraum — ruhend ist kein Austritt")
	}
	liste, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(liste) != 1 || liste[0].MitgliedID != id {
		t.Fatalf("List = %+v, erwartet das ruhende Mitglied %d", liste, id)
	}
	if !liste[0].Ruhend {
		t.Error("Ruhend der Zeile = false, erwartet true — sonst ist das Merkmal in der Liste unsichtbar")
	}

	if err := svc.SetRuhend(id, false); err != nil {
		t.Fatalf("SetRuhend(false): %v", err)
	}
	if m, err = svc.Get(id); err != nil {
		t.Fatalf("Get nach SetRuhend(false): %v", err)
	}
	if m.Mitgliedschaften[0].Ruhend {
		t.Error("Ruhend = true, erwartet false nach dem Zurücknehmen")
	}
}

// Ruhend und Rückstand sind unabhängig. Für ein ruhendes Mitglied wird keine
// Lastschrift losgeschickt, es kann also kein neuer Rückstand entstehen — ein
// bestehender bleibt aber stehen. Ruhend zu schalten ist kein Schuldenerlass.
func TestSetRuhend_LaesstEinenBestehendenRueckstandStehen(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Nina", "Klein")
	rueckstand := service.Rueckstand{Offen: true, Notiz: "Rücklastschrift Mai, angeschrieben am 05.06."}
	if err := svc.SetRueckstand(id, rueckstand); err != nil {
		t.Fatalf("SetRueckstand: %v", err)
	}

	if err := svc.SetRuhend(id, true); err != nil {
		t.Fatalf("SetRuhend(true): %v", err)
	}

	eintrag, err := svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag: %v", err)
	}
	if eintrag.Rueckstand != rueckstand {
		t.Errorf("Rueckstand = %+v, erwartet unverändert %+v — ruhend erlässt keine Schulden",
			eintrag.Rueckstand, rueckstand)
	}

	// Und umgekehrt: das Zurücknehmen rührt den Rückstand ebenso wenig an.
	if err := svc.SetRuhend(id, false); err != nil {
		t.Fatalf("SetRuhend(false): %v", err)
	}
	if eintrag, err = svc.Eintrag(id); err != nil {
		t.Fatalf("Eintrag nach SetRuhend(false): %v", err)
	}
	if eintrag.Rueckstand != rueckstand {
		t.Errorf("Rueckstand = %+v, erwartet unverändert %+v", eintrag.Rueckstand, rueckstand)
	}
}

// Ruhend schalten lässt sich nur, was läuft: ein ausgetretenes Mitglied hat keinen
// Zeitraum, für den ein Einzug ausgesetzt werden könnte.
func TestSetRuhend_AufAusgetretenemMitgliedIstEinFehler(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Nina", "Klein")
	if err := svc.SetKuendigung(id, austrittZum(datum(t, "2026-06-30"))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	if err := svc.SetRuhend(id, true); !errors.Is(err, service.ErrNichtAktiv) {
		t.Fatalf("SetRuhend nach dem Austritt = %v, erwartet ErrNichtAktiv", err)
	}
}

// Der neue Zeitraum beginnt unbelastet: weder mit der Kündigung des alten noch
// mit dessen Ruhend-Kennzeichen. Beides gehörte zu einem Zeitraum, der vorbei ist.
func TestRejoin_BeginntOhneKuendigungUndNichtRuhend(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegen(t, svc, "Nina", "Klein")
	if err := svc.SetRuhend(id, true); err != nil {
		t.Fatalf("SetRuhend: %v", err)
	}

	gekuendigt := datum(t, "2026-03-15")
	austritt := datum(t, "2026-06-30")
	if err := svc.SetKuendigung(id, service.Kuendigung{Datum: &gekuendigt, Austritt: &austritt}); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	if err := svc.Rejoin(id, datum(t, "2026-09-01")); err != nil {
		t.Fatalf("Rejoin: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	laufend := m.LaufendeMitgliedschaft()
	if laufend == nil {
		t.Fatal("LaufendeMitgliedschaft = nil, erwartet den neuen Zeitraum")
	}
	if laufend.Kuendigungsdatum != nil {
		t.Errorf("Kuendigungsdatum = %v, erwartet nil — die Kündigung galt dem alten Zeitraum", laufend.Kuendigungsdatum)
	}
	if laufend.Ruhend {
		t.Error("Ruhend = true, erwartet false — wer wieder eintritt, trainiert wieder")
	}

	// Der alte Zeitraum behält seine Angaben: er ist die Historie.
	alt := m.Mitgliedschaften[0]
	if alt.Kuendigungsdatum == nil || !alt.Kuendigungsdatum.Equal(gekuendigt) {
		t.Errorf("Kuendigungsdatum des alten Zeitraums = %v, erwartet unverändert %v", alt.Kuendigungsdatum, gekuendigt)
	}
	if !alt.Ruhend {
		t.Error("Ruhend des alten Zeitraums = false, erwartet unverändert true")
	}
}

// Zu einer unbekannten ID lässt sich nichts ruhend schalten. Das Gegenstück für die
// Kündigung steht bei den übrigen Lebenszyklus-Fehlern in lebenszyklus_test.go.
func TestSetRuhend_UnbekannteIDMeldetNichtGefunden(t *testing.T) {
	svc := neuerService(t)

	if err := svc.SetRuhend(4711, true); !errors.Is(err, service.ErrNichtGefunden) {
		t.Errorf("SetRuhend(4711) = %v, erwartet ErrNichtGefunden", err)
	}
}
