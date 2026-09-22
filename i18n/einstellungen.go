package i18n

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Einstellungen ist der Teil der Programmkonfiguration, der nichts mit den
// Mitgliederdaten zu tun hat und deshalb nicht in vereinsdaten steht
// (ADR-0017) — heute nur die Sprache, absichtlich als eigener Typ statt eines
// losen Strings, falls später weitere programmweite Einstellungen dazukommen.
type Einstellungen struct {
	Sprache Sprache `json:"sprache"`
}

// StandardEinstellungen gelten, solange keine Einstellungsdatei existiert
// oder sie eine unbekannte Sprache enthält: Deutsch, weil die App bislang
// ausschließlich deutsch war und eine bestehende Installation nicht
// unvermittelt englisch starten soll.
var StandardEinstellungen = Einstellungen{Sprache: Deutsch}

// EinstellungenLaden liest die Einstellungen aus pfad. Fehlt die Datei,
// liefert es StandardEinstellungen ohne Fehler — eine frische Installation
// hat noch keine geschrieben. Eine vorhandene, aber unlesbare Datei ist ein
// Fehler; eine vorhandene Datei mit unbekannter Sprache dagegen nicht — sie
// liefert ebenfalls StandardEinstellungen, damit eine von Hand verstümmelte
// Sprachangabe die App nicht am Start hindert (siehe main.go).
func EinstellungenLaden(pfad string) (Einstellungen, error) {
	inhalt, err := os.ReadFile(pfad)
	if errors.Is(err, os.ErrNotExist) {
		return StandardEinstellungen, nil
	}
	if err != nil {
		return Einstellungen{}, err
	}

	var e Einstellungen
	if err := json.Unmarshal(inhalt, &e); err != nil {
		return Einstellungen{}, err
	}
	if !e.Sprache.Gueltig() {
		return StandardEinstellungen, nil
	}

	return e, nil
}

// Speichern schreibt e nach pfad und legt das Elternverzeichnis bei Bedarf
// an, genau wie main.go es für den Datenbankordner tut.
func (e Einstellungen) Speichern(pfad string) error {
	inhalt, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(pfad), 0o700); err != nil {
		return err
	}

	return os.WriteFile(pfad, inhalt, 0o600)
}
