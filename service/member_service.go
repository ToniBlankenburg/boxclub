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
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// isoDatum ist das Layout, in dem alle Datumsspalten als TEXT in SQLite liegen.
// Es sortiert lexikografisch korrekt und lässt sich in SQL direkt vergleichen.
const isoDatum = "2006-01-02"

// Beitragsklasse ist eine Preisstufe. Die Staffelung ergibt sich ausschließlich
// aus der Trainingsfrequenz; der Preis liegt in Cent vor, um
// Fließkomma-Ungenauigkeiten zu vermeiden.
type Beitragsklasse struct {
	ID                  int64
	Name                string
	PreisMonatlichCents int64
	Aktiv               bool
}

// Mitglied ist die natürliche Person, die dem Verein bekannt ist. Der Datensatz
// bleibt derselbe, auch wenn die Person zwischenzeitlich aus- und wieder eintritt
// — die zeitliche Zuordnung steckt in Mitgliedschaften.
type Mitglied struct {
	ID               int64
	Vorname          string
	Nachname         string
	Geburtsdatum     *time.Time
	Adresse          string
	Email            string
	Telefon          string
	BeitragsklasseID int64
	BezahltBis       *time.Time

	// Mitgliedschaften sind alle Zeiträume dieser Person, aufsteigend nach Eintritt.
	Mitgliedschaften []Mitgliedschaft
}

// Mitgliedschaft ist ein Zeitraum, in dem ein Mitglied aktiv im Verein ist.
// Austritt == nil bedeutet: die Mitgliedschaft läuft aktuell.
type Mitgliedschaft struct {
	ID         int64
	MitgliedID int64
	Eintritt   time.Time
	Austritt   *time.Time
}

// NeuesMitglied sind die Stammdaten, die beim Anlegen eines Mitglieds erfasst
// werden. Eintritt eröffnet zugleich die erste Mitgliedschaft.
type NeuesMitglied struct {
	Vorname          string
	Nachname         string
	Geburtsdatum     *time.Time
	Adresse          string
	Email            string
	Telefon          string
	BeitragsklasseID int64
	Eintritt         time.Time
}

// ErrNichtGefunden meldet, dass zu einer ID kein Datensatz existiert.
var ErrNichtGefunden = errors.New("nicht gefunden")

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

// MemberService kapselt alle Operationen auf Mitgliedern, Mitgliedschaften und
// Beitragsklassen.
type MemberService struct {
	db *sql.DB
}

// Open öffnet die SQLite-Datei unter dbPath, legt sie bei Bedarf samt Schema an
// und seedet die Standard-Beitragsklassen. Der Aufrufer ist für Close
// verantwortlich.
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
CREATE TABLE IF NOT EXISTS beitragsklasse (
	id                    INTEGER PRIMARY KEY AUTOINCREMENT,
	name                  TEXT    NOT NULL UNIQUE,
	preis_monatlich_cents INTEGER NOT NULL,
	aktiv                 INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS mitglied (
	id                INTEGER PRIMARY KEY AUTOINCREMENT,
	vorname           TEXT    NOT NULL,
	nachname          TEXT    NOT NULL,
	geburtsdatum      TEXT,
	adresse           TEXT    NOT NULL DEFAULT '',
	email             TEXT    NOT NULL DEFAULT '',
	telefon           TEXT    NOT NULL DEFAULT '',
	beitragsklasse_id INTEGER NOT NULL REFERENCES beitragsklasse(id),
	bezahlt_bis       TEXT
);

CREATE TABLE IF NOT EXISTS mitgliedschaft (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	mitglied_id INTEGER NOT NULL REFERENCES mitglied(id) ON DELETE CASCADE,
	eintritt    TEXT    NOT NULL,
	austritt    TEXT
);

CREATE INDEX IF NOT EXISTS idx_mitgliedschaft_mitglied ON mitgliedschaft(mitglied_id);
`

// seedKlassen sind die beiden Klassen, mit denen eine frische Datenbank startet.
var seedKlassen = []Beitragsklasse{
	{Name: "Erwachsen 1×/Woche", PreisMonatlichCents: 6000, Aktiv: true},
	{Name: "Erwachsen 2×/Woche", PreisMonatlichCents: 8000, Aktiv: true},
}

func (s *MemberService) migrate() error {
	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("schema anlegen: %w", err)
	}

	for _, k := range seedKlassen {
		_, err := s.db.Exec(
			`INSERT OR IGNORE INTO beitragsklasse (name, preis_monatlich_cents, aktiv)
			 VALUES (?, ?, ?)`,
			k.Name, k.PreisMonatlichCents, k.Aktiv)
		if err != nil {
			return fmt.Errorf("beitragsklasse %q seeden: %w", k.Name, err)
		}
	}

	return nil
}

// AktiveBeitragsklassen liefert die Klassen, die aktuell zur Auswahl stehen,
// aufsteigend nach Preis. Deaktivierte Klassen bleiben außen vor, damit ein
// Formular sie nicht mehr anbietet — ihre Preishistorie bleibt aber erhalten.
func (s *MemberService) AktiveBeitragsklassen() ([]Beitragsklasse, error) {
	rows, err := s.db.Query(
		`SELECT id, name, preis_monatlich_cents, aktiv
		 FROM beitragsklasse
		 WHERE aktiv = 1
		 ORDER BY preis_monatlich_cents, id`)
	if err != nil {
		return nil, fmt.Errorf("beitragsklassen lesen: %w", err)
	}
	defer rows.Close()

	var klassen []Beitragsklasse
	for rows.Next() {
		var k Beitragsklasse
		if err := rows.Scan(&k.ID, &k.Name, &k.PreisMonatlichCents, &k.Aktiv); err != nil {
			return nil, fmt.Errorf("beitragsklasse lesen: %w", err)
		}
		klassen = append(klassen, k)
	}

	return klassen, rows.Err()
}

// Beitragsklasse liefert eine einzelne Klasse — auch eine deaktivierte, denn
// bestehende Mitglieder können ihr weiterhin zugeordnet sein. Existiert die ID
// nicht, ist der Fehler ErrNichtGefunden.
func (s *MemberService) Beitragsklasse(id int64) (Beitragsklasse, error) {
	var k Beitragsklasse

	err := s.db.QueryRow(
		`SELECT id, name, preis_monatlich_cents, aktiv
		 FROM beitragsklasse WHERE id = ?`, id).
		Scan(&k.ID, &k.Name, &k.PreisMonatlichCents, &k.Aktiv)
	if errors.Is(err, sql.ErrNoRows) {
		return Beitragsklasse{}, fmt.Errorf("beitragsklasse %d: %w", id, ErrNichtGefunden)
	}
	if err != nil {
		return Beitragsklasse{}, fmt.Errorf("beitragsklasse lesen: %w", err)
	}

	return k, nil
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
		 	(vorname, nachname, geburtsdatum, adresse, email, telefon, beitragsklasse_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		n.Vorname, n.Nachname, alsDatumsText(n.Geburtsdatum),
		n.Adresse, n.Email, n.Telefon, n.BeitragsklasseID)
	if err != nil {
		return 0, fmt.Errorf("mitglied anlegen: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("mitglied-id lesen: %w", err)
	}

	_, err = tx.Exec(
		`INSERT INTO mitgliedschaft (mitglied_id, eintritt) VALUES (?, ?)`,
		id, n.Eintritt.Format(isoDatum))
	if err != nil {
		return 0, fmt.Errorf("mitgliedschaft anlegen: %w", err)
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
	if n.BeitragsklasseID == 0 {
		fehler = append(fehler, "Bitte eine Beitragsklasse wählen.")
	}

	if len(fehler) > 0 {
		return &ValidierungsFehler{Meldungen: fehler}
	}

	return nil
}

// Get liefert ein Mitglied samt allen seinen Mitgliedschaften. Existiert die ID
// nicht, ist der Fehler ErrNichtGefunden.
func (s *MemberService) Get(id int64) (Mitglied, error) {
	var (
		m            Mitglied
		geburtsdatum sql.NullString
		bezahltBis   sql.NullString
	)

	err := s.db.QueryRow(
		`SELECT id, vorname, nachname, geburtsdatum, adresse, email, telefon,
		 	beitragsklasse_id, bezahlt_bis
		 FROM mitglied WHERE id = ?`, id).
		Scan(&m.ID, &m.Vorname, &m.Nachname, &geburtsdatum, &m.Adresse, &m.Email,
			&m.Telefon, &m.BeitragsklasseID, &bezahltBis)
	if errors.Is(err, sql.ErrNoRows) {
		return Mitglied{}, fmt.Errorf("mitglied %d: %w", id, ErrNichtGefunden)
	}
	if err != nil {
		return Mitglied{}, fmt.Errorf("mitglied lesen: %w", err)
	}

	if m.Geburtsdatum, err = ausDatumsText(geburtsdatum); err != nil {
		return Mitglied{}, fmt.Errorf("geburtsdatum von mitglied %d: %w", id, err)
	}
	if m.BezahltBis, err = ausDatumsText(bezahltBis); err != nil {
		return Mitglied{}, fmt.Errorf("bezahlt_bis von mitglied %d: %w", id, err)
	}

	if m.Mitgliedschaften, err = s.mitgliedschaften(id); err != nil {
		return Mitglied{}, err
	}

	return m, nil
}

func (s *MemberService) mitgliedschaften(mitgliedID int64) ([]Mitgliedschaft, error) {
	rows, err := s.db.Query(
		`SELECT id, mitglied_id, eintritt, austritt
		 FROM mitgliedschaft WHERE mitglied_id = ?
		 ORDER BY eintritt, id`, mitgliedID)
	if err != nil {
		return nil, fmt.Errorf("mitgliedschaften lesen: %w", err)
	}
	defer rows.Close()

	var alle []Mitgliedschaft
	for rows.Next() {
		var (
			ms       Mitgliedschaft
			eintritt string
			austritt sql.NullString
		)
		if err := rows.Scan(&ms.ID, &ms.MitgliedID, &eintritt, &austritt); err != nil {
			return nil, fmt.Errorf("mitgliedschaft lesen: %w", err)
		}

		if ms.Eintritt, err = time.Parse(isoDatum, eintritt); err != nil {
			return nil, fmt.Errorf("eintritt von mitgliedschaft %d: %w", ms.ID, err)
		}
		if ms.Austritt, err = ausDatumsText(austritt); err != nil {
			return nil, fmt.Errorf("austritt von mitgliedschaft %d: %w", ms.ID, err)
		}

		alle = append(alle, ms)
	}

	return alle, rows.Err()
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
