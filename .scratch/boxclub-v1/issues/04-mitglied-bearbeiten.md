Status: ready-for-human

# 04: Mitglied bearbeiten

**What to build:** Aus der Mitgliederliste kann der Nutzer ein Mitglied anklicken, ein Bearbeitungsformular öffnen, Stammdaten ändern und speichern. Die Liste zeigt die aktualisierten Werte sofort.

**Blocked by:** 03 (Mitgliederliste anzeigen)

## Acceptance Criteria

- [x] Klick auf einen Listeneintrag öffnet ein Formular, vorbefüllt mit den aktuellen Daten des Mitglieds
- [x] Bearbeitbar sind: Vorname, Nachname, Adresse, E-Mail, Telefon, Beitragsklasse
- [x] Geburtsdatum und Eintrittsdatum können in v1 **nicht** verändert werden (bewusst — Tippfehler-Schutz)
- [x] Nach dem Speichern zeigt die Liste die aktualisierten Werte ohne Full-Page-Reload
- [x] `MemberService.Update(id, patch)` schreibt nur die im Patch enthaltenen Felder
- [x] Test am Seam: Create → Update → Get liefert die aktualisierten Werte, unveränderte Felder bleiben unverändert
- [x] Test am Seam: `Update` auf nicht existierende ID gibt einen erkennbaren Fehler (kein Silent-Fail, kein Neuanlegen)
- [x] Test: Beitragsklassen-Wechsel funktioniert und die neue Klasse ist in der Listenansicht sichtbar

## Comments

### Umsetzung (Code-Review eingearbeitet)

- Seam: `MemberService.Update(id, MitgliedPatch)`. Im Patch bedeutet ein nicht gesetztes (nil-)Feld "unverändert"; Geburtsdatum und Eintritt fehlen im Patch-Typ bewusst ganz, damit die Unveränderlichkeit strukturell gilt und nicht nur im Formular.
- `Mitglied.LaufendeMitgliedschaft()` kam neu dazu: das Bearbeitungsformular zeigt den Eintritt des laufenden Zeitraums an, und diese Lebenszyklus-Frage gehört laut ADR-0002 in `service/`, nicht in den Adapter. Ticket 07 (Aus-/Wiedereintritt) baut darauf auf.
- Eine unbekannte ID liefert im htmx-Pfad **200 mit der Liste plus Warnhinweis**, keinen 404: htmx tauscht Antworten mit Fehlerstatus nicht ein, ein Klick auf eine veraltete Zeile bliebe sonst sichtbar wirkungslos.
- Verworfen: Sonderbehandlung deaktivierter Beitragsklassen im Formular. In v1 kann keine Klasse deaktiviert werden (Ticket 08 ist read-only), der Pfad wäre unerreichbar.
- Verifiziert: `go test ./...` grün, `wails build` unter Linux grün. `wails dev` und Windows wurden **nicht** gegengeprüft.
