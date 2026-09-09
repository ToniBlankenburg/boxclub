Status: ready-for-agent

# 04: Mitglied bearbeiten

**What to build:** Aus der Mitgliederliste kann der Nutzer ein Mitglied anklicken, ein Bearbeitungsformular öffnen, Stammdaten ändern und speichern. Die Liste zeigt die aktualisierten Werte sofort.

**Blocked by:** 03 (Mitgliederliste anzeigen)

## Acceptance Criteria

- [ ] Klick auf einen Listeneintrag öffnet ein Formular, vorbefüllt mit den aktuellen Daten des Mitglieds
- [ ] Bearbeitbar sind: Vorname, Nachname, Adresse, E-Mail, Telefon, Beitragsklasse
- [ ] Geburtsdatum und Eintrittsdatum können in v1 **nicht** verändert werden (bewusst — Tippfehler-Schutz)
- [ ] Nach dem Speichern zeigt die Liste die aktualisierten Werte ohne Full-Page-Reload
- [ ] `MemberService.Update(id, patch)` schreibt nur die im Patch enthaltenen Felder
- [ ] Test am Seam: Create → Update → Get liefert die aktualisierten Werte, unveränderte Felder bleiben unverändert
- [ ] Test am Seam: `Update` auf nicht existierende ID gibt einen erkennbaren Fehler (kein Silent-Fail, kein Neuanlegen)
- [ ] Test: Beitragsklassen-Wechsel funktioniert und die neue Klasse ist in der Listenansicht sichtbar
