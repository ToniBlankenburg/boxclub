Status: ready-for-human

# 14: Adresse in Straße, Postleitzahl und Ort trennen

**What to build:** Statt eines einzigen Adressfeldes pflegt der Vereinsadmin Straße samt Hausnummer, Postleitzahl und Ort **getrennt** — so wie in seiner bestehenden Excel-Tabelle, die dafür drei Spalten führt.

**Blocked by:** None (kann sofort starten)

## Acceptance Criteria

- [x] Am Mitglied gibt es drei getrennte Felder: Adresse (Straße und Hausnummer), Postleitzahl, Ort
- [x] Alle drei sind im Anlegen- und im Bearbeiten-Formular als eigene Eingaben vorhanden
- [x] Die Mitgliederliste stellt die Anschrift lesbar dar, ohne die Zeile zu überfüllen
- [x] Alle drei Felder sind optional — ein Mitglied ohne Anschrift bleibt anlegbar
- [x] Der Suchumfang bleibt unverändert (Vorname, Nachname, E-Mail, Telefon, Mitglieds-ID); Adressfelder werden **nicht** durchsucht
- [x] Tests am Service-Seam: Anlegen und Bearbeiten mit allen drei Feldern, sowie mit leeren Feldern
- [ ] `go test ./...` grün; `wails dev` und `wails build` unter Windows und Linux — Linux grün, **Windows steht aus** (siehe Kommentar)

## Notes

Dies ist das einzige Ticket dieser Reihe, das ein **bestehendes** Feld umformt statt neue hinzuzufügen. Deshalb steht es eigenständig und nicht im Sammelposten von Ticket 16.

## Comments

### Umgesetzt (Claude, 2026-09-11)

- `service.Anschrift` (`service/anschrift.go`) bündelt die drei Felder `Adresse` (Straße + Hausnummer), `Postleitzahl` und `Ort` samt den Schreibweisen, die die Oberfläche zeigt (`Leer`, `OrtZeile`, `Einzeilig`). Sie hängt an `Mitglied`, `NeuesMitglied` und `Listeneintrag`; `MitgliedPatch` ändert sie als Ganzes, weil ein Umzug alle drei Angaben betrifft.
- Schema: `mitglied.postleitzahl` und `mitglied.ort` neu, beide `TEXT NOT NULL DEFAULT ''`. Glossareintrag in `CONTEXT.md` → Anschrift, Schemazeile in `CLAUDE.md` nachgezogen.
- Liste: neue Spalte „Anschrift", zweizeilig (Straße / PLZ Ort), abgeschnitten mit vollständigem Tooltip, „—" wenn nichts erfasst ist. Die aufklappenden Zeilen für Rückstand und Mitgliedschaft spannen jetzt `colspan="4"`.
- Der Suchumfang ist unverändert und durch `TestSearch_TrifftKeineAnschriftsfelder` gegen Regression abgesichert.

**Achtung beim Aktualisieren:** Es gibt keinen Migrationsmechanismus — eine bestehende Entwicklungs-Datenbank läuft ab jetzt in `no such column: postleitzahl`. Die `boxclub.db` (bzw. `BOXCLUB_DB`) muss gelöscht werden.

**Offen:** `go test ./...`, `go vet` und `wails build` sind unter **Linux** grün. **Windows ist auf dieser Maschine nicht prüfbar** und steht noch aus.
