package service

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Dokumente sind die PDFs, die der Verein zu einem Mitglied aufhebt (CONTEXT.md
// → Dokument). Es gibt genau zwei Arten, und beide haben ihren eigenen Ort: der
// Vertrag hängt an der Mitgliedschaft, die Rechnung an der Person.
//
// Sie liegen als Blob in der Datenbank und nicht als Dateien daneben
// (ADR-0007). Daraus folgt die eine Regel, an die sich dieses Package halten
// muss: der Blob darf in keiner Abfrage stehen, die mehr als eine Zeile liefert.
// Das Schema erzwingt sie nicht — die Typen hier tun es, so weit Typen das
// können: Dokument trägt keinen Inhalt, Dokumentinhalt trägt ihn, und den holt
// nur VertragInhalt.

// Dokumentart sagt, was für ein Papier abgelegt ist. Mehr als diese zwei kennt
// v1 nicht: die Ablage ist ein Ort für Vereinsunterlagen und kein allgemeines
// Dateifach.
type Dokumentart string

const (
	// ArtVertrag ist die unterschriebene Vereinbarung über eine Mitgliedschaft,
	// eingescannt und hochgeladen.
	ArtVertrag Dokumentart = "vertrag"

	// ArtRechnung ist die vom Verein geschriebene Rechnung über eine Leistung
	// neben dem Beitrag. Sie erzeugt die App selbst und legt sie am Mitglied ab
	// (service.RechnungErstellen).
	ArtRechnung Dokumentart = "rechnung"
)

// ErrKeinVertrag meldet, dass zu einem Zeitraum kein Vertrag abgelegt ist. Es
// ist ein ErrNichtGefunden — wer nur wissen will, ob es die Sache gibt, prüft
// weiterhin darauf —, sagt aber zusätzlich, *was* fehlt: die Ansicht soll nicht
// "diesen Zeitraum gibt es nicht mehr" melden, wenn bloß kein Vertrag daran
// hängt.
var ErrKeinVertrag = fmt.Errorf("kein vertrag abgelegt: %w", ErrNichtGefunden)

// Dokument ist die Auskunft *über* ein abgelegtes PDF — ohne das PDF. Das ist
// die Form, in der Dokumente in Listen und Ansichten reisen; das Blob bleibt in
// der Datenbank, bis jemand es ausdrücklich holt.
type Dokument struct {
	ID   int64
	Art  Dokumentart
	Name string

	// AbgelegtAm ist der Zeitpunkt, zu dem dieses PDF in die Ablage kam. Beim
	// Ersetzen ist es der Zeitpunkt des Ersetzens: abgelegt ist, was jetzt da
	// liegt, und eine Versionshistorie führt v1 nicht.
	AbgelegtAm time.Time
}

// Dokumentinhalt ist ein Dokument samt seinem PDF. Es gibt genau einen Weg
// hierher — eine Abfrage nach genau einer ID —, und das ist Absicht.
type Dokumentinhalt struct {
	Dokument
	Inhalt []byte
}

// NeuesDokument ist eine hochgeladene Datei, bevor die Ablage sie annimmt.
type NeuesDokument struct {
	// Name ist der Dateiname, wie der Browser ihn mitgeschickt hat. Er wird
	// aufgehoben, um ihn beim Export wieder vorzuschlagen — sonst hieße die
	// exportierte Datei nach nichts.
	Name string

	// Inhalt ist die Datei. Ob sie ein PDF ist, entscheidet ihr Anfang und
	// nicht ihr Name (siehe pruefen).
	Inhalt []byte
}

// MaxDokumentBytes ist die Obergrenze für ein abgelegtes PDF: 10 MB. Ein
// gescannter Vertrag liegt bei ein paar hundert Kilobyte (ADR-0007), ein
// versehentlich ausgewähltes Video bei einigen hundert Megabyte — die Grenze
// trennt das eine vom anderen, lange bevor die Datenbank es merkt.
const MaxDokumentBytes = 10 << 20

// standardname springt ein, wenn eine Datei ohne brauchbaren Namen ankommt.
// Deswegen ein gültiges PDF abzuweisen wäre unverhältnismäßig: der Name ist
// eine Bequemlichkeit für den Export und keine Angabe, die der Verein pflegt.
const standardname = "dokument.pdf"

// pdfKennung sind die Zeichen, mit denen jede PDF-Datei beginnt. Geprüft wird
// der Anfang der Datei und nicht die Endung des Namens: eine umbenannte
// Word-Datei hieße sonst ".pdf" und läge unlesbar in der Datenbank, und
// umgekehrt ist ein PDF ohne Endung eines.
var pdfKennung = []byte("%PDF-")

// keinPDF sagt, warum eine Datei nicht angenommen wurde, und nennt sie beim
// Namen — hochgeladen wird aus einem Dateidialog heraus, und welche der
// ausgewählten Dateien gemeint ist, soll niemand raten müssen.
func keinPDF(name string) Meldung {
	return meldung("validierung.dokument.kein_pdf", name)
}

// zuGross ist die eigene Meldung für die Größengrenze. Sie nennt beide Zahlen:
// ohne die tatsächliche Größe bleibt unklar, ob die Datei knapp darüber liegt
// oder ob die falsche ausgewählt wurde.
func zuGross(name string, groesse int) Meldung {
	return meldung("validierung.dokument.zu_gross", name, megabyte(groesse), Dokumentgrenze())
}

// Dokumentgrenze schreibt die Obergrenze so, wie sie dem Verein gezeigt wird.
//
// Sie steht hier und nicht bei jedem, der sie zeigt: die Meldung des Service,
// der Hinweis unter dem Datei-Dialog und die Absage bei zu großem Upload
// sprechen von derselben Grenze, und drei Schreibweisen derselben Zahl sind
// eine zu viel — spätestens, wenn jemand die Grenze ändert.
func Dokumentgrenze() string {
	return megabyte(MaxDokumentBytes)
}

// megabyte schreibt eine Byte-Zahl so, wie sie in der Meldung stehen soll — mit
// Dezimalkomma, und ohne Nachkomma, wo es nichts zu sagen hat: die Grenze heißt
// "10 MB" und nicht "10,0 MB", eine tatsächliche Dateigröße dagegen "10,4 MB".
func megabyte(bytes int) string {
	if bytes%(1<<20) == 0 {
		return fmt.Sprintf("%d MB", bytes>>20)
	}

	return strings.Replace(fmt.Sprintf("%.1f MB", float64(bytes)/(1<<20)), ".", ",", 1)
}

// bereinigt macht aus dem, was der Browser mitgeschickt hat, einen Namen, den
// die Ablage aufheben kann.
//
// Verzeichnisse fallen weg: der Name wird beim Export als Ziel vorgeschlagen,
// und ein Pfad darin führte dort hin, wo die Datei einmal herkam. Getrennt wird
// an beiden Zeichen, weil der Name von einem fremden Betriebssystem stammen
// kann und nicht von dem, auf dem die App gerade läuft.
func (n NeuesDokument) bereinigt() NeuesDokument {
	name := strings.TrimSpace(n.Name)
	if schnitt := strings.LastIndexAny(name, `/\`); schnitt >= 0 {
		name = name[schnitt+1:]
	}
	if name == "" {
		name = standardname
	}

	n.Name = name

	return n
}

// pruefen sammelt beide Gründe, aus denen eine Datei nicht in die Ablage kommt.
// Beide auf einmal, weil sie unabhängig voneinander sind: wer ein 400 MB großes
// Video auswählt, soll nicht erst erfahren, dass es zu groß ist, und nach dem
// Verkleinern, dass es kein PDF war.
func (n NeuesDokument) pruefen() error {
	var meldungen []Meldung

	if len(n.Inhalt) > MaxDokumentBytes {
		meldungen = append(meldungen, zuGross(n.Name, len(n.Inhalt)))
	}
	if !bytes.HasPrefix(n.Inhalt, pdfKennung) {
		meldungen = append(meldungen, keinPDF(n.Name))
	}

	if len(meldungen) > 0 {
		return &ValidierungsFehler{Meldungen: meldungen}
	}

	return nil
}

// isoZeitpunkt ist das Layout, in dem der Zeitpunkt der Ablage als TEXT in
// SQLite liegt. Anders als die Datumsspalten braucht er die Uhrzeit: zwei
// Verträge desselben Tages sollen sich unterscheiden lassen.
const isoZeitpunkt = time.RFC3339

// VertragAblegen legt den eingescannten Vertrag über einen Zeitraum ab und
// ersetzt dabei den, der schon da liegt.
//
// Ersetzen ist kein eigener Aufruf, weil es kein eigener Vorgang ist: ein
// Zeitraum hat einen Vertrag, und wer einen zweiten hochlädt, hat den ersten
// falsch eingescannt. Papierkorb und Versionen gibt es nicht — das Papier liegt
// beim Verein ohnehin noch (CONTEXT.md → Dokument).
func (s *MemberService) VertragAblegen(mitgliedschaftID int64, neu NeuesDokument) error {
	neu = neu.bereinigt()
	if err := neu.pruefen(); err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("transaktion starten: %w", err)
	}
	defer tx.Rollback()

	if err := mitgliedschaftPruefen(tx, mitgliedschaftID); err != nil {
		return err
	}

	// Wegnehmen und neu einsetzen statt eines Upserts: es gibt keine ID, auf die
	// sich ein Upsert beziehen könnte, und der eindeutige Index über
	// mitgliedschaft_id ließe ein zweites INSERT ohnehin nicht zu.
	if _, err := tx.Exec(
		`DELETE FROM dokument WHERE mitgliedschaft_id = ? AND art = ?`,
		mitgliedschaftID, string(ArtVertrag)); err != nil {
		return fmt.Errorf("vertrag von mitgliedschaft %d entfernen: %w", mitgliedschaftID, err)
	}

	if _, err := tx.Exec(
		`INSERT INTO dokument (mitgliedschaft_id, art, dateiname, inhalt, erstellt_am)
		 VALUES (?, ?, ?, ?, ?)`,
		mitgliedschaftID, string(ArtVertrag), neu.Name, neu.Inhalt,
		time.Now().Format(isoZeitpunkt)); err != nil {
		return fmt.Errorf("vertrag zu mitgliedschaft %d ablegen: %w", mitgliedschaftID, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("vertrag ablegen abschließen: %w", err)
	}

	return nil
}

// VertragEntfernen nimmt den Vertrag eines Zeitraums aus der Ablage. Liegt
// keiner da, ist der Fehler ErrNichtGefunden — die Schaltfläche steht nur, wo
// einer liegt, und ein Klick aus einer veralteten Ansicht soll sich als solcher
// zu erkennen geben.
//
// Es gibt keine Rückfrage und keinen Papierkorb: das Papier liegt beim Verein,
// und ein zweites Einscannen ist weniger Aufwand als eine Wiederherstellung, die
// niemand gebaut hat.
func (s *MemberService) VertragEntfernen(mitgliedschaftID int64) error {
	ergebnis, err := s.db.Exec(
		`DELETE FROM dokument WHERE mitgliedschaft_id = ? AND art = ?`,
		mitgliedschaftID, string(ArtVertrag))
	if err != nil {
		return fmt.Errorf("vertrag von mitgliedschaft %d entfernen: %w", mitgliedschaftID, err)
	}

	entfernt, err := ergebnis.RowsAffected()
	if err != nil {
		return fmt.Errorf("vertrag von mitgliedschaft %d entfernen: %w", mitgliedschaftID, err)
	}
	if entfernt == 0 {
		return fmt.Errorf("mitgliedschaft %d: %w", mitgliedschaftID, ErrKeinVertrag)
	}

	return nil
}

// VertragInhalt holt den Vertrag eines Zeitraums samt seinem PDF. Liegt keiner
// da, ist der Fehler ErrNichtGefunden.
//
// Dies ist die einzige Abfrage im ganzen Service, die inhalt liest — und sie
// liefert genau eine Zeile. Ohne sie käme niemand mehr an seinen Vertrag: die
// Datenbank ist der einzige Ort, an dem er liegt (ADR-0007).
func (s *MemberService) VertragInhalt(mitgliedschaftID int64) (Dokumentinhalt, error) {
	var (
		dokument   Dokumentinhalt
		art        string
		abgelegtAm string
	)

	err := s.db.QueryRow(
		`SELECT id, art, dateiname, erstellt_am, inhalt
		 FROM dokument WHERE mitgliedschaft_id = ? AND art = ?`,
		mitgliedschaftID, string(ArtVertrag)).
		Scan(&dokument.ID, &art, &dokument.Name, &abgelegtAm, &dokument.Inhalt)
	if errors.Is(err, sql.ErrNoRows) {
		return Dokumentinhalt{}, fmt.Errorf("mitgliedschaft %d: %w", mitgliedschaftID, ErrKeinVertrag)
	}
	if err != nil {
		return Dokumentinhalt{}, fmt.Errorf("vertrag von mitgliedschaft %d lesen: %w", mitgliedschaftID, err)
	}

	dokument.Art = Dokumentart(art)
	if dokument.AbgelegtAm, err = time.Parse(isoZeitpunkt, abgelegtAm); err != nil {
		return Dokumentinhalt{}, fmt.Errorf("ablagezeitpunkt von dokument %d: %w", dokument.ID, err)
	}

	return dokument, nil
}

// vertraegeLesen liefert die Verträge der angegebenen Mitgliedschaften, je
// Zeitraum höchstens einen.
//
// Ohne inhalt — das ist der Punkt. Es ist die einzige Abfrage auf dokument, die
// mehrere Zeilen liefern kann, und sie holt nur, was darüber zu sagen ist: Name,
// Art und Zeitpunkt. Ein SELECT *, das die Blobs mitzöge, machte jedes Öffnen
// eines Mitglieds so teuer wie den Export all seiner Verträge — und zwar erst
// dann, wenn Daten drin sind (ADR-0007).
func (s *MemberService) vertraegeLesen(mitgliedschaftIDs []int64) (map[int64]Dokument, error) {
	vertraege := make(map[int64]Dokument, len(mitgliedschaftIDs))
	if len(mitgliedschaftIDs) == 0 {
		return vertraege, nil
	}

	platzhalter := strings.Repeat(", ?", len(mitgliedschaftIDs)-1)
	werte := make([]any, 0, len(mitgliedschaftIDs))
	for _, id := range mitgliedschaftIDs {
		werte = append(werte, id)
	}

	// Der Filter auf die Art ist heute überflüssig — an einer Mitgliedschaft
	// hängt nur der Vertrag, eine Rechnung hängt an der Person (ADR-0009). Er
	// steht trotzdem da, an allen vier Stellen: eine Funktion namens "Vertrag"
	// soll den Vertrag liefern und löschen und nicht, was immer dort hängt.
	werte = append(werte, string(ArtVertrag))

	zeilen, err := s.db.Query(
		`SELECT mitgliedschaft_id, id, art, dateiname, erstellt_am
		 FROM dokument WHERE mitgliedschaft_id IN (?`+platzhalter+`) AND art = ?`, werte...)
	if err != nil {
		return nil, fmt.Errorf("verträge der mitgliedschaften lesen: %w", err)
	}
	defer zeilen.Close()

	for zeilen.Next() {
		var (
			mitgliedschaftID int64
			dokument         Dokument
			art              string
			abgelegtAm       string
		)
		if err := zeilen.Scan(&mitgliedschaftID, &dokument.ID, &art, &dokument.Name, &abgelegtAm); err != nil {
			return nil, fmt.Errorf("vertrag einer mitgliedschaft lesen: %w", err)
		}

		dokument.Art = Dokumentart(art)
		if dokument.AbgelegtAm, err = time.Parse(isoZeitpunkt, abgelegtAm); err != nil {
			return nil, fmt.Errorf("ablagezeitpunkt von dokument %d: %w", dokument.ID, err)
		}

		vertraege[mitgliedschaftID] = dokument
	}

	return vertraege, zeilen.Err()
}

// rechnungAblegen legt ein erzeugtes Rechnungs-PDF am Mitglied ab und liefert
// die vergebene Dokument-ID.
//
// Anders als beim Vertrag gibt es hier kein Ersetzen und keinen eindeutigen
// Index: von Rechnungen gibt es je Mitglied beliebig viele (schema), jede ihr
// eigenes Dokument. Aufgerufen wird das nur aus RechnungErstellen heraus, das
// zuerst das PDF erzeugt und es dann, wenn ein Mitglied dahintersteht, hier
// ablegt.
func (s *MemberService) rechnungAblegen(mitgliedID int64, nummer string, pdf []byte) (int64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("transaktion starten: %w", err)
	}
	defer tx.Rollback()

	if err := mitgliedPruefen(tx, mitgliedID); err != nil {
		return 0, err
	}

	res, err := tx.Exec(
		`INSERT INTO dokument (mitglied_id, art, dateiname, inhalt, erstellt_am)
		 VALUES (?, ?, ?, ?, ?)`,
		mitgliedID, string(ArtRechnung), rechnungsDateiname(nummer), pdf, time.Now().Format(isoZeitpunkt))
	if err != nil {
		return 0, fmt.Errorf("rechnung zu mitglied %d ablegen: %w", mitgliedID, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("dokument-id lesen: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("rechnung ablegen abschließen: %w", err)
	}

	return id, nil
}

// rechnungsDateiname baut den Dateinamen aus der Rechnungsnummer — derselben
// Nummer, unter der der Verein die Rechnung in seiner Buchhaltung führt.
func rechnungsDateiname(nummer string) string {
	return fmt.Sprintf("Rechnung %s.pdf", strings.TrimSpace(nummer))
}

// RechnungenDesMitglieds liefert die Rechnungen, die an einem Mitglied abgelegt
// sind, neueste zuerst — ohne ihr PDF, aus demselben Grund wie vertraegeLesen:
// es ist die einzige Abfrage auf dokument, die mehrere Zeilen liefern kann, und
// ein SELECT *, das die Blobs mitzöge, machte jedes Öffnen eines Mitglieds so
// teuer wie den Export all seiner Rechnungen (ADR-0007).
func (s *MemberService) RechnungenDesMitglieds(mitgliedID int64) ([]Dokument, error) {
	zeilen, err := s.db.Query(
		`SELECT id, art, dateiname, erstellt_am
		 FROM dokument WHERE mitglied_id = ? AND art = ?
		 ORDER BY erstellt_am DESC`, mitgliedID, string(ArtRechnung))
	if err != nil {
		return nil, fmt.Errorf("rechnungen von mitglied %d lesen: %w", mitgliedID, err)
	}
	defer zeilen.Close()

	var rechnungen []Dokument
	for zeilen.Next() {
		var (
			d          Dokument
			art        string
			abgelegtAm string
		)
		if err := zeilen.Scan(&d.ID, &art, &d.Name, &abgelegtAm); err != nil {
			return nil, fmt.Errorf("rechnung eines mitglieds lesen: %w", err)
		}

		d.Art = Dokumentart(art)
		if d.AbgelegtAm, err = time.Parse(isoZeitpunkt, abgelegtAm); err != nil {
			return nil, fmt.Errorf("ablagezeitpunkt von dokument %d: %w", d.ID, err)
		}

		rechnungen = append(rechnungen, d)
	}

	return rechnungen, zeilen.Err()
}

// RechnungInhalt holt eine abgelegte Rechnung samt ihrem PDF — dieselbe Bauart
// wie VertragInhalt: die einzige Abfrage, die inhalt für eine Rechnung liest,
// und sie liefert genau eine Zeile.
func (s *MemberService) RechnungInhalt(dokumentID int64) (Dokumentinhalt, error) {
	var (
		dokument   Dokumentinhalt
		art        string
		abgelegtAm string
	)

	err := s.db.QueryRow(
		`SELECT id, art, dateiname, erstellt_am, inhalt
		 FROM dokument WHERE id = ? AND art = ?`,
		dokumentID, string(ArtRechnung)).
		Scan(&dokument.ID, &art, &dokument.Name, &abgelegtAm, &dokument.Inhalt)
	if errors.Is(err, sql.ErrNoRows) {
		return Dokumentinhalt{}, fmt.Errorf("rechnung %d: %w", dokumentID, ErrNichtGefunden)
	}
	if err != nil {
		return Dokumentinhalt{}, fmt.Errorf("rechnung %d lesen: %w", dokumentID, err)
	}

	dokument.Art = Dokumentart(art)
	if dokument.AbgelegtAm, err = time.Parse(isoZeitpunkt, abgelegtAm); err != nil {
		return Dokumentinhalt{}, fmt.Errorf("ablagezeitpunkt von dokument %d: %w", dokument.ID, err)
	}

	return dokument, nil
}

// mitgliedschaftPruefen meldet ErrNichtGefunden, wenn zu der ID kein Zeitraum
// vorliegt. Ohne diese Prüfung liefe das Ablegen in einen Fremdschlüsselfehler,
// und der sagte dem Verein nichts.
func mitgliedschaftPruefen(q abfrager, id int64) error {
	var vorhanden int

	err := q.QueryRow(`SELECT 1 FROM mitgliedschaft WHERE id = ?`, id).Scan(&vorhanden)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("mitgliedschaft %d: %w", id, ErrNichtGefunden)
	}
	if err != nil {
		return fmt.Errorf("mitgliedschaft lesen: %w", err)
	}

	return nil
}
