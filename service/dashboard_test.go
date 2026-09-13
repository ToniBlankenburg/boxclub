package service_test

import (
	"testing"
	"time"

	"github.com/ToniBlankenburg/boxclub/service"
)

// Monatsuebersicht am MemberService-Seam (Ticket 26). Wie bei Status liegen
// alle Fixtures relativ zu heute (heuteVersetzt): das Monatssoll rechnet gegen
// den Tag, an dem hingesehen wird, und feste Kalendertage wären ab morgen
// etwas anderes.

// ersterDesFolgemonats liefert den ersten Tag des Monats nach heute — die im
// Ticket ausdrücklich genannte Zeile: ein Eintritt zum Monatsanfang, der aber
// noch nicht begonnen hat.
func ersterDesFolgemonats(heute time.Time) time.Time {
	return time.Date(heute.Year(), heute.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0)
}

func TestMonatsuebersicht_ZaehltDenBeitragEinerLaufendenMitgliedschaftInsMonatssoll(t *testing.T) {
	svc := neuerService(t)
	mitgliedAnlegenZum(t, svc, "Aktiv", heuteVersetzt(-30))

	u, err := svc.Monatsuebersicht()
	if err != nil {
		t.Fatalf("Monatsuebersicht: %v", err)
	}

	if u.MonatssollCents != beitragImTest {
		t.Errorf("MonatssollCents = %d, erwartet %d", u.MonatssollCents, beitragImTest)
	}
	if u.Mitgliederzahlen.Aktiv != 1 {
		t.Errorf("Aktiv = %d, erwartet 1", u.Mitgliederzahlen.Aktiv)
	}
	if got := u.Mitgliederzahlen.Gesamt(); got != 1 {
		t.Errorf("Gesamt = %d, erwartet 1", got)
	}
}

// Der Kern des Tickets: die Kündigungsfrist zählt weiterhin ins Monatssoll,
// denn wer noch trainiert, zahlt auch noch (CONTEXT.md → Kündigungsfrist).
func TestMonatsuebersicht_KuendigungsfristZaehltInsMonatssoll(t *testing.T) {
	svc := neuerService(t)
	id := mitgliedAnlegenZum(t, svc, "Frist", heuteVersetzt(-100))

	if err := svc.SetKuendigung(id, austrittZum(heuteVersetzt(30))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	u, err := svc.Monatsuebersicht()
	if err != nil {
		t.Fatalf("Monatsuebersicht: %v", err)
	}

	if u.MonatssollCents != beitragImTest {
		t.Errorf("MonatssollCents = %d, erwartet %d — die Kündigungsfrist zählt mit", u.MonatssollCents, beitragImTest)
	}
	if u.Mitgliederzahlen.InKuendigungsfrist != 1 {
		t.Errorf("InKuendigungsfrist = %d, erwartet 1", u.Mitgliederzahlen.InKuendigungsfrist)
	}
	if u.Mitgliederzahlen.Aktiv != 0 {
		t.Errorf("Aktiv = %d, erwartet 0", u.Mitgliederzahlen.Aktiv)
	}
}

// Ruhend zählt nicht ins Monatssoll — steht aber daneben, mit Anzahl und
// Betrag, sonst wäre die Summe unerklärt kleiner als die Mitgliederzahl
// vermuten lässt.
func TestMonatsuebersicht_RuhendZaehltNichtInsMonatssollSondernDaneben(t *testing.T) {
	svc := neuerService(t)
	id := mitgliedAnlegenZum(t, svc, "Pause", heuteVersetzt(-30))

	if err := svc.SetRuhend(id, true); err != nil {
		t.Fatalf("SetRuhend: %v", err)
	}

	u, err := svc.Monatsuebersicht()
	if err != nil {
		t.Fatalf("Monatsuebersicht: %v", err)
	}

	if u.MonatssollCents != 0 {
		t.Errorf("MonatssollCents = %d, erwartet 0 — ruhend zieht nichts ein", u.MonatssollCents)
	}
	if u.RuhendCents != beitragImTest {
		t.Errorf("RuhendCents = %d, erwartet %d", u.RuhendCents, beitragImTest)
	}
	if u.Mitgliederzahlen.Ruhend != 1 {
		t.Errorf("Mitgliederzahlen.Ruhend = %d, erwartet 1", u.Mitgliederzahlen.Ruhend)
	}
	if u.Mitgliederzahlen.Aktiv != 0 {
		t.Errorf("Mitgliederzahlen.Aktiv = %d, erwartet 0 — ruhend steht an seiner Stelle, nicht daneben", u.Mitgliederzahlen.Aktiv)
	}
}

// Ruhend gilt auch in der Kündigungsfrist (CONTEXT.md → Ruhend), und auch dort
// zählt es nicht ins Monatssoll.
func TestMonatsuebersicht_RuhendInDerKuendigungsfristZaehltEbenfallsNicht(t *testing.T) {
	svc := neuerService(t)
	id := mitgliedAnlegenZum(t, svc, "PauseFrist", heuteVersetzt(-100))

	if err := svc.SetKuendigung(id, austrittZum(heuteVersetzt(30))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}
	if err := svc.SetRuhend(id, true); err != nil {
		t.Fatalf("SetRuhend: %v", err)
	}

	u, err := svc.Monatsuebersicht()
	if err != nil {
		t.Fatalf("Monatsuebersicht: %v", err)
	}

	if u.MonatssollCents != 0 {
		t.Errorf("MonatssollCents = %d, erwartet 0", u.MonatssollCents)
	}
	if u.Mitgliederzahlen.InKuendigungsfrist != 0 {
		t.Errorf("InKuendigungsfrist = %d, erwartet 0 — ruhend steht an seiner Stelle", u.Mitgliederzahlen.InKuendigungsfrist)
	}
	if u.Mitgliederzahlen.Ruhend != 1 {
		t.Errorf("Ruhend = %d, erwartet 1", u.Mitgliederzahlen.Ruhend)
	}
}

// Der seltene Grenzfall, in dem sich Neu und Ruhend überschneiden könnten:
// SetRuhend prüft wie SetKuendigung nur, ob der Austritt noch bevorsteht
// (laeuftNoch), nicht ob der Eintritt schon begonnen hat — ruhend stellen
// lässt sich also auch ein Zeitraum, der noch gar nicht läuft. Ein noch nicht
// begonnener Zeitraum zieht aber ohnehin nichts ein, ruhend gestellt oder
// nicht, und zählt deshalb weiterhin als Neu.
func TestMonatsuebersicht_RuhendVorDemEintrittZaehltAlsNeuUndNichtAlsRuhend(t *testing.T) {
	svc := neuerService(t)
	id := mitgliedAnlegenZum(t, svc, "FruehePause", heuteVersetzt(7))

	if err := svc.SetRuhend(id, true); err != nil {
		t.Fatalf("SetRuhend vor dem Eintritt: %v", err)
	}

	u, err := svc.Monatsuebersicht()
	if err != nil {
		t.Fatalf("Monatsuebersicht: %v", err)
	}

	if u.Mitgliederzahlen.Neu != 1 {
		t.Errorf("Neu = %d, erwartet 1", u.Mitgliederzahlen.Neu)
	}
	if u.Mitgliederzahlen.Ruhend != 0 {
		t.Errorf("Ruhend = %d, erwartet 0 — ein noch nicht begonnener Zeitraum zieht nichts ein", u.Mitgliederzahlen.Ruhend)
	}
	if u.RuhendCents != 0 {
		t.Errorf("RuhendCents = %d, erwartet 0", u.RuhendCents)
	}
}

// Ein Eintritt zum Ersten des Folgemonats zählt nicht ins Monatssoll — der
// Zeitraum hat noch nicht begonnen (Status Neu), auch wenn "nächster Monat"
// im Alltag oft der übliche Beginn ist.
func TestMonatsuebersicht_EintrittAmErstenDesFolgemonatsZaehltNicht(t *testing.T) {
	svc := neuerService(t)
	mitgliedAnlegenZum(t, svc, "Folgemonat", ersterDesFolgemonats(heuteVersetzt(0)))

	u, err := svc.Monatsuebersicht()
	if err != nil {
		t.Fatalf("Monatsuebersicht: %v", err)
	}

	if u.MonatssollCents != 0 {
		t.Errorf("MonatssollCents = %d, erwartet 0 — der Eintritt steht noch bevor", u.MonatssollCents)
	}
	if u.Mitgliederzahlen.Neu != 1 {
		t.Errorf("Neu = %d, erwartet 1", u.Mitgliederzahlen.Neu)
	}
}

// Ein Austritt heute zählt nicht mehr ins Monatssoll: der Austrittstag selbst
// ist schon erreicht (CONTEXT.md → Status), und ab dann wird nichts mehr
// eingezogen.
func TestMonatsuebersicht_AustrittHeuteZaehltNichtMehr(t *testing.T) {
	svc := neuerService(t)
	id := mitgliedAnlegenZum(t, svc, "Stichtag", heuteVersetzt(-100))

	if err := svc.SetKuendigung(id, austrittZum(heuteVersetzt(0))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	u, err := svc.Monatsuebersicht()
	if err != nil {
		t.Fatalf("Monatsuebersicht: %v", err)
	}

	if u.MonatssollCents != 0 {
		t.Errorf("MonatssollCents = %d, erwartet 0 — der Austritt ist heute erreicht", u.MonatssollCents)
	}
	if u.Mitgliederzahlen.Ausgetreten != 1 {
		t.Errorf("Ausgetreten = %d, erwartet 1", u.Mitgliederzahlen.Ausgetreten)
	}
	if u.Mitgliederzahlen.InKuendigungsfrist != 0 {
		t.Errorf("InKuendigungsfrist = %d, erwartet 0", u.Mitgliederzahlen.InKuendigungsfrist)
	}
}

// Der Rückstand hängt am Mitglied und nicht an der Mitgliedschaft (ADR-0006):
// er zählt auch bei einem Ausgetretenen mit, denn ein Austritt erlässt keine
// Schulden.
func TestMonatsuebersicht_RueckstandZaehltAuchBeiEinemAusgetretenen(t *testing.T) {
	svc := neuerService(t)
	id := mitgliedAnlegenZum(t, svc, "Schuldner", heuteVersetzt(-100))

	if err := svc.SetKuendigung(id, austrittZum(heuteVersetzt(-1))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}
	if err := svc.SetRueckstand(id, service.Rueckstand{Offen: true, Notiz: "Rücklastschrift"}); err != nil {
		t.Fatalf("SetRueckstand: %v", err)
	}

	u, err := svc.Monatsuebersicht()
	if err != nil {
		t.Fatalf("Monatsuebersicht: %v", err)
	}

	if u.RueckstandAnzahl != 1 {
		t.Errorf("RueckstandAnzahl = %d, erwartet 1 — ein Austritt erlässt keine Schulden", u.RueckstandAnzahl)
	}
	if u.Mitgliederzahlen.Ausgetreten != 1 {
		t.Errorf("Ausgetreten = %d, erwartet 1", u.Mitgliederzahlen.Ausgetreten)
	}
}

// Neu eingetreten und ausgetreten "in diesem Monat" ist eine Bewegung und kein
// Zustand: heute liegt garantiert im laufenden Monat, ein Datum ein Jahr
// entfernt garantiert nicht.
func TestMonatsuebersicht_NeuEingetretenUndAusgetretenDiesenMonat(t *testing.T) {
	svc := neuerService(t)

	// Eintritt heute: fällt in den laufenden Monat.
	mitgliedAnlegenZum(t, svc, "EintrittHeute", heuteVersetzt(0))

	// Eintritt vor einem Jahr: fällt sicher nicht in den laufenden Monat.
	mitgliedAnlegenZum(t, svc, "EintrittFrueher", heuteVersetzt(-365))

	// Austritt heute: fällt in den laufenden Monat, auch wenn der Zeitraum
	// selbst vor einem Jahr begann.
	idAustrittHeute := mitgliedAnlegenZum(t, svc, "AustrittHeute", heuteVersetzt(-365))
	if err := svc.SetKuendigung(idAustrittHeute, austrittZum(heuteVersetzt(0))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	// Austritt vor einem Jahr: fällt sicher nicht in den laufenden Monat.
	idAustrittFrueher := mitgliedAnlegenZum(t, svc, "AustrittFrueher", heuteVersetzt(-730))
	if err := svc.SetKuendigung(idAustrittFrueher, austrittZum(heuteVersetzt(-365))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	u, err := svc.Monatsuebersicht()
	if err != nil {
		t.Fatalf("Monatsuebersicht: %v", err)
	}

	if u.NeuEingetretenDiesenMonat != 1 {
		t.Errorf("NeuEingetretenDiesenMonat = %d, erwartet 1", u.NeuEingetretenDiesenMonat)
	}
	if u.AusgetretenDiesenMonat != 1 {
		t.Errorf("AusgetretenDiesenMonat = %d, erwartet 1", u.AusgetretenDiesenMonat)
	}
}

// Mehrere Mitglieder zusammen: die fünf Zustände schließen sich gegenseitig
// aus und ergeben in der Summe den ganzen Bestand.
func TestMonatsuebersicht_GesamtErgibtSichAusDenFuenfZustaenden(t *testing.T) {
	svc := neuerService(t)

	mitgliedAnlegenZum(t, svc, "Aktiv", heuteVersetzt(-30))
	mitgliedAnlegenZum(t, svc, "Neu", heuteVersetzt(7))

	ruhend := mitgliedAnlegenZum(t, svc, "Ruhend", heuteVersetzt(-30))
	if err := svc.SetRuhend(ruhend, true); err != nil {
		t.Fatalf("SetRuhend: %v", err)
	}

	frist := mitgliedAnlegenZum(t, svc, "Frist", heuteVersetzt(-100))
	if err := svc.SetKuendigung(frist, austrittZum(heuteVersetzt(30))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	ausgetreten := mitgliedAnlegenZum(t, svc, "Ausgetreten", heuteVersetzt(-100))
	if err := svc.SetKuendigung(ausgetreten, austrittZum(heuteVersetzt(-1))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}

	u, err := svc.Monatsuebersicht()
	if err != nil {
		t.Fatalf("Monatsuebersicht: %v", err)
	}

	if got := u.Mitgliederzahlen.Gesamt(); got != 5 {
		t.Fatalf("Gesamt = %d, erwartet 5", got)
	}
	z := u.Mitgliederzahlen
	if z.Aktiv != 1 || z.Neu != 1 || z.Ruhend != 1 || z.InKuendigungsfrist != 1 || z.Ausgetreten != 1 {
		t.Errorf("Mitgliederzahlen = %+v, erwartet je einmal Aktiv, Neu, Ruhend, InKuendigungsfrist, Ausgetreten", z)
	}
	// Monatssoll trägt nur Aktiv und InKuendigungsfrist, macht hier also zwei
	// Beiträge — nicht die fünf angelegten Mitglieder.
	if u.MonatssollCents != beitragImTest*2 {
		t.Errorf("MonatssollCents = %d, erwartet %d", u.MonatssollCents, beitragImTest*2)
	}
}
