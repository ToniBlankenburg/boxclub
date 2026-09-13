package service

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
)

// Eine Mitgliedschaft ist für null bis drei Termine des Stundenplans angemeldet
// (ADR-0008). Die Zuordnung hat keinen eigenen Fachbegriff — sie heißt im Code
// nach ihrer Tabelle, in der Oberfläche steht „Training".
//
// Die Trainingsfrequenz steht hier daneben, weil sie nichts anderes ist als die
// Anzahl dieser Zuordnungen (CONTEXT.md → Trainingsfrequenz): etwas, das man
// außerhalb der Zuordnung noch einmal festhalten könnte, soll sie gerade nicht
// sein.

// MaxTrainingstermine ist die Obergrenze: für mehr als drei wöchentliche
// Termine meldet der Verein niemanden an. Ein vierter wäre eine Frequenz, die
// es nicht gibt, und wird deshalb abgewiesen statt abgeschnitten.
const MaxTrainingstermine = 3

// Die Meldungen zu den drei Regeln, die eine Zuordnung kennt — fertige Sätze
// für das Formular, wie überall im Service.
var (
	zuVieleTermine = fmt.Sprintf("Es sind höchstens %d Trainingstermine möglich.", MaxTrainingstermine)

	// Ein Termin, den es nicht gibt, kann nur aus einer veralteten Ansicht oder
	// einem selbstgebauten Request stammen. Benennen lässt er sich nicht — es
	// steht ja nichts mehr da, was einen Namen hätte.
	unbekannterTermin = "Einen ausgewählten Trainingstermin gibt es nicht."
)

// archivierterTermin sagt, warum ein archivierter Termin nicht neu vergeben
// werden kann. Er wird beim Namen genannt: in der Auswahl stehen drei bis zehn
// Zeilen, und welche davon gemeint ist, soll niemand raten müssen.
func archivierterTermin(t Trainingstermin) string {
	return fmt.Sprintf("»%s« ist archiviert und kann nicht neu vergeben werden.", t.Anzeige())
}

// Trainingsfrequenz ist, wie oft pro Woche ein Mitglied trainiert. Sie wird
// nirgends gespeichert, sondern ist die Anzahl der Termine, für die seine
// Mitgliedschaft angemeldet ist — dadurch können Frequenz und Termine gar nicht
// erst auseinanderlaufen.
//
// Der Nullwert heißt "keine Frequenz" und nicht "1×": ein Mitglied ohne
// eingetragenen Termin trainiert nicht einmal die Woche, sondern es ist
// schlicht nichts vereinbart.
type Trainingsfrequenz int

// Bezeichnung ist der Text, den die Oberfläche zur Frequenz zeigt. Er steht wie
// bei Rueckstand.Bezeichnung im Service, damit Liste und spätere Ansichten
// dieselben Worte benutzen.
func (f Trainingsfrequenz) Bezeichnung() string {
	if f <= 0 {
		return "keine Frequenz"
	}

	return fmt.Sprintf("%d× pro Woche", int(f))
}

// trainingsfrequenzAus liest die Frequenz aus einer Terminliste ab. Die
// Ableitung steht hier ein einziges Mal: Mitgliedschaft und Listeneintrag
// zeigen beide dieselbe Frequenz, und zwei Zählungen könnten wieder
// auseinanderlaufen — genau das, was die Ableitung verhindern soll.
//
// Archivierte Termine zählen dabei mit: der Termin ist aus dem Stundenplan
// genommen, die Vereinbarung über ihn steht weiter (ADR-0008). Den Stundenplan
// aufzuräumen darf nicht stillschweigend ändern, wozu Mitglieder angemeldet sind.
func trainingsfrequenzAus(termine []Trainingstermin) Trainingsfrequenz {
	return Trainingsfrequenz(len(termine))
}

// Frequenzfilter grenzt die Ergebnisliste nach der Trainingsfrequenz ein. Die
// Stufen tragen ihre Frequenz als Wert, deshalb ist der Vergleich unten ein
// Vergleich und keine Zuordnungstabelle.
//
// Eine Stufe für "kein Training" gibt es bewusst nicht: gefragt wird nach den
// Trainierenden einer Frequenz ("alle 2×-Trainierenden auf einmal"), und wer
// gar keinen Termin hat, ist keine solche Gruppe.
type Frequenzfilter int

const (
	// FrequenzfilterAlle ist der Nullwert und grenzt nicht ein.
	FrequenzfilterAlle Frequenzfilter = 0
	// FrequenzfilterEinmal: nur Mitglieder mit genau einem Trainingstermin.
	FrequenzfilterEinmal Frequenzfilter = 1
	// FrequenzfilterZweimal: nur Mitglieder mit genau zwei Trainingsterminen.
	FrequenzfilterZweimal Frequenzfilter = 2
	// FrequenzfilterDreimal: nur Mitglieder mit genau drei Trainingsterminen.
	FrequenzfilterDreimal Frequenzfilter = 3
)

// trifft entscheidet, ob eine Frequenz durch diesen Filter kommt.
func (f Frequenzfilter) trifft(frequenz Trainingsfrequenz) bool {
	if f == FrequenzfilterAlle {
		return true
	}

	return int(f) == int(frequenz)
}

// trainingsterminIDsNormalisieren wirft weg, was keine Auswahl ist: den
// Nullwert und jede Wiederholung. Doppelt angekreuzt ist derselbe Termin und
// nicht zwei — sonst zählte er zweimal für die Frequenz und einmal zu viel
// gegen die Obergrenze.
//
// Normalisiert wird hier, wo geprüft und gespeichert wird: sonst prüfte die
// Regel etwas anderes, als hinterher in der Datenbank steht.
func trainingsterminIDsNormalisieren(ids []int64) []int64 {
	gesaeubert := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id > 0 && !slices.Contains(gesaeubert, id) {
			gesaeubert = append(gesaeubert, id)
		}
	}

	return gesaeubert
}

// trainingsterminIDsPruefen meldet den Verstoß gegen die Obergrenze — als Text
// und nicht als Fehler, damit der Aufrufer ihn zu den übrigen Meldungen seiner
// Eingabe legen kann.
//
// Ob es die Termine gibt und ob sie noch zu haben sind, steht hier nicht: das
// weiß nur die Datenbank und prüft deshalb trainingstermineSchreiben.
func trainingsterminIDsPruefen(ids []int64) []string {
	if len(trainingsterminIDsNormalisieren(ids)) > MaxTrainingstermine {
		return []string{zuVieleTermine}
	}

	return nil
}

// trainingstermineSchreiben meldet eine Mitgliedschaft für genau die
// übergebenen Termine an: erst alles weg, dann neu. Eine Zuordnung hat außer
// ihren beiden Verweisen nichts, woran sich eine einzelne wiedererkennen ließe
// — ein Abgleich Zeile für Zeile wäre deshalb aufwendiger und nicht genauer.
//
// Zwei Regeln prüft erst diese Stelle, weil beide die Datenbank brauchen: es
// muss den Termin geben, und ein archivierter lässt sich nicht neu vergeben.
// Bereits zugeordnet bleibt er dagegen erlaubt — sonst wäre eine beliebige
// Änderung am Mitglied nicht mehr speicherbar, solange ein alter Termin
// angekreuzt ist, und das Aufräumen des Stundenplans risse Löcher in Formulare,
// die damit nichts zu tun haben.
//
// Dass es nicht mehr als MaxTrainingstermine sind, hat die Validierung des
// Aufrufers bereits geprüft.
func trainingstermineSchreiben(tx *sql.Tx, mitgliedschaftID int64, ids []int64) error {
	ids = trainingsterminIDsNormalisieren(ids)

	bisher, err := zugeordneteTerminIDs(tx, mitgliedschaftID)
	if err != nil {
		return err
	}

	if err := terminzuordnungPruefen(tx, ids, bisher); err != nil {
		return err
	}

	if _, err := tx.Exec(
		`DELETE FROM mitgliedschaft_trainingstermin WHERE mitgliedschaft_id = ?`,
		mitgliedschaftID); err != nil {
		return fmt.Errorf("trainingstermine von mitgliedschaft %d abmelden: %w", mitgliedschaftID, err)
	}

	for _, id := range ids {
		if _, err := tx.Exec(
			`INSERT INTO mitgliedschaft_trainingstermin (mitgliedschaft_id, trainingstermin_id)
			 VALUES (?, ?)`, mitgliedschaftID, id); err != nil {
			return fmt.Errorf("mitgliedschaft %d für trainingstermin %d anmelden: %w",
				mitgliedschaftID, id, err)
		}
	}

	return nil
}

// terminzuordnungPruefen geht die gewünschten Termine durch und sammelt alle
// Verstöße auf einmal — wer zwei archivierte angekreuzt hat, soll das nicht
// nacheinander erfahren.
//
// bisher sind die Termine, für die die Mitgliedschaft schon angemeldet ist. Nur
// sie dürfen archiviert sein.
func terminzuordnungPruefen(tx *sql.Tx, ids []int64, bisher []int64) error {
	var meldungen []string

	for _, id := range ids {
		termin, err := trainingsterminLesen(tx.QueryRow(
			`SELECT id, wochentag, beginn, ende, bezeichnung, archiviert
			 FROM trainingstermin WHERE id = ?`, id))
		if errors.Is(err, sql.ErrNoRows) {
			meldungen = append(meldungen, unbekannterTermin)

			continue
		}
		if err != nil {
			return fmt.Errorf("trainingstermin %d lesen: %w", id, err)
		}

		if termin.Archiviert && !slices.Contains(bisher, id) {
			meldungen = append(meldungen, archivierterTermin(termin))
		}
	}

	if len(meldungen) > 0 {
		return &ValidierungsFehler{Meldungen: meldungen}
	}

	return nil
}

// zugeordneteTerminIDs liefert die Termine einer einzelnen Mitgliedschaft als
// bloße IDs. Gefragt wird danach nur, um einen bereits vergebenen archivierten
// Termin wiederzuerkennen — dafür sind die vier Angaben des Termins nichts wert.
func zugeordneteTerminIDs(tx *sql.Tx, mitgliedschaftID int64) ([]int64, error) {
	zeilen, err := tx.Query(
		`SELECT trainingstermin_id FROM mitgliedschaft_trainingstermin
		 WHERE mitgliedschaft_id = ?`, mitgliedschaftID)
	if err != nil {
		return nil, fmt.Errorf("trainingstermine von mitgliedschaft %d lesen: %w", mitgliedschaftID, err)
	}
	defer zeilen.Close()

	var ids []int64
	for zeilen.Next() {
		var id int64
		if err := zeilen.Scan(&id); err != nil {
			return nil, fmt.Errorf("trainingstermin von mitgliedschaft %d lesen: %w", mitgliedschaftID, err)
		}
		ids = append(ids, id)
	}

	return ids, zeilen.Err()
}

// trainingstermineLesen liefert die Termine der angegebenen Mitgliedschaften,
// je Mitgliedschaft in Wochenreihenfolge.
//
// Gelesen wird in zwei Abfragen statt in einem Verbund: der Verbund
// vervielfachte die Zeilen der Mitgliedschaft, und die Spaltenliste des Termins
// stünde ein zweites Mal in Go. Stattdessen kommen die Verweise für sich und
// die Termine als ganzer Stundenplan — der hat eine Handvoll Zeilen, und seine
// Reihenfolge ist dann genau die, die auch die Terminansicht zeigt.
func (s *MemberService) trainingstermineLesen(mitgliedschaftIDs []int64) (map[int64][]Trainingstermin, error) {
	zugeordnet := make(map[int64][]Trainingstermin, len(mitgliedschaftIDs))

	angemeldet, err := s.terminanmeldungenLesen(mitgliedschaftIDs)
	if err != nil {
		return nil, err
	}
	if len(angemeldet) == 0 {
		return zugeordnet, nil
	}

	// Mit den archivierten: sie zählen weiter zur Frequenz und müssen deshalb
	// auch lesbar bleiben.
	termine, err := s.ListTrainingstermine(true)
	if err != nil {
		return nil, err
	}

	for _, termin := range termine {
		for _, mitgliedschaftID := range angemeldet[termin.ID] {
			zugeordnet[mitgliedschaftID] = append(zugeordnet[mitgliedschaftID], termin)
		}
	}

	return zugeordnet, nil
}

// terminanmeldungenLesen liefert die Zuordnungen der angegebenen
// Mitgliedschaften, geordnet nach dem Termin: zu jedem Termin die
// Mitgliedschaften, die für ihn angemeldet sind. Diese Richtung, weil der
// Aufrufer anschließend den Stundenplan durchgeht und nicht die Mitglieder.
func (s *MemberService) terminanmeldungenLesen(mitgliedschaftIDs []int64) (map[int64][]int64, error) {
	angemeldet := map[int64][]int64{}
	if len(mitgliedschaftIDs) == 0 {
		return angemeldet, nil
	}

	// Die Platzhalter entstehen aus der Anzahl der IDs, die Werte gehen als
	// Parameter in die Anweisung.
	platzhalter := strings.Repeat(", ?", len(mitgliedschaftIDs)-1)
	werte := make([]any, 0, len(mitgliedschaftIDs))
	for _, id := range mitgliedschaftIDs {
		werte = append(werte, id)
	}

	zeilen, err := s.db.Query(
		`SELECT mitgliedschaft_id, trainingstermin_id FROM mitgliedschaft_trainingstermin
		 WHERE mitgliedschaft_id IN (?`+platzhalter+`)`, werte...)
	if err != nil {
		return nil, fmt.Errorf("trainingstermine der mitgliedschaften lesen: %w", err)
	}
	defer zeilen.Close()

	for zeilen.Next() {
		var mitgliedschaftID, terminID int64
		if err := zeilen.Scan(&mitgliedschaftID, &terminID); err != nil {
			return nil, fmt.Errorf("trainingstermin einer mitgliedschaft lesen: %w", err)
		}
		angemeldet[terminID] = append(angemeldet[terminID], mitgliedschaftID)
	}

	return angemeldet, zeilen.Err()
}
