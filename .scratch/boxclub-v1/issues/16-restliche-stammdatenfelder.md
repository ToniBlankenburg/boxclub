Status: ready-for-agent

# 16: Restliche Stammdatenfelder

**What to build:** Die letzten sechs Felder der bestehenden Excel-Tabelle bekommen einen Ort in der App: **IBAN**, **Geschlecht**, **Google-Bewertung** und **Digital** am Mitglied, **Anmeldedatum** und **Anmeldegebühr** an der Mitgliedschaft. Damit hält die App jede Spalte der alten Tabelle — das ist die Voraussetzung dafür, dass die Excel danach wirklich abgeschaltet werden kann.

**Blocked by:** 13 (Beitrag individuell) — Anmeldedatum und Anmeldegebühr hängen an der Mitgliedschaft, die dort umstrukturiert wird

## Acceptance Criteria

- [ ] Am Mitglied: IBAN, Geschlecht, Google-Bewertung, Digital
- [ ] An der Mitgliedschaft: Anmeldedatum, Anmeldegebühr
- [ ] Alle sechs sind im Anlegen- und im Bearbeiten-Formular vorhanden und optional
- [ ] Die **IBAN** wird als reiner Text gespeichert: **keine Validierung**, keine Formatprüfung, keine SEPA-Erzeugung ([ADR-0006](../../../docs/adr/0006-rueckstand-statt-bezahlt-bis.md))
- [ ] **Geschlecht** ist ein Freitextfeld, kein erzwungenes Dropdown — die Excel-Werteliste ist eine Eintipphilfe, keine Einschränkung
- [ ] **Google-Bewertung** ist zweiwertig: hat bewertet / hat nicht bewertet
- [ ] **Digital** wird als reiner Freitext geführt, **ohne Semantik im Modell**. Die Bedeutung der Spalte ist unbekannt; sie fährt mit, damit beim Import keine Daten verloren gehen. Keine Prüfung, kein Dropdown, kein Glossareintrag
- [ ] Die **Anmeldegebühr** wird in Cent gespeichert und als Euro-Betrag eingegeben; sie ist ein historischer Wert und wird nicht neu berechnet
- [ ] Das **Anmeldedatum** ist vom Eintritt getrennt und liegt in der Regel davor
- [ ] Die Listenzeile bleibt lesbar: die neuen Felder erscheinen im Formular, in der Liste nur dort, wo sie dem Überblick dienen
- [ ] Tests am Service-Seam: Anlegen und Bearbeiten mit allen Feldern, mit leeren Feldern, Euro-nach-Cent bei der Anmeldegebühr
- [ ] `go test ./...` grün; `wails dev` und `wails build` unter Windows und Linux

## Notes

Sobald die Bedeutung von `Digital` bekannt ist, ist das eine eigene, kleine Änderung — Feld umbenennen und einen Typ geben. Bis dahin ist Freitext bewusst die richtige Antwort: Daten behalten, ohne eine Bedeutung zu erfinden.

Entwicklungs-Datenbank vor dem ersten Start löschen.
