Status: ready-for-agent

# 14: Adresse in Straße, Postleitzahl und Ort trennen

**What to build:** Statt eines einzigen Adressfeldes pflegt der Vereinsadmin Straße samt Hausnummer, Postleitzahl und Ort **getrennt** — so wie in seiner bestehenden Excel-Tabelle, die dafür drei Spalten führt.

**Blocked by:** None (kann sofort starten)

## Acceptance Criteria

- [ ] Am Mitglied gibt es drei getrennte Felder: Adresse (Straße und Hausnummer), Postleitzahl, Ort
- [ ] Alle drei sind im Anlegen- und im Bearbeiten-Formular als eigene Eingaben vorhanden
- [ ] Die Mitgliederliste stellt die Anschrift lesbar dar, ohne die Zeile zu überfüllen
- [ ] Alle drei Felder sind optional — ein Mitglied ohne Anschrift bleibt anlegbar
- [ ] Der Suchumfang bleibt unverändert (Vorname, Nachname, E-Mail, Telefon, Mitglieds-ID); Adressfelder werden **nicht** durchsucht
- [ ] Tests am Service-Seam: Anlegen und Bearbeiten mit allen drei Feldern, sowie mit leeren Feldern
- [ ] `go test ./...` grün; `wails dev` und `wails build` unter Windows und Linux

## Notes

Dies ist das einzige Ticket dieser Reihe, das ein **bestehendes** Feld umformt statt neue hinzuzufügen. Deshalb steht es eigenständig und nicht im Sammelposten von Ticket 16.
