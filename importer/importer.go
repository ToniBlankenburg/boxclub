// Package importer liest die bestehende Excel-Tabelle des Vereins und übersetzt
// sie in Importsätze des Service.
//
// Der Importer schreibt nichts: er liefert je lesbarer Zeile einen Satz und je
// unlesbarer einen Eintrag im Fehlerbericht. Eingesetzt werden die Sätze
// woanders (siehe .scratch/boxclub-v1/spec.md → Seams). Das trennt zwei Dinge,
// die getrennt gehören: was in der Tabelle steht, und was davon in die
// Datenbank kommt.
//
// Die zweite Leitlinie ist, dass der Importer nichts errät. Was er nicht
// zuordnen kann, macht die Zeile zum Fehlerfall statt sie mit einem Standardwert
// durchzuwinken — der Bericht ist das Werkzeug, mit dem der Verein den Schmutz
// in seiner Tabelle überhaupt zu sehen bekommt.
//
// Eine Ausnahme davon sind die Trainingstermine: sie werden gegen den
// Stundenplan gehalten, den der Aufrufer mitbringt (siehe Stundenplan), und ein
// Freitext ohne Termin hält die Zeile nicht auf. Er kommt in den Bericht, das
// Mitglied kommt ohne diesen Termin herein — eine nicht zugeordnete
// Trainingszeit ist eine unvollständige Vereinbarung und kein unlesbares
// Mitglied.
package importer

import (
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"golang.org/x/text/unicode/norm"

	"github.com/ToniBlankenburg/boxclub/service"
)

// blatt ist das einzige Blatt, das gelesen wird. Das zweite Blatt „Quelle" hält
// nur die Dropdown-Wertelisten und geht den Import nichts an.
const blatt = "Verwaltung"

// kopfzeile ist die Zeile mit den Überschriften; die Daten beginnen darunter.
const kopfzeile = 1

// Zeilenmeldung ist, was zu einer Zeile zu sagen war — mit der Zeilennummer,
// unter der sie in Excel steht, damit der Verein sie dort wiederfindet und
// korrigiert.
//
// Ein Eintrag ist eine Zeile und nicht eine Meldung: eine Zeile, an der drei
// Dinge fehlen, ist eine gescheiterte Zeile und keine drei. Sonst zählte der
// Bericht mehr Fehlschläge, als die Tabelle Zeilen hat.
//
// Ob die Zeile an diesen Meldungen gescheitert ist, sagt nicht der Eintrag,
// sondern die Liste, in der er steht (siehe Ergebnis).
type Zeilenmeldung struct {
	Zeile     int
	Meldungen []string
}

// Zeilensatz ist ein gelesener Satz samt seiner Herkunft. Die Zeilennummer
// reist mit, weil auch das Einsetzen noch scheitern kann und der Bericht dann
// dieselbe Zeile nennen können muss.
type Zeilensatz struct {
	Zeile int
	Satz  service.Importsatz
}

// Ergebnis ist, was aus einer Datei herauskommt: die lesbaren Zeilen und die
// Gründe für die übrigen. Beides zusammen, nicht das eine oder das andere — eine
// einzelne kaputte Zeile darf die 199 gesunden nicht aufhalten.
type Ergebnis struct {
	Saetze []Zeilensatz
	Fehler []Zeilenmeldung

	// Hinweise stehen zu Zeilen, die trotzdem in Saetze stehen: an ihnen war
	// etwas zu bemängeln, das die Zeile nicht unbrauchbar macht. Das ist heute
	// alles, was die Trainingsspalten betrifft — ein Freitext ohne Termin im
	// Stundenplan und die Gegenprobe zur Spalte „1x 2x Woche".
	//
	// Getrennt von Fehler, weil beide verschieden zählen: die eine Liste sind
	// die gescheiterten Zeilen, die andere sind übernommene mit einer
	// Nacharbeit. Im Bericht stehen sie nebeneinander an derselben Stelle.
	Hinweise []Zeilenmeldung
}

// ExcelImporter liest .xlsx-Dateien im Format der Vereinstabelle. Er hält keinen
// Zustand; der leere Wert ist einsatzbereit.
type ExcelImporter struct{}

// Lesen parst die Mappe und liefert Sätze und Fehlerbericht.
//
// Der Stundenplan ist der Katalog, dem die Freitexte der Trainingsspalten
// zugeordnet werden. Ein leerer ist erlaubt und kein Fehler der Datei: dann
// kommt jedes Mitglied ohne Termine herein, und der Bericht sagt zu jeder
// Trainingszelle, dass sie nirgends steht.
//
// Der zurückgegebene Fehler ist dem Fehlerbericht nicht gleichrangig: er meldet,
// dass die Datei als Ganzes nicht zu gebrauchen ist (kein .xlsx, kein Blatt
// „Verwaltung", fehlende Spalte). Dann wird gar nichts übernommen, statt halbe
// Zeilen zu schreiben.
func (ExcelImporter) Lesen(r io.Reader, plan Stundenplan) (Ergebnis, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return Ergebnis{}, fmt.Errorf("die Datei ließ sich nicht als .xlsx öffnen: %w", err)
	}
	defer f.Close()

	// RawCellValue liefert die Zellen so, wie sie gespeichert sind: ein
	// Datum als Seriennummer statt in der Formatierung, die der Rechner des
	// Vereins gerade anzeigt. Gedeutet wird hier und nicht von Excelize.
	// Erst nachsehen, ob es das Blatt gibt, und dann lesen: sonst bekäme jeder
	// Lesefehler die Auskunft „das Blatt fehlt", auch wenn das Blatt da ist und
	// etwas ganz anderes klemmt.
	index, err := f.GetSheetIndex(blatt)
	if err != nil || index == -1 {
		return Ergebnis{}, fmt.Errorf("das Blatt „%s“ fehlt in der Datei", blatt)
	}

	zeilen, err := f.GetRows(blatt, excelize.Options{RawCellValue: true})
	if err != nil {
		return Ergebnis{}, fmt.Errorf("das Blatt „%s“ ließ sich nicht lesen: %w", blatt, err)
	}
	if len(zeilen) < kopfzeile {
		return Ergebnis{}, fmt.Errorf("das Blatt „%s“ hat keine Kopfzeile", blatt)
	}

	kopf, err := kopfLesen(zeilen[kopfzeile-1])
	if err != nil {
		return Ergebnis{}, err
	}

	return kopf.zeilenLesen(zeilen[kopfzeile:], plan), nil
}

// kopf ordnet jeder erwarteten Überschrift ihre Spaltennummer zu. Zugeordnet
// wird über die Überschrift und nicht über die Position: die Tabelle ist über
// Jahre gewachsen, und eine verschobene Spalte darf nicht die Daten verschieben.
type kopf map[string]int

// pflichtspalten sind die Überschriften, ohne die der Import nicht laufen kann.
// Sie stehen in der Reihenfolge der Tabelle; die Tippfehler „Eintrit" und die
// irreführende „Mandatsreferenz" gehören dazu, weil in der Datei genau das
// steht (siehe .scratch/boxclub-v1/excel-vorlage.md).
var pflichtspalten = []string{
	spalteTraining1, spalteNummer, spalteVorname, spalteNachname, spalteBeitrag,
	spalteStatus, spalteEintritt, spalteGekuendigt, spalteIBAN, spalteAnmeldedatum,
	spalteFrequenz, spalteTelefon, spalteEmail, spalteAdresse, spaltePLZ, spalteOrt,
	spalteGeburtstag, spalteGeschlecht, spalteAnmeldegebuehr, spalteBewertung,
}

// Die Überschriften der Tabelle, wortwörtlich wie sie dort stehen.
const (
	spalteTraining1      = "Training - 1"
	spalteTraining2      = "Training - 2"
	spalteTraining3      = "Training - 3"
	spalteNummer         = "Mandatsreferenz"
	spalteVorname        = "Vorname"
	spalteNachname       = "Nachname"
	spalteBeitrag        = "Beitrag"
	spalteStatus         = "Status"
	spalteEintritt       = "Mitgliedschaft"
	spalteGekuendigt     = "Gekündigt"
	spalteIBAN           = "IBAN"
	spalteAnmeldedatum   = "Eintrit"
	spalteFrequenz       = "1x 2x Woche"
	spalteTelefon        = "Telefonnummer"
	spalteEmail          = "E-Mail"
	spalteAdresse        = "Adresse"
	spaltePLZ            = "Postleitzahl"
	spalteOrt            = "Ort"
	spalteGeburtstag     = "Geburtstag"
	spalteGeschlecht     = "Geschlecht"
	spalteAnmeldegebuehr = "Anmeldegebühr"
	spalteBewertung      = "Bewertung"
)

// trainingsspalten sind die Termine in der Reihenfolge, in der sie in der
// Tabelle stehen. Nur die erste ist Pflicht: die Mustertabelle führt sie allein, und
// eine zweite und dritte kommen vor, ohne dass ihr Fehlen den Import aufhalten
// dürfte. Weicht die Zahl der so gefundenen Termine von der Spalte
// „1x 2x Woche" ab, sagt der Bericht das (siehe frequenzPruefen).
var trainingsspalten = []string{spalteTraining1, spalteTraining2, spalteTraining3}

// kopfLesen ordnet die Überschriften ihren Spalten zu und meldet die erste, die
// fehlt. Abgebrochen wird dabei sofort: eine fehlende Spalte hieße, jede Zeile
// halb zu übernehmen, und ein Bericht mit 200 gleichlautenden Einträgen hilft
// niemandem.
func kopfLesen(zeile []string) (kopf, error) {
	k := make(kopf, len(zeile))
	for i, ueberschrift := range zeile {
		k[norm.NFC.String(strings.TrimSpace(ueberschrift))] = i
	}

	for _, pflicht := range pflichtspalten {
		if _, ok := k[pflicht]; !ok {
			return nil, fmt.Errorf(
				"im Blatt „%s“ fehlt die Spalte „%s“ — die Datei passt nicht zur erwarteten Tabelle",
				blatt, pflicht)
		}
	}

	return k, nil
}

// zeilenLesen geht die Datenzeilen durch und sammelt Sätze und Fehler ein.
func (k kopf) zeilenLesen(zeilen [][]string, plan Stundenplan) Ergebnis {
	var ergebnis Ergebnis

	// vergeben merkt sich, welche Mitglieds-Nummer in dieser Datei schon
	// vorkam. Eine doppelte Nummer ist kein Detail: unter ihr führt der Verein
	// seine Mitglieder, und zwei Zeilen mit derselben wären zwei Personen mit
	// einer Identität.
	vergeben := map[int64]int{}

	for i, zeile := range zeilen {
		nummer := kopfzeile + 1 + i

		if leer(zeile) {
			continue
		}

		satz, gruende, hinweise := k.zeileLesen(zeile, plan)
		if len(gruende) == 0 {
			if vorher, doppelt := vergeben[satz.ID]; doppelt {
				gruende = append(gruende, fmt.Sprintf(
					"Mitglieds-ID %d doppelt — sie steht schon in Zeile %d", satz.ID, vorher))
			} else {
				vergeben[satz.ID] = nummer
			}
		}

		// Die Hinweise einer gescheiterten Zeile gehen zu ihren Gründen: die
		// Zeile ist nicht übernommen, und an zwei Stellen im Bericht zu stehen
		// hieße, sie zweimal zu zählen. Beides zusammen ist dann die Liste der
		// Meldungen zu dieser Zeile — daher der Name des Feldes.
		if len(gruende) > 0 {
			ergebnis.Fehler = append(ergebnis.Fehler,
				Zeilenmeldung{Zeile: nummer, Meldungen: append(gruende, hinweise...)})

			continue
		}

		if len(hinweise) > 0 {
			ergebnis.Hinweise = append(ergebnis.Hinweise,
				Zeilenmeldung{Zeile: nummer, Meldungen: hinweise})
		}

		ergebnis.Saetze = append(ergebnis.Saetze, Zeilensatz{Zeile: nummer, Satz: satz})
	}

	return ergebnis
}

// leer sagt, ob in der Zeile überhaupt nichts steht. Solche Zeilen entstehen
// unterhalb der Daten, wenn jemand in Excel weiter unten einmal etwas
// eingetragen und wieder gelöscht hat; sie sind keine Fehler.
func leer(zeile []string) bool {
	for _, zelle := range zeile {
		if strings.TrimSpace(zelle) != "" {
			return false
		}
	}

	return true
}

// zeileLesen übersetzt eine Datenzeile in einen Importsatz oder in die Gründe,
// aus denen das nicht geht. Gesammelt werden alle Gründe einer Zeile und nicht
// nur der erste: wer seine Tabelle korrigiert, will alles auf einmal sehen.
//
// Zurück kommen zwei Listen: die Gründe, an denen die Zeile scheitert, und die
// Hinweise, mit denen sie trotzdem durchgeht.
func (k kopf) zeileLesen(zeile []string, plan Stundenplan) (service.Importsatz, []string, []string) {
	z := zeilenleser{kopf: k, zeile: zeile, plan: plan}

	satz := service.Importsatz{
		ID: z.nummer(),
		NeuesMitglied: service.NeuesMitglied{
			Vorname:         z.pflichttext(spalteVorname),
			Nachname:        z.pflichttext(spalteNachname),
			Geburtsdatum:    z.datum(spalteGeburtstag),
			Anschrift:       z.anschrift(),
			Email:           z.text(spalteEmail),
			Telefon:         z.text(spalteTelefon),
			IBAN:            z.text(spalteIBAN),
			Geschlecht:      z.text(spalteGeschlecht),
			GoogleBewertung: z.bewertung(),
			BeitragCents:    z.beitrag(),
			Anmeldung: service.Anmeldung{
				Datum:        z.datum(spalteAnmeldedatum),
				GebuehrCents: z.anmeldegebuehr(),
			},
		},
	}

	if eintritt := z.datum(spalteEintritt); eintritt != nil {
		satz.Eintritt = *eintritt
	} else if z.text(spalteEintritt) == "" {
		z.melden("in der Spalte „%s“ fehlt das Eintrittsdatum", spalteEintritt)
	}

	z.lebenszyklus(&satz)

	satz.TrainingsterminIDs = z.trainingstermine()
	z.frequenzPruefen(len(satz.TrainingsterminIDs))

	return satz, z.fehler, z.hinweise
}

// zeilenleser liest die Zellen einer Zeile und sammelt dabei die Gründe ein,
// aus denen sie nicht zu deuten waren. Er sammelt statt abzubrechen, damit eine
// Zeile mit drei Problemen auch drei Einträge im Bericht ergibt.
type zeilenleser struct {
	kopf     kopf
	zeile    []string
	plan     Stundenplan
	fehler   []string
	hinweise []string
}

// melden legt einen Grund zum Bericht dieser Zeile. Die Zeile ist damit
// gescheitert.
func (z *zeilenleser) melden(format string, args ...any) {
	z.fehler = append(z.fehler, fmt.Sprintf(format, args...))
}

// hinweisen legt eine Meldung zum Bericht, die die Zeile nicht aufhält.
func (z *zeilenleser) hinweisen(format string, args ...any) {
	z.hinweise = append(z.hinweise, fmt.Sprintf(format, args...))
}

// text liest eine Zelle als Freitext. Fehlt die Spalte in dieser Zeile — Excel
// schneidet leere Zellen am Zeilenende ab —, ist das Ergebnis leer.
//
// Der Text kommt in Normalform NFC zurück: Umlaute stehen in einer über Jahre
// von Hand gepflegten Excel mal vorkomponiert ("ü"), mal zerlegt ("u" + Trema)
// da, je nachdem, auf welchem Rechner die Zeile zuletzt getippt wurde — für das
// Auge derselbe Text, für einen exakten Vergleich zwei verschiedene. Ohne diese
// Vereinheitlichung schlüge z. B. der Statusvergleich in lebenszyklus an genau
// den Zeilen fehl, an denen es niemand bemerkt.
func (z *zeilenleser) text(spalte string) string {
	i, ok := z.kopf[spalte]
	if !ok || i >= len(z.zeile) {
		return ""
	}

	return norm.NFC.String(strings.TrimSpace(z.zeile[i]))
}

// pflichttext liest eine Angabe, die dastehen muss.
func (z *zeilenleser) pflichttext(spalte string) string {
	wert := z.text(spalte)
	if wert == "" {
		z.melden("in der Spalte „%s“ fehlt die Angabe", spalte)
	}

	return wert
}

func (z *zeilenleser) anschrift() service.Anschrift {
	return service.Anschrift{
		Adresse:      z.text(spalteAdresse),
		Postleitzahl: z.text(spaltePLZ),
		Ort:          z.text(spalteOrt),
	}
}

// nummer liest die Mitglieds-ID aus der Spalte „Mandatsreferenz". Trotz des
// Namens steht dort keine SEPA-Mandatsreferenz, sondern die laufende Nummer, und
// sie wird als Mitglieds-ID übernommen (CONTEXT.md → Mitglieds-ID).
func (z *zeilenleser) nummer() int64 {
	wert := z.text(spalteNummer)
	if wert == "" {
		z.melden("in der Spalte „%s“ fehlt die Mitglieds-ID", spalteNummer)

		return 0
	}

	nummer, err := strconv.ParseInt(wert, 10, 64)
	if err != nil || nummer <= 0 {
		z.melden("die Mitglieds-ID „%s“ ist keine positive Zahl", wert)

		return 0
	}

	return nummer
}

// datumsformate sind die Schreibweisen, in denen ein getipptes Datum in der
// Tabelle stehen kann. Geraten wird daran nichts: was in keine davon passt und
// auch keine Seriennummer ist, kommt in den Bericht.
var datumsformate = []string{"2006-01-02", "02.01.2006", "2.1.2006", "02.01.06"}

// datum liest eine Datumszelle. Sie steht dort entweder als Excel-Seriennummer
// (Epoche 1899-12-30) oder als getippter Text — in der gewachsenen Tabelle des
// Vereins kommt beides vor, und beides muss denselben Tag ergeben.
//
// Eine leere Zelle ist kein Fehler: welche Datumsangaben Pflicht sind,
// entscheidet der Aufrufer.
func (z *zeilenleser) datum(spalte string) *time.Time {
	wert := z.text(spalte)
	if wert == "" {
		return nil
	}

	if serie, err := strconv.ParseFloat(wert, 64); err == nil {
		return z.seriennummer(spalte, wert, serie)
	}

	for _, format := range datumsformate {
		if d, err := time.Parse(format, wert); err == nil {
			return &d
		}
	}

	z.melden("in der Spalte „%s“ ist „%s“ kein lesbares Datum", spalte, wert)

	return nil
}

// Die Jahre, zwischen denen ein Datum aus dieser Tabelle liegen muss. Sie
// begrenzen keine Fachlichkeit, sondern fangen eine Zahl ab, die gar keine
// Seriennummer ist: ein Verein hat keine 1905 geborenen Mitglieder und keine
// Eintritte im Jahr 3000, und jede Zahl, die auf ein solches Datum führt, war in
// der Zelle etwas anderes gemeint.
const (
	fruehestesJahr = 1920
	spaetestesJahr = 2100
)

// seriennummer deutet eine Zahl als Excel-Seriennummer — aber nur, wenn dabei
// ein Datum herauskommt, das in dieser Tabelle stehen kann.
//
// Ohne diese Grenze wird jede Zahl klaglos zu einem Datum: „1990" ergäbe den
// 12. Juni 1905 und „60" den 28. Februar 1900, beides ohne ein Wort im Bericht.
// Das ist der stille Standardwert, den das Ticket ausschließt — nur besonders
// schwer zu bemerken, weil am Ende ein plausibel aussehendes Datum steht.
func (z *zeilenleser) seriennummer(spalte, wert string, serie float64) *time.Time {
	if jahreszahl(wert) {
		z.melden("in der Spalte „%s“ ist „%s“ nur eine Jahreszahl und kein vollständiges Datum",
			spalte, wert)

		return nil
	}

	d, err := excelize.ExcelDateToTime(serie, false)
	if err != nil || d.Year() < fruehestesJahr || d.Year() > spaetestesJahr {
		z.melden("in der Spalte „%s“ ergibt „%s“ kein Datum, das in dieser Tabelle stehen kann",
			spalte, wert)

		return nil
	}

	// Auf den Kalendertag zurückgeschnitten: eine Seriennummer kann einen
	// Nachkommateil für die Uhrzeit tragen, und der Verein führt Tage.
	tag := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)

	return &tag
}

// jahreszahl sagt, ob die Zelle nur ein Jahr enthält. Sie bekommt eine eigene
// Meldung, weil die Grenze oben sie ohnehin abfinge — aber mit der Auskunft, aus
// „1990" ließe sich kein Datum machen, könnte niemand etwas anfangen. Mit der
// Auskunft, dort fehle der Tag, schon.
//
// Die Bereiche überschneiden sich nicht: die Seriennummern 1900 bis 2100 liegen
// alle im Jahr 1905 und damit ohnehin außerhalb der Grenze.
func jahreszahl(wert string) bool {
	if len(wert) != 4 {
		return false
	}

	jahr, err := strconv.Atoi(wert)

	return err == nil && jahr >= 1900 && jahr <= spaetestesJahr
}

func (z *zeilenleser) beitrag() int64 {
	cents, err := service.BeitragAusEuro(z.text(spalteBeitrag))
	if err != nil {
		z.melden("in der Spalte „%s“: %s", spalteBeitrag, err)

		return 0
	}

	return cents
}

func (z *zeilenleser) anmeldegebuehr() int64 {
	cents, err := service.AnmeldegebuehrAusEuro(z.text(spalteAnmeldegebuehr))
	if err != nil {
		z.melden("in der Spalte „%s“: %s", spalteAnmeldegebuehr, err)

		return 0
	}

	return cents
}

// bewertungswerte ordnet den Zellinhalt der zweiwertigen Google-Bewertung zu.
// Welche Zeichen der Verein für „hat bewertet" benutzt, war beim Schreiben
// dieses Imports nicht bekannt — deshalb steht hier eine kurze Liste, und alles
// andere kommt in den Bericht, statt auf „nein" geraten zu werden.
var bewertungswerte = map[string]bool{
	"":         false,
	"❌":        false,
	"✖":        false,
	"x":        false,
	"nein":     false,
	"✅":        true,
	"✔":        true,
	"☑":        true,
	"ja":       true,
	"bewertet": true,
}

func (z *zeilenleser) bewertung() service.GoogleBewertung {
	// Das Variantenselektor-Zeichen hängt an Emoji, die aus Word oder aus dem
	// Browser in die Tabelle geraten sind; es gehört nicht zum Wert.
	wert := strings.ToLower(strings.ReplaceAll(z.text(spalteBewertung), "️", ""))

	bewertet, bekannt := bewertungswerte[wert]
	if !bekannt {
		z.melden("in der Spalte „%s“ ist „%s“ kein bekannter Wert", spalteBewertung, wert)

		return false
	}

	return service.GoogleBewertung(bewertet)
}

// keinTraining ist der Wert, mit dem die Tabelle „kein Termin" ausdrückt. Er
// ergibt keinen Termin — ein Trainingstermin namens „Kein" wäre einer, den es
// nicht gibt, und würde die Frequenz um eins zu hoch ablesen lassen.
const keinTraining = "kein"

// trainingstermine ordnet die Freitexte der drei Trainingsspalten den Terminen
// des Stundenplans zu (ADR-0008). Was sich nicht zuordnen lässt, kommt als
// Hinweis in den Bericht und fehlt der Mitgliedschaft — aufgehalten wird die
// Zeile davon nicht: eine fehlende Trainingszeit trägt der Verein in der App
// nach, ein nicht importiertes Mitglied müsste er ganz von Hand anlegen.
//
// Derselbe Termin in zwei Spalten zählt einmal. Er ist dieselbe Vereinbarung
// und keine zwei, und der Service hielte es ohnehin so
// (trainingsterminIDsNormalisieren); zweimal gezählt stünde er der Gegenprobe
// zur Frequenzspalte im Weg.
func (z *zeilenleser) trainingstermine() []int64 {
	var ids []int64

	for _, spalte := range trainingsspalten {
		wert := z.text(spalte)
		if wert == "" || strings.EqualFold(wert, keinTraining) {
			continue
		}

		id, meldung := z.plan.zuordnen(spalte, wert)
		if meldung != "" {
			z.hinweisen("%s", meldung)

			continue
		}

		if !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
	}

	return ids
}

// frequenzPruefen hält die Spalte „1x 2x Woche" gegen die Zahl der zugeordneten
// Termine. Übernommen wird sie nicht: die Trainingsfrequenz ist die Anzahl der
// Termine und nichts daneben (CONTEXT.md → Trainingsfrequenz). Weicht sie ab,
// steht das im Bericht — stillschweigend zu korrigieren hieße, eine der beiden
// Angaben wegzuwerfen, ohne zu wissen, welche stimmt.
//
// Der Widerspruch ist ein Hinweis und hält die Zeile nicht auf: seit dem
// Abgleich gegen den Stundenplan kann er allein daher rühren, dass ein Freitext
// dort nicht zu finden war — und dann wäre entgegen ADR-0008 doch wieder die
// ganze Zeile draußen.
//
// Eine unlesbare Frequenz bleibt dagegen ein Grund und damit ein Fehlerfall, wie
// vor diesem Abgleich: sie ist kein Widerspruch zwischen zwei Angaben, sondern
// eine Zelle, die niemand deuten kann — und der Importer errät nichts.
func (z *zeilenleser) frequenzPruefen(zugeordnet int) {
	wert := z.text(spalteFrequenz)
	if wert == "" {
		return
	}

	angegeben, err := strconv.Atoi(strings.TrimSpace(strings.Split(wert, "x")[0]))
	if err != nil {
		z.melden("in der Spalte „%s“ ist „%s“ keine lesbare Frequenz", spalteFrequenz, wert)

		return
	}

	if angegeben != zugeordnet {
		z.hinweisen("Frequenz %d× widerspricht %d zugeordneten Trainingsterminen",
			angegeben, zugeordnet)
	}
}

// Die Werte der Status-Spalte, wortwörtlich wie sie in der Werteliste des
// Blattes „Quelle" stehen — samt dem Tippfehler „Inakiv".
const (
	statusMitglied          = "mitglied"
	statusNeu               = "neu"
	statusAktiv             = "aktiv"
	statusStillgelegt       = "stillgelegt"
	statusInaktivTippfehler = "inakiv"
	statusInaktiv           = "inaktiv"
	statusGekuendigt        = "gekündigt"
	statusKuendigungsfrist  = "kündigungsfrist"
)

// lebenszyklus deutet die Spalten „Status" und „Gekündigt" zusammen.
//
// Der Status wird nicht gespeichert — er wird aus den Datumsfeldern abgelesen
// (CONTEXT.md → Status). Er entscheidet aber, wohin das Datum aus „Gekündigt"
// gehört: bei „Gekündigt" ist es der Austritt, bei „Kündigungsfrist" das
// Kündigungsdatum. Dieselbe Zahl bedeutet also je nach Nachbarzelle etwas
// anderes, und genau deshalb steht die Deutung an einer Stelle.
func (z *zeilenleser) lebenszyklus(satz *service.Importsatz) {
	roh := z.text(spalteStatus)
	if roh == "" {
		z.melden("in der Spalte „%s“ fehlt der Status", spalteStatus)

		return
	}

	gekuendigt := z.datum(spalteGekuendigt)

	// War die Zelle gefüllt, aber unlesbar, hat z.datum das bereits gemeldet.
	// Dann noch einmal „fehlt das Datum" zu sagen, wäre nicht nur doppelt,
	// sondern falsch: es steht ja etwas da.
	unlesbar := gekuendigt == nil && z.text(spalteGekuendigt) != ""

	switch status := strings.ToLower(roh); status {
	case statusMitglied, statusNeu, statusAktiv:
		// Ohne Wirkung: diese Zustände liest der Status ohnehin aus Eintritt
		// und Austritt ab. Ein Datum daneben wäre aber widersprüchlich.
		if gekuendigt != nil {
			z.melden("Status „%s“, aber in der Spalte „%s“ steht ein Datum", roh, spalteGekuendigt)
		}
	case statusStillgelegt, statusInaktivTippfehler, statusInaktiv:
		satz.Ruhend = true
		if gekuendigt != nil {
			z.melden("Status „%s“, aber in der Spalte „%s“ steht ein Datum", roh, spalteGekuendigt)
		}
	case statusGekuendigt:
		if gekuendigt == nil {
			if !unlesbar {
				z.melden("Status „%s“, aber in der Spalte „%s“ fehlt das Datum", roh, spalteGekuendigt)
			}

			return
		}
		satz.Kuendigung.Austritt = gekuendigt
	case statusKuendigungsfrist:
		if gekuendigt == nil {
			if !unlesbar {
				z.melden("Status „%s“, aber in der Spalte „%s“ fehlt das Datum", roh, spalteGekuendigt)
			}

			return
		}
		satz.Kuendigung.Datum = gekuendigt
	default:
		z.melden("unbekannter Status „%s“", roh)
	}
}
