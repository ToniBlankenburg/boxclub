package service_test

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/ToniBlankenburg/boxclub/service"
)

func tagVonHeute(tage int) time.Time {
	return time.Now().AddDate(0, 0, tage)
}

func stilllegungen(t *testing.T, svc *service.MemberService, id int64) []service.Stilllegung {
	t.Helper()

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	return m.Mitgliedschaften[len(m.Mitgliedschaften)-1].Stilllegungen
}

func fehlerSchluessel(t *testing.T, err error) []string {
	t.Helper()

	var v *service.ValidierungsFehler
	if !asValidierung(err, &v) {
		t.Fatalf("err = %v, erwartet ValidierungsFehler", err)
	}

	return schluessel(v.Meldungen)
}

func TestCreateStilllegung_PflichtfelderUndReihenfolge(t *testing.T) {
	svc := neuerService(t)
	id := mitgliedAnlegen(t, svc, "Nina", "Klein")

	_, err := svc.CreateStilllegung(id, service.Stilllegungsangabe{})
	if got := fehlerSchluessel(t, err); !slices.Equal(got, []string{
		"validierung.stilllegung.von_leer", "validierung.stilllegung.bis_leer"}) {
		t.Errorf("Schlüssel = %v", got)
	}

	_, err = svc.CreateStilllegung(id, service.Stilllegungsangabe{Von: tagVonHeute(5), Bis: tagVonHeute(4)})
	if got := fehlerSchluessel(t, err); !slices.Equal(got, []string{"validierung.stilllegung.bis_vor_von"}) {
		t.Errorf("Schlüssel = %v", got)
	}

	// Ein Tag lang (von = bis) ist erlaubt.
	if _, err := svc.CreateStilllegung(id, service.Stilllegungsangabe{Von: tagVonHeute(40), Bis: tagVonHeute(40)}); err != nil {
		t.Errorf("eintägige Stilllegung: %v", err)
	}
}

func TestCreateStilllegung_KeineNeueSolangeEineLaeuftUndKeineUeberlappung(t *testing.T) {
	svc := neuerService(t)
	id := mitgliedAnlegen(t, svc, "Nina", "Klein")

	if _, err := svc.CreateStilllegung(id, stilllegungHeute()); err != nil {
		t.Fatalf("erste: %v", err)
	}

	// Eine zweite, die gar nicht überlappt, ist trotzdem gesperrt, solange eine läuft.
	_, err := svc.CreateStilllegung(id, service.Stilllegungsangabe{Von: tagVonHeute(200), Bis: tagVonHeute(210)})
	if got := fehlerSchluessel(t, err); !slices.Equal(got, []string{"validierung.stilllegung.laeuft_bereits"}) {
		t.Errorf("Schlüssel = %v", got)
	}

	if err := wiederAktiv(svc, id); err != nil {
		t.Fatalf("wiederAktiv: %v", err)
	}

	// Danach überlappt nichts mehr mit der beendeten ...
	if _, err := svc.CreateStilllegung(id, service.Stilllegungsangabe{Von: tagVonHeute(200), Bis: tagVonHeute(210)}); err != nil {
		t.Fatalf("zweite: %v", err)
	}

	// ... aber eine dritte, die die zweite berührt (Randtag einschließend), schon.
	_, err = svc.CreateStilllegung(id, service.Stilllegungsangabe{Von: tagVonHeute(210), Bis: tagVonHeute(220)})
	if got := fehlerSchluessel(t, err); !slices.Equal(got, []string{"validierung.stilllegung.ueberlappt"}) {
		t.Errorf("Schlüssel = %v", got)
	}
}

func TestStilllegung_BisHinterAustrittWirdSoGespeichert(t *testing.T) {
	svc := neuerService(t)
	id := mitgliedAnlegen(t, svc, "Nina", "Klein")

	austritt := tagVonHeute(30)
	if err := svc.SetKuendigung(id, service.Kuendigung{Austritt: &austritt}); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	// Gekündigt zu sein verbietet die Stilllegung nicht, und bis darf hinter dem Austritt liegen.
	bis := tagVonHeute(90)
	if _, err := svc.CreateStilllegung(id, service.Stilllegungsangabe{Von: tagVonHeute(10), Bis: bis}); err != nil {
		t.Fatalf("CreateStilllegung: %v", err)
	}
	if got := stilllegungen(t, svc, id); len(got) != 1 || !got[0].Bis.Equal(time.Date(bis.Year(), bis.Month(), bis.Day(), 0, 0, 0, 0, time.UTC)) {
		t.Errorf("Stilllegungen = %v", got)
	}
}

func TestStilllegung_AendernLoeschenUndGrund(t *testing.T) {
	svc := neuerService(t)
	id := mitgliedAnlegen(t, svc, "Nina", "Klein")

	sid, err := svc.CreateStilllegung(id, service.Stilllegungsangabe{
		Von: tagVonHeute(10), Bis: tagVonHeute(20), Grund: "  Knie  "})
	if err != nil {
		t.Fatalf("CreateStilllegung: %v", err)
	}
	if got := stilllegungen(t, svc, id); got[0].Grund != "Knie" {
		t.Errorf("Grund = %q, erwartet getrimmt", got[0].Grund)
	}

	// Die eigene Stilllegung zählt bei der Überlappungsprüfung nicht.
	if err := svc.UpdateStilllegung(sid, service.Stilllegungsangabe{Von: tagVonHeute(12), Bis: tagVonHeute(25)}); err != nil {
		t.Fatalf("UpdateStilllegung: %v", err)
	}
	if got := stilllegungen(t, svc, id); got[0].Grund != "" {
		t.Errorf("Grund = %q, erwartet leer", got[0].Grund)
	}

	if err := svc.DeleteStilllegung(sid); err != nil {
		t.Fatalf("DeleteStilllegung: %v", err)
	}
	if got := stilllegungen(t, svc, id); len(got) != 0 {
		t.Errorf("Stilllegungen = %v, erwartet keine", got)
	}
	if err := svc.DeleteStilllegung(sid); !isNichtGefunden(err) {
		t.Errorf("zweites Löschen = %v, erwartet ErrNichtGefunden", err)
	}
}

func TestBeendeStilllegung_LaufendeEndetGesternBevorstehendeFaelltWeg(t *testing.T) {
	svc := neuerService(t)
	id := mitgliedAnlegen(t, svc, "Nina", "Klein")

	laufend, err := svc.CreateStilllegung(id, stilllegungHeute())
	if err != nil {
		t.Fatalf("CreateStilllegung: %v", err)
	}
	if err := svc.BeendeStilllegung(laufend); err != nil {
		t.Fatalf("BeendeStilllegung: %v", err)
	}
	// Idempotent: die schon beendete bleibt.
	if err := svc.BeendeStilllegung(laufend); err != nil {
		t.Fatalf("zweites Beenden: %v", err)
	}

	eintrag, err := svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag: %v", err)
	}
	if eintrag.Stillgelegt {
		t.Error("Stillgelegt = true, erwartet false direkt nach „wieder aktiv“")
	}
	if got := stilllegungen(t, svc, id); len(got) != 1 {
		t.Fatalf("Stilllegungen = %v, erwartet die beendete als Historie", got)
	}

	bevorstehend, err := svc.CreateStilllegung(id, service.Stilllegungsangabe{Von: tagVonHeute(30), Bis: tagVonHeute(40)})
	if err != nil {
		t.Fatalf("bevorstehende: %v", err)
	}
	if err := svc.BeendeStilllegung(bevorstehend); err != nil {
		t.Fatalf("BeendeStilllegung: %v", err)
	}
	if got := stilllegungen(t, svc, id); len(got) != 1 {
		t.Errorf("Stilllegungen = %v, erwartet die bevorstehende gelöscht", got)
	}
}

func TestStilllegung_KuendigungInStilllegungGesperrt(t *testing.T) {
	svc := neuerService(t)
	id := mitgliedAnlegen(t, svc, "Nina", "Klein")

	if _, err := svc.CreateStilllegung(id, stilllegungHeute()); err != nil {
		t.Fatalf("CreateStilllegung: %v", err)
	}

	austritt := tagVonHeute(90)
	err := svc.SetKuendigung(id, service.Kuendigung{Austritt: &austritt})
	if got := fehlerSchluessel(t, err); !slices.Equal(got, []string{"validierung.kuendigung.stillgelegt"}) {
		t.Errorf("Schlüssel = %v", got)
	}

	if err := wiederAktiv(svc, id); err != nil {
		t.Fatalf("wiederAktiv: %v", err)
	}
	if err := svc.SetKuendigung(id, service.Kuendigung{Austritt: &austritt}); err != nil {
		t.Errorf("SetKuendigung nach „wieder aktiv“: %v", err)
	}
}

func TestStilllegung_StillgelegtErstAbBeginn(t *testing.T) {
	svc := neuerService(t)
	id := mitgliedAnlegen(t, svc, "Nina", "Klein")

	// Beginnt morgen: heute noch nicht stillgelegt, die Teilnehmerliste führt sie noch.
	if _, err := svc.CreateStilllegung(id, service.Stilllegungsangabe{Von: tagVonHeute(1), Bis: tagVonHeute(30)}); err != nil {
		t.Fatalf("CreateStilllegung: %v", err)
	}
	eintrag, err := svc.Eintrag(id)
	if err != nil {
		t.Fatalf("Eintrag: %v", err)
	}
	if eintrag.Stillgelegt {
		t.Error("Stillgelegt = true, erwartet false vor dem Beginn")
	}
}

func TestMindestlaufzeit_EndeMitStilllegungsmonaten(t *testing.T) {
	svc := neuerService(t)

	id, err := svc.Create(service.NeuesMitglied{
		Vorname: "Nina", Nachname: "Klein", BeitragCents: beitragImTest,
		Eintritt: datum(t, "2026-01-31"), MindestlaufzeitMonate: 12,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	ms := m.Mitgliedschaften[0]
	if ms.MindestlaufzeitMonate != 12 {
		t.Fatalf("MindestlaufzeitMonate = %d", ms.MindestlaufzeitMonate)
	}
	if ende := ms.MindestlaufzeitEnde(); ende == nil || !ende.Equal(datum(t, "2027-01-31")) {
		t.Errorf("Ende = %v, erwartet 2027-01-31", ende)
	}

	// 20.03. bis 05.04. macht März und April zu Stilllegungsmonaten: zwei Monate dazu.
	// Eine zweite im April zählt ihn nicht doppelt.
	for _, z := range [][2]string{{"2026-03-20", "2026-04-05"}, {"2026-04-20", "2026-04-25"}} {
		if _, err := svc.CreateStilllegung(id, service.Stilllegungsangabe{Von: datum(t, z[0]), Bis: datum(t, z[1])}); err != nil {
			t.Fatalf("CreateStilllegung: %v", err)
		}
	}
	m, _ = svc.Get(id)
	ms = m.Mitgliedschaften[0]
	if ms.Stilllegungsmonate() != 2 {
		t.Errorf("Stilllegungsmonate = %d, erwartet 2", ms.Stilllegungsmonate())
	}
	// 31.01. + 14 Monate = 31.03.2027
	if ende := ms.MindestlaufzeitEnde(); ende == nil || !ende.Equal(datum(t, "2027-03-31")) {
		t.Errorf("Ende = %v, erwartet 2027-03-31", ende)
	}

	// Ändern und Aufheben über den Patch.
	null := 0
	if err := svc.Update(id, service.MitgliedPatch{MindestlaufzeitMonate: &null}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	m, _ = svc.Get(id)
	if m.Mitgliedschaften[0].MindestlaufzeitEnde() != nil {
		t.Error("Ende ≠ nil, erwartet nil ohne Mindestlaufzeit")
	}
	negativ := -1
	err = svc.Update(id, service.MitgliedPatch{MindestlaufzeitMonate: &negativ})
	if got := fehlerSchluessel(t, err); !slices.Equal(got, []string{"validierung.mitglied.mindestlaufzeit_negativ"}) {
		t.Errorf("Schlüssel = %v", got)
	}
}

func TestRejoin_BeginntOhneMindestlaufzeit(t *testing.T) {
	svc := neuerService(t)

	id, err := svc.Create(service.NeuesMitglied{
		Vorname: "Nina", Nachname: "Klein", BeitragCents: beitragImTest,
		Eintritt: datum(t, "2025-01-05"), MindestlaufzeitMonate: 6,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	austritt := datum(t, "2025-09-30")
	if err := svc.SetKuendigung(id, service.Kuendigung{Austritt: &austritt}); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}
	if err := svc.Rejoin(id, datum(t, "2026-02-01")); err != nil {
		t.Fatalf("Rejoin: %v", err)
	}

	m, _ := svc.Get(id)
	if m.Mitgliedschaften[0].MindestlaufzeitMonate != 6 || m.Mitgliedschaften[1].MindestlaufzeitMonate != 0 {
		t.Errorf("Mindestlaufzeiten = %d / %d, erwartet 6 / 0",
			m.Mitgliedschaften[0].MindestlaufzeitMonate, m.Mitgliedschaften[1].MindestlaufzeitMonate)
	}
}

// Ein Monat zählt voll, sobald irgendein Tag darin stillgelegt ist: wer erst ab
// dem Monatsletzten stillgelegt ist, fehlt im Monatssoll, zählt aber heute
// noch nicht als stillgelegt.
func TestMonatsuebersicht_StilllegungsmonatLaesstBeitragAusObwohlHeuteNochNichtStillgelegt(t *testing.T) {
	svc := neuerService(t)
	id := mitgliedAnlegen(t, svc, "Nina", "Klein")

	jetzt := time.Now()
	letzter := time.Date(jetzt.Year(), jetzt.Month()+1, 0, 0, 0, 0, 0, time.UTC)
	if letzter.Format("2006-01-02") == jetzt.Format("2006-01-02") {
		t.Skip("heute ist der Monatsletzte: der Fall lässt sich nicht von „heute stillgelegt“ trennen")
	}
	if _, err := svc.CreateStilllegung(id, service.Stilllegungsangabe{Von: letzter, Bis: letzter.AddDate(0, 0, 5)}); err != nil {
		t.Fatalf("CreateStilllegung: %v", err)
	}

	u, err := svc.Monatsuebersicht()
	if err != nil {
		t.Fatalf("Monatsuebersicht: %v", err)
	}
	if u.MonatssollCents != 0 || u.StillgelegtCents != beitragImTest {
		t.Errorf("Soll/Stillgelegt = %d/%d, erwartet 0/%d", u.MonatssollCents, u.StillgelegtCents, beitragImTest)
	}
	if u.Mitgliederzahlen.Stillgelegt != 0 || u.Mitgliederzahlen.Aktiv != 1 {
		t.Errorf("Mitgliederzahlen = %+v, erwartet heute noch aktiv", u.Mitgliederzahlen)
	}
}

// Eine Stilllegung im nächsten Monat berührt den laufenden nicht.
func TestMonatsuebersicht_StilllegungNaechsterMonatLaesstBeitragStehen(t *testing.T) {
	svc := neuerService(t)
	id := mitgliedAnlegen(t, svc, "Nina", "Klein")

	jetzt := time.Now()
	naechster := time.Date(jetzt.Year(), jetzt.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	if _, err := svc.CreateStilllegung(id, service.Stilllegungsangabe{Von: naechster, Bis: naechster.AddDate(0, 0, 10)}); err != nil {
		t.Fatalf("CreateStilllegung: %v", err)
	}

	u, err := svc.Monatsuebersicht()
	if err != nil {
		t.Fatalf("Monatsuebersicht: %v", err)
	}
	if u.MonatssollCents != beitragImTest {
		t.Errorf("Monatssoll = %d, erwartet %d", u.MonatssollCents, beitragImTest)
	}
}

func asValidierung(err error, ziel **service.ValidierungsFehler) bool {
	return errors.As(err, ziel)
}

func isNichtGefunden(err error) bool {
	return errors.Is(err, service.ErrNichtGefunden)
}
