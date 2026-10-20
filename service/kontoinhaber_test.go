package service_test

import (
	"errors"
	"testing"

	"github.com/ToniBlankenburg/boxclub/service"
)

// kindMitKontoinhaber legt ein Kind (Handkennzeichen, kein Geburtsdatum) an,
// dessen Eltern zahlen.
func kindMitKontoinhaber(t *testing.T, svc *service.MemberService, vorname, kontoinhaber, iban string) int64 {
	t.Helper()

	id, err := svc.Create(service.NeuesMitglied{
		Vorname: vorname, Nachname: "Berger", KindManuell: true,
		Kontoinhaber: kontoinhaber, IBAN: iban,
		BeitragCents: beitragImTest, Eintritt: heuteVersetzt(-100),
	})
	if err != nil {
		t.Fatalf("Create(%s): %v", vorname, err)
	}

	return id
}

func exportLesen(t *testing.T, svc *service.MemberService) service.MoneyMoneyExport {
	t.Helper()

	export, err := svc.MoneyMoneyExportieren()
	if err != nil {
		t.Fatalf("MoneyMoneyExportieren: %v", err)
	}

	return export
}

func TestKontoinhaber_KindMitAbweichendemKontoinhaberZahltNichtSelbst(t *testing.T) {
	svc := neuerService(t)

	id := kindMitKontoinhaber(t, svc, "Lena", "Petra Berger", "de02 1203 0000 0000 2020 51")

	z := zeileZuMitglied(t, exportLesen(t, svc), id)
	if z.Zahlungspflichtiger != "Petra Berger" || z.IBAN != "DE02120300000000202051" {
		t.Errorf("Zeile = %+v, erwartet Petra Berger mit normalisierter IBAN", z)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if m.Kontoinhaber != "Petra Berger" {
		t.Errorf("Kontoinhaber = %q, erwartet %q", m.Kontoinhaber, "Petra Berger")
	}
}

func TestKontoinhaber_ErwachsenerDarfKeinenAbweichendenHaben(t *testing.T) {
	svc := neuerService(t)

	_, err := svc.Create(service.NeuesMitglied{
		Vorname: "Anna", Nachname: "Berger", Geburtsdatum: zeiger(datum(t, "1990-01-01")),
		Kontoinhaber: "Petra Berger", BeitragCents: beitragImTest, Eintritt: heuteVersetzt(-1),
	})
	var fehler *service.ValidierungsFehler
	if !errors.As(err, &fehler) {
		t.Fatalf("Create: Fehler = %v, erwartet ValidierungsFehler", err)
	}

	id := moneyMoneyMitglied(t, svc, "Bernd", "Cramer", "")
	name := "Petra Cramer"
	err = svc.Update(id, service.MitgliedPatch{Kontoinhaber: &name})
	if !errors.As(err, &fehler) {
		t.Fatalf("Update: Fehler = %v, erwartet ValidierungsFehler", err)
	}
	if m, _ := svc.Get(id); m.Kontoinhaber != "" {
		t.Errorf("Kontoinhaber = %q, erwartet unberührt", m.Kontoinhaber)
	}
}

func TestKontoinhaber_GeschwisterMitGleicherIBANTeilenSichEinenZahler(t *testing.T) {
	svc := neuerService(t)

	a := kindMitKontoinhaber(t, svc, "Lena", "Petra Berger", "DE02 1203 0000 0000 2020 51")
	b := kindMitKontoinhaber(t, svc, "Tim", "Petra Berger", "de02120300000000202051")

	neu := "Petra Weber"
	if err := svc.Update(a, service.MitgliedPatch{Kontoinhaber: &neu}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	export := exportLesen(t, svc)
	if z := zeileZuMitglied(t, export, b); z.Zahlungspflichtiger != "Petra Weber" {
		t.Errorf("Zahlungspflichtiger von Tim = %q, erwartet %q", z.Zahlungspflichtiger, "Petra Weber")
	}
	if len(export.Zeilen) != 2 {
		t.Errorf("Zeilen = %d, erwartet eine je Mitgliedschaft", len(export.Zeilen))
	}
}

func TestKontoinhaber_GleicheIBANMitAnderemNamenZuletztGespeicherterGewinnt(t *testing.T) {
	svc := neuerService(t)

	a := kindMitKontoinhaber(t, svc, "Lena", "Petra Berger", "DE02120300000000202051")
	b := kindMitKontoinhaber(t, svc, "Tim", "P. Berger", "DE02120300000000202051")

	export := exportLesen(t, svc)
	for _, id := range []int64{a, b} {
		if z := zeileZuMitglied(t, export, id); z.Zahlungspflichtiger != "P. Berger" {
			t.Errorf("Zahlungspflichtiger = %q, erwartet %q für beide", z.Zahlungspflichtiger, "P. Berger")
		}
	}
}

func TestKontoinhaber_OhneIBANTeilenSichZweiKinderKeinenZahler(t *testing.T) {
	svc := neuerService(t)

	a := kindMitKontoinhaber(t, svc, "Lena", "Petra Berger", "")
	b := kindMitKontoinhaber(t, svc, "Tim", "Paul Cramer", "")

	neu := "Petra Weber"
	if err := svc.Update(a, service.MitgliedPatch{Kontoinhaber: &neu}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	export := exportLesen(t, svc)
	if z := zeileZuMitglied(t, export, a); z.Zahlungspflichtiger != "Petra Weber" {
		t.Errorf("Zahlungspflichtiger von Lena = %q", z.Zahlungspflichtiger)
	}
	if z := zeileZuMitglied(t, export, b); z.Zahlungspflichtiger != "Paul Cramer" {
		t.Errorf("Zahlungspflichtiger von Tim = %q, erwartet unberührt", z.Zahlungspflichtiger)
	}
}

func TestKontoinhaber_LeerenLaesstDasMitgliedWiederSelbstZahlen(t *testing.T) {
	svc := neuerService(t)

	id := kindMitKontoinhaber(t, svc, "Lena", "Petra Berger", "DE02120300000000202051")

	leer := ""
	if err := svc.Update(id, service.MitgliedPatch{Kontoinhaber: &leer}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	z := zeileZuMitglied(t, exportLesen(t, svc), id)
	if z.Zahlungspflichtiger != "Lena Berger" || z.IBAN != "DE02120300000000202051" {
		t.Errorf("Zeile = %+v, erwartet Lena Berger mit unveränderter IBAN", z)
	}
	if m, _ := svc.Get(id); m.Kontoinhaber != "" {
		t.Errorf("Kontoinhaber = %q, erwartet leer", m.Kontoinhaber)
	}
}

func TestKontoinhaber_NachSetzenZahltDasKindNichtMehrSelbst(t *testing.T) {
	svc := neuerService(t)

	id, err := svc.Create(service.NeuesMitglied{
		Vorname: "Lena", Nachname: "Berger", KindManuell: true, IBAN: "DE02120300000000202051",
		BeitragCents: beitragImTest, Eintritt: heuteVersetzt(-100),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	name, iban := "Petra Berger", "DE89370400440532013000"
	if err := svc.Update(id, service.MitgliedPatch{Kontoinhaber: &name, IBAN: &iban}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	z := zeileZuMitglied(t, exportLesen(t, svc), id)
	if z.Zahlungspflichtiger != "Petra Berger" || z.IBAN != iban {
		t.Errorf("Zeile = %+v", z)
	}
}

// Ein Zahler, auf den nichts mehr zeigt, verschwindet: beobachtbar daran, dass
// ein später neu eingetragener gleicher Kontoinhaber ohne IBAN nichts Altes
// erbt und dass ein Wechsel zurück auf „selbst“ keine Altlast hinterlässt.
func TestKontoinhaber_WechselUndZurueckHinterlaesstKeineAltlast(t *testing.T) {
	svc := neuerService(t)

	id := kindMitKontoinhaber(t, svc, "Lena", "Petra Berger", "DE02120300000000202051")

	leer := ""
	if err := svc.Update(id, service.MitgliedPatch{Kontoinhaber: &leer}); err != nil {
		t.Fatalf("Update leeren: %v", err)
	}
	neu, iban := "Paul Cramer", ""
	if err := svc.Update(id, service.MitgliedPatch{Kontoinhaber: &neu, IBAN: &iban}); err != nil {
		t.Fatalf("Update setzen: %v", err)
	}

	z := zeileZuMitglied(t, exportLesen(t, svc), id)
	if z.Zahlungspflichtiger != "Paul Cramer" || z.IBAN != "" {
		t.Errorf("Zeile = %+v, erwartet Paul Cramer ohne IBAN", z)
	}
}

func TestKontoinhaber_WiedereintrittUebernimmtDenZahler(t *testing.T) {
	svc := neuerService(t)

	id := kindMitKontoinhaber(t, svc, "Lena", "Petra Berger", "DE02120300000000202051")
	if err := svc.SetKuendigung(id, austrittZum(heuteVersetzt(-10))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}
	if err := svc.Rejoin(id, heuteVersetzt(0)); err != nil {
		t.Fatalf("Rejoin: %v", err)
	}

	if z := zeileZuMitglied(t, exportLesen(t, svc), id); z.Zahlungspflichtiger != "Petra Berger" {
		t.Errorf("Zahlungspflichtiger = %q", z.Zahlungspflichtiger)
	}
}

func TestKontoinhaber_AenderungNachWiedereintrittLaesstDieAlteMitgliedschaftUnveraendert(t *testing.T) {
	svc := neuerService(t)

	id := kindMitKontoinhaber(t, svc, "Lena", "Petra Berger", "DE02120300000000202051")
	if err := svc.SetKuendigung(id, austrittZum(heuteVersetzt(-10))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}
	if err := svc.Rejoin(id, heuteVersetzt(0)); err != nil {
		t.Fatalf("Rejoin: %v", err)
	}

	neu := "Paul Berger"
	if err := svc.Update(id, service.MitgliedPatch{Kontoinhaber: &neu}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(m.Mitgliedschaften) != 2 {
		t.Fatalf("Mitgliedschaften = %d, erwartet 2", len(m.Mitgliedschaften))
	}
	if alt, neue := m.Mitgliedschaften[0].Kontoinhaber, m.Mitgliedschaften[1].Kontoinhaber; alt != "Petra Berger" || neue != "Paul Berger" {
		t.Errorf("Kontoinhaber alt = %q, neu = %q, erwartet %q und %q", alt, neue, "Petra Berger", "Paul Berger")
	}
}

func TestKontoinhaber_WiedereintrittBehaeltKontoinhaberAuchBeiVolljaehrigem(t *testing.T) {
	svc := neuerService(t)

	id := kindMitKontoinhaber(t, svc, "Lena", "Petra Berger", "DE02120300000000202051")
	// Das Handkennzeichen gilt nur ohne Geburtsdatum; ohne es ist Lena kein Kind mehr.
	nein := false
	if err := svc.Update(id, service.MitgliedPatch{KindManuell: &nein}); err != nil {
		t.Fatalf("Update KindManuell: %v", err)
	}
	if err := svc.SetKuendigung(id, austrittZum(heuteVersetzt(-10))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}
	if err := svc.Rejoin(id, heuteVersetzt(0)); err != nil {
		t.Fatalf("Rejoin: %v", err)
	}

	m, err := svc.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if m.IstKind() || m.Kontoinhaber != "Petra Berger" {
		t.Errorf("IstKind = %v, Kontoinhaber = %q, erwartet erwachsen mit %q", m.IstKind(), m.Kontoinhaber, "Petra Berger")
	}
}

func TestKontoinhaber_ZweiKinderDerselbenElternErgebenZweiZeilen(t *testing.T) {
	svc := neuerService(t)

	a := kindMitKontoinhaber(t, svc, "Lena", "Petra Berger", "DE02 1203 0000 0000 2020 51")
	b := kindMitKontoinhaber(t, svc, "Tom", "Petra Berger", "de02120300000000202051")

	export := exportLesen(t, svc)
	if len(export.Zeilen) != 2 {
		t.Fatalf("Zeilen = %d, erwartet 2", len(export.Zeilen))
	}

	za, zb := zeileZuMitglied(t, export, a), zeileZuMitglied(t, export, b)
	if za.Zahlungspflichtiger != zb.Zahlungspflichtiger || za.IBAN != zb.IBAN {
		t.Errorf("Zeilen %+v und %+v, erwartet derselbe Zahlungspflichtige und dieselbe IBAN", za, zb)
	}
	if za.MitgliedID == zb.MitgliedID || za.BetragCents != beitragImTest || zb.BetragCents != beitragImTest {
		t.Errorf("Zeilen %+v und %+v, erwartet je eigene Referenz und eigener Betrag", za, zb)
	}
}

func TestKontoinhaber_SelbstZahlerErscheintNachUmbenennungMitNeuemNamen(t *testing.T) {
	svc := neuerService(t)

	id := moneyMoneyMitglied(t, svc, "Anna", "Alt", "DE89370400440532013000")

	neu := "Neu"
	if err := svc.Update(id, service.MitgliedPatch{Nachname: &neu}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if z := zeileZuMitglied(t, exportLesen(t, svc), id); z.Zahlungspflichtiger != "Anna Neu" {
		t.Errorf("Zahlungspflichtiger = %q, erwartet %q", z.Zahlungspflichtiger, "Anna Neu")
	}
}
