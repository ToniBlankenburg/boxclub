// Package service enthält die gesamte Business-Logik der Mitgliederverwaltung.
//
// Der MemberService spricht direkt mit database/sql gegen SQLite — bewusst ohne
// Repository-Interface und ohne getrenntes Domain-/Persistenz-Modell
// (siehe .scratch/boxclub-v1/spec.md → "Bewusst nicht modelliert").
package service

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
	_ "modernc.org/sqlite"
)

// isoDatum ist das Layout, in dem alle Datumsspalten als TEXT in SQLite liegen.
// Es sortiert lexikografisch korrekt und lässt sich in SQL direkt vergleichen.
const isoDatum = "2006-01-02"

// Mitglied ist die natürliche Person, die dem Verein bekannt ist. Der Datensatz
// bleibt derselbe, auch wenn die Person zwischenzeitlich aus- und wieder eintritt
// — die zeitliche Zuordnung steckt in Mitgliedschaften.
type Mitglied struct {
	ID           int64
	Vorname      string
	Nachname     string
	Geburtsdatum *time.Time
	Anschrift    Anschrift
	Email        string
	Telefon      string

	// IBAN steht als reiner Text da: die App speichert sie, prüft sie aber
	// nicht und erzeugt keine Lastschrift-Dateien (ADR-0006). Eine Prüfziffer
	// zu verlangen hieße, eine unvollständig abgetippte IBAN gar nicht erst
	// festhalten zu können — und genau das tut die Excel heute.
	IBAN string

	// Geschlecht ist Freitext. Die Werteliste der Excel ist eine Eintipphilfe
	// (siehe GeschlechtVorschlaege), keine Einschränkung.
	Geschlecht string

	// GoogleBewertung ist zweiwertig (CONTEXT.md → Google-Bewertung).
	GoogleBewertung GoogleBewertung

	// Digital ist die gleichnamige Excel-Spalte, wortwörtlich als Freitext.
	// Ihre Bedeutung ist unbekannt; sie fährt mit, damit beim Import keine
	// Daten verloren gehen, und bekommt bis dahin bewusst keine Semantik —
	// weder Prüfung noch Auswahl noch Glossareintrag. Sobald klar ist, was sie
	// bedeutet, ist das eine eigene, kleine Änderung: umbenennen und einen Typ
	// geben.
	Digital string

	// Rueckstand hängt am Mitglied und nicht an der Mitgliedschaft: ein Austritt
	// erlässt keine Schulden (ADR-0006).
	Rueckstand Rueckstand

	// Mitgliedschaften sind alle Zeiträume dieser Person, aufsteigend nach Eintritt.
	Mitgliedschaften []Mitgliedschaft
}

// Mitgliedschaft ist ein Zeitraum, in dem ein Mitglied aktiv im Verein ist. Ob
// er läuft, sagt nicht das Vorhandensein eines Austritts, sondern sein
// Erreichtsein: ein erfasster, aber noch bevorstehender Austritt ist die
// Kündigungsfrist und beendet nichts (siehe Status).
//
// Der Beitrag hängt hier und nicht am Mitglied: er ist Teil der Vereinbarung,
// die mit dem Eintritt zustande kam. Ein Wiedereintritt bekommt dadurch seinen
// eigenen Beitrag, und der alte bleibt an der alten Mitgliedschaft stehen
// (ADR-0005).
type Mitgliedschaft struct {
	ID         int64
	MitgliedID int64
	Eintritt   time.Time

	// Kuendigungsdatum ist der Tag, an dem gekündigt wurde; nil heißt "keine
	// Kündigung erfasst". Es steht neben dem Austritt und nicht in ihm, weil es
	// etwas anderes tut: es hält einen Vorgang fest, während der Austritt den
	// Zeitraum begrenzt. Erfasst werden beide zusammen (siehe Kuendigung).
	Kuendigungsdatum *time.Time

	Austritt *time.Time

	// Ruhend sagt, dass das Mitglied vorübergehend nicht trainiert (Verletzung,
	// Auslandsaufenthalt) und für diese Zeit kein Beitrag eingezogen wird. Die
	// Mitgliedschaft läuft weiter — ruhend ist kein Austritt und auch kein
	// Lebenszyklus-Zustand, sondern ein Merkmal daneben (CONTEXT.md → Ruhend).
	//
	// Es hängt an der Mitgliedschaft und nicht am Mitglied, weil sich nur
	// ruhend schalten lässt, was läuft. Zum Rückstand ist es unabhängig: für ein
	// ruhendes Mitglied geht keine Lastschrift los, es kann also kein neuer
	// Rückstand entstehen — ein bestehender bleibt aber stehen.
	Ruhend bool

	// BeitragCents ist der monatliche Beitrag in Cent. 0 ist ein gültiger
	// Betrag — Trainer zahlen nichts — und keine fehlende Angabe.
	BeitragCents int64

	// Anmeldung ist, was beim Zustandekommen dieses Zeitraums einmalig anfiel:
	// Anmeldedatum und Anmeldegebühr. Sie hängt an der Mitgliedschaft und nicht
	// am Mitglied, weil jeder Zeitraum seine eigene Anmeldung hatte.
	Anmeldung Anmeldung

	// Trainingstermine sind die Termine des Stundenplans, für die dieser
	// Zeitraum angemeldet ist — null bis MaxTrainingstermine, in
	// Wochenreihenfolge. Sie hängen wie der Beitrag an der Mitgliedschaft: für
	// welche Zeiten angemeldet wurde, gehört zu der Vereinbarung, die mit dem
	// Eintritt zustande kam.
	//
	// Archivierte Termine stehen mit darin: sie sind aus dem Stundenplan
	// genommen, aus der Vereinbarung nicht (ADR-0008).
	Trainingstermine []Trainingstermin
}

// Trainingsfrequenz ist die Anzahl der Trainingstermine dieses Zeitraums.
// Gespeichert wird sie nicht (CONTEXT.md → Trainingsfrequenz); abgelesen wird
// sie in trainingsfrequenzAus.
func (ms Mitgliedschaft) Trainingsfrequenz() Trainingsfrequenz {
	return trainingsfrequenzAus(ms.Trainingstermine)
}

// Status ist der Lebenszyklus-Zustand dieses Zeitraums am heutigen Tag.
// Gespeichert wird er nicht (CONTEXT.md → Status); abgelesen wird er in
// statusAus aus Eintritt, Kündigungsdatum und Austritt.
func (ms Mitgliedschaft) Status() Status {
	return ms.statusAm(heute())
}

// statusAm ist dasselbe zu einem bestimmten Tag. Den Tag durchzureichen lohnt
// sich, wo mehrere Zeiträume nacheinander zu beurteilen sind: sonst fragte jede
// Runde die Uhr erneut und könnte über Mitternacht eine andere Antwort bekommen.
func (ms Mitgliedschaft) statusAm(heute string) Status {
	return statusAus(ms.Eintritt, ms.Kuendigungsdatum, ms.Austritt, heute)
}

// LaufendeMitgliedschaft liefert den Zeitraum, in dem das Mitglied aktuell aktiv
// ist, oder nil, wenn es derzeit keinem angehört. Gibt es — regulär
// ausgeschlossen — mehrere offene Zeiträume, gilt der zuletzt begonnene; das ist
// dieselbe Regel, nach der List ein Mitglied als aktiv führt.
//
// "Läuft" heißt: der Austritt ist nicht erreicht. Ein erfasster, aber noch
// bevorstehender Austritt beendet nichts — das ist die Kündigungsfrist, in der
// weiter trainiert und weiter gezahlt wird. Auch ein Zeitraum, dessen Eintritt
// erst bevorsteht (Status Neu), läuft in diesem Sinn: begonnen hat er nicht,
// beendet ist er aber ebenso wenig.
func (m Mitglied) LaufendeMitgliedschaft() *Mitgliedschaft {
	// Einmal die Uhr fragen und den Tag durchreichen: sonst könnten zwei
	// Zeiträume derselben Schleife gegen verschiedene Tage beurteilt werden.
	heute := heute()

	// Mitgliedschaften liegen aufsteigend nach Eintritt vor, der zuletzt
	// begonnene offene Zeitraum ist also der letzte passende.
	for i := len(m.Mitgliedschaften) - 1; i >= 0; i-- {
		if !m.Mitgliedschaften[i].statusAm(heute).Ausgetreten() {
			return &m.Mitgliedschaften[i]
		}
	}

	return nil
}

// LetzteMitgliedschaft liefert den Zeitraum, der für das Mitglied gerade
// maßgeblich ist: den laufenden, und wenn keiner läuft, den zuletzt begonnenen.
// Das ist die Angabe, die eine Ansicht zeigt, wenn sie nach dem Eintritt fragt —
// bei einem Ehemaligen wäre "kein Eintritt" die falsche Auskunft, denn
// eingetreten war er ja.
//
// nil liefert nur ein Mitglied ganz ohne Mitgliedschaft; regulär gibt es das
// nicht, weil Create beides zusammen anlegt.
func (m Mitglied) LetzteMitgliedschaft() *Mitgliedschaft {
	if laufend := m.LaufendeMitgliedschaft(); laufend != nil {
		return laufend
	}

	if len(m.Mitgliedschaften) == 0 {
		return nil
	}

	// Mitgliedschaften liegen aufsteigend nach Eintritt vor.
	return &m.Mitgliedschaften[len(m.Mitgliedschaften)-1]
}

// Rueckstand ist das Kennzeichen, dass bei einem Mitglied nach einer
// Rücklastschrift Geld offen ist, samt der Notiz zum Vorgang. Beides gehört
// zusammen und reist zusammen: das Glossar kennt genau einen Begriff dafür
// (CONTEXT.md → Rückstand).
//
// Der Nullwert ist "in Ordnung, nichts notiert" — der Normalfall beim
// Lastschrifteinzug und damit auch der Zustand eines frisch angelegten
// Mitglieds.
type Rueckstand struct {
	// Offen sagt, ob gerade Geld offen ist. Einen dritten, neutralen Zustand
	// gibt es nicht: wer selbst einzieht, darf die Abwesenheit einer Rückgabe
	// als "gezahlt" lesen (ADR-0006).
	Offen bool

	// Notiz hält den Vorgang fest ("Rücklastschrift Oktober, angeschrieben am
	// 05.10."). Sie überlebt das Aufheben — sie ist die einzige Spur, die ein
	// erledigter Vorgang in v1 hinterlässt, denn eine Zahlungshistorie gibt es
	// nicht.
	Notiz string
}

// Bezeichnung ist der Text, den die Oberfläche am Kennzeichen zeigt. Er steht
// hier und nicht im Template, damit Liste und spätere Ansichten nicht
// auseinanderlaufen.
func (r Rueckstand) Bezeichnung() string {
	if r.Offen {
		return "im Rückstand"
	}

	return "in Ordnung"
}

// NeuesMitglied sind die Stammdaten, die beim Anlegen eines Mitglieds erfasst
// werden. Eintritt eröffnet zugleich die erste Mitgliedschaft.
type NeuesMitglied struct {
	Vorname      string
	Nachname     string
	Geburtsdatum *time.Time
	Anschrift    Anschrift
	Email        string
	Telefon      string
	Eintritt     time.Time

	// Die vier freiwilligen Angaben am Mitglied — siehe Mitglied.
	IBAN            string
	Geschlecht      string
	GoogleBewertung GoogleBewertung
	Digital         string

	// BeitragCents ist der individuell vereinbarte Monatsbeitrag in Cent. Er
	// landet an der Mitgliedschaft, die mit dem Eintritt beginnt.
	BeitragCents int64

	// Anmeldung landet wie der Beitrag an der Mitgliedschaft, die mit dem
	// Eintritt beginnt. Beide Angaben darin sind freiwillig.
	Anmeldung Anmeldung

	// TrainingsterminIDs sind die Termine des Stundenplans, für die die erste
	// Mitgliedschaft angemeldet wird — Verweise und kein Text, höchstens
	// MaxTrainingstermine. Doppelte zählen einmal (siehe
	// trainingsterminIDsNormalisieren).
	TrainingsterminIDs []int64
}

// negativerBeitrag ist die Meldung zum einzigen Beitrag, den es nicht geben
// kann. 0 € ist ausdrücklich erlaubt.
const negativerBeitrag = "Der Beitrag darf nicht negativ sein."

// negativeGebuehr ist das Gegenstück für die Anmeldegebühr. 0 € heißt hier
// "keine erhoben" und ist der Normalfall, nicht die Ausnahme.
const negativeGebuehr = "Die Anmeldegebühr darf nicht negativ sein."

// ErrNichtGefunden meldet, dass zu einer ID kein Datensatz existiert.
var ErrNichtGefunden = errors.New("nicht gefunden")

// ErrNichtAktiv meldet, dass das Mitglied derzeit keine laufende Mitgliedschaft
// hat — ein Austritt braucht aber einen Zeitraum, den er beenden kann.
var ErrNichtAktiv = errors.New("keine laufende mitgliedschaft")

// ErrBereitsAktiv meldet, dass das Mitglied bereits eine laufende Mitgliedschaft
// hat — ein Wiedereintritt setzt einen Austritt voraus.
var ErrBereitsAktiv = errors.New("bereits aktives mitglied")

// ValidierungsFehler bündelt alle Regelverstöße einer Eingabe. Bewusst als
// Liste: das Formular soll alle fehlenden Pflichtangaben auf einmal anzeigen
// können, statt den Nutzer eine nach der anderen abarbeiten zu lassen. Die
// Meldungen sind fertige, dem Nutzer zeigbare Sätze.
type ValidierungsFehler struct {
	Meldungen []string
}

func (f *ValidierungsFehler) Error() string {
	return strings.Join(f.Meldungen, " ")
}

// MemberService kapselt alle Operationen auf Mitgliedern und ihren
// Mitgliedschaften.
type MemberService struct {
	db *sql.DB
}

// Open öffnet die SQLite-Datei unter dbPath und legt sie bei Bedarf samt Schema
// an. Geseedet wird nichts: eine frische Datenbank ist leer, alles darin ist
// vom Verein erfasst. Der Aufrufer ist für Close verantwortlich.
func Open(dbPath string) (*MemberService, error) {
	// Fremdschlüssel sind in SQLite pro Verbindung abzuschalten bzw. -zuschalten;
	// _pragma im DSN sorgt dafür, dass jede Verbindung aus dem Pool sie aktiviert.
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)", dbPath)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("datenbank öffnen: %w", err)
	}

	svc := &MemberService{db: db}
	if err := svc.migrate(); err != nil {
		db.Close()
		return nil, err
	}

	return svc, nil
}

// Close gibt die Datenbankverbindung frei.
func (s *MemberService) Close() error {
	return s.db.Close()
}

const schema = `
CREATE TABLE IF NOT EXISTS mitglied (
	id               INTEGER PRIMARY KEY AUTOINCREMENT,
	vorname          TEXT    NOT NULL,
	nachname         TEXT    NOT NULL,
	geburtsdatum     TEXT,
	adresse          TEXT    NOT NULL DEFAULT '',
	postleitzahl     TEXT    NOT NULL DEFAULT '',
	ort              TEXT    NOT NULL DEFAULT '',
	email            TEXT    NOT NULL DEFAULT '',
	telefon          TEXT    NOT NULL DEFAULT '',
	iban             TEXT    NOT NULL DEFAULT '',
	geschlecht       TEXT    NOT NULL DEFAULT '',
	google_bewertung INTEGER NOT NULL DEFAULT 0,
	digital          TEXT    NOT NULL DEFAULT '',
	rueckstand       INTEGER NOT NULL DEFAULT 0,
	rueckstand_notiz TEXT    NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS mitgliedschaft (
	id                      INTEGER PRIMARY KEY AUTOINCREMENT,
	mitglied_id             INTEGER NOT NULL REFERENCES mitglied(id) ON DELETE CASCADE,
	anmeldedatum            TEXT,
	eintritt                TEXT    NOT NULL,
	kuendigungsdatum        TEXT,
	austritt                TEXT,
	anmeldegebuehr_cents    INTEGER NOT NULL DEFAULT 0,
	beitrag_monatlich_cents INTEGER NOT NULL DEFAULT 0,
	ruhend                  INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_mitgliedschaft_mitglied ON mitgliedschaft(mitglied_id);

-- Der Stundenplan des Vereins (ADR-0008). Er steht für sich: kein Fremdschlüssel
-- auf eine Mitgliedschaft, denn ein Termin gibt es auch dann, wenn niemand dafür
-- angemeldet ist. Der Wochentag ist eine Zahl (Montag = 1 … Sonntag = 7) und
-- keine Bezeichnung, sonst sortierte die Liste nach dem Alphabet statt nach der
-- Woche; die Uhrzeiten stehen als "HH:MM" da, was in dieser Form ebenfalls
-- richtig sortiert. Ein Index fehlt bewusst: ein Verein hat eine Handvoll
-- Termine, und die liest die Ansicht ohnehin am Stück.
CREATE TABLE IF NOT EXISTS trainingstermin (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	wochentag   INTEGER NOT NULL,
	beginn      TEXT    NOT NULL,
	ende        TEXT,
	bezeichnung TEXT    NOT NULL DEFAULT '',
	archiviert  INTEGER NOT NULL DEFAULT 0
);

-- Wofür eine Mitgliedschaft angemeldet ist (ADR-0008). Der Schlüssel aus beiden
-- Spalten macht die Doppelanmeldung unmöglich: derselbe Termin zweimal wäre eine
-- Frequenz, die um eins zu hoch abgelesen würde.
--
-- Auf den Termin wirkt kein ON DELETE: er wird nie gelöscht, sondern archiviert,
-- und bestehende Anmeldungen bleiben dabei bestehen. Auf die Mitgliedschaft
-- dagegen schon — mit ihr ist auch die Vereinbarung fort.
CREATE TABLE IF NOT EXISTS mitgliedschaft_trainingstermin (
	mitgliedschaft_id  INTEGER NOT NULL REFERENCES mitgliedschaft(id) ON DELETE CASCADE,
	trainingstermin_id INTEGER NOT NULL REFERENCES trainingstermin(id),
	PRIMARY KEY (mitgliedschaft_id, trainingstermin_id)
);

CREATE INDEX IF NOT EXISTS idx_mitgliedschaft_trainingstermin_termin
	ON mitgliedschaft_trainingstermin(trainingstermin_id);
`

func (s *MemberService) migrate() error {
	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("schema anlegen: %w", err)
	}

	return nil
}

// Create legt ein Mitglied samt seiner ersten, noch laufenden Mitgliedschaft an
// und liefert die vergebene Mitglied-ID. Beides passiert in einer Transaktion:
// ein Mitglied ohne Mitgliedschaft darf nicht entstehen.
//
// Bewusst gibt es keinen Eindeutigkeitsschutz über Name und Geburtsdatum —
// Namensgleichheit im Verein ist möglich und darf die Anlage nicht blockieren.
func (s *MemberService) Create(n NeuesMitglied) (int64, error) {
	if err := n.validieren(); err != nil {
		return 0, err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("transaktion starten: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		`INSERT INTO mitglied
		 	(vorname, nachname, geburtsdatum, adresse, postleitzahl, ort, email, telefon,
		 	 iban, geschlecht, google_bewertung, digital)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		n.Vorname, n.Nachname, alsDatumsText(n.Geburtsdatum),
		n.Anschrift.Adresse, n.Anschrift.Postleitzahl, n.Anschrift.Ort,
		n.Email, n.Telefon,
		n.IBAN, n.Geschlecht, bool(n.GoogleBewertung), n.Digital)
	if err != nil {
		return 0, fmt.Errorf("mitglied anlegen: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("mitglied-id lesen: %w", err)
	}

	res, err = tx.Exec(
		`INSERT INTO mitgliedschaft
		 	(mitglied_id, anmeldedatum, eintritt, anmeldegebuehr_cents, beitrag_monatlich_cents)
		 VALUES (?, ?, ?, ?, ?)`,
		id, alsDatumsText(n.Anmeldung.Datum), n.Eintritt.Format(isoDatum),
		n.Anmeldung.GebuehrCents, n.BeitragCents)
	if err != nil {
		return 0, fmt.Errorf("mitgliedschaft anlegen: %w", err)
	}

	mitgliedschaftID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("mitgliedschaft-id lesen: %w", err)
	}

	if err := trainingstermineSchreiben(tx, mitgliedschaftID, n.TrainingsterminIDs); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("transaktion abschließen: %w", err)
	}

	return id, nil
}

// validieren sammelt alle Regelverstöße auf einmal ein. Hier liegen die
// Pflichtfeld-Regeln als einzige Quelle — die Adapter-Schicht prüft sie nicht
// noch einmal, sondern zeigt die Meldungen nur an.
func (n NeuesMitglied) validieren() error {
	var fehler []string

	if n.Vorname == "" {
		fehler = append(fehler, "Vorname darf nicht leer sein.")
	}
	if n.Nachname == "" {
		fehler = append(fehler, "Nachname darf nicht leer sein.")
	}
	if n.Eintritt.IsZero() {
		fehler = append(fehler, "Eintrittsdatum darf nicht leer sein.")
	}
	if n.BeitragCents < 0 {
		fehler = append(fehler, negativerBeitrag)
	}
	if n.Anmeldung.GebuehrCents < 0 {
		fehler = append(fehler, negativeGebuehr)
	}
	fehler = append(fehler, trainingsterminIDsPruefen(n.TrainingsterminIDs)...)

	if len(fehler) > 0 {
		return &ValidierungsFehler{Meldungen: fehler}
	}

	return nil
}

// MitgliedPatch beschreibt eine Änderung an den Stammdaten eines Mitglieds. Ein
// Feld, das nil ist, bleibt unangetastet — nur gesetzte Felder werden
// geschrieben. Das erlaubt es, einzelne Angaben zu korrigieren, ohne den Rest
// des Datensatzes vorher lesen und wieder mitschicken zu müssen.
//
// Geburtsdatum und Eintrittsdatum fehlen bewusst: sie sind in v1 nach der
// Anlage nicht mehr änderbar (Tippfehler-Schutz). Der Eintritt gehört ohnehin
// zur Mitgliedschaft, nicht zu den Stammdaten.
type MitgliedPatch struct {
	Vorname  *string
	Nachname *string
	Email    *string
	Telefon  *string

	// Die vier freiwilligen Angaben am Mitglied — siehe Mitglied. Ein Zeiger
	// auf den leeren Wert heißt "leeren", nil heißt "nicht angerührt".
	IBAN            *string
	Geschlecht      *string
	GoogleBewertung *GoogleBewertung
	Digital         *string

	// Anschrift ändert sich als Ganzes und nicht feldweise: ein Umzug betrifft
	// alle drei Angaben, und im Formular stehen sie zusammen. Wer nur den Ort
	// korrigieren will, schickt die anderen beiden unverändert mit.
	Anschrift *Anschrift

	// BeitragCents gehört nicht zu den Stammdaten, sondern zur maßgeblichen
	// Mitgliedschaft. Er steht trotzdem hier: das Bearbeitungsformular zeigt
	// beides zusammen, und dann soll auch beides zusammen gespeichert werden —
	// ganz oder gar nicht.
	BeitragCents *int64

	// Anmeldung ändert Anmeldedatum und -gebühr der maßgeblichen Mitgliedschaft
	// als Ganzes, wie die Anschrift ihre drei Felder: beide beschreiben denselben
	// Vorgang und stehen im Formular nebeneinander. Ein Zeiger auf die leere
	// Anmeldung heißt deshalb "nichts mehr erfasst", nil dagegen "nicht angerührt".
	//
	// Anders als Geburtsdatum und Eintritt ist das Anmeldedatum nachträglich
	// änderbar: es begrenzt keinen Zeitraum, gegen den der Lebenszyklus prüft,
	// sondern hält einen Vorgang fest — und ein falsch abgetippter Vorgang muss
	// sich korrigieren lassen.
	Anmeldung *Anmeldung

	// TrainingsterminIDs ersetzen die Anmeldungen der maßgeblichen
	// Mitgliedschaft vollständig: Ankreuzen und Abwählen sind für den Nutzer
	// derselbe Vorgang, er schickt die Termine, die danach gelten sollen. Ein
	// Zeiger auf eine leere Liste heißt deshalb "gar keine mehr", nil dagegen
	// "nicht angerührt" — wie bei jedem anderen Feld des Patches.
	TrainingsterminIDs *[]int64
}

// Update schreibt die im Patch gesetzten Felder auf das Mitglied mit dieser ID.
// Existiert die ID nicht, ist der Fehler ErrNichtGefunden — es wird in keinem
// Fall ein neuer Datensatz angelegt.
//
// Stammdaten und Beitrag liegen in zwei Tabellen, gehören aber zu einem
// Formular: geschrieben werden sie deshalb in einer Transaktion, damit keine
// halbe Änderung stehen bleibt.
func (s *MemberService) Update(id int64, patch MitgliedPatch) error {
	if err := patch.validieren(); err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("transaktion starten: %w", err)
	}
	defer tx.Rollback()

	// Geprüft wird immer, auch wenn der Patch nichts zu schreiben hat: ein
	// Aufruf auf ein nicht existierendes Mitglied soll auch dann auffallen.
	if err := mitgliedPruefen(tx, id); err != nil {
		return err
	}

	if stammdaten := patch.zuweisungen(); !stammdaten.leer() {
		if _, err := tx.Exec(
			`UPDATE mitglied SET `+stammdaten.klausel()+` WHERE id = ?`,
			append(stammdaten.werte, id)...); err != nil {
			return fmt.Errorf("mitglied %d aktualisieren: %w", id, err)
		}
	}

	if mitgliedschaft := patch.mitgliedschaftZuweisungen(); !mitgliedschaft.leer() {
		if err := mitgliedschaftSchreiben(tx, id, mitgliedschaft); err != nil {
			return err
		}
	}

	if patch.TrainingsterminIDs != nil {
		mitgliedschaftID, err := massgeblicheMitgliedschaftLesen(tx, id)
		if err != nil {
			return err
		}
		if err := trainingstermineSchreiben(tx, mitgliedschaftID, *patch.TrainingsterminIDs); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("transaktion abschließen: %w", err)
	}

	return nil
}

// mitgliedschaftSchreiben setzt die übergebenen Felder auf der maßgeblichen
// Mitgliedschaft — der laufenden, und wenn keine läuft, der zuletzt begonnenen.
// Das ist derselbe Zeitraum, den die Liste zeigt und den LetzteMitgliedschaft
// liefert: geändert wird, was der Nutzer vor sich sieht.
func mitgliedschaftSchreiben(tx *sql.Tx, mitgliedID int64, z zuweisungssatz) error {
	res, err := tx.Exec(
		`UPDATE mitgliedschaft SET `+z.klausel()+`
		 WHERE id = (`+massgeblicheMitgliedschaft+`)`, append(z.werte, mitgliedID, heute())...)
	if err != nil {
		return fmt.Errorf("mitgliedschaft von mitglied %d aktualisieren: %w", mitgliedID, err)
	}

	betroffen, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("betroffene zeilen lesen: %w", err)
	}
	if betroffen == 0 {
		// Regulär unerreichbar: Create legt Mitglied und Mitgliedschaft zusammen an.
		return fmt.Errorf("mitglied %d hat keine mitgliedschaft: %w", mitgliedID, ErrNichtGefunden)
	}

	return nil
}

// massgeblicheMitgliedschaft wählt den Zeitraum aus, der für ein Mitglied gerade
// gilt. Die Sortierung stellt eine laufende Mitgliedschaft vor jede beendete und
// unter mehreren die zuletzt begonnene nach vorn — buchstäblich dieselbe Regel
// wie in eintraegeAbfrage, bis hin zum gemeinsamen laeuftNoch, damit Liste und
// Änderung denselben Zeitraum meinen. Liefen die beiden Sortierungen
// auseinander, änderte der Nutzer einen anderen Zeitraum als den, den er sieht.
//
// Der zweite Platzhalter nimmt den heutigen Tag auf, der erste die Mitglieds-ID.
const massgeblicheMitgliedschaft = `
	SELECT id FROM mitgliedschaft
	WHERE mitglied_id = ?
	ORDER BY ` + laeuftNoch + ` DESC, eintritt DESC, id DESC
	LIMIT 1`

// massgeblicheMitgliedschaftLesen liefert die ID des Zeitraums, der für das
// Mitglied gerade gilt — derselbe, den die Liste zeigt und dessen Beitrag
// beitragSchreiben ändert.
func massgeblicheMitgliedschaftLesen(q abfrager, mitgliedID int64) (int64, error) {
	var id int64

	err := q.QueryRow(massgeblicheMitgliedschaft, mitgliedID, heute()).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		// Regulär unerreichbar: Create legt Mitglied und Mitgliedschaft zusammen an.
		return 0, fmt.Errorf("mitglied %d hat keine mitgliedschaft: %w", mitgliedID, ErrNichtGefunden)
	}
	if err != nil {
		return 0, fmt.Errorf("maßgebliche mitgliedschaft von mitglied %d lesen: %w", mitgliedID, err)
	}

	return id, nil
}

// validieren prüft die gesetzten Felder gegen dieselben Pflichtfeld-Regeln, die
// auch bei der Neuanlage gelten: ein Pflichtfeld darf nachträglich nicht leer
// gemacht werden. Nicht gesetzte Felder sind keine Aussage und damit auch kein
// Regelverstoß.
func (p MitgliedPatch) validieren() error {
	var fehler []string

	if p.Vorname != nil && *p.Vorname == "" {
		fehler = append(fehler, "Vorname darf nicht leer sein.")
	}
	if p.Nachname != nil && *p.Nachname == "" {
		fehler = append(fehler, "Nachname darf nicht leer sein.")
	}
	if p.BeitragCents != nil && *p.BeitragCents < 0 {
		fehler = append(fehler, negativerBeitrag)
	}
	if p.Anmeldung != nil && p.Anmeldung.GebuehrCents < 0 {
		fehler = append(fehler, negativeGebuehr)
	}
	if p.TrainingsterminIDs != nil {
		fehler = append(fehler, trainingsterminIDsPruefen(*p.TrainingsterminIDs)...)
	}

	if len(fehler) > 0 {
		return &ValidierungsFehler{Meldungen: fehler}
	}

	return nil
}

// zuweisungssatz sammelt die SET-Fragmente einer Änderung samt ihrer Werte. Die
// Spaltennamen sind dabei ausnahmslos Literale aus diesem Package; die Werte
// gehen als Parameter in die Anweisung und werden nie in sie hineingeschrieben.
type zuweisungssatz struct {
	fragmente []string
	werte     []any
}

func (z *zuweisungssatz) setze(spalte string, wert any) {
	z.fragmente = append(z.fragmente, spalte+" = ?")
	z.werte = append(z.werte, wert)
}

// leer sagt, ob nichts zu schreiben ist — dann unterbleibt die Anweisung ganz.
func (z zuweisungssatz) leer() bool {
	return len(z.fragmente) == 0
}

// klausel ist der Teil hinter SET.
func (z zuweisungssatz) klausel() string {
	return strings.Join(z.fragmente, ", ")
}

// zuweisungen übersetzt die gesetzten Felder in SET-Fragmente samt ihrer Werte.
func (p MitgliedPatch) zuweisungen() zuweisungssatz {
	var z zuweisungssatz

	if p.Vorname != nil {
		z.setze("vorname", *p.Vorname)
	}
	if p.Nachname != nil {
		z.setze("nachname", *p.Nachname)
	}
	if p.Anschrift != nil {
		z.setze("adresse", p.Anschrift.Adresse)
		z.setze("postleitzahl", p.Anschrift.Postleitzahl)
		z.setze("ort", p.Anschrift.Ort)
	}
	if p.Email != nil {
		z.setze("email", *p.Email)
	}
	if p.Telefon != nil {
		z.setze("telefon", *p.Telefon)
	}
	if p.IBAN != nil {
		z.setze("iban", *p.IBAN)
	}
	if p.Geschlecht != nil {
		z.setze("geschlecht", *p.Geschlecht)
	}
	if p.GoogleBewertung != nil {
		z.setze("google_bewertung", bool(*p.GoogleBewertung))
	}
	if p.Digital != nil {
		z.setze("digital", *p.Digital)
	}

	return z
}

// mitgliedschaftZuweisungen übersetzt die Patch-Felder, die nicht am Mitglied
// hängen, sondern an seiner maßgeblichen Mitgliedschaft. Getrennt von
// zuweisungen, weil sie in eine andere Tabelle gehen — nicht, weil sie im
// Formular woanders stünden.
func (p MitgliedPatch) mitgliedschaftZuweisungen() zuweisungssatz {
	var z zuweisungssatz

	if p.BeitragCents != nil {
		z.setze("beitrag_monatlich_cents", *p.BeitragCents)
	}
	if p.Anmeldung != nil {
		z.setze("anmeldedatum", alsDatumsText(p.Anmeldung.Datum))
		z.setze("anmeldegebuehr_cents", p.Anmeldung.GebuehrCents)
	}

	return z
}

// abfrager ist die Teilmenge von *sql.DB und *sql.Tx, die die Prüfungen unten
// brauchen — so laufen dieselben Abfragen innerhalb wie außerhalb einer
// Transaktion.
type abfrager interface {
	QueryRow(query string, args ...any) *sql.Row
}

// mitgliedPruefen meldet ErrNichtGefunden, wenn zu der ID kein Mitglied
// vorliegt.
func mitgliedPruefen(q abfrager, id int64) error {
	var vorhanden int

	err := q.QueryRow(`SELECT 1 FROM mitglied WHERE id = ?`, id).Scan(&vorhanden)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("mitglied %d: %w", id, ErrNichtGefunden)
	}
	if err != nil {
		return fmt.Errorf("mitglied lesen: %w", err)
	}

	return nil
}

// Get liefert ein Mitglied samt allen seinen Mitgliedschaften. Existiert die ID
// nicht, ist der Fehler ErrNichtGefunden.
func (s *MemberService) Get(id int64) (Mitglied, error) {
	var (
		m            Mitglied
		geburtsdatum sql.NullString
	)

	err := s.db.QueryRow(
		`SELECT id, vorname, nachname, geburtsdatum, adresse, postleitzahl, ort,
		 	email, telefon, iban, geschlecht, google_bewertung, digital,
		 	rueckstand, rueckstand_notiz
		 FROM mitglied WHERE id = ?`, id).
		Scan(&m.ID, &m.Vorname, &m.Nachname, &geburtsdatum,
			&m.Anschrift.Adresse, &m.Anschrift.Postleitzahl, &m.Anschrift.Ort,
			&m.Email, &m.Telefon,
			&m.IBAN, &m.Geschlecht, &m.GoogleBewertung, &m.Digital,
			&m.Rueckstand.Offen, &m.Rueckstand.Notiz)
	if errors.Is(err, sql.ErrNoRows) {
		return Mitglied{}, fmt.Errorf("mitglied %d: %w", id, ErrNichtGefunden)
	}
	if err != nil {
		return Mitglied{}, fmt.Errorf("mitglied lesen: %w", err)
	}

	if m.Geburtsdatum, err = ausDatumsText(geburtsdatum); err != nil {
		return Mitglied{}, fmt.Errorf("geburtsdatum von mitglied %d: %w", id, err)
	}

	if m.Mitgliedschaften, err = s.mitgliedschaften(id); err != nil {
		return Mitglied{}, err
	}

	return m, nil
}

func (s *MemberService) mitgliedschaften(mitgliedID int64) ([]Mitgliedschaft, error) {
	rows, err := s.db.Query(
		`SELECT id, mitglied_id, anmeldedatum, eintritt, kuendigungsdatum, austritt,
		 	anmeldegebuehr_cents, beitrag_monatlich_cents, ruhend
		 FROM mitgliedschaft WHERE mitglied_id = ?
		 ORDER BY eintritt, id`, mitgliedID)
	if err != nil {
		return nil, fmt.Errorf("mitgliedschaften lesen: %w", err)
	}
	defer rows.Close()

	var alle []Mitgliedschaft
	for rows.Next() {
		var (
			ms               Mitgliedschaft
			anmeldedatum     sql.NullString
			eintritt         string
			kuendigungsdatum sql.NullString
			austritt         sql.NullString
		)
		if err := rows.Scan(&ms.ID, &ms.MitgliedID, &anmeldedatum, &eintritt, &kuendigungsdatum, &austritt,
			&ms.Anmeldung.GebuehrCents, &ms.BeitragCents, &ms.Ruhend); err != nil {
			return nil, fmt.Errorf("mitgliedschaft lesen: %w", err)
		}

		if ms.Anmeldung.Datum, err = ausDatumsText(anmeldedatum); err != nil {
			return nil, fmt.Errorf("anmeldedatum von mitgliedschaft %d: %w", ms.ID, err)
		}
		if ms.Eintritt, err = time.Parse(isoDatum, eintritt); err != nil {
			return nil, fmt.Errorf("eintritt von mitgliedschaft %d: %w", ms.ID, err)
		}
		if ms.Kuendigungsdatum, err = ausDatumsText(kuendigungsdatum); err != nil {
			return nil, fmt.Errorf("kündigungsdatum von mitgliedschaft %d: %w", ms.ID, err)
		}
		if ms.Austritt, err = ausDatumsText(austritt); err != nil {
			return nil, fmt.Errorf("austritt von mitgliedschaft %d: %w", ms.ID, err)
		}

		alle = append(alle, ms)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("mitgliedschaften lesen: %w", err)
	}

	ids := make([]int64, 0, len(alle))
	for _, ms := range alle {
		ids = append(ids, ms.ID)
	}

	termine, err := s.trainingstermineLesen(ids)
	if err != nil {
		return nil, err
	}
	for i := range alle {
		alle[i].Trainingstermine = termine[alle[i].ID]
	}

	return alle, nil
}

// SetKuendigung erfasst die Kündigung an der laufenden Mitgliedschaft: den Tag,
// an dem gekündigt wurde, und den Tag, zu dem der Austritt wirksam wird. Der
// Mitglied-Datensatz bleibt unangetastet — ausgetreten ist die Mitgliedschaft,
// nicht die Person.
//
// Beide Angaben dürfen einzeln fehlen: eine Kündigung ohne Termin ist der Fall
// "Kündigung liegt vor, Datum noch offen", ein Austritt ohne Kündigungsdatum der
// Normalfall der Altbestände. Ein Austritt in der Zukunft ist ausdrücklich
// erlaubt — das ist die laufende Kündigungsfrist, in der das Mitglied weiter
// trainiert und weiter zahlt.
//
// Geschrieben werden beide Spalten, auch die nicht gesetzte: erfasst wird die
// Kündigung als Ganzes, wie die Anmeldung im Patch. Das gilt ausdrücklich auch
// fürs Leeren — ein Aufruf, der nur den Austritt mitbringt, löscht ein zuvor
// erfasstes Kündigungsdatum. So lässt sich eine irrtümlich eingetragene
// Kündigungserklärung wieder wegnehmen, ohne dass es dafür einen zweiten
// Vorgang braucht; das Formular schickt den erfassten Wert deshalb immer mit
// (siehe app.kuendigungsformular).
//
// Solange der Austritt bevorsteht, läuft der Zeitraum weiter und lässt sich
// erneut erfassen: ein falsch abgetippter Termin ist während der Kündigungsfrist
// korrigierbar. Erst ab dem Austrittstag läuft nichts mehr, und der nächste
// Aufruf endet in ErrNichtAktiv — ab da ist der Weg zurück der Wiedereintritt.
func (s *MemberService) SetKuendigung(id int64, k Kuendigung) error {
	// Der laufende Zeitraum wird gelesen, geprüft und beendet — das gehört in
	// eine Transaktion, damit dazwischen keine andere Änderung dazwischenfunkt.
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("transaktion starten: %w", err)
	}
	defer tx.Rollback()

	if err := mitgliedPruefen(tx, id); err != nil {
		return err
	}

	mitgliedschaftID, eintritt, err := laufendeMitgliedschaftLesen(tx, id)
	if err != nil {
		return err
	}

	if fehler := k.pruefen(eintritt); len(fehler) > 0 {
		return &ValidierungsFehler{Meldungen: fehler}
	}

	if _, err := tx.Exec(
		`UPDATE mitgliedschaft SET kuendigungsdatum = ?, austritt = ? WHERE id = ?`,
		alsDatumsText(k.Datum), alsDatumsText(k.Austritt), mitgliedschaftID); err != nil {
		return fmt.Errorf("kündigung von mitglied %d eintragen: %w", id, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("transaktion abschließen: %w", err)
	}

	return nil
}

// SetRuhend schaltet die laufende Mitgliedschaft ruhend oder wieder aktiv. Für
// die Zeit, in der sie ruht, wird kein Beitrag eingezogen; Mitglied bleibt das
// Mitglied trotzdem (CONTEXT.md → Ruhend).
//
// Nur ein laufender Zeitraum lässt sich ruhend schalten: bei einem ausgetretenen
// Mitglied gibt es keinen Einzug, den man aussetzen könnte — der Fehler ist dann
// ErrNichtAktiv. Ein Mitglied in der Kündigungsfrist läuft dagegen noch und lässt
// sich sehr wohl ruhend schalten; sein Beitrag wird ja bis zum Austritt eingezogen.
//
// Der Rückstand bleibt dabei unberührt. Für ein ruhendes Mitglied geht keine
// Lastschrift los, es kann also kein *neuer* Rückstand entstehen; ein
// bestehender bleibt stehen, denn ruhend zu schalten ist kein Schuldenerlass
// (ADR-0006). Der Aufruf ist idempotent.
func (s *MemberService) SetRuhend(id int64, ruhend bool) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("transaktion starten: %w", err)
	}
	defer tx.Rollback()

	if err := mitgliedPruefen(tx, id); err != nil {
		return err
	}

	mitgliedschaftID, _, err := laufendeMitgliedschaftLesen(tx, id)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(
		`UPDATE mitgliedschaft SET ruhend = ? WHERE id = ?`,
		ruhend, mitgliedschaftID); err != nil {
		return fmt.Errorf("ruhend von mitglied %d setzen: %w", id, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("transaktion abschließen: %w", err)
	}

	return nil
}

// Rejoin lässt ein ehemaliges Mitglied wieder eintreten: es entsteht eine neue
// Mitgliedschaft, kein neues Mitglied. Stammdaten, ID und die bisherigen
// Zeiträume bleiben dabei unberührt.
func (s *MemberService) Rejoin(id int64, eintritt time.Time) error {
	if eintritt.IsZero() {
		return &ValidierungsFehler{Meldungen: []string{"Eintrittsdatum darf nicht leer sein."}}
	}

	// Prüfen und Einfügen gehören zusammen: sonst könnten zwischen beiden zwei
	// laufende Zeiträume entstehen.
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("transaktion starten: %w", err)
	}
	defer tx.Rollback()

	if err := mitgliedPruefen(tx, id); err != nil {
		return err
	}

	if _, _, err := laufendeMitgliedschaftLesen(tx, id); err == nil {
		return fmt.Errorf("mitglied %d: %w", id, ErrBereitsAktiv)
	} else if !errors.Is(err, ErrNichtAktiv) {
		return err
	}

	eintrittsText := eintritt.Format(isoDatum)

	letzterAustritt, err := letztenAustrittLesen(tx, id)
	if err != nil {
		return err
	}
	if eintrittsText < letzterAustritt {
		return &ValidierungsFehler{Meldungen: []string{
			"Das Eintrittsdatum darf nicht vor dem letzten Austritt liegen.",
		}}
	}

	// Der zuletzt vereinbarte Beitrag ist der Startwert der neuen Mitgliedschaft
	// — änderbar wie jeder andere. Die alte behält ihren eigenen: das ist der
	// Zweck der Aufteilung (ADR-0005).
	//
	// Die Trainingstermine wandern dabei ausdrücklich nicht mit: der neue
	// Zeitraum beginnt ohne Anmeldung, weil die alte für einen Zeitraum galt,
	// der vorbei ist — an welchen Tagen jemand nach Jahren wieder trainiert,
	// wird neu vereinbart. Bis dahin liest sich die Frequenz als "kein
	// Training", und die Anmeldungen der alten Mitgliedschaft bleiben
	// unangetastet stehen.
	//
	// Aus demselben Grund bleibt die Anmeldung leer: Anmeldedatum und -gebühr
	// halten fest, was bei *diesem* Eintritt geschah. Die Werte des alten
	// Zeitraums abzuschreiben hieße, eine Gebühr zu behaupten, die niemand
	// gezahlt hat.
	//
	// Ebenso beginnt der neue Zeitraum ohne Kündigung und nicht ruhend: beides
	// galt dem alten. Die Spalten stehen deshalb gar nicht erst in der
	// Anweisung — sie bekommen ihren Vorgabewert.
	if _, err := tx.Exec(
		`INSERT INTO mitgliedschaft (mitglied_id, eintritt, beitrag_monatlich_cents)
		 VALUES (?, ?, (SELECT beitrag_monatlich_cents FROM mitgliedschaft
		                WHERE mitglied_id = ? ORDER BY eintritt DESC, id DESC LIMIT 1))`,
		id, eintrittsText, id); err != nil {
		return fmt.Errorf("wiedereintritt von mitglied %d eintragen: %w", id, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("transaktion abschließen: %w", err)
	}

	return nil
}

// letztenAustrittLesen liefert den spätesten Austritt des Mitglieds als
// ISO-Text, oder den leeren Text, wenn es noch keinen gibt. Aufgerufen wird das
// nur, wenn kein Zeitraum mehr läuft — dann sind alle Austritte gesetzt und das
// Maximum ist der Rand, hinter dem ein Wiedereintritt liegen muss.
func letztenAustrittLesen(q abfrager, mitgliedID int64) (string, error) {
	var austritt sql.NullString

	if err := q.QueryRow(
		`SELECT MAX(austritt) FROM mitgliedschaft WHERE mitglied_id = ?`,
		mitgliedID).Scan(&austritt); err != nil {
		return "", fmt.Errorf("letzten austritt von mitglied %d lesen: %w", mitgliedID, err)
	}

	return austritt.String, nil
}

// laeuftNoch ist die Bedingung "Austritt nicht erreicht" in SQL — die Lesart von
// "aktiv", seit die Kündigungsfrist dazugehört (CONTEXT.md → Status). Der
// Platzhalter nimmt den heutigen Tag in ISO-Form auf; ein fehlender Austritt
// zählt mit, denn was kein Ende hat, hat es auch nicht erreicht.
//
// Es ist das SQL-Gegenstück zu statusAus' erster Prüfung und muss mit ihr
// übereinstimmen: hier entscheidet die Bedingung, welcher Zeitraum eine Zeile
// überhaupt bekommt, dort, wie er heißt. Getrennt sind sie, weil die Auswahl in
// die Abfrage gehört und die Ableitung nach Go (ADR-0004); wer eine ändert,
// ändert die andere mit. Beide Ränder sind einschließend — der Austrittstag
// selbst zählt bereits als erreicht.
const laeuftNoch = `(austritt IS NULL OR austritt > ?)`

// laufendeMitgliedschaftLesen liefert ID und Eintritt des Zeitraums, in dem das
// Mitglied gerade aktiv ist; gibt es keinen, ist der Fehler ErrNichtAktiv.
//
// Der Eintritt kommt als ISO-Text zurück und nicht als time.Time: verglichen
// wird er ohnehin nur mit anderen Kalendertagen in derselben Form. Die ISO-Form
// sortiert lexikografisch genau wie kalendarisch, und ein Zeitpunktvergleich
// läge je nach Zonenversatz um einen Tag daneben — aus der Datenbank gelesene
// Daten liegen in UTC, "heute" kommt aus der lokalen Uhr.
func laufendeMitgliedschaftLesen(q abfrager, mitgliedID int64) (int64, string, error) {
	var (
		id       int64
		eintritt string
	)

	err := q.QueryRow(
		`SELECT id, eintritt FROM mitgliedschaft
		 WHERE mitglied_id = ? AND `+laeuftNoch+`
		 ORDER BY eintritt DESC, id DESC
		 LIMIT 1`, mitgliedID, heute()).Scan(&id, &eintritt)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", fmt.Errorf("mitglied %d: %w", mitgliedID, ErrNichtAktiv)
	}
	if err != nil {
		return 0, "", fmt.Errorf("laufende mitgliedschaft von mitglied %d lesen: %w", mitgliedID, err)
	}

	return id, eintritt, nil
}

// Listeneintrag ist eine Zeile der Mitgliederliste. Er trägt bewusst nur die
// Angaben, die die Liste anzeigt — samt Beitrag und Eintritt der maßgeblichen
// Mitgliedschaft, damit die Adapter-Schicht die Zeile ohne weitere Abfragen
// rendern kann. Das vollständige Mitglied mit seiner Mitgliedschafts-Historie
// liefert Get.
type Listeneintrag struct {
	MitgliedID   int64
	Vorname      string
	Nachname     string
	BeitragCents int64
	Eintritt     time.Time

	// Anschrift steht in der Liste, weil der Verein sie beim Durchsehen
	// braucht. Gesucht wird gegen sie nicht (siehe suchzeile).
	Anschrift Anschrift

	// Rueckstand färbt das Kennzeichen der Zeile und erklärt es.
	Rueckstand Rueckstand

	// GoogleBewertung ist die einzige der freiwilligen Angaben, die in der Liste
	// steht: die Frage "wen kann ich noch fragen" beantwortet man durchsehend.
	// IBAN, Geschlecht, Digital, Anmeldedatum und Anmeldegebühr stehen nur im
	// Formular — sie helfen beim Überblick nicht und machten die Zeile nur breiter.
	GoogleBewertung GoogleBewertung

	// Kuendigungsdatum ist der Tag, an dem gekündigt wurde, oder nil. Es steht
	// in der Liste, weil eine erfasste Kündigung ohne Termin sonst unsichtbar
	// wäre — die Zeile sähe aus wie jede andere aktive. Zusammen mit Eintritt
	// und Austritt ergibt es den Status.
	Kuendigungsdatum *time.Time

	// Austritt ist der Tag, zu dem die Mitgliedschaft endet, oder nil. Er steht
	// auch in der Standardansicht: ein noch bevorstehender Austritt ist die
	// Kündigungsfrist, und dann zeigt die Zeile, wann der Zeitraum ausläuft.
	// Abgelesen wird daraus der Status.
	Austritt *time.Time

	// Ruhend ist das Merkmal der maßgeblichen Mitgliedschaft. Es steht neben
	// dem Lebenszyklus und nicht in ihm: ein ruhendes Mitglied ist ein aktives,
	// von dem gerade nichts eingezogen wird.
	Ruhend bool

	// Trainingstermine sind die Termine der maßgeblichen Mitgliedschaft — bei
	// einem Ehemaligen also die seines letzten Zeitraums, wie Beitrag und
	// Eintritt daneben.
	Trainingstermine []Trainingstermin
}

// Trainingsfrequenz ist die Anzahl der Trainingstermine dieser Zeile — dieselbe
// Ableitung wie an der Mitgliedschaft, aus derselben Quelle.
func (e Listeneintrag) Trainingsfrequenz() Trainingsfrequenz {
	return trainingsfrequenzAus(e.Trainingstermine)
}

// Status ist der Lebenszyklus-Zustand dieser Zeile — dieselbe Ableitung wie an
// der Mitgliedschaft, aus denselben drei Datumsfeldern. Er steht bewusst nicht
// als Feld in der Zeile: ein abgelesener Wert, der einmal mitgeschrieben wird,
// ist morgen falsch.
func (e Listeneintrag) Status() Status {
	return statusAus(e.Eintritt, e.Kuendigungsdatum, e.Austritt, heute())
}

// List liefert alle aktiven Mitglieder, sortiert nach Nachname und Vorname.
// Aktiv heißt: es existiert eine Mitgliedschaft, deren Austritt nicht erreicht
// ist. Wer nur beendete Mitgliedschaften hat, taucht hier nicht auf.
//
// Die Liste räumt sich damit von selbst auf: niemand muss ein Mitglied
// "auf ausgetreten setzen", das Austrittsdatum erledigt das am Stichtag. Bis
// dahin bleibt die Zeile stehen, auch wenn die Kündigung längst erfasst ist —
// in der Kündigungsfrist wird weiter trainiert und weiter gezahlt.
//
// Das ist die Standardansicht — dieselbe, die Search ohne Suchbegriff und mit
// dem Nullwert des Filters liefert.
func (s *MemberService) List() ([]Listeneintrag, error) {
	return s.Search("", Suchfilter{})
}

// Rueckstandsfilter grenzt die Ergebnisliste nach dem Rückstands-Kennzeichen
// ein. Das Kennzeichen ist zweiwertig, der Filter hat deshalb genau zwei Stufen
// plus "alle" — die beiden Stufen zerlegen den Bestand vollständig, ein
// Mitglied fällt immer in eine von beiden (ADR-0006).
type Rueckstandsfilter int

const (
	// RueckstandsfilterAlle ist der Nullwert und grenzt nicht ein.
	RueckstandsfilterAlle Rueckstandsfilter = iota
	// RueckstandsfilterImRueckstand: nur Mitglieder, bei denen Geld offen ist —
	// die Ausnahmeliste, die der Verein tatsächlich abarbeitet.
	RueckstandsfilterImRueckstand
	// RueckstandsfilterInOrdnung: nur Mitglieder ohne offenen Rückstand.
	RueckstandsfilterInOrdnung
)

// trifft entscheidet, ob ein Kennzeichen durch diesen Filter kommt.
func (f Rueckstandsfilter) trifft(r Rueckstand) bool {
	switch f {
	case RueckstandsfilterImRueckstand:
		return r.Offen
	case RueckstandsfilterInOrdnung:
		return !r.Offen
	default:
		return true
	}
}

// Suchfilter grenzt die Mitgliederliste ein. Der Nullwert ist bewusst die
// Standardansicht: jeder Rückstandswert, nur aktive Mitglieder. Damit ist
// "Filter zurücksetzen" nichts anderes als ein Suchfilter{}.
type Suchfilter struct {
	Rueckstand Rueckstandsfilter
	// Frequenz grenzt auf eine Trainingsfrequenz ein. Sie ist abgeleitet und
	// wird deshalb wie der Rückstand in Go ausgewertet, nicht in SQL.
	Frequenz Frequenzfilter
	// AuchEhemalige nimmt Mitglieder ohne laufende Mitgliedschaft mit auf — also
	// die, deren Austritt erreicht ist. Ein bevorstehender Austritt macht
	// niemanden ehemalig; solche Zeilen stehen auch ohne dieses Kennzeichen da.
	AuchEhemalige bool
}

// Search liefert die Mitglieder, auf die Suchbegriff und Filter gemeinsam
// zutreffen — sortiert wie List. Der Suchbegriff trifft als Teilzeichenkette in
// Vorname, Nachname, E-Mail oder Telefon; Groß-/Kleinschreibung entscheidet
// nicht. Zusätzlich trifft er die Mitglieds-ID, dort aber genau und nicht als
// Teilzeichenkette (siehe suchzeile). Ein leerer Begriff (auch einer aus lauter
// Leerraum) grenzt nichts ein, das Ergebnis ist dann das der Filter allein.
//
// Gesucht und gefiltert wird vollständig hier im Service — die Oberfläche
// bekommt bereits das fertige Ergebnis und siebt nichts nach.
func (s *MemberService) Search(query string, filter Suchfilter) ([]Listeneintrag, error) {
	// Die Aktivität entscheidet, welche Mitgliedschaft eine Zeile überhaupt
	// hat, und gehört deshalb in die Abfrage. Die übrigen Dimensionen sind
	// reine Auswahl auf den gelesenen Zeilen (siehe ADR-0004).
	zeilen, err := s.eintraegeLesen(filter.AuchEhemalige, "")
	if err != nil {
		return nil, err
	}

	begriff := strings.ToLower(strings.TrimSpace(query))

	var treffer []Listeneintrag
	for _, z := range zeilen {
		if z.passtZu(begriff, filter) {
			treffer = append(treffer, z.eintrag)
		}
	}

	// Sortiert wird nicht in SQL, sondern in Go: siehe nachNamenSortieren.
	nachNamenSortieren(treffer)

	return treffer, nil
}

// Eintrag liefert die Listenzeile eines einzelnen Mitglieds — dieselbe Zeile,
// die auch Search liefert. Damit kann die Adapter-Schicht nach einer Änderung
// genau eine Zeile neu rendern, statt die ganze Liste auszutauschen.
//
// Auch ein ausgetretenes Mitglied hat eine Zeile: seit Search Ehemalige
// einschließen kann, stehen sie in der Liste, und was in der Liste steht, muss
// sich auch einzeln neu rendern lassen. Nur zu einer unbekannten ID — oder zu
// einem Mitglied ganz ohne Mitgliedschaft — gibt es keine Zeile; dann ist der
// Fehler ErrNichtGefunden.
func (s *MemberService) Eintrag(id int64) (Listeneintrag, error) {
	zeilen, err := s.eintraegeLesen(true, " WHERE m.id = ?", id)
	if err != nil {
		return Listeneintrag{}, err
	}
	if len(zeilen) == 0 {
		return Listeneintrag{}, fmt.Errorf("listeneintrag zu mitglied %d: %w", id, ErrNichtGefunden)
	}

	return zeilen[0].eintrag, nil
}

// eintraegeAbfrage liest die Zeilen der Mitgliederliste. Die Unterabfrage wählt
// genau eine Mitgliedschaft je Mitglied aus. Regulär gibt es nie mehr als eine
// laufende; sollte doch einmal eine zweite entstehen, erscheint das Mitglied
// trotzdem nur einmal in der Liste.
//
// Der erste Parameter entscheidet über die Aktivität: ist er falsch, bleiben
// nur laufende Mitgliedschaften übrig und Ausgetretene fallen mangels
// Verbundpartner ganz aus der Liste. Ist er wahr, kommen sie mit ihrem zuletzt
// beendeten Zeitraum dazu — die Sortierung stellt eine laufende Mitgliedschaft
// dabei immer vor eine beendete, damit ein Wiedereintritt als aktiv erscheint.
//
// Laufend heißt hier wie überall "Austritt nicht erreicht" (siehe laeuftNoch):
// ein Mitglied in der Kündigungsfrist hat ein Austrittsdatum und steht trotzdem
// in der Standardansicht, weil es weiter trainiert und weiter zahlt. Deshalb
// steht der heutige Tag zweimal in der Abfrage — einmal im Filter, einmal in
// der Sortierung.
const eintraegeAbfrage = `
	SELECT m.id, m.vorname, m.nachname, m.email, m.telefon,
		m.adresse, m.postleitzahl, m.ort,
		m.google_bewertung, m.rueckstand, m.rueckstand_notiz,
		ms.id, ms.eintritt, ms.kuendigungsdatum, ms.austritt,
		ms.beitrag_monatlich_cents, ms.ruhend
	FROM mitglied m
	JOIN mitgliedschaft ms ON ms.id = (
		SELECT id FROM mitgliedschaft
		WHERE mitglied_id = m.id AND (? OR ` + laeuftNoch + `)
		ORDER BY ` + laeuftNoch + ` DESC, eintritt DESC, id DESC
		LIMIT 1
	)`

// suchzeile ist eine Listenzeile samt der Felder, gegen die gesucht wird.
//
// E-Mail und Telefon gehören nicht in den Listeneintrag: die Liste zeigt sie
// nicht an, und was sie nicht anzeigt, soll sie auch nicht mitschleppen. Für
// die Suche braucht es sie trotzdem — hier liegen sie klein geschrieben bereit.
type suchzeile struct {
	eintrag    Listeneintrag
	suchfelder []string

	// mitgliedschaftID ist der Zeitraum, aus dem die Zeile ihre Angaben bezieht.
	// Er verlässt den Service nicht und dient allein dazu, die Trainingstermine
	// nachzuladen.
	mitgliedschaftID int64
}

// passtZu entscheidet, ob die Zeile ins Ergebnis gehört. Die Aktivität steht
// hier nicht zur Debatte: über die entscheidet bereits die Abfrage.
func (z suchzeile) passtZu(begriff string, filter Suchfilter) bool {
	if !filter.Rueckstand.trifft(z.eintrag.Rueckstand) {
		return false
	}

	if !filter.Frequenz.trifft(z.eintrag.Trainingsfrequenz()) {
		return false
	}

	return z.trifftBegriff(begriff)
}

// trifftBegriff prüft den bereits klein geschriebenen Suchbegriff gegen die
// Suchfelder und die Mitglieds-ID. Ein leerer Begriff trifft jede Zeile.
//
// Die Mitglieds-ID ist dabei der Sonderfall: sie trifft nur, wenn der Begriff
// die ganze Nummer ist. Als Teiltreffer wäre sie bei 200 Mitgliedern wertlos —
// "7" brächte die 7, die 17, die 27 und die 70er zurück, und damit wäre die
// Nummer als Sprungmarke zu einer bekannten Zeile gerade nicht mehr zu
// gebrauchen.
func (z suchzeile) trifftBegriff(begriff string) bool {
	if begriff == "" {
		return true
	}

	if begriff == strconv.FormatInt(z.eintrag.MitgliedID, 10) {
		return true
	}

	for _, feld := range z.suchfelder {
		if strings.Contains(feld, begriff) {
			return true
		}
	}

	return false
}

// eintraegeLesen führt die Zeilenabfrage aus, optional um eine Bedingung
// ergänzt. Die Bedingung ist immer ein Literal aus diesem Package; die Werte
// gehen als Parameter in die Anweisung.
func (s *MemberService) eintraegeLesen(auchEhemalige bool, bedingung string, werte ...any) ([]suchzeile, error) {
	// Die Parameter der Abfrage stehen vor denen der Bedingung und gehören
	// deshalb auch in der Parameterliste nach vorn: erst die Aktivität, dann
	// der heutige Tag für Filter und Sortierung (siehe eintraegeAbfrage).
	heute := heute()

	rows, err := s.db.Query(eintraegeAbfrage+bedingung,
		append([]any{auchEhemalige, heute, heute}, werte...)...)
	if err != nil {
		return nil, fmt.Errorf("mitgliederliste lesen: %w", err)
	}
	defer rows.Close()

	var zeilen []suchzeile
	for rows.Next() {
		var (
			e                Listeneintrag
			email, telefon   string
			mitgliedschaftID int64
			eintritt         string
			kuendigungsdatum sql.NullString
			austritt         sql.NullString
		)
		if err := rows.Scan(&e.MitgliedID, &e.Vorname, &e.Nachname, &email, &telefon,
			&e.Anschrift.Adresse, &e.Anschrift.Postleitzahl, &e.Anschrift.Ort,
			&e.GoogleBewertung, &e.Rueckstand.Offen, &e.Rueckstand.Notiz,
			&mitgliedschaftID, &eintritt, &kuendigungsdatum, &austritt,
			&e.BeitragCents, &e.Ruhend); err != nil {
			return nil, fmt.Errorf("listeneintrag lesen: %w", err)
		}

		if e.Eintritt, err = time.Parse(isoDatum, eintritt); err != nil {
			return nil, fmt.Errorf("eintritt von mitglied %d: %w", e.MitgliedID, err)
		}
		if e.Kuendigungsdatum, err = ausDatumsText(kuendigungsdatum); err != nil {
			return nil, fmt.Errorf("kündigungsdatum von mitglied %d: %w", e.MitgliedID, err)
		}
		if e.Austritt, err = ausDatumsText(austritt); err != nil {
			return nil, fmt.Errorf("austritt von mitglied %d: %w", e.MitgliedID, err)
		}

		zeilen = append(zeilen, suchzeile{
			eintrag: e,
			suchfelder: []string{
				strings.ToLower(e.Vorname),
				strings.ToLower(e.Nachname),
				strings.ToLower(email),
				strings.ToLower(telefon),
			},
			mitgliedschaftID: mitgliedschaftID,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("mitgliederliste lesen: %w", err)
	}

	// Die Termine kommen für alle Zeilen auf einmal dazu: eine Abfrage je Zeile
	// wären bei 200 Mitgliedern 200 Abfragen für eine Ansicht.
	ids := make([]int64, 0, len(zeilen))
	for _, z := range zeilen {
		ids = append(ids, z.mitgliedschaftID)
	}

	termine, err := s.trainingstermineLesen(ids)
	if err != nil {
		return nil, err
	}
	for i := range zeilen {
		zeilen[i].eintrag.Trainingstermine = termine[zeilen[i].mitgliedschaftID]
	}

	return zeilen, nil
}

// SetRueckstand setzt Kennzeichen und Notiz eines Mitglieds in einem Zug —
// beides zusammen, weil der Nutzer beides im selben Formular vor sich hat.
//
// Die Notiz wird beim Aufheben nicht geleert. Sie ist die einzige Spur, die ein
// erledigter Vorgang in v1 hinterlässt — eine Zahlungshistorie gibt es nicht
// (ADR-0006). Wer sie loswerden will, übergibt einen leeren Text. Umschließender
// Leerraum wird abgeschnitten: normalisiert wird, wo gespeichert wird.
//
// Der Aufruf ist idempotent: derselbe Wert noch einmal geschrieben ändert
// nichts und ist kein Fehler. Andere Felder des Mitglieds bleiben unberührt,
// insbesondere seine Mitgliedschaften — der Rückstand hängt an der Person und
// überlebt deshalb Austritt und Wiedereintritt. Existiert die ID nicht, ist der
// Fehler ErrNichtGefunden.
func (s *MemberService) SetRueckstand(id int64, r Rueckstand) error {
	res, err := s.db.Exec(
		`UPDATE mitglied SET rueckstand = ?, rueckstand_notiz = ? WHERE id = ?`,
		r.Offen, strings.TrimSpace(r.Notiz), id)
	if err != nil {
		return fmt.Errorf("rückstand von mitglied %d setzen: %w", id, err)
	}

	betroffen, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("betroffene zeilen lesen: %w", err)
	}
	if betroffen == 0 {
		return fmt.Errorf("mitglied %d: %w", id, ErrNichtGefunden)
	}

	return nil
}

// nachNamenSortieren ordnet die Liste nach Nachname, dann Vorname — nach
// deutschen Regeln, in denen Umlaute wie ihr Grundbuchstabe zählen ("Ärmel"
// zwischen "Adler" und "Berger") und Groß-/Kleinschreibung nicht entscheidet.
//
// Das muss Go erledigen: SQLite kennt nur binäre Sortierung und COLLATE NOCASE,
// das ausschließlich ASCII faltet — dort landeten Umlaute hinter "Z". Bei der
// Größenordnung dieses Vereins (200 Mitglieder) ist Sortieren im Speicher
// ohnehin unmerklich.
func nachNamenSortieren(liste []Listeneintrag) {
	// Ein Collator ist nicht nebenläufigkeitssicher, deshalb einer pro Aufruf.
	sortierung := collate.New(language.German)

	slices.SortStableFunc(liste, func(a, b Listeneintrag) int {
		if c := sortierung.CompareString(a.Nachname, b.Nachname); c != 0 {
			return c
		}
		if c := sortierung.CompareString(a.Vorname, b.Vorname); c != 0 {
			return c
		}

		return int(a.MitgliedID - b.MitgliedID)
	})
}

// alsDatumsText bereitet ein optionales Datum für die Ablage in SQLite auf.
func alsDatumsText(d *time.Time) any {
	if d == nil {
		return nil
	}

	return d.Format(isoDatum)
}

// ausDatumsText liest eine optionale Datumsspalte zurück.
func ausDatumsText(s sql.NullString) (*time.Time, error) {
	if !s.Valid {
		return nil, nil
	}

	d, err := time.Parse(isoDatum, s.String)
	if err != nil {
		return nil, err
	}

	return &d, nil
}
