package service

import (
	"fmt"
	"time"
)

// AustrittFuerTest beendet die laufende Mitgliedschaft eines Mitglieds. Die
// Funktion existiert ausschließlich für Tests (sie steht in einer _test.go-Datei
// und ist außerhalb des Testlaufs nicht Teil des Package): die Liste muss
// beweisen können, dass ausgetretene Mitglieder verschwinden, bevor es einen
// fachlichen Austritt gibt.
//
// Sobald Ticket 07 MarkExit einführt, ersetzt dieser Aufruf sich selbst — dann
// bauen die Tests die Fixture über die reguläre API auf und diese Datei fällt weg.
func (s *MemberService) AustrittFuerTest(mitgliedID int64, austritt time.Time) error {
	res, err := s.db.Exec(
		`UPDATE mitgliedschaft SET austritt = ?
		 WHERE mitglied_id = ? AND austritt IS NULL`,
		austritt.Format(isoDatum), mitgliedID)
	if err != nil {
		return fmt.Errorf("austritt eintragen: %w", err)
	}

	betroffen, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("betroffene zeilen lesen: %w", err)
	}
	if betroffen == 0 {
		return fmt.Errorf("mitglied %d hat keine laufende mitgliedschaft: %w", mitgliedID, ErrNichtGefunden)
	}

	return nil
}
