package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ToniBlankenburg/boxclub/service"
)

// Der abgeleitete Status am MemberService-Seam. Beobachtet wird ausschließlich
// über die Service-API: dass nichts gespeichert wird, zeigt sich daran, dass
// keine Operation ihn setzt und er sich trotzdem mit den Datumsfeldern ändert.
//
// Alle Fixtures liegen relativ zu heute (heuteVersetzt), denn genau gegen
// diesen Tag wird abgelesen — feste Kalendertage wären ab morgen etwas anderes.

// mitgliedAnlegenZum legt ein Mitglied mit einem bestimmten Eintritt an. Für den
// Status ist der Eintritt die erste der drei Angaben, aus denen abgelesen wird.
func mitgliedAnlegenZum(t *testing.T, svc *service.MemberService, nachname string, eintritt time.Time) int64 {
	t.Helper()

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:      "Test",
		Nachname:     nachname,
		BeitragCents: beitragImTest,
		Eintritt:     eintritt,
	})
	if err != nil {
		t.Fatalf("Create(%s): %v", nachname, err)
	}

	return id
}

// statusVon liest den Status aus der Listenzeile — dieselbe Angabe, die die
// Oberfläche zeigt.
func statusVon(t *testing.T, svc *service.MemberService, id int64) service.Status {
	t.Helper()

	eintrag, err := svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag(%d): %v", id, err)
	}

	return eintrag.Status()
}

// Neu: das Mitglied ist erfasst, der Verein erwartet es — angefangen hat es noch
// nicht.
func TestStatus_NeuSolangeDerEintrittBevorsteht(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegenZum(t, svc, "Morgen", heuteVersetzt(1))

	if got := statusVon(t, svc, id); got != service.StatusNeu {
		t.Errorf("Status = %v, erwartet Neu — der Eintritt liegt in der Zukunft", got.Bezeichnung())
	}
}

// Der Eintrittstag selbst zählt schon: "Eintritt erreicht" schließt den Tag
// ein, an dem er erreicht wird.
func TestStatus_AktivAbDemEintrittstagSelbst(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegenZum(t, svc, "Heute", heuteVersetzt(0))

	if got := statusVon(t, svc, id); got != service.StatusAktiv {
		t.Errorf("Status = %v, erwartet Aktiv — der Eintritt ist heute erreicht", got.Bezeichnung())
	}
}

func TestStatus_AktivNachDemEintritt(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegenZum(t, svc, "Gestern", heuteVersetzt(-30))

	if got := statusVon(t, svc, id); got != service.StatusAktiv {
		t.Errorf("Status = %v, erwartet Aktiv", got.Bezeichnung())
	}
}

// Der Kern des Tickets: ein erfasster, aber noch nicht erreichter Austritt ist
// die Kündigungsfrist — das Mitglied trainiert und zahlt weiter.
func TestStatus_InKuendigungsfristSolangeDerAustrittBevorsteht(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegenZum(t, svc, "Frist", heuteVersetzt(-100))

	kuendigungsdatum := heuteVersetzt(-3)
	austritt := heuteVersetzt(30)
	if err := svc.SetKuendigung(id, service.Kuendigung{Datum: &kuendigungsdatum, Austritt: &austritt}); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	if got := statusVon(t, svc, id); got != service.StatusInKuendigungsfrist {
		t.Errorf("Status = %v, erwartet In Kündigungsfrist", got.Bezeichnung())
	}
}

// Eine Kündigung, deren Termin noch nicht feststeht, ist ebenfalls die
// Kündigungsfrist: erklärt ist sie, beendet ist noch nichts.
func TestStatus_InKuendigungsfristAuchOhneAustrittstermin(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegenZum(t, svc, "Offen", heuteVersetzt(-100))

	kuendigungsdatum := heuteVersetzt(0)
	if err := svc.SetKuendigung(id, service.Kuendigung{Datum: &kuendigungsdatum}); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	if got := statusVon(t, svc, id); got != service.StatusInKuendigungsfrist {
		t.Errorf("Status = %v, erwartet In Kündigungsfrist — erklärt ist die Kündigung, offen nur der Termin", got.Bezeichnung())
	}
}

// Die Gegenprobe zum Fall davor: ein erfasster Austritt, dessen Tag der
// Erklärung niemand notiert hat — der Normalfall der Altbestände. Auch das ist
// die Kündigungsfrist und nicht "Aktiv", denn das Ende steht fest.
func TestStatus_InKuendigungsfristAuchOhneKuendigungsdatum(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegenZum(t, svc, "Altbestand", heuteVersetzt(-100))

	if err := svc.SetKuendigung(id, austrittZum(heuteVersetzt(30))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	if got := statusVon(t, svc, id); got != service.StatusInKuendigungsfrist {
		t.Errorf("Status = %v, erwartet In Kündigungsfrist — der Austritt steht fest, nur die Erklärung ist nicht notiert", got.Bezeichnung())
	}
}

// Der Austrittstag selbst zählt schon: "Austritt erreicht" schließt ihn ein, wie
// der Eintrittstag beim Aktiv-Werden.
func TestStatus_AusgetretenAmAustrittstagSelbst(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegenZum(t, svc, "Stichtag", heuteVersetzt(-100))

	if err := svc.SetKuendigung(id, austrittZum(heuteVersetzt(0))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	if got := statusVon(t, svc, id); got != service.StatusAusgetreten {
		t.Errorf("Status = %v, erwartet Ausgetreten — der Austritt ist heute erreicht", got.Bezeichnung())
	}
}

func TestStatus_AusgetretenNachDemAustritt(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegenZum(t, svc, "Ehemalig", heuteVersetzt(-100))

	if err := svc.SetKuendigung(id, austrittZum(heuteVersetzt(-1))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	if got := statusVon(t, svc, id); got != service.StatusAusgetreten {
		t.Errorf("Status = %v, erwartet Ausgetreten", got.Bezeichnung())
	}
}

// Ruhend steht neben dem Status und nicht in ihm (CONTEXT.md → Ruhend): ein
// ruhendes Mitglied ist ein aktives, von dem gerade nichts eingezogen wird.
func TestStatus_RuhendTrittNebenDenStatusUndNichtAnSeineStelle(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegenZum(t, svc, "Pause", heuteVersetzt(-100))

	if err := svc.SetRuhend(id, true); err != nil {
		t.Fatalf("SetRuhend: %v", err)
	}

	eintrag, err := svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag: %v", err)
	}
	if eintrag.Status() != service.StatusAktiv {
		t.Errorf("Status = %v, erwartet Aktiv — ruhend ist kein Lebenszyklus-Zustand", eintrag.Status().Bezeichnung())
	}
	if !eintrag.Ruhend {
		t.Error("Ruhend = false, erwartet true — das Merkmal muss zusätzlich ablesbar sein")
	}
}

// Der einzige Fall, in dem sich zwei Regeln überschneiden: gekündigt, bevor der
// Eintritt überhaupt erreicht ist. Kuendigung.pruefen lässt das ausdrücklich zu
// (wer zurücktritt, bevor seine Mitgliedschaft beginnt), und dann gilt der
// spätere Zustand — dass jemand geht, ist die Auskunft, auf die es ankommt.
func TestStatus_KuendigungVorDemEintrittSchlaegtNeu(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegenZum(t, svc, "Ruecktritt", heuteVersetzt(14))

	kuendigungsdatum := heuteVersetzt(0)
	austritt := heuteVersetzt(14)
	if err := svc.SetKuendigung(id, service.Kuendigung{Datum: &kuendigungsdatum, Austritt: &austritt}); err != nil {
		t.Fatalf("SetKuendigung vor dem Eintritt: %v", err)
	}

	if got := statusVon(t, svc, id); got != service.StatusInKuendigungsfrist {
		t.Errorf("Status = %v, erwartet In Kündigungsfrist und nicht Neu", got.Bezeichnung())
	}
}

// Der Status wird nirgends gespeichert: dasselbe Mitglied, dieselben Spalten,
// und trotzdem ein anderer Zustand, sobald der Austritt erreicht ist. Beide
// Zeiträume hier tragen ihren eigenen Status.
func TestStatus_MehrereMitgliedschaftenNacheinanderHabenJeIhrenEigenen(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegenZum(t, svc, "Zweimal", heuteVersetzt(-400))

	if err := svc.SetKuendigung(id, austrittZum(heuteVersetzt(-200))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}
	if err := svc.Rejoin(id, heuteVersetzt(-100)); err != nil {
		t.Fatalf("Rejoin: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(m.Mitgliedschaften) != 2 {
		t.Fatalf("Mitgliedschaften = %d, erwartet 2", len(m.Mitgliedschaften))
	}
	if got := m.Mitgliedschaften[0].Status(); got != service.StatusAusgetreten {
		t.Errorf("Status der ersten Mitgliedschaft = %v, erwartet Ausgetreten", got.Bezeichnung())
	}
	if got := m.Mitgliedschaften[1].Status(); got != service.StatusAktiv {
		t.Errorf("Status der zweiten Mitgliedschaft = %v, erwartet Aktiv", got.Bezeichnung())
	}

	// Die Zeile zeigt den maßgeblichen Zeitraum — den laufenden, nicht den
	// beendeten davor.
	if got := statusVon(t, svc, id); got != service.StatusAktiv {
		t.Errorf("Status der Zeile = %v, erwartet Aktiv", got.Bezeichnung())
	}
}

// Die schwierige Reihenfolge: ein ausgetretener Zeitraum und daneben ein
// laufender, der bereits gekündigt ist — beide tragen dann ein Austrittsdatum.
// Die Zeile muss trotzdem den laufenden zeigen, sonst stünde ein
// Wiedereingetretener als Ehemaliger in der Liste.
func TestStatus_KuendigungsfristNachEinemFruherenAustritt(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegenZum(t, svc, "Zurueck", heuteVersetzt(-400))

	if err := svc.SetKuendigung(id, austrittZum(heuteVersetzt(-200))); err != nil {
		t.Fatalf("erste SetKuendigung: %v", err)
	}
	if err := svc.Rejoin(id, heuteVersetzt(-100)); err != nil {
		t.Fatalf("Rejoin: %v", err)
	}
	if err := svc.SetKuendigung(id, austrittZum(heuteVersetzt(30))); err != nil {
		t.Fatalf("zweite SetKuendigung: %v", err)
	}

	eintrag, err := svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag: %v", err)
	}
	if got := eintrag.Status(); got != service.StatusInKuendigungsfrist {
		t.Errorf("Status = %v, erwartet In Kündigungsfrist aus dem laufenden Zeitraum", got.Bezeichnung())
	}
	if !eintrag.Eintritt.Equal(heuteVersetzt(-100)) {
		t.Errorf("Eintritt = %v, erwartet den des laufenden Zeitraums %v",
			eintrag.Eintritt, heuteVersetzt(-100))
	}

	// Und die Zeile bleibt in der Standardansicht: der zweite Zeitraum läuft.
	liste, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(liste) != 1 || liste[0].MitgliedID != id {
		t.Fatalf("List = %+v, erwartet Mitglied %d", liste, id)
	}

	// Geändert wird derselbe Zeitraum, den die Liste zeigt: ein Beitrag, der
	// beim alten landete, wäre für den Nutzer spurlos verschwunden.
	neuerBeitrag := int64(7700)
	if err := svc.Update(id, service.MitgliedPatch{BeitragCents: &neuerBeitrag}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(m.Mitgliedschaften) != 2 {
		t.Fatalf("Mitgliedschaften = %d, erwartet 2", len(m.Mitgliedschaften))
	}
	if m.Mitgliedschaften[1].BeitragCents != neuerBeitrag {
		t.Errorf("Beitrag des laufenden Zeitraums = %d, erwartet %d",
			m.Mitgliedschaften[1].BeitragCents, neuerBeitrag)
	}
	if m.Mitgliedschaften[0].BeitragCents != beitragImTest {
		t.Errorf("Beitrag des beendeten Zeitraums = %d, erwartet unverändert %d",
			m.Mitgliedschaften[0].BeitragCents, beitragImTest)
	}
}

// "Aktiv" hat seine Definition geändert: nicht mehr "kein Austrittsdatum",
// sondern "Austritt nicht erreicht". Damit bleibt die Kündigungsfrist in der
// Standardansicht — das Mitglied trainiert dort ja weiter.
func TestList_BehaeltMitgliedInKuendigungsfristInDerStandardansicht(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegenZum(t, svc, "Frist", heuteVersetzt(-100))

	kuendigungsdatum := heuteVersetzt(-3)
	austritt := heuteVersetzt(30)
	if err := svc.SetKuendigung(id, service.Kuendigung{Datum: &kuendigungsdatum, Austritt: &austritt}); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	liste, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(liste) != 1 || liste[0].MitgliedID != id {
		t.Fatalf("List = %+v, erwartet Mitglied %d — in der Kündigungsfrist wird weiter trainiert und gezahlt", liste, id)
	}
	if liste[0].Austritt == nil {
		t.Error("Austritt = nil, erwartet den erfassten Termin — er gehört in die Zeile, auch solange er bevorsteht")
	}
}

// Und die Kehrseite: ab dem Austrittstag verschwindet die Zeile von selbst. Das
// ist der eigentliche Gewinn gegenüber der Excel — niemand zieht eine
// Statusspalte von Hand nach.
func TestList_LaesstAbDemAustrittstagAus(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegenZum(t, svc, "Stichtag", heuteVersetzt(-100))

	if err := svc.SetKuendigung(id, austrittZum(heuteVersetzt(0))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	liste, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(liste) != 0 {
		t.Errorf("List = %+v, erwartet leer — der Austritt ist heute erreicht", liste)
	}

	ehemalige, err := svc.Search("", service.Suchfilter{AuchEhemalige: true})
	if err != nil {
		t.Fatalf("Search (auch Ehemalige): %v", err)
	}
	if len(ehemalige) != 1 || ehemalige[0].MitgliedID != id {
		t.Fatalf("Search (auch Ehemalige) = %+v, erwartet Mitglied %d", ehemalige, id)
	}
}

// Ein noch nicht eingetretenes Mitglied steht in der Standardansicht: erfasst
// ist es, und der Verein will es sehen, bevor es das erste Mal kommt.
func TestList_ZeigtNeueMitgliederVorIhremEintritt(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegenZum(t, svc, "Morgen", heuteVersetzt(7))

	liste, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(liste) != 1 || liste[0].MitgliedID != id {
		t.Fatalf("List = %+v, erwartet Mitglied %d", liste, id)
	}
}

// Was in der Kündigungsfrist noch läuft, lässt sich auch noch ruhend schalten —
// die Mitgliedschaft ist ja nicht beendet.
func TestSetRuhend_GehtAuchInDerKuendigungsfrist(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegenZum(t, svc, "Frist", heuteVersetzt(-100))

	if err := svc.SetKuendigung(id, austrittZum(heuteVersetzt(30))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}
	if err := svc.SetRuhend(id, true); err != nil {
		t.Fatalf("SetRuhend in der Kündigungsfrist: %v", err)
	}

	eintrag, err := svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag: %v", err)
	}
	if !eintrag.Ruhend {
		t.Error("Ruhend = false, erwartet true")
	}
	if got := eintrag.Status(); got != service.StatusInKuendigungsfrist {
		t.Errorf("Status = %v, erwartet In Kündigungsfrist", got.Bezeichnung())
	}
}

// Aus der Sperre gegen das zweite Kündigen wird während der Frist eine
// Korrekturmöglichkeit: solange der Austritt bevorsteht, läuft der Zeitraum und
// ein Tippfehler im Termin lässt sich richtigstellen.
func TestSetKuendigung_LaesstSichWaehrendDerFristKorrigieren(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegenZum(t, svc, "Tippfehler", heuteVersetzt(-100))

	if err := svc.SetKuendigung(id, austrittZum(heuteVersetzt(30))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	richtig := heuteVersetzt(60)
	if err := svc.SetKuendigung(id, austrittZum(richtig)); err != nil {
		t.Fatalf("SetKuendigung (Korrektur): %v", err)
	}

	eintrag, err := svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag: %v", err)
	}
	if eintrag.Austritt == nil || !eintrag.Austritt.Equal(richtig) {
		t.Errorf("Austritt = %v, erwartet %v", eintrag.Austritt, richtig)
	}
}

// Ist der Austritt dagegen erreicht, läuft nichts mehr — dann ist ein zweiter
// Kündigungsversuch wie bisher ErrNichtAktiv.
func TestSetKuendigung_NachErreichtemAustrittMeldetNichtAktiv(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegenZum(t, svc, "Ehemalig", heuteVersetzt(-100))

	if err := svc.SetKuendigung(id, austrittZum(heuteVersetzt(-1))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	if err := svc.SetKuendigung(id, austrittZum(heuteVersetzt(0))); !errors.Is(err, service.ErrNichtAktiv) {
		t.Errorf("SetKuendigung = %v, erwartet ErrNichtAktiv", err)
	}
}

// Wiedereintreten kann nur, wer draußen ist. In der Kündigungsfrist ist das
// Mitglied noch drin — ein zweiter Zeitraum daneben wäre ein Widerspruch.
func TestRejoin_WaehrendDerKuendigungsfristMeldetBereitsAktiv(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegenZum(t, svc, "Frist", heuteVersetzt(-100))

	if err := svc.SetKuendigung(id, austrittZum(heuteVersetzt(30))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	if err := svc.Rejoin(id, heuteVersetzt(31)); !errors.Is(err, service.ErrBereitsAktiv) {
		t.Errorf("Rejoin = %v, erwartet ErrBereitsAktiv", err)
	}
}

// LaufendeMitgliedschaft zieht mit der neuen Lesart von "aktiv" nach: sie
// liefert den Zeitraum, solange der Austritt bevorsteht, und erst ab dem
// Austrittstag nichts mehr.
func TestLaufendeMitgliedschaft_EndetErstAmAustrittstag(t *testing.T) {
	svc := neuerService(t)

	id := mitgliedAnlegenZum(t, svc, "Frist", heuteVersetzt(-100))

	if err := svc.SetKuendigung(id, austrittZum(heuteVersetzt(30))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if m.LaufendeMitgliedschaft() == nil {
		t.Error("LaufendeMitgliedschaft = nil, erwartet den Zeitraum — der Austritt steht noch bevor")
	}
}

// Die Bezeichnungen stehen im Service, damit Liste und spätere Ansichten
// dieselben Worte benutzen — wie bei Rueckstand und Trainingsfrequenz.
func TestStatus_BezeichnungIstVierwertig(t *testing.T) {
	faelle := map[service.Status]string{
		service.StatusNeu:                "Neu",
		service.StatusAktiv:              "Aktiv",
		service.StatusInKuendigungsfrist: "In Kündigungsfrist",
		service.StatusAusgetreten:        "Ausgetreten",
	}

	for status, erwartet := range faelle {
		if got := status.Bezeichnung(); got != erwartet {
			t.Errorf("Bezeichnung von %d = %q, erwartet %q", int(status), got, erwartet)
		}
	}
}
