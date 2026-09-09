Status: ready-for-agent

# 03: Mitgliederliste anzeigen

**What to build:** Statt eines einzelnen Mitglieds zeigt die App jetzt eine Tabelle **aller aktiven Mitglieder**. Neue Mitglieder erscheinen nach dem Anlegen sofort in der Liste. Das ist die Arbeitspferd-Ansicht der App.

**Blocked by:** 02 (SQLite-Bootstrap & erstes Mitglied anlegen)

## Acceptance Criteria

- [ ] Tabelle zeigt für jedes aktive Mitglied mindestens: Name, Beitragsklasse, `bezahlt_bis`, Eintrittsdatum
- [ ] "Aktiv" heißt: es existiert eine `mitgliedschaft` mit `austritt IS NULL`
- [ ] Nach dem Anlegen eines neuen Mitglieds erscheint es sofort in der Liste (htmx-Fragment-Swap, kein Full-Page-Reload)
- [ ] `MemberService.List` liefert nur aktive Mitgliedschaften zurück, sortiert nach Nachname/Vorname
- [ ] Test am Seam: mehrere Mitglieder angelegt → `List` gibt sie in konsistenter Reihenfolge zurück
- [ ] Test am Seam: ein Mitglied ohne aktive Mitgliedschaft (z. B. später ausgetreten, aus Test-Fixture) erscheint **nicht** in `List`
- [ ] Leere Liste rendert eine sinnvolle Leer-Ansicht ("Keine Mitglieder erfasst")

## Notes

Die Liste ist das Herzstück, an dem alle späteren Slices (Suche, Filter, Bearbeiten, Status) andocken. Sauberes Fragment-Muster hier zahlt sich mehrfach aus.
