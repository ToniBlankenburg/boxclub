package service

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Stilllegung ist ein Zeitraum, in dem eine Mitgliedschaft ruht: das Mitglied
// trainiert nicht, kündigt aber auch nicht, und der Beitrag wird für die Zeit
// nicht eingezogen (CONTEXT.md → Stilllegung, ADR-0022).
//
// Beide Ränder sind Pflicht und zählen einschließend. Eine Mitgliedschaft kann
// mehrere haben; sie überlappen nie und bilden zusammen die Historie.
type Stilllegung struct {
	ID               int64
	MitgliedschaftID int64
	Von              time.Time
	Bis              time.Time

	// Grund ist ein Freitext und darf fehlen.
	Grund string
}

// Stilllegungsangabe ist, was beim Erfassen oder Ändern einer Stilllegung
// eingegeben wird.
type Stilllegungsangabe struct {
	Von   time.Time
	Bis   time.Time
	Grund string
}

// liegtAm sagt, ob der Tag (ISO-Text) in der Stilllegung liegt. Beide Ränder
// zählen.
func (s Stilllegung) liegtAm(tag string) bool {
	return s.Von.Format(isoDatum) <= tag && tag <= s.Bis.Format(isoDatum)
}

// stillgelegtAm sagt, ob einer der Zeiträume den Tag enthält — die Ableitung
// von „stillgelegt“, die nirgends gespeichert wird.
func stillgelegtAm(liste []Stilllegung, tag string) bool {
	for _, s := range liste {
		if s.liegtAm(tag) {
			return true
		}
	}

	return false
}

// stilllegungsmonatAm sagt, ob der Kalendermonat des Tages ein
// Stilllegungsmonat ist: sobald irgendein Tag des Monats in einer Stilllegung
// liegt, zählt der Monat voll (CONTEXT.md → Stilllegung). Verglichen wird auf
// die ersten sieben Zeichen der ISO-Form, „JJJJ-MM“ — ein Monat liegt genau
// dann in der Stilllegung, wenn er zwischen deren Anfangs- und Endmonat liegt.
func stilllegungsmonatAm(liste []Stilllegung, tag string) bool {
	monat := tag[:7]
	for _, s := range liste {
		if s.Von.Format(isoDatum)[:7] <= monat && monat <= s.Bis.Format(isoDatum)[:7] {
			return true
		}
	}

	return false
}

// anzahlStilllegungsmonate zählt die verschiedenen Kalendermonate, die eine der
// Stilllegungen berührt. Zwei Zeiträume im selben Monat zählen ihn einmal.
func anzahlStilllegungsmonate(liste []Stilllegung) int {
	monate := map[string]bool{}
	for _, s := range liste {
		for m := monatBeginn(s.Von); !m.After(s.Bis); m = m.AddDate(0, 1, 0) {
			monate[m.Format("2006-01")] = true
		}
	}

	return len(monate)
}

func monatBeginn(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// monateAddieren rechnet Kalendermonate auf einen Tag auf. Fällt das Ziel
// hinter das Monatsende (31.01. + 1 Monat), gilt der letzte Tag des Zielmonats —
// time.AddDate schöbe ihn sonst in den Folgemonat.
func monateAddieren(t time.Time, monate int) time.Time {
	ziel := time.Date(t.Year(), t.Month()+time.Month(monate), 1, 0, 0, 0, 0, t.Location())
	letzter := time.Date(ziel.Year(), ziel.Month()+1, 0, 0, 0, 0, 0, t.Location()).Day()

	return time.Date(ziel.Year(), ziel.Month(), min(t.Day(), letzter), 0, 0, 0, 0, t.Location())
}

// Stillgelegt sagt, ob heute in einer Stilllegung dieser Mitgliedschaft liegt.
// Es wird abgelesen und nie gespeichert; einen Schalter gibt es nicht.
func (ms Mitgliedschaft) Stillgelegt() bool {
	return stillgelegtAm(ms.Stilllegungen, heute())
}

// Stilllegungsmonate ist die Zahl der Kalendermonate, in denen irgendein Tag
// stillgelegt war oder ist. Sie verlängern die Mindestlaufzeit.
func (ms Mitgliedschaft) Stilllegungsmonate() int {
	return anzahlStilllegungsmonate(ms.Stilllegungen)
}

// MindestlaufzeitEnde ist der Tag, bis zu dem sich das Mitglied gebunden hat:
// Eintritt plus Mindestlaufzeit plus Stilllegungsmonate. nil heißt „keine
// Bindung“. Abgelesen und nie gespeichert; es bindet auch nichts — das
// Kündigungsformular warnt nur (CONTEXT.md → Mindestlaufzeit).
func (ms Mitgliedschaft) MindestlaufzeitEnde() *time.Time {
	if ms.MindestlaufzeitMonate <= 0 {
		return nil
	}

	ende := monateAddieren(ms.Eintritt, ms.MindestlaufzeitMonate+ms.Stilllegungsmonate())

	return &ende
}

var (
	stilllegungVonLeer     = meldung("validierung.stilllegung.von_leer")
	stilllegungBisLeer     = meldung("validierung.stilllegung.bis_leer")
	stilllegungBisVorVon   = meldung("validierung.stilllegung.bis_vor_von")
	stilllegungLaeuft      = meldung("validierung.stilllegung.laeuft_bereits")
	stilllegungUeberlappt  = meldung("validierung.stilllegung.ueberlappt")
	kuendigungStillgelegt  = meldung("validierung.kuendigung.stillgelegt")
	mindestlaufzeitNegativ = meldung("validierung.mitglied.mindestlaufzeit_negativ")
)

// CreateStilllegung erfasst eine Stilllegung an der laufenden Mitgliedschaft.
// Ohne laufende Mitgliedschaft ist der Fehler ErrNichtAktiv: bei einem
// Ausgetretenen gibt es keinen Einzug, den man aussetzen könnte.
//
// Solange eine Stilllegung läuft (heute liegt in ihr), kommt keine neue hinzu —
// erst nach „wieder aktiv“ (BeendeStilllegung). Überlappen darf sie nie. Ein
// Ende hinter dem Austritt wird so gespeichert, wie es eingegeben wurde.
func (s *MemberService) CreateStilllegung(mitgliedID int64, a Stilllegungsangabe) (int64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("transaktion starten: %w", err)
	}
	defer tx.Rollback()

	if err := mitgliedPruefen(tx, mitgliedID); err != nil {
		return 0, err
	}

	mitgliedschaftID, _, err := laufendeMitgliedschaftLesen(tx, mitgliedID)
	if err != nil {
		return 0, err
	}

	fehler := a.pruefen()
	if len(fehler) == 0 {
		laeuft, err := stilllegungVorhanden(tx,
			`mitgliedschaft_id = ? AND von <= ? AND bis >= ?`, mitgliedschaftID, heute(), heute())
		if err != nil {
			return 0, err
		}
		if laeuft {
			fehler = []Meldung{stilllegungLaeuft}
		} else if fehler, err = a.ueberlappungPruefen(tx, mitgliedschaftID, 0); err != nil {
			return 0, err
		}
	}
	if len(fehler) > 0 {
		return 0, &ValidierungsFehler{Meldungen: fehler}
	}

	res, err := tx.Exec(
		`INSERT INTO stilllegung (mitgliedschaft_id, von, bis, grund) VALUES (?, ?, ?, ?)`,
		mitgliedschaftID, a.Von.Format(isoDatum), a.Bis.Format(isoDatum), grundAlsWert(a.Grund))
	if err != nil {
		return 0, fmt.Errorf("stilllegung von mitglied %d eintragen: %w", mitgliedID, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("id der stilllegung lesen: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("transaktion abschließen: %w", err)
	}

	return id, nil
}

// UpdateStilllegung ersetzt Zeitraum und Grund einer Stilllegung. Auch eine
// vergangene lässt sich korrigieren: sie ist Historie, aber ein Tippfehler darin
// soll behebbar bleiben. Die Regeln sind die des Erfassens, ohne die Sperre
// „es läuft schon eine“ — sonst ließe sich die laufende nicht ändern.
func (s *MemberService) UpdateStilllegung(id int64, a Stilllegungsangabe) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("transaktion starten: %w", err)
	}
	defer tx.Rollback()

	mitgliedschaftID, err := stilllegungLesen(tx, id)
	if err != nil {
		return err
	}

	fehler := a.pruefen()
	if len(fehler) == 0 {
		if fehler, err = a.ueberlappungPruefen(tx, mitgliedschaftID, id); err != nil {
			return err
		}
	}
	if len(fehler) > 0 {
		return &ValidierungsFehler{Meldungen: fehler}
	}

	if _, err := tx.Exec(
		`UPDATE stilllegung SET von = ?, bis = ?, grund = ? WHERE id = ?`,
		a.Von.Format(isoDatum), a.Bis.Format(isoDatum), grundAlsWert(a.Grund), id); err != nil {
		return fmt.Errorf("stilllegung %d ändern: %w", id, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("transaktion abschließen: %w", err)
	}

	return nil
}

// DeleteStilllegung löscht eine Stilllegung ganz, auch eine vergangene.
func (s *MemberService) DeleteStilllegung(id int64) error {
	res, err := s.db.Exec(`DELETE FROM stilllegung WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("stilllegung %d löschen: %w", id, err)
	}

	betroffen, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("betroffene zeilen lesen: %w", err)
	}
	if betroffen == 0 {
		return fmt.Errorf("stilllegung %d: %w", id, ErrNichtGefunden)
	}

	return nil
}

// BeendeStilllegung ist „wieder aktiv“: eine laufende Stilllegung endet gestern,
// eine, die heute oder später beginnt, wird gelöscht, eine schon vergangene
// bleibt unberührt. Der Aufruf ist idempotent.
//
// Gestern und nicht heute, weil beide Ränder einschließend zählen: mit
// bis = heute wäre das Mitglied heute noch stillgelegt, und der Knopf täte bis
// Mitternacht nichts Sichtbares.
func (s *MemberService) BeendeStilllegung(id int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("transaktion starten: %w", err)
	}
	defer tx.Rollback()

	var von, bis string
	err = tx.QueryRow(`SELECT von, bis FROM stilllegung WHERE id = ?`, id).Scan(&von, &bis)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("stilllegung %d: %w", id, ErrNichtGefunden)
	}
	if err != nil {
		return fmt.Errorf("stilllegung %d lesen: %w", id, err)
	}

	heute := heute()
	switch {
	case von >= heute:
		_, err = tx.Exec(`DELETE FROM stilllegung WHERE id = ?`, id)
	case bis >= heute:
		// Aus demselben Tag wie heute abgeleitet: zwei Uhrablesungen könnten über
		// Mitternacht auseinanderlaufen.
		tag, _ := time.Parse(isoDatum, heute)
		gestern := tag.AddDate(0, 0, -1).Format(isoDatum)
		_, err = tx.Exec(`UPDATE stilllegung SET bis = ? WHERE id = ?`, gestern, id)
	}
	if err != nil {
		return fmt.Errorf("stilllegung %d beenden: %w", id, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("transaktion abschließen: %w", err)
	}

	return nil
}

// pruefen sammelt die Verstöße, die allein in der Eingabe liegen.
func (a Stilllegungsangabe) pruefen() []Meldung {
	var fehler []Meldung

	if a.Von.IsZero() {
		fehler = append(fehler, stilllegungVonLeer)
	}
	if a.Bis.IsZero() {
		fehler = append(fehler, stilllegungBisLeer)
	}
	if !a.Von.IsZero() && !a.Bis.IsZero() && a.Bis.Format(isoDatum) < a.Von.Format(isoDatum) {
		fehler = append(fehler, stilllegungBisVorVon)
	}

	return fehler
}

// ueberlappungPruefen meldet, wenn der Zeitraum einen anderen der Mitgliedschaft
// berührt. ausser ist die Stilllegung, die gerade geändert wird (0 beim Erfassen).
func (a Stilllegungsangabe) ueberlappungPruefen(q abfrager, mitgliedschaftID, ausser int64) ([]Meldung, error) {
	ueberlappt, err := stilllegungVorhanden(q,
		`mitgliedschaft_id = ? AND id <> ? AND von <= ? AND bis >= ?`,
		mitgliedschaftID, ausser, a.Bis.Format(isoDatum), a.Von.Format(isoDatum))
	if err != nil {
		return nil, err
	}
	if ueberlappt {
		return []Meldung{stilllegungUeberlappt}, nil
	}

	return nil, nil
}

// stilllegungVorhanden fragt, ob es eine Stilllegung mit der Bedingung gibt. Die
// Bedingung ist immer ein Literal aus diesem Package.
func stilllegungVorhanden(q abfrager, bedingung string, werte ...any) (bool, error) {
	var vorhanden int
	err := q.QueryRow(`SELECT 1 FROM stilllegung WHERE `+bedingung+` LIMIT 1`, werte...).Scan(&vorhanden)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("stilllegungen prüfen: %w", err)
	}

	return true, nil
}

// stilllegungLesen liefert die Mitgliedschaft einer Stilllegung oder
// ErrNichtGefunden.
func stilllegungLesen(q abfrager, id int64) (int64, error) {
	var mitgliedschaftID int64
	err := q.QueryRow(`SELECT mitgliedschaft_id FROM stilllegung WHERE id = ?`, id).Scan(&mitgliedschaftID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("stilllegung %d: %w", id, ErrNichtGefunden)
	}
	if err != nil {
		return 0, fmt.Errorf("stilllegung %d lesen: %w", id, err)
	}

	return mitgliedschaftID, nil
}

func grundAlsWert(grund string) any {
	if g := strings.TrimSpace(grund); g != "" {
		return g
	}

	return nil
}

// stilllegungenLesen liefert die Stilllegungen der angegebenen Mitgliedschaften
// auf einmal, je Mitgliedschaft aufsteigend nach Beginn — eine Abfrage für die
// ganze Liste statt einer je Zeile.
func (s *MemberService) stilllegungenLesen(mitgliedschaftIDs []int64) (map[int64][]Stilllegung, error) {
	liste := map[int64][]Stilllegung{}
	if len(mitgliedschaftIDs) == 0 {
		return liste, nil
	}

	platzhalter := strings.Repeat(", ?", len(mitgliedschaftIDs)-1)
	werte := make([]any, len(mitgliedschaftIDs))
	for i, id := range mitgliedschaftIDs {
		werte[i] = id
	}

	zeilen, err := s.db.Query(
		`SELECT id, mitgliedschaft_id, von, bis, grund FROM stilllegung
		 WHERE mitgliedschaft_id IN (?`+platzhalter+`) ORDER BY von, id`, werte...)
	if err != nil {
		return nil, fmt.Errorf("stilllegungen lesen: %w", err)
	}
	defer zeilen.Close()

	for zeilen.Next() {
		var (
			st       Stilllegung
			von, bis string
			grund    sql.NullString
		)
		if err := zeilen.Scan(&st.ID, &st.MitgliedschaftID, &von, &bis, &grund); err != nil {
			return nil, fmt.Errorf("stilllegung lesen: %w", err)
		}
		if st.Von, err = time.Parse(isoDatum, von); err != nil {
			return nil, fmt.Errorf("von der stilllegung %d: %w", st.ID, err)
		}
		if st.Bis, err = time.Parse(isoDatum, bis); err != nil {
			return nil, fmt.Errorf("bis der stilllegung %d: %w", st.ID, err)
		}
		st.Grund = grund.String
		liste[st.MitgliedschaftID] = append(liste[st.MitgliedschaftID], st)
	}

	return liste, zeilen.Err()
}
