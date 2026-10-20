package service

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// zahlerFremdenFinden liefert den Zahler, der unter diesem Namen für jemand
// anderen zahlt (CONTEXT.md → Kontoinhaber). Dieselbe normalisierte IBAN heißt
// derselbe Zahler: ein bestehender wird wiederverwendet und bekommt den neuen
// Namen — der zuletzt gespeicherte gilt für alle, die auf ihn zeigen. Eine
// leere IBAN verbindet nie: jeder Aufruf legt dann einen eigenen an.
func zahlerFremdenFinden(tx *sql.Tx, name, iban string) (int64, error) {
	name = strings.TrimSpace(name)
	iban = ibanNormalisieren(iban)

	if iban != "" {
		var id int64

		err := tx.QueryRow(
			`SELECT id FROM zahler WHERE mitglied_id IS NULL AND iban = ?`, iban).Scan(&id)
		switch {
		case err == nil:
			if _, err := tx.Exec(`UPDATE zahler SET name = ? WHERE id = ?`, name, id); err != nil {
				return 0, fmt.Errorf("name von zahler %d aktualisieren: %w", id, err)
			}

			return id, nil
		case !errors.Is(err, sql.ErrNoRows):
			return 0, fmt.Errorf("zahler zur iban suchen: %w", err)
		}
	}

	return zahlerNeuAnlegen(tx, name, iban)
}

// zahlerNeuAnlegen legt einen Zahler für jemand anderen an, ohne nach einem
// bestehenden mit derselben IBAN zu suchen. Die IBAN ist schon normalisiert.
func zahlerNeuAnlegen(tx *sql.Tx, name, iban string) (int64, error) {
	res, err := tx.Exec(`INSERT INTO zahler (name, iban) VALUES (?, ?)`, name, iban)
	if err != nil {
		return 0, fmt.Errorf("zahler anlegen: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("zahler-id lesen: %w", err)
	}

	return id, nil
}

// zahlerMitFruehererMitgliedschaftGeteilt sagt, ob außer der maßgeblichen
// Mitgliedschaft noch eine andere derselben Person auf diesen Zahler zeigt.
func zahlerMitFruehererMitgliedschaftGeteilt(tx *sql.Tx, mitgliedID, mitgliedschaftID, zahlerID int64) (bool, error) {
	var geteilt bool
	if err := tx.QueryRow(
		`SELECT EXISTS (SELECT 1 FROM mitgliedschaft
		                WHERE mitglied_id = ? AND zahler_id = ? AND id <> ?)`,
		mitgliedID, zahlerID, mitgliedschaftID).Scan(&geteilt); err != nil {
		return false, fmt.Errorf("zahler von mitglied %d auf geteilte nutzung prüfen: %w", mitgliedID, err)
	}

	return geteilt, nil
}

// zahlerSelbstFinden liefert den Zahler „selbst“ des Mitglieds und legt ihn an,
// falls er nach einem Wechsel zu einem Kontoinhaber schon aufgeräumt wurde.
func zahlerSelbstFinden(tx *sql.Tx, mitgliedID int64, iban string) (int64, error) {
	var id int64

	err := tx.QueryRow(`SELECT id FROM zahler WHERE mitglied_id = ?`, mitgliedID).Scan(&id)
	switch {
	case err == nil:
		if _, err := tx.Exec(`UPDATE zahler SET iban = ? WHERE id = ?`, ibanNormalisieren(iban), id); err != nil {
			return 0, fmt.Errorf("iban von zahler %d aktualisieren: %w", id, err)
		}

		return id, nil
	case errors.Is(err, sql.ErrNoRows):
		return zahlerSelbstAnlegen(tx, mitgliedID, iban)
	default:
		return 0, fmt.Errorf("zahler von mitglied %d lesen: %w", mitgliedID, err)
	}
}

// zahlerSchreiben wendet IBAN und Kontoinhaber eines Patches auf die
// maßgebliche Mitgliedschaft an. nil heißt „nicht angerührt“: das Feld behält
// den Wert des bisherigen Zahlers. Danach zeigt die Mitgliedschaft auf den
// passenden Zahler, und ein Zahler, auf den nichts mehr zeigt, ist gelöscht.
//
// Ältere Mitgliedschaften behalten ihren Zahler: er hält fest, wer damals
// gezahlt hat.
func zahlerSchreiben(tx *sql.Tx, mitgliedID int64, ibanNeu, kontoinhaberNeu *string) error {
	mitgliedschaftID, err := massgeblicheMitgliedschaftLesen(tx, mitgliedID)
	if err != nil {
		return err
	}

	var (
		aktuellerID   int64
		selbst        bool
		aktuellerName string
		aktuelleIBAN  string
	)
	if err := tx.QueryRow(
		`SELECT z.id, z.mitglied_id IS NOT NULL, z.name, z.iban
		 FROM mitgliedschaft ms JOIN zahler z ON z.id = ms.zahler_id WHERE ms.id = ?`,
		mitgliedschaftID).Scan(&aktuellerID, &selbst, &aktuellerName, &aktuelleIBAN); err != nil {
		return fmt.Errorf("zahler von mitglied %d lesen: %w", mitgliedID, err)
	}

	if selbst {
		aktuellerName = ""
	}

	iban, name := aktuelleIBAN, aktuellerName
	if ibanNeu != nil {
		iban = ibanNormalisieren(*ibanNeu)
	}
	if kontoinhaberNeu != nil {
		name = strings.TrimSpace(*kontoinhaberNeu)
	}

	if name == aktuellerName && iban == aktuelleIBAN {
		return nil
	}

	if name != aktuellerName && name != "" {
		if err := kontoinhaberErlaubt(tx, mitgliedID); err != nil {
			return err
		}
	}

	// Zeigt auch eine frühere Mitgliedschaft derselben Person auf diesen Zahler
	// (Wiedereintritt), gilt die Änderung nur der maßgeblichen: die alte behält
	// ihren Zahler. Dafür bekommt die neue einen eigenen, statt den geteilten
	// umzubenennen.
	geteilt, err := zahlerMitFruehererMitgliedschaftGeteilt(tx, mitgliedID, mitgliedschaftID, aktuellerID)
	if err != nil {
		return err
	}

	var neuerID int64
	switch {
	case name != "" && !selbst && geteilt:
		neuerID, err = zahlerNeuAnlegen(tx, name, iban)
	case name == "":
		neuerID, err = zahlerSelbstFinden(tx, mitgliedID, iban)
	case !selbst && iban == "" && aktuelleIBAN == "":
		// Ein Zahler ohne IBAN gehört nur dieser Mitgliedschaft: umbenennen.
		neuerID = aktuellerID
		_, err = tx.Exec(`UPDATE zahler SET name = ? WHERE id = ?`, name, aktuellerID)
	default:
		neuerID, err = zahlerFremdenFinden(tx, name, iban)
	}
	if err != nil {
		return err
	}

	if _, err := tx.Exec(
		`UPDATE mitgliedschaft SET zahler_id = ? WHERE id = ?`, neuerID, mitgliedschaftID); err != nil {
		return fmt.Errorf("zahler von mitgliedschaft %d setzen: %w", mitgliedschaftID, err)
	}

	return verwaisteZahlerLoeschen(tx)
}

// kontoinhaberErlaubt lehnt einen neuen abweichenden Kontoinhaber für jemanden
// ab, der kein Kind ist. Gelesen wird nach den Stammdaten des Patches, damit ein
// im selben Speichern gesetztes Handkennzeichen schon zählt.
func kontoinhaberErlaubt(tx *sql.Tx, mitgliedID int64) error {
	var (
		geburtsdatum sql.NullString
		m            Mitglied
	)
	if err := tx.QueryRow(
		`SELECT geburtsdatum, kind_manuell FROM mitglied WHERE id = ?`, mitgliedID).
		Scan(&geburtsdatum, &m.KindManuell); err != nil {
		return fmt.Errorf("mitglied %d lesen: %w", mitgliedID, err)
	}

	var err error
	if m.Geburtsdatum, err = ausDatumsText(geburtsdatum); err != nil {
		return fmt.Errorf("geburtsdatum von mitglied %d: %w", mitgliedID, err)
	}

	if !m.IstKind() {
		return &ValidierungsFehler{Meldungen: []Meldung{kontoinhaberKeinKind}}
	}

	return nil
}

// verwaisteZahlerLoeschen entfernt jeden Zahler, auf den keine Mitgliedschaft
// mehr zeigt. Es gibt keine Verwaltung von Hand (Spec → Zusammenführen und
// Aufräumen).
func verwaisteZahlerLoeschen(tx *sql.Tx) error {
	if _, err := tx.Exec(
		`DELETE FROM zahler WHERE id NOT IN (SELECT zahler_id FROM mitgliedschaft)`); err != nil {
		return fmt.Errorf("verwaiste zahler löschen: %w", err)
	}

	return nil
}

// kontoinhaberLesen liefert zu den angegebenen Mitgliedschaften den Namen des
// abweichenden Zahlers. Wer selbst zahlt, hat keinen Eintrag.
func (s *MemberService) kontoinhaberLesen(mitgliedschaftIDs []int64) (map[int64]string, error) {
	namen := map[int64]string{}
	if len(mitgliedschaftIDs) == 0 {
		return namen, nil
	}

	platzhalter := strings.Repeat(", ?", len(mitgliedschaftIDs)-1)
	werte := make([]any, 0, len(mitgliedschaftIDs))
	for _, id := range mitgliedschaftIDs {
		werte = append(werte, id)
	}

	zeilen, err := s.db.Query(
		`SELECT ms.id, z.name FROM mitgliedschaft ms JOIN zahler z ON z.id = ms.zahler_id
		 WHERE z.mitglied_id IS NULL AND ms.id IN (?`+platzhalter+`)`, werte...)
	if err != nil {
		return nil, fmt.Errorf("kontoinhaber der mitgliedschaften lesen: %w", err)
	}
	defer zeilen.Close()

	for zeilen.Next() {
		var id int64
		var name string
		if err := zeilen.Scan(&id, &name); err != nil {
			return nil, fmt.Errorf("kontoinhaber einer mitgliedschaft lesen: %w", err)
		}
		namen[id] = name
	}

	return namen, zeilen.Err()
}
