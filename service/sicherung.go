package service

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
)

// SicherungErstellen schreibt eine vollständige Kopie der Datenbank nach ziel
// (CONTEXT.md → Sicherung). VACUUM INTO liefert dabei einen in sich
// stimmigen Stand, auch wenn gerade geschrieben wird, und eine Datei, die
// sich unverändert als boxclub.db öffnen lässt.
//
// Eine vorhandene Datei am Ziel wird ersetzt — den Überschreiben-Dialog stellt
// der Datei-Dialog des Betriebssystems. VACUUM INTO selbst verweigert ein
// vorhandenes Ziel, deshalb wird erst daneben geschrieben und dann umbenannt:
// scheitert die Kopie, bleibt eine ältere Sicherung heil.
func (s *MemberService) SicherungErstellen(ziel string) error {
	zwischen := ziel + ".neu"
	if err := os.Remove(zwischen); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("zwischendatei entfernen: %w", err)
	}

	if _, err := s.db.Exec(`VACUUM INTO ?`, zwischen); err != nil {
		return fmt.Errorf("sicherung schreiben: %w", err)
	}

	// 0600: die Datenbank ist personenbezogen wie ihr Verzeichnis (ADR-0003).
	if err := os.Chmod(zwischen, 0o600); err != nil {
		os.Remove(zwischen)
		return fmt.Errorf("rechte der sicherung setzen: %w", err)
	}

	if err := os.Rename(zwischen, ziel); err != nil {
		os.Remove(zwischen)
		return fmt.Errorf("sicherung ablegen: %w", err)
	}

	return nil
}

// EndungVorWiederherstellung hängt an dem Pfad der Datenbank, unter dem
// Wiederherstellen den bisherigen Stand ablegt, bevor es ihn ersetzt.
const EndungVorWiederherstellung = ".vor-restore"

// ErrKeineSicherung meldet, dass die gewählte Datei keine Sicherung dieser App
// ist: keine SQLite-Datei, beschädigt oder ohne die Tabelle der Mitglieder.
var ErrKeineSicherung = errors.New("keine sicherung")

// Wiederherstellen ersetzt die gesamte Datenbank durch die Sicherung unter
// quelle — kein Zusammenführen (CONTEXT.md → Sicherung). Der bisherige Stand
// bleibt als Datei neben der Datenbank liegen (EndungVorWiederherstellung);
// eine ältere Kopie dieser Art wird dabei ersetzt.
//
// Danach ist der Service geschlossen: die Verbindung zeigte auf die alte Datei,
// und was die App im Speicher hält, gehört zu ihr. Der Aufrufer beendet die
// App; beim nächsten Start ergänzt migrate() eine ältere Sicherung um
// fehlende Spalten (ADR-0019). Eine abgelehnte Quelle lässt den Service
// unberührt.
func (s *MemberService) Wiederherstellen(quelle string) error {
	if err := sicherungPruefen(quelle); err != nil {
		return err
	}

	if err := s.SicherungErstellen(s.dbPath + EndungVorWiederherstellung); err != nil {
		return fmt.Errorf("bisherigen stand sichern: %w", err)
	}

	zwischen := s.dbPath + ".neu"
	if err := dateiKopieren(quelle, zwischen); err != nil {
		os.Remove(zwischen)
		return fmt.Errorf("sicherung kopieren: %w", err)
	}

	if err := s.db.Close(); err != nil {
		os.Remove(zwischen)
		return fmt.Errorf("datenbank schließen: %w", err)
	}

	if err := os.Rename(zwischen, s.dbPath); err != nil {
		os.Remove(zwischen)
		return fmt.Errorf("datenbank ersetzen: %w", err)
	}

	return nil
}

// sicherungPruefen liest quelle schreibgeschützt und verlangt eine heile
// SQLite-Datei mit der Tabelle der Mitglieder.
func sicherungPruefen(quelle string) error {
	if _, err := os.Stat(quelle); err != nil {
		return fmt.Errorf("%w: %v", ErrKeineSicherung, err)
	}

	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=ro", quelle))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrKeineSicherung, err)
	}
	defer db.Close()

	var ergebnis string
	if err := db.QueryRow(`PRAGMA integrity_check`).Scan(&ergebnis); err != nil {
		return fmt.Errorf("%w: %v", ErrKeineSicherung, err)
	}
	if ergebnis != "ok" {
		return fmt.Errorf("%w: integrity_check meldet %q", ErrKeineSicherung, ergebnis)
	}

	var tabellen int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'mitglied'`).Scan(&tabellen); err != nil {
		return fmt.Errorf("%w: %v", ErrKeineSicherung, err)
	}
	if tabellen == 0 {
		return fmt.Errorf("%w: keine tabelle mitglied", ErrKeineSicherung)
	}

	return nil
}

func dateiKopieren(quelle, ziel string) error {
	ein, err := os.Open(quelle)
	if err != nil {
		return err
	}
	defer ein.Close()

	aus, err := os.OpenFile(ziel, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}

	if _, err := io.Copy(aus, ein); err != nil {
		aus.Close()
		return err
	}

	return aus.Close()
}
