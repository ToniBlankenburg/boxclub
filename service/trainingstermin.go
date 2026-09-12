package service

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Der Trainingstermin ist der wöchentliche Termin des Vereins, aus dem der
// Stundenplan besteht (CONTEXT.md → Trainingstermin, ADR-0008). Er existiert
// für sich: dass niemand dafür angemeldet ist, macht ihn nicht ungültig.
//
// Er ist ausdrücklich **kein Datum**. "Samstag 10:30" gilt bis auf Weiteres;
// den Termin am 15.09. gibt es als eigene Sache nicht, und damit gibt es auch
// keine Stelle, an der stehen könnte, wer da war — das ist die Grenze zum für
// v1 ausgeschlossenen attendance tracking, und sie liegt genau hier.

// Wochentag ist der Tag, an dem ein Termin stattfindet — gezählt wie in
// ISO 8601 (Montag = 1 … Sonntag = 7).
//
// Eine Zahl und kein Text, weil die Liste in Wochenreihenfolge steht: nach dem
// Alphabet sortiert stünde Dienstag vor Donnerstag vor Freitag vor Mittwoch,
// und der Stundenplan wäre als Stundenplan nicht mehr lesbar.
//
// Der Nullwert ist kein Tag: er heißt "nicht ausgewählt" und wird abgewiesen.
type Wochentag int

const (
	Montag Wochentag = iota + 1
	Dienstag
	Mittwoch
	Donnerstag
	Freitag
	Samstag
	Sonntag
)

// wochentagsnamen ist die Beschriftung je Tag, indiziert mit dem Wochentag
// selbst — Feld 0 bleibt deshalb leer.
var wochentagsnamen = [...]string{Montag: "Montag", Dienstag: "Dienstag", Mittwoch: "Mittwoch",
	Donnerstag: "Donnerstag", Freitag: "Freitag", Samstag: "Samstag", Sonntag: "Sonntag"}

// Bezeichnung ist der Text, den die Oberfläche zum Tag zeigt. Er steht wie bei
// Status.Bezeichnung im Service, damit Auswahlliste, Stundenplan und die
// spätere Anmeldung dieselben Worte benutzen.
func (w Wochentag) Bezeichnung() string {
	if !w.gueltig() {
		return "unbekannt"
	}

	return wochentagsnamen[w]
}

// gueltig sagt, ob der Wert einer der sieben Tage ist.
func (w Wochentag) gueltig() bool {
	return w >= Montag && w <= Sonntag
}

// Wochentage sind alle sieben Tage in Wochenreihenfolge — die Auswahlliste des
// Formulars. Sie wird hier aufgezählt und nicht im Template, damit Reihenfolge,
// Beschriftung und gespeicherter Wert aus derselben Quelle kommen.
func Wochentage() []Wochentag {
	return []Wochentag{Montag, Dienstag, Mittwoch, Donnerstag, Freitag, Samstag, Sonntag}
}

// uhrzeitLayout ist das Format, in dem Uhrzeiten als TEXT in SQLite liegen und
// in dem <input type="time"> sie liefert. Zweistellig und mit Doppelpunkt
// sortiert es lexikografisch richtig — deshalb kann die Liste nach der Spalte
// sortieren und pruefen mit einem Textvergleich auskommen.
const uhrzeitLayout = "15:04"

// Uhrzeit ist eine Tageszeit ohne Datum, in der Form "18:00". Der Nullwert ist
// der leere Text und heißt "nicht erfasst" — das ist beim Ende der Normalfall,
// nicht die Ausnahme.
//
// Ein eigener Typ und kein string, weil daran die Zusicherung hängt, dass jede
// gespeicherte Uhrzeit durch uhrzeitAus gegangen und damit vereinheitlicht ist:
// "9:05" steht als "09:05" da und nicht hinter "10:30".
type Uhrzeit string

// Leer sagt, ob die Uhrzeit nicht erfasst ist. Auch für das Template gedacht,
// das den Nullwert sonst über einen Textvergleich erkennen müsste.
func (u Uhrzeit) Leer() bool {
	return u == ""
}

// uhrzeitAus liest eine getippte Uhrzeit und gibt sie vereinheitlicht zurück.
// Umschließender Leerraum fällt weg, ein leerer Text ergibt die leere Uhrzeit —
// "nicht erfasst" ist keine Fehleingabe, sondern beim Ende die Regel.
func uhrzeitAus(roh string) (Uhrzeit, error) {
	roh = strings.TrimSpace(roh)
	if roh == "" {
		return "", nil
	}

	t, err := time.Parse(uhrzeitLayout, roh)
	if err != nil {
		return "", fmt.Errorf("%q ist keine Uhrzeit", roh)
	}

	return Uhrzeit(t.Format(uhrzeitLayout)), nil
}

// Trainingstermin ist ein Eintrag des Stundenplans, so wie er gelesen wird.
type Trainingstermin struct {
	ID        int64
	Wochentag Wochentag

	// Beginn ist Pflicht, Ende freiwillig: wann das Training anfängt, weiß der
	// Verein immer, wann es endet nicht unbedingt.
	Beginn Uhrzeit
	Ende   Uhrzeit

	// Bezeichnung ist freier Text ("Anfänger", "Wettkampf"). Trainer und Ort
	// stehen bewusst nicht daneben, sondern passen hier hinein, solange es eine
	// Halle gibt (ADR-0008).
	Bezeichnung string

	// Archiviert heißt: es gibt den Termin nicht mehr, gelöscht wird er trotzdem
	// nie (ADR-0008). Er verschwindet aus Liste und Auswahl, bestehende
	// Anmeldungen bleiben lesbar und zählen weiter zur Trainingsfrequenz.
	Archiviert bool
}

// Anzeige ist die eine Schreibweise des Termins: "Samstag 10:30 – 12:00 ·
// Anfänger", bei fehlenden Angaben entsprechend kürzer. Sie steht im Service
// und nicht im Template, weil derselbe Text im Stundenplan, in der Auswahl der
// Mitgliedschaft und in der Mitgliederliste erscheint — drei Stellen, an denen
// er nicht dreimal verschieden entstehen soll.
func (t Trainingstermin) Anzeige() string {
	zeit := string(t.Beginn)
	if !t.Ende.Leer() {
		zeit += " – " + string(t.Ende)
	}

	anzeige := t.Wochentag.Bezeichnung() + " " + zeit
	if t.Bezeichnung != "" {
		anzeige += " · " + t.Bezeichnung
	}

	return anzeige
}

// Trainingsterminangabe ist, was über einen Termin erfasst wird — die Rohwerte
// des Formulars. Beginn und Ende stehen als Text darin und nicht als Uhrzeit:
// geprüft und vereinheitlicht wird im Service, damit dieselbe Regel nicht auch
// noch im Handler steht.
//
// Anlegen und Ändern nehmen dieselbe Angabe: beide erfassen den Termin
// vollständig, ein Patch-Typ wie bei den Stammdaten wäre hier vier Felder für
// nichts.
type Trainingsterminangabe struct {
	Wochentag   Wochentag
	Beginn      string
	Ende        string
	Bezeichnung string
}

// Die Meldungen zu den Regeln, die ein Termin kennt — fertige Sätze für das
// Formular, wie überall im Service.
const (
	fehlenderWochentag = "Bitte einen Wochentag auswählen."
	fehlenderBeginn    = "Der Beginn ist eine Pflichtangabe."
	unlesbarerBeginn   = "Der Beginn ist keine gültige Uhrzeit."
	unlesbaresEnde     = "Das Ende ist keine gültige Uhrzeit."
	endeVorBeginn      = "Das Ende darf nicht vor dem Beginn liegen."
)

// pruefen macht aus den Rohwerten einen Termin oder sammelt alle Verstöße auf
// einmal — wer sich in zwei Feldern vertan hat, soll das nicht nacheinander
// erfahren.
//
// Verglichen wird die vereinheitlichte Textform: "09:05" < "10:30" gilt
// lexikografisch wie zeitlich, dafür ist das Format gemacht. Ein Termin über
// Mitternacht ("22:00 – 00:30") fällt damit durch — das ist gewollt: im
// Boxverein ist das ein Tippfehler und keine Nachtschicht.
//
// Gleicher Zeitpunkt ist kein Ende *vor* dem Beginn und geht durch (Ticket 20).
func (a Trainingsterminangabe) pruefen() (Trainingstermin, error) {
	termin := Trainingstermin{
		Wochentag:   a.Wochentag,
		Bezeichnung: strings.TrimSpace(a.Bezeichnung),
	}

	var meldungen []string

	if !a.Wochentag.gueltig() {
		meldungen = append(meldungen, fehlenderWochentag)
	}

	beginn, err := uhrzeitAus(a.Beginn)
	switch {
	case err != nil:
		meldungen = append(meldungen, unlesbarerBeginn)
	case beginn.Leer():
		meldungen = append(meldungen, fehlenderBeginn)
	default:
		termin.Beginn = beginn
	}

	ende, err := uhrzeitAus(a.Ende)
	if err != nil {
		meldungen = append(meldungen, unlesbaresEnde)
	} else {
		termin.Ende = ende
	}

	// Nur vergleichbar, wenn beide Uhrzeiten gelesen werden konnten — sonst
	// stünde neben "keine gültige Uhrzeit" noch eine Aussage über sie.
	if !termin.Beginn.Leer() && !termin.Ende.Leer() && termin.Ende < termin.Beginn {
		meldungen = append(meldungen, endeVorBeginn)
	}

	if len(meldungen) > 0 {
		return Trainingstermin{}, &ValidierungsFehler{Meldungen: meldungen}
	}

	return termin, nil
}

// CreateTrainingstermin legt einen Eintrag des Stundenplans an und liefert
// dessen ID. Ein neuer Termin ist nie archiviert.
//
// Doppelte Termine weist nichts ab: derselbe Wochentag zur selben Zeit kann
// zweimal dastehen, wenn zwei Gruppen parallel trainieren. Zu entscheiden, ob
// das ein Versehen ist, kann nur der Verein.
func (s *MemberService) CreateTrainingstermin(a Trainingsterminangabe) (int64, error) {
	termin, err := a.pruefen()
	if err != nil {
		return 0, err
	}

	res, err := s.db.Exec(
		`INSERT INTO trainingstermin (wochentag, beginn, ende, bezeichnung, archiviert)
		 VALUES (?, ?, ?, ?, 0)`,
		int(termin.Wochentag), string(termin.Beginn), alsUhrzeitText(termin.Ende), termin.Bezeichnung)
	if err != nil {
		return 0, fmt.Errorf("trainingstermin anlegen: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("vergebene trainingstermin-id lesen: %w", err)
	}

	return id, nil
}

// UpdateTrainingstermin ersetzt alle vier Angaben eines Termins. Das Formular
// schickt sie immer vollständig; ein geleertes Ende heißt deshalb "es ist
// keines mehr erfasst" und nicht "unverändert".
//
// Das Archiv-Kennzeichen bleibt unberührt: ein archivierter Termin lässt sich
// korrigieren, ohne dadurch zurück in die Auswahl zu rutschen. Darüber
// entscheidet allein SetTrainingsterminArchiviert.
//
// Eine Zeitänderung wirkt auf alle angemeldeten Mitgliedschaften zugleich —
// genau dafür gibt es den Stundenplan (ADR-0008).
func (s *MemberService) UpdateTrainingstermin(id int64, a Trainingsterminangabe) error {
	termin, err := a.pruefen()
	if err != nil {
		return err
	}

	res, err := s.db.Exec(
		`UPDATE trainingstermin SET wochentag = ?, beginn = ?, ende = ?, bezeichnung = ? WHERE id = ?`,
		int(termin.Wochentag), string(termin.Beginn), alsUhrzeitText(termin.Ende), termin.Bezeichnung, id)
	if err != nil {
		return fmt.Errorf("trainingstermin %d ändern: %w", id, err)
	}

	return betroffenPruefen(res, id)
}

// SetTrainingsterminArchiviert nimmt einen Termin aus dem Stundenplan oder holt
// ihn zurück. Gelöscht wird nie (ADR-0008) — und weil ein Fehlgriff sonst
// endgültig wäre, führt derselbe Weg auch wieder zurück.
//
// Der Aufruf ist idempotent: denselben Wert noch einmal geschrieben ändert
// nichts und ist kein Fehler.
func (s *MemberService) SetTrainingsterminArchiviert(id int64, archiviert bool) error {
	res, err := s.db.Exec(`UPDATE trainingstermin SET archiviert = ? WHERE id = ?`, archiviert, id)
	if err != nil {
		return fmt.Errorf("trainingstermin %d archivieren: %w", id, err)
	}

	return betroffenPruefen(res, id)
}

// GetTrainingstermin liest einen einzelnen Termin — der Weg ins
// Bearbeitungsformular. Anders als die Liste liefert er auch archivierte
// Einträge: bearbeiten und reaktivieren lässt sich nur, was man vorher lesen
// kann.
func (s *MemberService) GetTrainingstermin(id int64) (Trainingstermin, error) {
	termin, err := trainingsterminLesen(s.db.QueryRow(
		`SELECT id, wochentag, beginn, ende, bezeichnung, archiviert
		 FROM trainingstermin WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return Trainingstermin{}, fmt.Errorf("trainingstermin %d: %w", id, ErrNichtGefunden)
	}
	if err != nil {
		return Trainingstermin{}, fmt.Errorf("trainingstermin %d lesen: %w", id, err)
	}

	return termin, nil
}

// ListTrainingstermine liefert den Stundenplan in Wochenreihenfolge: Montag vor
// Dienstag, innerhalb des Tages nach Beginn. Sortiert die Datenbank, denn beide
// Spalten sortieren in ihrer gespeicherten Form richtig — der Wochentag als
// Zahl, die Uhrzeit als "HH:MM". Bei den Namen geht das nicht und muss Go
// erledigen (nachNamenSortieren); hier wäre es überflüssige Arbeit.
//
// Archivierte Termine sind standardmäßig nicht dabei — sie sind aus dem
// Stundenplan genommen. Mit auchArchivierte stehen sie an ihrem Platz in der
// Woche mit darin: wer einen zurückholen will, sucht ihn dort, wo er hingehört.
func (s *MemberService) ListTrainingstermine(auchArchivierte bool) ([]Trainingstermin, error) {
	bedingung := "WHERE archiviert = 0"
	if auchArchivierte {
		bedingung = ""
	}

	zeilen, err := s.db.Query(fmt.Sprintf(
		`SELECT id, wochentag, beginn, ende, bezeichnung, archiviert
		 FROM trainingstermin %s
		 ORDER BY wochentag, beginn, bezeichnung, id`, bedingung))
	if err != nil {
		return nil, fmt.Errorf("trainingstermine lesen: %w", err)
	}
	defer zeilen.Close()

	termine := []Trainingstermin{}
	for zeilen.Next() {
		termin, err := trainingsterminLesen(zeilen)
		if err != nil {
			return nil, fmt.Errorf("trainingstermin lesen: %w", err)
		}
		termine = append(termine, termin)
	}

	if err := zeilen.Err(); err != nil {
		return nil, fmt.Errorf("trainingstermine lesen: %w", err)
	}

	return termine, nil
}

// trainingsterminLesen füllt einen Termin aus einer Ergebniszeile. Einzeilige
// und mehrzeilige Abfragen teilen sich die Funktion über das gemeinsame Scan —
// die Spaltenliste steht damit nur einmal in Go.
func trainingsterminLesen(zeile interface{ Scan(...any) error }) (Trainingstermin, error) {
	var (
		termin Trainingstermin
		ende   sql.NullString
	)

	if err := zeile.Scan(&termin.ID, &termin.Wochentag, &termin.Beginn, &ende,
		&termin.Bezeichnung, &termin.Archiviert); err != nil {
		return Trainingstermin{}, err
	}

	termin.Ende = Uhrzeit(ende.String)

	return termin, nil
}

// alsUhrzeitText bereitet eine optionale Uhrzeit für die Ablage in SQLite auf:
// die leere Uhrzeit wird NULL, wie bei alsDatumsText. So steht "nicht erfasst"
// in der Datenbank genauso da wie im Modell und nicht als leerer Text daneben.
func alsUhrzeitText(u Uhrzeit) any {
	if u.Leer() {
		return nil
	}

	return string(u)
}

// betroffenPruefen macht aus "kein Datensatz geändert" den Fehler
// ErrNichtGefunden — bei UPDATE der einzige Weg, eine unbekannte ID überhaupt
// zu bemerken.
func betroffenPruefen(res sql.Result, id int64) error {
	betroffen, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("betroffene zeilen lesen: %w", err)
	}
	if betroffen == 0 {
		return fmt.Errorf("trainingstermin %d: %w", id, ErrNichtGefunden)
	}

	return nil
}
