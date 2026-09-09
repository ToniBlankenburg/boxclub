Status: ready-for-agent

# 09: Excel-Import

**What to build:** Der Nutzer kann seine bestehende `.xlsx`-Tabelle in die App laden. Die App zeigt einen Bericht, welche Zeilen erfolgreich importiert wurden und welche mit welchem Grund gescheitert sind. Der Import ist **wiederholbar** — ein zweiter Lauf aktualisiert bestehende Mitglieder statt zu duplizieren.

**Blocked by:** 02 (SQLite-Bootstrap), **extern: Nutzer liefert die Spaltennamen (und 2–3 Beispielzeilen) der bestehenden Excel-Tabelle**

## Acceptance Criteria

- [ ] Import-Aktion öffnet einen Datei-Dialog für `.xlsx`
- [ ] Nach dem Import zeigt ein Bericht: Anzahl erfolgreich importierter Zeilen, Anzahl fehlgeschlagener Zeilen, jede fehlgeschlagene Zeile einzeln mit Grund (z. B. "fehlender Nachname", "unbekannte Beitragsklasse", "ungültiges Datumsformat")
- [ ] Ein zweiter Import derselben Datei aktualisiert bestehende Mitglieder (Match-Regel: exakter Vor-/Nachname + Geburtsdatum, alternativ E-Mail — die genaue Regel wird beim Ticket festgezurrt, sobald die Spaltennamen vorliegen)
- [ ] `ExcelImporter` liest von einem `io.Reader` und liefert die geparsten Zeilen + Fehlerbericht zurück, **ohne** selbst in die DB zu schreiben
- [ ] Ein separater Orchestrator im Wails-Layer setzt die geparsten Zeilen per `MemberService.Create` / `Update` in die DB
- [ ] Test am `ExcelImporter`-Seam mit Fixture-`.xlsx` in `testdata/`: gute Zeilen, kaputte Zeilen (fehlende Felder, ungültige Daten), Mix
- [ ] Test: Wiederholungslauf ist idempotent — keine Duplikate

## Notes

`ExcelImporter` liefert bewusst nur geparste Rows und Fehlerbericht — das entkoppelt Format-Parsing von Business-Regeln und legt das Muster für den späteren v2-MoneyMoney-CSV-Importer vor (siehe spec.md).

**Externer Blocker:** Bevor dieses Ticket gestartet werden kann, muss der Nutzer die Spaltennamen und 2–3 anonymisierte Beispielzeilen seiner heutigen Excel-Tabelle liefern. Ohne diese Info bleibt das `ColumnMapping` unklar.
