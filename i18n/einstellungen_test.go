package i18n

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEinstellungenLaden_fehlendeDatei(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "einstellungen.json")

	e, err := EinstellungenLaden(pfad)
	if err != nil {
		t.Fatalf("EinstellungenLaden: %v", err)
	}
	if e != StandardEinstellungen {
		t.Errorf("EinstellungenLaden(fehlende Datei) = %+v, want %+v", e, StandardEinstellungen)
	}
}

func TestEinstellungenLaden_kaputteDatei(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "einstellungen.json")
	if err := os.WriteFile(pfad, []byte("das ist kein JSON"), 0o600); err != nil {
		t.Fatalf("Testdatei schreiben: %v", err)
	}

	if _, err := EinstellungenLaden(pfad); err == nil {
		t.Error("EinstellungenLaden(kaputte Datei) soll einen Fehler liefern")
	}
}

func TestEinstellungenLaden_unbekannteSprache(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "einstellungen.json")
	if err := os.WriteFile(pfad, []byte(`{"sprache":"fr"}`), 0o600); err != nil {
		t.Fatalf("Testdatei schreiben: %v", err)
	}

	e, err := EinstellungenLaden(pfad)
	if err != nil {
		t.Fatalf("EinstellungenLaden: %v", err)
	}
	if e != StandardEinstellungen {
		t.Errorf("EinstellungenLaden(unbekannte Sprache) = %+v, want %+v (Standard)", e, StandardEinstellungen)
	}
}

func TestSpeichernUndLaden(t *testing.T) {
	// Verzeichnis existiert noch nicht — Speichern muss es anlegen, genau wie
	// main.go es für den Datenbankordner tut.
	pfad := filepath.Join(t.TempDir(), "Boxclub", "einstellungen.json")

	will := Einstellungen{Sprache: Englisch}
	if err := will.Speichern(pfad); err != nil {
		t.Fatalf("Speichern: %v", err)
	}

	e, err := EinstellungenLaden(pfad)
	if err != nil {
		t.Fatalf("EinstellungenLaden: %v", err)
	}
	if e != will {
		t.Errorf("EinstellungenLaden nach Speichern = %+v, want %+v", e, will)
	}
}
