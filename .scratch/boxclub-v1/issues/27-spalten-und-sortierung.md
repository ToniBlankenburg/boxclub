Status: ready-for-agent

# 27: Spalten ein- und ausblenden, nach Spalte sortieren

**What to build:** Die Mitgliederliste führt seit Ticket 19 acht Spalten und ist an der Fensterbreite angekommen. Der Admin soll selbst entscheiden, **welche** Spalten er sieht — und die Liste **nach jeder** davon sortieren können, nicht nur nach Nachname.

**Blocked by:** —
**Siehe:** [ADR-0004](../../../docs/adr/0004-suche-und-filter-im-speicher.md), Kommentar in Ticket 19

## Acceptance Criteria

- [ ] Klick auf eine Spaltenüberschrift sortiert danach; erneuter Klick dreht die Richtung um. Die aktive Spalte und die Richtung sind sichtbar
- [ ] Sortiert wird **serverseitig im Speicher** (ADR-0004), nicht in SQL und nicht im Browser — wie Suche und Filter heute
- [ ] Sortierung bleibt bei Suche und Filter erhalten und umgekehrt
- [ ] Sinnvoll sortiert wird je Spaltentyp: Namen alphabetisch, Beträge und Nummern numerisch, Daten chronologisch, Status in der Reihenfolge des Lebenszyklus und nicht alphabetisch
- [ ] Leere Werte landen **immer am Ende**, in beiden Richtungen. Ein Mitglied ohne Anschrift darf nicht die halbe erste Seite füllen
- [ ] Spalten lassen sich über ein Menü ein- und ausblenden. **Name bleibt immer sichtbar** — eine Liste ohne Namen ist keine
- [ ] Die Spaltenwahl überlebt den Neustart (`localStorage`, kein Go, keine Datenbank). Sie gilt pro Gerät; bei einer lokalen Einzelplatz-App ist das dasselbe
- [ ] Eine ausgeblendete Spalte lässt sich nicht als Sortierspalte auswählen; ist die aktive Sortierspalte ausgeblendet, fällt die Sortierung auf Nachname zurück
- [ ] Service-Tests für jede Sortierordnung inklusive Leerwerten
- [ ] `go test ./...` grün; `wails dev` und `wails build` unter Windows und Linux

## Notes

**Warum die Spaltenwahl im Browser und die Sortierung auf dem Server:** die Sortierung ändert, *welche* Zeilen oben stehen, und gehört damit zu derselben Verarbeitung wie Suche und Filter (ADR-0004). Die Spaltenwahl ändert nur, was man von einer Zeile sieht — eine Ansichtsvorliebe eines einzigen Benutzers an einem einzigen Rechner. Dafür muss nichts durch Go.

**`nachNamenSortieren` bleibt der Standard** (`member_service.go:1515`) und wird zum Sonderfall einer allgemeineren Sortierung, nicht ersetzt.

**Das Platzproblem wird damit gelöst, nicht verschoben.** Ticket 19 hatte es an Anschrift und Training weitergereicht, die beide schon abschneiden statt umzubrechen.
