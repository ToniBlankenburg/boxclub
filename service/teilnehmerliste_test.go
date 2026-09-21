package service_test

import (
	"slices"
	"testing"
	"time"

	"github.com/ToniBlankenburg/boxclub/service"
)

// Die Teilnehmerliste (CONTEXT.md → Teilnehmerliste, Ticket 28) wird wie alles
// hier über die MemberService-API beobachtet: angelegt wird über Create, Update,
// SetKuendigung und Rejoin, gelesen über Teilnehmerlisten.

// teilnehmerAnlegen legt ein Mitglied an, das ab eintritt läuft und für die
// angegebenen Termine angemeldet ist.
func teilnehmerAnlegen(t *testing.T, svc *service.MemberService, vorname, nachname string, eintritt time.Time, termine ...service.Trainingstermin) int64 {
	t.Helper()

	id, err := svc.Create(service.NeuesMitglied{
		Vorname:            vorname,
		Nachname:           nachname,
		BeitragCents:       beitragImTest,
		Eintritt:           eintritt,
		TrainingsterminIDs: ids(termine...),
	})
	if err != nil {
		t.Fatalf("Create(%s %s): %v", vorname, nachname, err)
	}

	return id
}

// listeZu sucht die Teilnehmerliste eines Termins heraus.
func listeZu(t *testing.T, listen []service.Teilnehmerliste, termin service.Trainingstermin) service.Teilnehmerliste {
	t.Helper()

	for _, l := range listen {
		if l.Termin.ID == termin.ID {
			return l
		}
	}
	t.Fatalf("keine Teilnehmerliste zu Termin %q; vorhanden: %v", termin.Anzeige(), terminAnzeigen(listen))

	return service.Teilnehmerliste{}
}

// terminAnzeigen verdichtet Listen auf die Anzeigetexte ihrer Termine.
func terminAnzeigen(listen []service.Teilnehmerliste) []string {
	texte := make([]string, 0, len(listen))
	for _, l := range listen {
		texte = append(texte, l.Termin.Anzeige())
	}

	return texte
}

// teilnehmerNamen verdichtet die Teilnehmer auf "Nachname, Vorname" — die Form, in der
// sich Inhalt und Reihenfolge in einem Zug vergleichen lassen.
func teilnehmerNamen(l service.Teilnehmerliste) []string {
	texte := make([]string, 0, len(l.Teilnehmer))
	for _, teilnehmer := range l.Teilnehmer {
		texte = append(texte, teilnehmer.Nachname+", "+teilnehmer.Vorname)
	}

	return texte
}

func TestTeilnehmerlisten_ZeigtAngemeldeteAlphabetischMitAnzahl(t *testing.T) {
	svc := neuerService(t)
	montag, mittwoch, samstag := stundenplan(t, svc)
	eintritt := heuteVersetzt(-90)

	// Angelegt nicht in der Reihenfolge, in der sie stehen sollen. Gleicher
	// Nachname bei den ersten beiden: dann entscheidet der Vorname. Der Umlaut
	// sortiert wie in der Mitgliederliste zum O und nicht hinter das Z.
	teilnehmerAnlegen(t, svc, "Zoe", "Berger", eintritt, samstag)
	teilnehmerAnlegen(t, svc, "Anna", "Berger", eintritt, samstag, montag)
	teilnehmerAnlegen(t, svc, "Chris", "Albers", eintritt, samstag)
	teilnehmerAnlegen(t, svc, "Ömer", "Özdemir", eintritt, samstag)
	teilnehmerAnlegen(t, svc, "Paula", "Pohl", eintritt, samstag)

	listen, err := svc.Teilnehmerlisten(false)
	if err != nil {
		t.Fatalf("Teilnehmerlisten: %v", err)
	}

	// Jeder Termin des Stundenplans steht einmal da, in Wochenreihenfolge —
	// auch der, für den niemand angemeldet ist.
	if got, want := terminAnzeigen(listen), []string{montag.Anzeige(), mittwoch.Anzeige(), samstag.Anzeige()}; !slices.Equal(got, want) {
		t.Fatalf("Termine = %v, erwartet %v", got, want)
	}

	sa := listeZu(t, listen, samstag)
	if got, want := teilnehmerNamen(sa), []string{"Albers, Chris", "Berger, Anna", "Berger, Zoe", "Özdemir, Ömer", "Pohl, Paula"}; !slices.Equal(got, want) {
		t.Errorf("Samstag = %v, erwartet %v", got, want)
	}
	if sa.Anzahl() != 5 {
		t.Errorf("Anzahl Samstag = %d, erwartet 5", sa.Anzahl())
	}

	// Ein Mitglied für zwei Termine steht in beiden Listen, je einmal.
	if got, want := teilnehmerNamen(listeZu(t, listen, montag)), []string{"Berger, Anna"}; !slices.Equal(got, want) {
		t.Errorf("Montag = %v, erwartet %v", got, want)
	}

	mi := listeZu(t, listen, mittwoch)
	if len(mi.Teilnehmer) != 0 || mi.Anzahl() != 0 {
		t.Errorf("Mittwoch = %v (Anzahl %d), erwartet leer", teilnehmerNamen(mi), mi.Anzahl())
	}
}

// Ein ruhendes Mitglied trainiert gerade nicht. Ohne Kennzeichen in der Liste
// muss schon die Auswahl stimmen: es fehlt, und mit der Rücknahme der Pause
// steht es wieder da — die Anmeldung selbst hat sich nie geändert.
func TestTeilnehmerlisten_RuhendeStehenNichtDarin(t *testing.T) {
	svc := neuerService(t)
	_, _, samstag := stundenplan(t, svc)
	eintritt := heuteVersetzt(-90)

	teilnehmerAnlegen(t, svc, "Anna", "Berger", eintritt, samstag)
	pause := teilnehmerAnlegen(t, svc, "Ben", "Conrad", eintritt, samstag)

	if err := svc.SetRuhend(pause, true); err != nil {
		t.Fatalf("SetRuhend(true): %v", err)
	}

	listen, err := svc.Teilnehmerlisten(false)
	if err != nil {
		t.Fatalf("Teilnehmerlisten: %v", err)
	}
	if got, want := teilnehmerNamen(listeZu(t, listen, samstag)), []string{"Berger, Anna"}; !slices.Equal(got, want) {
		t.Errorf("während der Pause = %v, erwartet %v", got, want)
	}

	if err := svc.SetRuhend(pause, false); err != nil {
		t.Fatalf("SetRuhend(false): %v", err)
	}

	listen, err = svc.Teilnehmerlisten(false)
	if err != nil {
		t.Fatalf("Teilnehmerlisten: %v", err)
	}
	if got, want := teilnehmerNamen(listeZu(t, listen, samstag)), []string{"Berger, Anna", "Conrad, Ben"}; !slices.Equal(got, want) {
		t.Errorf("nach der Pause = %v, erwartet %v", got, want)
	}
}

// Die Auswahl folgt dem Lebenszyklus: wer neu ist, aktiv ist oder in der
// Kündigungsfrist steht, gehört zum Termin; wer ausgetreten ist, nicht mehr —
// seine Anmeldung bleibt nur Historie und zählt weiter zur Frequenz.
func TestTeilnehmerlisten_NachStatus(t *testing.T) {
	svc := neuerService(t)
	_, _, samstag := stundenplan(t, svc)

	teilnehmerAnlegen(t, svc, "Anna", "Aktiv", heuteVersetzt(-90), samstag)
	teilnehmerAnlegen(t, svc, "Nele", "Neu", heuteVersetzt(30), samstag)

	fristlaeufer := teilnehmerAnlegen(t, svc, "Frida", "Frist", heuteVersetzt(-90), samstag)
	if err := svc.SetKuendigung(fristlaeufer, austrittZum(heuteVersetzt(30))); err != nil {
		t.Fatalf("SetKuendigung (Kündigungsfrist): %v", err)
	}

	weg := teilnehmerAnlegen(t, svc, "Egon", "Ehemalig", heuteVersetzt(-90), samstag)
	if err := svc.SetKuendigung(weg, austrittZum(heuteVersetzt(-1))); err != nil {
		t.Fatalf("SetKuendigung (Austritt erreicht): %v", err)
	}

	listen, err := svc.Teilnehmerlisten(false)
	if err != nil {
		t.Fatalf("Teilnehmerlisten: %v", err)
	}

	got := teilnehmerNamen(listeZu(t, listen, samstag))
	if want := []string{"Aktiv, Anna", "Frist, Frida", "Neu, Nele"}; !slices.Equal(got, want) {
		t.Errorf("Samstag = %v, erwartet %v", got, want)
	}

	// Der Ausgetretene ist aus der Liste, nicht aus dem Bestand: seine
	// Anmeldung steht noch.
	m, err := svc.Get(weg)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(m.Mitgliedschaften) != 1 || !slices.Equal(ids(m.Mitgliedschaften[0].Trainingstermine...), ids(samstag)) {
		t.Errorf("Anmeldung des Ausgetretenen = %+v, erwartet weiterhin Samstag", m.Mitgliedschaften)
	}
}

// Nach einem Wiedereintritt zählt nur der neue Zeitraum: er beginnt ohne
// Termine, und die des alten bleiben Historie. Wer wieder eintritt, steht also
// erst dann wieder in einer Liste, wenn neu ausgewählt wurde — und dann nur in
// der neuen.
func TestTeilnehmerlisten_NachWiedereintrittNurDerNeueZeitraum(t *testing.T) {
	svc := neuerService(t)
	montag, _, samstag := stundenplan(t, svc)

	id := teilnehmerAnlegen(t, svc, "Rita", "Rückkehr", heuteVersetzt(-200), samstag)
	if err := svc.SetKuendigung(id, austrittZum(heuteVersetzt(-60))); err != nil {
		t.Fatalf("SetKuendigung: %v", err)
	}
	if err := svc.Rejoin(id, heuteVersetzt(-10)); err != nil {
		t.Fatalf("Rejoin: %v", err)
	}

	listen, err := svc.Teilnehmerlisten(false)
	if err != nil {
		t.Fatalf("Teilnehmerlisten: %v", err)
	}
	if got := teilnehmerNamen(listeZu(t, listen, samstag)); len(got) != 0 {
		t.Errorf("Samstag nach Wiedereintritt = %v, erwartet leer (der neue Zeitraum hat keine Termine)", got)
	}

	neueTermine := ids(montag)
	if err := svc.Update(id, service.MitgliedPatch{TrainingsterminIDs: &neueTermine}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	listen, err = svc.Teilnehmerlisten(false)
	if err != nil {
		t.Fatalf("Teilnehmerlisten: %v", err)
	}
	if got, want := teilnehmerNamen(listeZu(t, listen, montag)), []string{"Rückkehr, Rita"}; !slices.Equal(got, want) {
		t.Errorf("Montag = %v, erwartet %v", got, want)
	}
	if got := teilnehmerNamen(listeZu(t, listen, samstag)); len(got) != 0 {
		t.Errorf("Samstag = %v, erwartet weiterhin leer", got)
	}

	// Meldet der neue Zeitraum sich wieder für den Termin des alten an, steht
	// die Person je Termin trotzdem nur einmal da: zwei Zeiträume mit
	// demselben Termin sind keine zwei Teilnehmer.
	beide := ids(montag, samstag)
	if err := svc.Update(id, service.MitgliedPatch{TrainingsterminIDs: &beide}); err != nil {
		t.Fatalf("Update (beide Termine): %v", err)
	}

	listen, err = svc.Teilnehmerlisten(false)
	if err != nil {
		t.Fatalf("Teilnehmerlisten: %v", err)
	}
	for _, termin := range []service.Trainingstermin{montag, samstag} {
		if got, want := teilnehmerNamen(listeZu(t, listen, termin)), []string{"Rückkehr, Rita"}; !slices.Equal(got, want) {
			t.Errorf("%s = %v, erwartet %v", termin.Anzeige(), got, want)
		}
	}
}

// Ein archivierter Termin ist aus dem Stundenplan genommen, seine Anmeldungen
// stehen aber weiter (ADR-0008). Die Liste folgt dem Schalter der Terminansicht:
// ohne ihn fehlt der Termin samt Teilnehmern, mit ihm steht er an seinem Platz
// in der Woche.
func TestTeilnehmerlisten_ArchivierteNurAufWunsch(t *testing.T) {
	svc := neuerService(t)
	montag, mittwoch, samstag := stundenplan(t, svc)

	teilnehmerAnlegen(t, svc, "Anna", "Berger", heuteVersetzt(-90), mittwoch)
	teilnehmerAnlegen(t, svc, "Ben", "Conrad", heuteVersetzt(-90), samstag)

	if err := svc.SetTrainingsterminArchiviert(mittwoch.ID, true); err != nil {
		t.Fatalf("SetTrainingsterminArchiviert: %v", err)
	}

	listen, err := svc.Teilnehmerlisten(false)
	if err != nil {
		t.Fatalf("Teilnehmerlisten(false): %v", err)
	}
	if got, want := terminAnzeigen(listen), []string{montag.Anzeige(), samstag.Anzeige()}; !slices.Equal(got, want) {
		t.Errorf("ohne Archivierte: Termine = %v, erwartet %v", got, want)
	}

	listen, err = svc.Teilnehmerlisten(true)
	if err != nil {
		t.Fatalf("Teilnehmerlisten(true): %v", err)
	}
	if got, want := terminAnzeigen(listen), []string{montag.Anzeige(), mittwoch.Anzeige(), samstag.Anzeige()}; !slices.Equal(got, want) {
		t.Errorf("mit Archivierten: Termine = %v, erwartet %v", got, want)
	}
	if got, want := teilnehmerNamen(listeZu(t, listen, mittwoch)), []string{"Berger, Anna"}; !slices.Equal(got, want) {
		t.Errorf("archivierter Mittwoch = %v, erwartet %v", got, want)
	}
}

// Die Serienmail an eine Teilnehmerliste (CONTEXT.md → Teilnehmerliste)
// braucht nur die E-Mail-Adressen der dort Angemeldeten — andere Termine
// bleiben außen vor.
func TestTeilnehmerEmails_NurDerAngegebeneTermin(t *testing.T) {
	svc := neuerService(t)
	montag, _, samstag := stundenplan(t, svc)
	eintritt := heuteVersetzt(-90)

	_, err := svc.Create(service.NeuesMitglied{
		Vorname: "Anna", Nachname: "Berger", Email: "anna@example.org",
		BeitragCents: beitragImTest, Eintritt: eintritt, TrainingsterminIDs: ids(samstag),
	})
	if err != nil {
		t.Fatalf("Create Anna: %v", err)
	}
	_, err = svc.Create(service.NeuesMitglied{
		Vorname: "Ben", Nachname: "Conrad", Email: "ben@example.org",
		BeitragCents: beitragImTest, Eintritt: eintritt, TrainingsterminIDs: ids(montag),
	})
	if err != nil {
		t.Fatalf("Create Ben: %v", err)
	}

	emails, err := svc.TeilnehmerEmails(samstag.ID)
	if err != nil {
		t.Fatalf("TeilnehmerEmails: %v", err)
	}
	if erwartet := []string{"anna@example.org"}; !slices.Equal(emails, erwartet) {
		t.Errorf("TeilnehmerEmails(Samstag) = %v, erwartet %v", emails, erwartet)
	}
}

// Ein Termin ohne Anmeldungen liefert keine Adressen und keinen Fehler.
func TestTeilnehmerEmails_OhneAnmeldungenLeer(t *testing.T) {
	svc := neuerService(t)
	_, mittwoch, _ := stundenplan(t, svc)

	emails, err := svc.TeilnehmerEmails(mittwoch.ID)
	if err != nil || len(emails) != 0 {
		t.Fatalf("TeilnehmerEmails(Mittwoch) = %v, %v, erwartet leer", emails, err)
	}
}
