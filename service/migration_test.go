package service_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/ToniBlankenburg/boxclub/service"
)

// TestOpenErgaenztFehlendeSpaltenEinerAelterenDatenbank bildet eine
// Datenbank nach, die seit vor ADR-0012/ADR-0016 durchläuft: vereinsdaten ohne
// logo/logo_mime/moneymoney_verwendungszweck, mitglied ohne die seither
// hinzugekommenen Angaben. CREATE TABLE IF NOT EXISTS allein lässt eine
// bestehende Tabelle unangetastet — ohne spaltenErgaenzen bricht
// GetVereinsdaten mit "no such column: logo" ab, und genau das war der Grund,
// warum sich der Verein-Reiter auf einer länger laufenden Installation nicht
// mehr öffnen ließ.
func TestOpenErgaenztFehlendeSpaltenEinerAelterenDatenbank(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "boxclub.db")

	altesSchemaAnlegen(t, pfad)

	svc, err := service.Open(pfad)
	if err != nil {
		t.Fatalf("service.Open einer älteren Datenbank: %v", err)
	}
	defer svc.Close()

	if _, err := svc.GetVereinsdaten(); err != nil {
		t.Errorf("GetVereinsdaten nach der Migration: %v", err)
	}

	mitglieder, err := svc.List()
	if err != nil {
		t.Fatalf("List nach der Migration: %v", err)
	}
	if len(mitglieder) != 1 {
		t.Fatalf("List = %d Mitglieder, erwartet 1", len(mitglieder))
	}

	// Die vor der Migration angelegte Zeile bleibt lesbar, und die neu
	// ergänzten Spalten zeigen den Nullwert — nicht geraten, sondern "diese
	// Angabe kannte die Zeile noch nicht".
	m := mitglieder[0]
	if m.Nachname != "Bestand" {
		t.Errorf("Nachname = %q, erwartet %q (die Zeile von vor der Migration)", m.Nachname, "Bestand")
	}
	if m.Rueckstand.Offen {
		t.Errorf("Rueckstand.Offen = true, erwartet false (Nullwert einer nachgezogenen Spalte)")
	}
}

// TestOpenSpaltenErgaenzenIstWiederholbar sichert zu, dass ein zweiter Start
// gegen dieselbe (bereits ergänzte) Datei kein "duplicate column" wirft — im
// echten Betrieb läuft migrate() bei jedem Programmstart erneut.
func TestOpenSpaltenErgaenzenIstWiederholbar(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "boxclub.db")

	altesSchemaAnlegen(t, pfad)

	erst, err := service.Open(pfad)
	if err != nil {
		t.Fatalf("erstes service.Open: %v", err)
	}
	if err := erst.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	zweit, err := service.Open(pfad)
	if err != nil {
		t.Fatalf("zweites service.Open (dieselbe, bereits ergänzte Datei): %v", err)
	}
	defer zweit.Close()

	if _, err := zweit.GetVereinsdaten(); err != nil {
		t.Errorf("GetVereinsdaten nach erneutem Open: %v", err)
	}
}

// altesSchemaAnlegen legt eine Datenbank mit dem Stand von kurz nach der
// Ablösung der Beitragsklasse an (siehe CONTEXT.md → Beitrag, ADR-0005) —
// mitgliedschaft kennt bereits beitrag_monatlich_cents (die vom Nutzer
// bewusst ausgeschlossene Grenze dieses Fixes), aber noch keine der später
// hinzugekommenen Spalten. Reines SQL statt der service-API, weil genau
// dieser veraltete Zustand nachgebildet werden soll, den service.Open() heute
// so nicht mehr anlegt.
func altesSchemaAnlegen(t *testing.T, pfad string) {
	t.Helper()

	db, err := sql.Open("sqlite", pfad)
	if err != nil {
		t.Fatalf("sqlite öffnen: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
CREATE TABLE mitglied (
	id           INTEGER PRIMARY KEY AUTOINCREMENT,
	vorname      TEXT    NOT NULL,
	nachname     TEXT    NOT NULL,
	geburtsdatum TEXT,
	adresse      TEXT    NOT NULL DEFAULT '',
	email        TEXT    NOT NULL DEFAULT '',
	telefon      TEXT    NOT NULL DEFAULT ''
);

CREATE TABLE mitgliedschaft (
	id                      INTEGER PRIMARY KEY AUTOINCREMENT,
	mitglied_id             INTEGER NOT NULL REFERENCES mitglied(id) ON DELETE CASCADE,
	eintritt                TEXT    NOT NULL,
	austritt                TEXT,
	beitrag_monatlich_cents INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE vereinsdaten (
	id             INTEGER PRIMARY KEY CHECK (id = 1),
	name           TEXT NOT NULL DEFAULT '',
	adresse        TEXT NOT NULL DEFAULT '',
	postleitzahl   TEXT NOT NULL DEFAULT '',
	ort            TEXT NOT NULL DEFAULT '',
	email          TEXT NOT NULL DEFAULT '',
	telefon        TEXT NOT NULL DEFAULT '',
	iban           TEXT NOT NULL DEFAULT '',
	bic            TEXT NOT NULL DEFAULT '',
	kreditinstitut TEXT NOT NULL DEFAULT '',
	fusszeile      TEXT NOT NULL DEFAULT ''
);
INSERT INTO vereinsdaten (id) VALUES (1);

INSERT INTO mitglied (vorname, nachname, adresse, email, telefon)
	VALUES ('Alt', 'Bestand', 'Musterweg 1', 'alt@example.com', '0123');
INSERT INTO mitgliedschaft (mitglied_id, eintritt, beitrag_monatlich_cents)
	VALUES (1, '2020-01-01', 6000);
`)
	if err != nil {
		t.Fatalf("altes Schema anlegen: %v", err)
	}
}
