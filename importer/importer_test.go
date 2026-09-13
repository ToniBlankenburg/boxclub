package importer_test

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"

	"github.com/ToniBlankenburg/boxclub/importer"
	"github.com/ToniBlankenburg/boxclub/service"
)

// Die Fixture ist bewusst im Test gebaut und liegt nicht als Datei im
// Repository: die echte Tabelle des Vereins enthält Klarnamen, IBANs und
// Kontaktdaten und darf nirgends hin. Gebaut wird sie aus denselben
// Spaltenüberschriften und derselben anonymisierten Musterzeile, die
// .scratch/boxclub-v1/excel-vorlage.md festhält — samt der Tippfehler-
// Überschrift „Eintrit".

// spalten sind die Überschriften des Blattes „Verwaltung" in der Reihenfolge,
// in der sie dort stehen.
var spalten = []string{
	"Training - 1", "Mandatsreferenz", "Vorname", "Nachname", "Beitrag",
	"Status", "Mitgliedschaft", "Gekündigt", "IBAN", "Eintrit", "1x 2x Woche",
	"Telefonnummer", "E-Mail", "Adresse", "Postleitzahl", "Ort", "Geburtstag",
	"Geschlecht", "Anmeldegebühr", "Bewertung", "Digital",
}

// musterzeile ist die anonymisierte Beispielzeile. Die Datumsspalten stehen als
// Excel-Seriennummern da, so wie in der echten Tabelle.
func musterzeile() map[string]any {
	return map[string]any{
		"Training - 1":    "Samstag 10:30 Uhr",
		"Mandatsreferenz": 1,
		"Vorname":         "Erika",
		"Nachname":        "Musterfrau",
		"Beitrag":         60,
		"Status":          "Neu",
		"Mitgliedschaft":  46296, // 2026-10-01
		"Gekündigt":       "",
		"IBAN":            "DE02120300000000202051",
		"Eintrit":         46270, // 2026-09-05
		"1x 2x Woche":     "1x Woche",
		"Telefonnummer":   "01511 0000000",
		"E-Mail":          "erika@example.org",
		"Adresse":         "Musterstrasse 15/1",
		"Postleitzahl":    "70174",
		"Ort":             "Stuttgart",
		"Geburtstag":      36601, // 2000-03-16
		"Geschlecht":      "Mann",
		"Anmeldegebühr":   60,
		"Bewertung":       "❌",
		"Digital":         "Digital",
	}
}

// mappe baut eine .xlsx mit dem Blatt „Verwaltung" und dem Blatt „Quelle", das
// der Importer ignorieren soll.
func mappe(t *testing.T, ueberschriften []string, zeilen ...map[string]any) []byte {
	t.Helper()

	f := excelize.NewFile()
	t.Cleanup(func() { f.Close() })

	if _, err := f.NewSheet("Verwaltung"); err != nil {
		t.Fatalf("Blatt anlegen: %v", err)
	}
	if _, err := f.NewSheet("Quelle"); err != nil {
		t.Fatalf("Blatt Quelle anlegen: %v", err)
	}
	if err := f.SetCellValue("Quelle", "A1", "Status"); err != nil {
		t.Fatalf("Quelle füllen: %v", err)
	}

	setzen := func(zeile int, werte func(spalte string) (any, bool)) {
		for i, spalte := range ueberschriften {
			wert, ok := werte(spalte)
			if !ok {
				continue
			}
			zelle, err := excelize.CoordinatesToCellName(i+1, zeile)
			if err != nil {
				t.Fatalf("Zelle bestimmen: %v", err)
			}
			if err := f.SetCellValue("Verwaltung", zelle, wert); err != nil {
				t.Fatalf("Zelle %s setzen: %v", zelle, err)
			}
		}
	}

	setzen(1, func(spalte string) (any, bool) { return spalte, true })
	for nr, zeile := range zeilen {
		setzen(nr+2, func(spalte string) (any, bool) {
			wert, ok := zeile[spalte]

			return wert, ok
		})
	}

	var puffer bytes.Buffer
	if err := f.Write(&puffer); err != nil {
		t.Fatalf("Mappe schreiben: %v", err)
	}

	return puffer.Bytes()
}

// termin baut einen Eintrag des Stundenplans. Er kommt ohne Datenbank aus: der
// Importer bekommt den Katalog übergeben und liest ihn nur.
func termin(id int64, tag service.Wochentag, beginn string) service.Trainingstermin {
	return service.Trainingstermin{ID: id, Wochentag: tag, Beginn: service.Uhrzeit(beginn)}
}

// Die Termine, auf die die Freitexte der Musterzeile passen. Die Excel schreibt
// sie „Samstag 10:30 Uhr", der Stundenplan „Samstag 10:30" — dass beides
// derselbe Termin ist, ist der Kern dieses Abgleichs.
var (
	samstag    = termin(1, service.Samstag, "10:30")
	dienstag   = termin(2, service.Dienstag, "19:30")
	donnerstag = termin(3, service.Donnerstag, "18:00")
)

// stundenplan ist der Katalog, gegen den die Tests lesen, wenn sie nichts
// anderes sagen.
func stundenplan() importer.Stundenplan {
	return importer.StundenplanAus([]service.Trainingstermin{samstag, dienstag, donnerstag})
}

// lesen ruft den Importer auf der gebauten Mappe auf.
func lesen(t *testing.T, inhalt []byte) importer.Ergebnis {
	t.Helper()

	return lesenMit(t, inhalt, stundenplan())
}

// lesenMit liest gegen einen eigenen Stundenplan — für die Fälle, in denen
// gerade er der Gegenstand ist.
func lesenMit(t *testing.T, inhalt []byte, plan importer.Stundenplan) importer.Ergebnis {
	t.Helper()

	ergebnis, err := importer.ExcelImporter{}.Lesen(bytes.NewReader(inhalt), plan)
	if err != nil {
		t.Fatalf("Lesen: %v", err)
	}

	return ergebnis
}

// datum parst ein ISO-Datum für Testfixtures.
func datum(t *testing.T, iso string) time.Time {
	t.Helper()

	d, err := time.Parse("2006-01-02", iso)
	if err != nil {
		t.Fatalf("Testfixture %q ist kein gültiges Datum: %v", iso, err)
	}

	return d
}

// gruende verdichtet den Fehlerbericht auf „Zeile: Grund" je Eintrag.
func gruende(ergebnis importer.Ergebnis) string {
	return meldungen(ergebnis.Fehler)
}

// hinweise ist dasselbe für die Meldungen zu Zeilen, die trotzdem durchgehen.
func hinweise(ergebnis importer.Ergebnis) string {
	return meldungen(ergebnis.Hinweise)
}

func meldungen(liste []importer.Zeilenmeldung) string {
	var b strings.Builder
	for _, m := range liste {
		fmt.Fprintf(&b, "Zeile %d: %s\n", m.Zeile, strings.Join(m.Meldungen, "; "))
	}

	return b.String()
}

func TestLesen_UebernimmtDieMusterzeile(t *testing.T) {
	ergebnis := lesen(t, mappe(t, spalten, musterzeile()))

	if len(ergebnis.Fehler) != 0 {
		t.Fatalf("Fehlerbericht nicht leer:\n%s", gruende(ergebnis))
	}
	if len(ergebnis.Saetze) != 1 {
		t.Fatalf("%d Sätze, erwartet 1", len(ergebnis.Saetze))
	}

	satz := ergebnis.Saetze[0].Satz

	if satz.ID != 1 {
		t.Errorf("Mitglieds-ID = %d, erwartet 1", satz.ID)
	}
	if satz.Vorname != "Erika" || satz.Nachname != "Musterfrau" {
		t.Errorf("Name = %q %q, erwartet Erika Musterfrau", satz.Vorname, satz.Nachname)
	}
	if !satz.Eintritt.Equal(datum(t, "2026-10-01")) {
		t.Errorf("Eintritt = %v, erwartet 2026-10-01", satz.Eintritt)
	}
	if satz.Anmeldung.Datum == nil || !satz.Anmeldung.Datum.Equal(datum(t, "2026-09-05")) {
		t.Errorf("Anmeldedatum = %v, erwartet 2026-09-05", satz.Anmeldung.Datum)
	}
	if satz.Geburtsdatum == nil || !satz.Geburtsdatum.Equal(datum(t, "2000-03-16")) {
		t.Errorf("Geburtsdatum = %v, erwartet 2000-03-16", satz.Geburtsdatum)
	}
	if satz.BeitragCents != 6000 {
		t.Errorf("Beitrag = %d Cent, erwartet 6000", satz.BeitragCents)
	}
	if satz.Anmeldung.GebuehrCents != 6000 {
		t.Errorf("Anmeldegebühr = %d Cent, erwartet 6000", satz.Anmeldung.GebuehrCents)
	}
	if satz.Anschrift != (service.Anschrift{Adresse: "Musterstrasse 15/1", Postleitzahl: "70174", Ort: "Stuttgart"}) {
		t.Errorf("Anschrift = %+v", satz.Anschrift)
	}
	if satz.Email != "erika@example.org" || satz.Telefon != "01511 0000000" {
		t.Errorf("Kontakt = %q / %q", satz.Email, satz.Telefon)
	}
	if satz.IBAN != "DE02120300000000202051" {
		t.Errorf("IBAN = %q", satz.IBAN)
	}
	if satz.Geschlecht != "Mann" {
		t.Errorf("Geschlecht = %q, erwartet Mann", satz.Geschlecht)
	}
	if bool(satz.GoogleBewertung) {
		t.Error("GoogleBewertung = true, erwartet false für ❌")
	}
	if satz.Digital != "Digital" {
		t.Errorf("Digital = %q, erwartet den Wert wortwörtlich", satz.Digital)
	}
	if satz.Ruhend {
		t.Error("Ruhend = true, erwartet false für Status „Neu“")
	}
	if satz.Kuendigung.Datum != nil || satz.Kuendigung.Austritt != nil {
		t.Errorf("Kündigung = %+v, erwartet leer", satz.Kuendigung)
	}
}

// mitZeile baut die Musterzeile mit einzelnen geänderten Zellen.
func mitZeile(aenderungen map[string]any) map[string]any {
	zeile := musterzeile()
	for spalte, wert := range aenderungen {
		zeile[spalte] = wert
	}

	return zeile
}

// einzigerSatz gibt den einen erwarteten Satz zurück und scheitert sonst.
func einzigerSatz(t *testing.T, ergebnis importer.Ergebnis) service.Importsatz {
	t.Helper()

	if len(ergebnis.Hinweise) != 0 {
		t.Fatalf("Hinweise, erwartet keine:\n%s", hinweise(ergebnis))
	}

	return satzMitHinweisen(t, ergebnis)
}

// satzMitHinweisen verlangt die eine übernommene Zeile, lässt aber Hinweise zu:
// ein nicht zugeordneter Trainingstermin hält die Zeile nicht auf.
func satzMitHinweisen(t *testing.T, ergebnis importer.Ergebnis) service.Importsatz {
	t.Helper()

	if len(ergebnis.Fehler) != 0 {
		t.Fatalf("Fehlerbericht nicht leer:\n%s", gruende(ergebnis))
	}
	if len(ergebnis.Saetze) != 1 {
		t.Fatalf("%d Sätze, erwartet 1", len(ergebnis.Saetze))
	}

	return ergebnis.Saetze[0].Satz
}

// nurFehler verlangt, dass keine Zeile übernommen wurde, und liefert den Bericht
// als Text.
func nurFehler(t *testing.T, ergebnis importer.Ergebnis) string {
	t.Helper()

	if len(ergebnis.Saetze) != 0 {
		t.Fatalf("%d Sätze übernommen, erwartet keinen", len(ergebnis.Saetze))
	}
	if len(ergebnis.Fehler) == 0 {
		t.Fatal("kein Eintrag im Fehlerbericht")
	}

	return gruende(ergebnis)
}

func TestLesen_SeriendatumUndTextdatumErgebenDenselbenTag(t *testing.T) {
	erwartet := datum(t, "2026-10-01")

	for _, geschrieben := range []any{46296, "2026-10-01", "01.10.2026", "1.10.2026"} {
		t.Run(fmt.Sprint(geschrieben), func(t *testing.T) {
			satz := einzigerSatz(t, lesen(t, mappe(t, spalten,
				mitZeile(map[string]any{"Mitgliedschaft": geschrieben}))))

			if !satz.Eintritt.Equal(erwartet) {
				t.Errorf("Eintritt = %v, erwartet %v", satz.Eintritt, erwartet)
			}
		})
	}
}

func TestLesen_UnlesbaresDatumMachtDieZeileZumFehlerfall(t *testing.T) {
	ergebnis := lesen(t, mappe(t, spalten,
		mitZeile(map[string]any{"Mitgliedschaft": "Oktober letztes Jahr"})))

	if bericht := nurFehler(t, ergebnis); !strings.Contains(bericht, "kein lesbares Datum") {
		t.Errorf("Bericht nennt das Datum nicht:\n%s", bericht)
	}
	if ergebnis.Fehler[0].Zeile != 2 {
		t.Errorf("Zeilennummer = %d, erwartet 2", ergebnis.Fehler[0].Zeile)
	}
}

func TestLesen_FehlenderNachnameMachtDieZeileZumFehlerfall(t *testing.T) {
	bericht := nurFehler(t, lesen(t, mappe(t, spalten,
		mitZeile(map[string]any{"Nachname": ""}))))

	if !strings.Contains(bericht, "Nachname") {
		t.Errorf("Bericht nennt den Nachnamen nicht:\n%s", bericht)
	}
}

func TestLesen_FehlendeUndUnlesbareMitgliedsIDMachenDieZeileZumFehlerfall(t *testing.T) {
	for name, wert := range map[string]any{"leer": "", "keine Zahl": "A-47"} {
		t.Run(name, func(t *testing.T) {
			bericht := nurFehler(t, lesen(t, mappe(t, spalten,
				mitZeile(map[string]any{"Mandatsreferenz": wert}))))

			if !strings.Contains(bericht, "Mitglieds-ID") {
				t.Errorf("Bericht nennt die Mitglieds-ID nicht:\n%s", bericht)
			}
		})
	}
}

func TestLesen_DoppelteMitgliedsIDMachtDieZweiteZeileZumFehlerfall(t *testing.T) {
	ergebnis := lesen(t, mappe(t, spalten,
		mitZeile(map[string]any{"Mandatsreferenz": 47}),
		mitZeile(map[string]any{"Mandatsreferenz": 47, "Vorname": "Hans"})))

	if len(ergebnis.Saetze) != 1 {
		t.Fatalf("%d Sätze, erwartet 1 — die erste Zeile bleibt übernehmbar", len(ergebnis.Saetze))
	}
	if len(ergebnis.Fehler) != 1 {
		t.Fatalf("%d Fehler, erwartet 1:\n%s", len(ergebnis.Fehler), gruende(ergebnis))
	}
	if ergebnis.Fehler[0].Zeile != 3 {
		t.Errorf("Zeilennummer = %d, erwartet 3", ergebnis.Fehler[0].Zeile)
	}
	if !strings.Contains(gruende(ergebnis), "47 doppelt") {
		t.Errorf("Bericht = %q, erwartet einen Hinweis auf die doppelte Nummer", gruende(ergebnis))
	}
}

// Die Trainingsspalten sind Freitext, der Satz führt Verweise: der Abgleich
// dagegen ist der Gegenstand dieser Gruppe (ADR-0008).

// mitDreiSpalten ist die Kopfzeile mit allen drei Trainingsspalten — die
// Mustertabelle führt nur die erste.
func mitDreiSpalten() []string {
	return append(slices.Clone(spalten), "Training - 2", "Training - 3")
}

func TestLesen_OrdnetDieFreitexteDemStundenplanZu(t *testing.T) {
	satz := einzigerSatz(t, lesen(t, mappe(t, spalten, musterzeile())))

	if !slices.Equal(satz.TrainingsterminIDs, []int64{samstag.ID}) {
		t.Errorf("Trainingstermine = %v, erwartet den Samstagstermin %d",
			satz.TrainingsterminIDs, samstag.ID)
	}
}

func TestLesen_VergleichtOhneRuecksichtAufSchreibweiseUndLeerraum(t *testing.T) {
	// Der Termin steht im Stundenplan als „Samstag 10:30“, mit Ende und
	// Bezeichnung — die Excel kennt beides nicht.
	plan := importer.StundenplanAus([]service.Trainingstermin{{
		ID: 7, Wochentag: service.Samstag, Beginn: "10:30", Ende: "12:00", Bezeichnung: "Anfänger",
	}})

	for _, geschrieben := range []string{
		"Samstag 10:30 Uhr",
		"samstag 10:30 uhr",
		"  Samstag   10:30  ",
		"Samstag 10:30 – 12:00 · Anfänger",
	} {
		t.Run(geschrieben, func(t *testing.T) {
			satz := satzMitHinweisen(t, lesenMit(t, mappe(t, spalten,
				mitZeile(map[string]any{"Training - 1": geschrieben})), plan))

			if !slices.Equal(satz.TrainingsterminIDs, []int64{7}) {
				t.Errorf("Trainingstermine = %v, erwartet den Termin 7", satz.TrainingsterminIDs)
			}
		})
	}
}

// „Sa 10:30“ ist keine Schreibvariante, die der Abgleich auflöst: alles außer
// Schreibweise und Leerraum muss stimmen. Das ist eingepreist (ADR-0008) — und
// es hält die Zeile nicht auf.
func TestLesen_UnbekannterFreitextWirdZumHinweisUndHaeltDieZeileNichtAuf(t *testing.T) {
	ergebnis := lesen(t, mappe(t, spalten, mitZeile(map[string]any{
		"Training - 1": "Sa 10:30",
		"1x 2x Woche":  "",
	})))

	satz := satzMitHinweisen(t, ergebnis)
	if len(satz.TrainingsterminIDs) != 0 {
		t.Errorf("Trainingstermine = %v, erwartet keine", satz.TrainingsterminIDs)
	}

	if len(ergebnis.Hinweise) != 1 || ergebnis.Hinweise[0].Zeile != 2 {
		t.Fatalf("Hinweise = %+v, erwartet einen zu Zeile 2", ergebnis.Hinweise)
	}

	bericht := hinweise(ergebnis)
	for _, erwartet := range []string{"Training - 1", "Sa 10:30", "Stundenplan"} {
		if !strings.Contains(bericht, erwartet) {
			t.Errorf("Hinweis nennt %q nicht:\n%s", erwartet, bericht)
		}
	}
}

// Der Wert „Kein“ ist kein Termin und kein Fehler: die Zelle ist ausgefüllt,
// und zwar mit „findet nicht statt“.
func TestLesen_KeinIstKeinTerminUndKeineMeldung(t *testing.T) {
	for _, geschrieben := range []string{"Kein", "kein", "KEIN"} {
		t.Run(geschrieben, func(t *testing.T) {
			satz := einzigerSatz(t, lesen(t, mappe(t, spalten, mitZeile(map[string]any{
				"Training - 1": geschrieben,
				"1x 2x Woche":  "",
			}))))

			if len(satz.TrainingsterminIDs) != 0 {
				t.Errorf("Trainingstermine = %v, erwartet keine", satz.TrainingsterminIDs)
			}
		})
	}
}

// Ein archivierter Termin wird nicht vergeben: der Import darf keine Zeiten
// austeilen, die es nicht mehr gibt (ADR-0008).
func TestLesen_VergibtArchivierteTermineNicht(t *testing.T) {
	stillgelegt := samstag
	stillgelegt.Archiviert = true

	ergebnis := lesenMit(t, mappe(t, spalten, mitZeile(map[string]any{"1x 2x Woche": ""})),
		importer.StundenplanAus([]service.Trainingstermin{stillgelegt}))

	satz := satzMitHinweisen(t, ergebnis)
	if len(satz.TrainingsterminIDs) != 0 {
		t.Errorf("Trainingstermine = %v, erwartet keine", satz.TrainingsterminIDs)
	}
	if bericht := hinweise(ergebnis); !strings.Contains(bericht, "archiviert") {
		t.Errorf("Hinweis sagt nicht, dass der Termin archiviert ist:\n%s", bericht)
	}
}

// Zwei Gruppen zur selben Zeit sind erlaubt (CreateTrainingstermin). Die kurze
// Schreibweise trifft dann beide — und geraten wird nicht.
func TestLesen_MeldetEinenMehrdeutigenFreitext(t *testing.T) {
	ergebnis := lesenMit(t, mappe(t, spalten, mitZeile(map[string]any{"1x 2x Woche": ""})),
		importer.StundenplanAus([]service.Trainingstermin{
			{ID: 1, Wochentag: service.Samstag, Beginn: "10:30", Bezeichnung: "Anfänger"},
			{ID: 2, Wochentag: service.Samstag, Beginn: "10:30", Bezeichnung: "Wettkampf"},
		}))

	satz := satzMitHinweisen(t, ergebnis)
	if len(satz.TrainingsterminIDs) != 0 {
		t.Errorf("Trainingstermine = %v, erwartet keine — geraten wird nicht", satz.TrainingsterminIDs)
	}
	if bericht := hinweise(ergebnis); !strings.Contains(bericht, "mehrere") {
		t.Errorf("Hinweis nennt die Mehrdeutigkeit nicht:\n%s", bericht)
	}
}

// Ein leerer Stundenplan ist kein Fehler der Datei: die Mitglieder kommen
// herein, ihre Trainingszeiten nicht.
func TestLesen_LeererStundenplanOrdnetNichtsZuUndLaesstDieZeilenDurch(t *testing.T) {
	ergebnis := lesenMit(t, mappe(t, spalten, musterzeile()), importer.StundenplanAus(nil))

	satz := satzMitHinweisen(t, ergebnis)
	if len(satz.TrainingsterminIDs) != 0 {
		t.Errorf("Trainingstermine = %v, erwartet keine", satz.TrainingsterminIDs)
	}
	if len(ergebnis.Hinweise) != 1 {
		t.Fatalf("Hinweise = %+v, erwartet einen zur Zeile", ergebnis.Hinweise)
	}
}

func TestLesen_OrdnetAlleDreiTrainingsspaltenZu(t *testing.T) {
	satz := einzigerSatz(t, lesen(t, mappe(t, mitDreiSpalten(), mitZeile(map[string]any{
		"Training - 2": "Dienstag 19:30 Uhr",
		"Training - 3": "Donnerstag 18:00 Uhr",
		"1x 2x Woche":  "3x Woche",
	}))))

	erwartet := []int64{samstag.ID, dienstag.ID, donnerstag.ID}
	if !slices.Equal(satz.TrainingsterminIDs, erwartet) {
		t.Errorf("Trainingstermine = %v, erwartet %v", satz.TrainingsterminIDs, erwartet)
	}
}

// Eine leere Trainingszelle ist kein Termin: sie zählt nicht mit.
func TestLesen_LeereTrainingszelleZaehltNicht(t *testing.T) {
	satz := einzigerSatz(t, lesen(t, mappe(t, mitDreiSpalten(), mitZeile(map[string]any{
		"Training - 2": "",
		"Training - 3": "Donnerstag 18:00 Uhr",
		"1x 2x Woche":  "2x Woche",
	}))))

	if len(satz.TrainingsterminIDs) != 2 {
		t.Errorf("Trainingstermine = %v, erwartet zwei", satz.TrainingsterminIDs)
	}
}

// Derselbe Termin in zwei Spalten ist dieselbe Vereinbarung und keine zwei.
func TestLesen_ZaehltDenselbenTerminNurEinmal(t *testing.T) {
	ergebnis := lesen(t, mappe(t, mitDreiSpalten(), mitZeile(map[string]any{
		"Training - 2": "samstag 10:30",
		"1x 2x Woche":  "1x Woche",
	})))

	satz := einzigerSatz(t, ergebnis)
	if !slices.Equal(satz.TrainingsterminIDs, []int64{samstag.ID}) {
		t.Errorf("Trainingstermine = %v, erwartet den Samstagstermin einmal",
			satz.TrainingsterminIDs)
	}
}

// Die Gegenprobe zur Spalte „1x 2x Woche" zählt die zugeordneten Termine — und
// hält die Zeile nicht auf: aus dieser Spalte wird nichts gespeichert.
func TestLesen_WidersprechendeFrequenzWirdZumHinweis(t *testing.T) {
	ergebnis := lesen(t, mappe(t, mitDreiSpalten(), mitZeile(map[string]any{
		"Training - 2": "Dienstag 19:30 Uhr",
		"Training - 3": "Donnerstag 18:00 Uhr",
		"1x 2x Woche":  "1x Woche",
	})))

	satzMitHinweisen(t, ergebnis)

	if bericht := hinweise(ergebnis); !strings.Contains(bericht,
		"Frequenz 1× widerspricht 3 zugeordneten Trainingsterminen") {
		t.Errorf("Bericht benennt den Widerspruch nicht:\n%s", bericht)
	}
}

// Eine unlesbare Frequenz ist kein Widerspruch zwischen zwei Angaben, sondern
// eine Zelle, die niemand deuten kann — und bleibt deshalb ein Fehlerfall.
func TestLesen_UnlesbareFrequenzMachtDieZeileZumFehlerfall(t *testing.T) {
	bericht := nurFehler(t, lesen(t, mappe(t, spalten,
		mitZeile(map[string]any{"1x 2x Woche": "gelegentlich"}))))

	if !strings.Contains(bericht, "keine lesbare Frequenz") {
		t.Errorf("Bericht nennt die unlesbare Frequenz nicht:\n%s", bericht)
	}
}

// Was der Stundenplan nicht kennt, zählt auch nicht: die Frequenzprüfung sieht
// die zugeordneten Termine und nicht die gefüllten Zellen.
func TestLesen_FrequenzpruefungZaehltNurZugeordneteTermine(t *testing.T) {
	ergebnis := lesen(t, mappe(t, mitDreiSpalten(), mitZeile(map[string]any{
		"Training - 2": "Dienstag 19:30 Uhr",
		"Training - 3": "Mitwoch 18:00 Uhr",
		"1x 2x Woche":  "2x Woche",
	})))

	satz := satzMitHinweisen(t, ergebnis)
	if len(satz.TrainingsterminIDs) != 2 {
		t.Errorf("Trainingstermine = %v, erwartet zwei", satz.TrainingsterminIDs)
	}

	bericht := hinweise(ergebnis)
	if strings.Contains(bericht, "widerspricht") {
		t.Errorf("Bericht meldet einen Frequenzwiderspruch, erwartet keinen:\n%s", bericht)
	}
	if !strings.Contains(bericht, "Mitwoch 18:00 Uhr") {
		t.Errorf("Bericht nennt den nicht zugeordneten Freitext nicht:\n%s", bericht)
	}
}

func TestLesen_DeutetDenStatusFuerDasGekuendigtDatum(t *testing.T) {
	gekuendigtAm := datum(t, "2026-11-03")

	t.Run("Gekündigt wird zum Austritt", func(t *testing.T) {
		satz := einzigerSatz(t, lesen(t, mappe(t, spalten, mitZeile(map[string]any{
			"Status": "Gekündigt", "Gekündigt": "03.11.2026",
		}))))

		if satz.Kuendigung.Austritt == nil || !satz.Kuendigung.Austritt.Equal(gekuendigtAm) {
			t.Errorf("Austritt = %v, erwartet %v", satz.Kuendigung.Austritt, gekuendigtAm)
		}
		if satz.Kuendigung.Datum != nil {
			t.Errorf("Kündigungsdatum = %v, erwartet leer", satz.Kuendigung.Datum)
		}
	})

	t.Run("Kündigungsfrist wird zum Kündigungsdatum", func(t *testing.T) {
		satz := einzigerSatz(t, lesen(t, mappe(t, spalten, mitZeile(map[string]any{
			"Status": "Kündigungsfrist", "Gekündigt": "03.11.2026",
		}))))

		if satz.Kuendigung.Datum == nil || !satz.Kuendigung.Datum.Equal(gekuendigtAm) {
			t.Errorf("Kündigungsdatum = %v, erwartet %v", satz.Kuendigung.Datum, gekuendigtAm)
		}
		if satz.Kuendigung.Austritt != nil {
			t.Errorf("Austritt = %v, erwartet leer", satz.Kuendigung.Austritt)
		}
	})

	for _, status := range []string{"Stillgelegt", "Inakiv"} {
		t.Run(status+" schaltet ruhend", func(t *testing.T) {
			satz := einzigerSatz(t, lesen(t, mappe(t, spalten,
				mitZeile(map[string]any{"Status": status}))))

			if !satz.Ruhend {
				t.Error("Ruhend = false, erwartet true")
			}
		})
	}
}

func TestLesen_UnbekannterStatusMachtDieZeileZumFehlerfall(t *testing.T) {
	bericht := nurFehler(t, lesen(t, mappe(t, spalten,
		mitZeile(map[string]any{"Status": "Halbtot"}))))

	if !strings.Contains(bericht, `unbekannter Status „Halbtot“`) {
		t.Errorf("Bericht nennt den Status nicht:\n%s", bericht)
	}
}

func TestLesen_UnbekannteBewertungMachtDieZeileZumFehlerfall(t *testing.T) {
	bericht := nurFehler(t, lesen(t, mappe(t, spalten,
		mitZeile(map[string]any{"Bewertung": "vielleicht"}))))

	if !strings.Contains(bericht, "Bewertung") {
		t.Errorf("Bericht nennt die Bewertung nicht:\n%s", bericht)
	}
}

func TestLesen_BrichtBeiFehlenderSpalteAb(t *testing.T) {
	ohneNachname := slices.DeleteFunc(slices.Clone(spalten),
		func(s string) bool { return s == "Nachname" })

	_, err := importer.ExcelImporter{}.Lesen(
		bytes.NewReader(mappe(t, ohneNachname, musterzeile())), stundenplan())
	if err == nil {
		t.Fatal("Lesen ohne Spalte „Nachname“ lief durch, erwartet Abbruch")
	}
	if !strings.Contains(err.Error(), "Nachname") {
		t.Errorf("Meldung = %q, erwartet einen Hinweis auf die fehlende Spalte", err)
	}
}

func TestLesen_BrichtOhneBlattVerwaltungAb(t *testing.T) {
	f := excelize.NewFile()
	defer f.Close()

	var puffer bytes.Buffer
	if err := f.Write(&puffer); err != nil {
		t.Fatalf("Mappe schreiben: %v", err)
	}

	_, err := importer.ExcelImporter{}.Lesen(bytes.NewReader(puffer.Bytes()), stundenplan())
	if err == nil {
		t.Fatal("Lesen ohne Blatt „Verwaltung“ lief durch, erwartet Abbruch")
	}
	if !strings.Contains(err.Error(), "Verwaltung") {
		t.Errorf("Meldung = %q, erwartet einen Hinweis auf das fehlende Blatt", err)
	}
}

func TestLesen_UeberspringtLeereZeilen(t *testing.T) {
	ergebnis := lesen(t, mappe(t, spalten,
		musterzeile(),
		map[string]any{},
		mitZeile(map[string]any{"Mandatsreferenz": 2, "Vorname": "Hans"})))

	if len(ergebnis.Fehler) != 0 {
		t.Fatalf("Fehlerbericht nicht leer:\n%s", gruende(ergebnis))
	}
	if len(ergebnis.Saetze) != 2 {
		t.Errorf("%d Sätze, erwartet 2 — die leere Zeile zählt nicht", len(ergebnis.Saetze))
	}
}

func TestLesen_FasstAlleMeldungenEinerZeileZusammen(t *testing.T) {
	ergebnis := lesen(t, mappe(t, spalten, mitZeile(map[string]any{
		"Nachname": "", "Status": "Halbtot",
	})))

	if len(ergebnis.Fehler) != 1 {
		t.Fatalf("%d Einträge im Bericht, erwartet 1 — eine Zeile ist eine gescheiterte Zeile:\n%s",
			len(ergebnis.Fehler), gruende(ergebnis))
	}
	if len(ergebnis.Fehler[0].Meldungen) != 2 {
		t.Errorf("Gründe = %q, erwartet beide Probleme der Zeile", ergebnis.Fehler[0].Meldungen)
	}
}

func TestLesen_WeistJahreszahlenUndUnplausibleSeriennummernAb(t *testing.T) {
	for name, wert := range map[string]any{
		"nackte Jahreszahl":      "1990",
		"Jahreszahl als Zahl":    1990,
		"zu kleine Seriennummer": 60,
		"zu große Seriennummer":  999999,
	} {
		t.Run(name, func(t *testing.T) {
			bericht := nurFehler(t, lesen(t, mappe(t, spalten,
				mitZeile(map[string]any{"Geburtstag": wert}))))

			if !strings.Contains(bericht, "Geburtstag") {
				t.Errorf("Bericht nennt die Spalte nicht:\n%s", bericht)
			}
		})
	}
}

func TestLesen_LaesstPlausibleGeburtsdatenDurch(t *testing.T) {
	// Ein 1935 geborenes Mitglied ist unwahrscheinlich, aber möglich — die
	// Plausibilitätsgrenze darf es nicht abweisen.
	for _, fall := range []struct {
		wert     any
		erwartet string
	}{
		{36601, "2000-03-16"},
		{12784, "1934-12-31"},
		{"01.01.1935", "1935-01-01"},
	} {
		satz := einzigerSatz(t, lesen(t, mappe(t, spalten,
			mitZeile(map[string]any{"Geburtstag": fall.wert}))))

		if satz.Geburtsdatum.Format("2006-01-02") != fall.erwartet {
			t.Errorf("Geburtstag %v = %s, erwartet %s",
				fall.wert, satz.Geburtsdatum.Format("2006-01-02"), fall.erwartet)
		}
	}
}

func TestLesen_MeldetEinUnlesbaresKuendigungsdatumNurEinmal(t *testing.T) {
	ergebnis := lesen(t, mappe(t, spalten, mitZeile(map[string]any{
		"Status": "Gekündigt", "Gekündigt": "irgendwann",
	})))

	bericht := nurFehler(t, ergebnis)
	if strings.Contains(bericht, "fehlt das Datum") {
		t.Errorf("Bericht sagt zugleich „nicht lesbar“ und „fehlt“:\n%s", bericht)
	}
	if len(ergebnis.Fehler[0].Meldungen) != 1 {
		t.Errorf("Gründe = %q, erwartet genau einen", ergebnis.Fehler[0].Meldungen)
	}
}
