package service

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// Der Import übernimmt die Excel-Tabelle des Vereins zeilenweise in die
// Datenbank. Er ist wiederholbar: derselbe Lauf ein zweites Mal legt nichts
// doppelt an, sondern schreibt die geänderten Werte auf den vorhandenen
// Datensatz. Das ist die Arbeitsweise, für die der Import gedacht ist — Fehler
// in Excel korrigieren und noch einmal importieren.
//
// Geparst wird hier nichts: das tut der ExcelImporter, der einen Importsatz je
// lesbarer Zeile liefert. Hier steht nur, was ein solcher Satz in der Datenbank
// bewirkt.

// Importsatz ist eine Zeile der Excel-Tabelle in der Sprache des Modells: ein
// Mitglied samt der Mitgliedschaft, die die Zeile beschreibt.
//
// Die Stammdaten stecken in NeuesMitglied und nicht in eigenen Feldern, weil
// eine importierte Zeile dasselbe beschreibt wie ein von Hand angelegtes
// Mitglied — dieselben Pflichtangaben, dieselben Regeln, dieselbe Prüfung.
type Importsatz struct {
	// ID ist die Mitglieds-ID aus der Spalte „Mandatsreferenz". Sie wird
	// übernommen und nicht neu vergeben: der Verein kennt seine Mitglieder unter
	// diesen Nummern, und während der Umstellung müssen beide Systeme
	// vergleichbar bleiben (CONTEXT.md → Mitglieds-ID).
	ID int64

	NeuesMitglied

	// Kuendigung ist, was die Spalten „Status" und „Gekündigt" zusammen
	// aussagen. Der Nullwert heißt „keine Kündigung erfasst" — anders als im
	// Formular, wo eine leere Eingabe nichts festhielte und deshalb abgewiesen
	// wird.
	Kuendigung Kuendigung

	// Ruhend kommt aus den Status-Werten „Stillgelegt" und „Inakiv".
	Ruhend bool
}

// Importwirkung sagt, was eine übernommene Zeile bewirkt hat. Der Bericht zählt
// beides getrennt: „12 neu, 3 aktualisiert" ist die Auskunft, an der der Admin
// einen zweiten Lauf von einem ersten unterscheidet.
type Importwirkung int

const (
	// ImportNeu: die Zeile hat ein Mitglied angelegt.
	ImportNeu Importwirkung = iota
	// ImportAktualisiert: die Zeile hat ein vorhandenes Mitglied überschrieben.
	ImportAktualisiert
)

// Uebernehmen schreibt eine Zeile in die Datenbank und sagt, ob sie angelegt
// oder aktualisiert wurde.
func (s *MemberService) Uebernehmen(satz Importsatz) (Importwirkung, error) {
	if err := satz.validieren(); err != nil {
		return 0, err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("transaktion starten: %w", err)
	}
	defer tx.Rollback()

	ziel, vorhanden, err := satz.zielMitglied(tx)
	if err != nil {
		return 0, err
	}

	wirkung := ImportNeu
	if vorhanden {
		wirkung = ImportAktualisiert
		err = satz.aktualisieren(tx, ziel)
	} else {
		err = satz.anlegen(tx)
	}
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("transaktion abschließen: %w", err)
	}

	return wirkung, nil
}

// validieren prüft den Satz nach denselben Regeln wie eine Eingabe von Hand.
func (satz Importsatz) validieren() error {
	var fehler []string

	if satz.ID <= 0 {
		fehler = append(fehler, "Die Mitglieds-ID muss eine positive Zahl sein.")
	}

	var vf *ValidierungsFehler
	if err := satz.NeuesMitglied.validieren(); errors.As(err, &vf) {
		fehler = append(fehler, vf.Meldungen...)
	} else if err != nil {
		return err
	}

	// Geprüft wird die Kündigung nur, wenn überhaupt eine erfasst ist. Kuendigung
	// selbst weist zwei leere Daten als nichtssagende Eingabe ab (siehe
	// fehlendesDatum) — im Import ist genau das aber der Normalfall: bei den
	// allermeisten Zeilen steht in der Spalte „Gekündigt" nichts.
	if satz.Kuendigung.Datum != nil || satz.Kuendigung.Austritt != nil {
		fehler = append(fehler, satz.Kuendigung.pruefen(satz.Eintritt.Format(isoDatum))...)
	}

	if len(fehler) > 0 {
		return &ValidierungsFehler{Meldungen: fehler}
	}

	return nil
}

// zielMitglied sucht das Mitglied, das diese Zeile meint.
//
// Maßgeblich ist die Mitglieds-ID: sie ist die fachliche Identität, und ein
// zweiter Lauf derselben Datei findet darüber jede Zeile wieder. Findet sie
// niemanden, greift der Rückfall auf Vorname, Nachname und Geburtsdatum — er
// fängt den Fall ab, dass dasselbe Mitglied zwischendurch von Hand angelegt
// wurde und dabei eine andere Nummer bekommen hat. Der Rückfall behält dessen
// Nummer bei: einen Primärschlüssel umzuschreiben, an dem Mitgliedschaften
// hängen, wäre der teurere Fehler als eine abweichende Nummer.
//
// Mehrdeutig heißt „nicht gefunden": Namensgleichheit im Verein ist möglich
// (siehe Create), und dann ist eine neue Zeile die ehrlichere Antwort als ein
// geratener Treffer.
func (satz Importsatz) zielMitglied(q abfrager) (int64, bool, error) {
	var vorname, nachname string

	err := q.QueryRow(
		`SELECT vorname, nachname FROM mitglied WHERE id = ?`, satz.ID).Scan(&vorname, &nachname)
	if err == nil {
		// Die Nummer allein genügt nicht: die App vergibt ihre IDs aus demselben
		// Zahlenraum, in dem auch die Excel zählt, und beide fangen bei 1 an. Ein
		// von Hand angelegtes Mitglied trüge also dieselbe Nummer wie eine
		// wildfremde Excel-Zeile — und würde stillschweigend von ihr überschrieben.
		// Der Name entscheidet deshalb mit, ob wirklich dieselbe Person gemeint ist.
		if !gleicherName(vorname, satz.Vorname) || !gleicherName(nachname, satz.Nachname) {
			return 0, false, &ValidierungsFehler{Meldungen: []string{fmt.Sprintf(
				"Die Mitglieds-ID %d gehört in der App bereits zu %s %s. "+
					"Entweder die Nummer in der Excel oder den Namen in der App korrigieren.",
				satz.ID, vorname, nachname)}}
		}

		return satz.ID, true, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, false, fmt.Errorf("mitglied %d suchen: %w", satz.ID, err)
	}

	var id int64

	// COALESCE und COUNT stehen zusammen in einer Abfrage, weil die Aggregatform
	// immer genau eine Zeile liefert: so unterscheidet ein Blick „nicht
	// gefunden" (0) von „mehrdeutig" (mehr als 1).
	var treffer int
	if err := q.QueryRow(
		`SELECT COALESCE(MIN(id), 0), COUNT(*) FROM mitglied
		 WHERE vorname = ? AND nachname = ? AND geburtsdatum IS ?`,
		satz.Vorname, satz.Nachname, alsDatumsText(satz.Geburtsdatum),
	).Scan(&id, &treffer); err != nil {
		return 0, false, fmt.Errorf("mitglied %s %s suchen: %w", satz.Vorname, satz.Nachname, err)
	}

	if treffer != 1 {
		return 0, false, nil
	}

	return id, true, nil
}

// anlegen schreibt Mitglied und Mitgliedschaft neu.
func (satz Importsatz) anlegen(tx *sql.Tx) error {
	if _, err := tx.Exec(
		`INSERT INTO mitglied
		 	(id, vorname, nachname, geburtsdatum, adresse, postleitzahl, ort, email, telefon,
		 	 iban, geschlecht, google_bewertung, digital)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		satz.ID, satz.Vorname, satz.Nachname, alsDatumsText(satz.Geburtsdatum),
		satz.Anschrift.Adresse, satz.Anschrift.Postleitzahl, satz.Anschrift.Ort,
		satz.Email, satz.Telefon,
		satz.IBAN, satz.Geschlecht, bool(satz.GoogleBewertung), satz.Digital); err != nil {
		return fmt.Errorf("mitglied %d anlegen: %w", satz.ID, err)
	}

	res, err := tx.Exec(
		`INSERT INTO mitgliedschaft
		 	(mitglied_id, anmeldedatum, eintritt, kuendigungsdatum, austritt,
		 	 anmeldegebuehr_cents, beitrag_monatlich_cents, ruhend)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		satz.ID, alsDatumsText(satz.Anmeldung.Datum), satz.Eintritt.Format(isoDatum),
		alsDatumsText(satz.Kuendigung.Datum), alsDatumsText(satz.Kuendigung.Austritt),
		satz.Anmeldung.GebuehrCents, satz.BeitragCents, satz.Ruhend)
	if err != nil {
		return fmt.Errorf("mitgliedschaft für %d anlegen: %w", satz.ID, err)
	}

	mitgliedschaftID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("mitgliedschaft-id lesen: %w", err)
	}

	return trainingsslotsSchreiben(tx, mitgliedschaftID, satz.Trainingsslots)
}

// aktualisieren überschreibt Stammdaten und Mitgliedschaft eines vorhandenen
// Mitglieds mit dem, was in der Zeile steht.
//
// Der Rückstand bleibt dabei unangetastet: die Excel führt keine Quellspalte
// dafür (ADR-0006), und was der Verein in der App von Hand gesetzt hat, darf
// ein erneuter Import nicht stillschweigend zurücksetzen.
func (satz Importsatz) aktualisieren(tx *sql.Tx, id int64) error {
	if _, err := tx.Exec(
		`UPDATE mitglied SET
		 	vorname = ?, nachname = ?, geburtsdatum = ?,
		 	adresse = ?, postleitzahl = ?, ort = ?, email = ?, telefon = ?,
		 	iban = ?, geschlecht = ?, google_bewertung = ?, digital = ?
		 WHERE id = ?`,
		satz.Vorname, satz.Nachname, alsDatumsText(satz.Geburtsdatum),
		satz.Anschrift.Adresse, satz.Anschrift.Postleitzahl, satz.Anschrift.Ort,
		satz.Email, satz.Telefon,
		satz.IBAN, satz.Geschlecht, bool(satz.GoogleBewertung), satz.Digital,
		id); err != nil {
		return fmt.Errorf("mitglied %d aktualisieren: %w", id, err)
	}

	mitgliedschaftID, err := satz.zielMitgliedschaft(tx, id)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(
		`UPDATE mitgliedschaft SET
		 	anmeldedatum = ?, eintritt = ?, kuendigungsdatum = ?, austritt = ?,
		 	anmeldegebuehr_cents = ?, beitrag_monatlich_cents = ?, ruhend = ?
		 WHERE id = ?`,
		alsDatumsText(satz.Anmeldung.Datum), satz.Eintritt.Format(isoDatum),
		alsDatumsText(satz.Kuendigung.Datum), alsDatumsText(satz.Kuendigung.Austritt),
		satz.Anmeldung.GebuehrCents, satz.BeitragCents, satz.Ruhend,
		mitgliedschaftID); err != nil {
		return fmt.Errorf("mitgliedschaft %d aktualisieren: %w", mitgliedschaftID, err)
	}

	return trainingsslotsSchreiben(tx, mitgliedschaftID, satz.Trainingsslots)
}

// gleicherName vergleicht zwei Namensteile. Groß- und Kleinschreibung
// unterscheiden keine Personen: wer „müller" tippt, wo „Müller" steht, hat
// keinen anderen Menschen gemeint, und der Import soll daran nicht scheitern.
func gleicherName(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}

// zielMitgliedschaft sucht den Zeitraum, den die Zeile beschreibt.
//
// Zuerst über den Eintritt: eine Excel-Zeile führt genau einen Zeitraum, und
// hat der Verein in der App zwischenzeitlich einen Wiedereintritt erfasst, meint
// die Zeile weiterhin den alten und nicht den neuen.
//
// Findet sich keiner, hat sich der Eintritt in der Excel geändert. Hat das
// Mitglied nur einen Zeitraum, ist klar, welcher gemeint war, und die Zeile
// korrigiert ihn. Hat es mehrere, ist es nicht klar — und dann wird nichts
// angefasst: die Excel kennt nur einen Zeitraum je Person und kann gar nicht
// sagen, welchen sie meint. Zu raten hieße hier, einen Wiedereintritt mit den
// Daten der alten Mitgliedschaft zu überschreiben und ihn damit zu verlieren.
func (satz Importsatz) zielMitgliedschaft(q abfrager, mitgliedID int64) (int64, error) {
	var id int64

	err := q.QueryRow(
		`SELECT id FROM mitgliedschaft WHERE mitglied_id = ? AND eintritt = ?`,
		mitgliedID, satz.Eintritt.Format(isoDatum)).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("mitgliedschaft von mitglied %d suchen: %w", mitgliedID, err)
	}

	// Wie in zielMitglied liefert die Aggregatform immer genau eine Zeile und
	// beantwortet damit „wie viele" und „welche" in einem Zug.
	var anzahl int
	if err := q.QueryRow(
		`SELECT COUNT(*), COALESCE(MIN(id), 0) FROM mitgliedschaft WHERE mitglied_id = ?`,
		mitgliedID).Scan(&anzahl, &id); err != nil {
		return 0, fmt.Errorf("mitgliedschaften von mitglied %d zählen: %w", mitgliedID, err)
	}

	if anzahl != 1 {
		return 0, &ValidierungsFehler{Meldungen: []string{fmt.Sprintf(
			"%s %s hat in der App mehrere Mitgliedschaften, und das Eintrittsdatum %s "+
				"passt zu keiner davon. Bitte in der App korrigieren.",
			satz.Vorname, satz.Nachname, satz.Eintritt.Format(isoDatum))}}
	}

	return id, nil
}
