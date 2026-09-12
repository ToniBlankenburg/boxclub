Status: ready-for-agent

# 21: Mitgliedschaft auf Termine aus dem Stundenplan umstellen

**What to build:** Die Trainingszeiten einer Mitgliedschaft werden nicht mehr getippt, sondern **ausgewählt**. Statt drei Freitextfeldern zeigt das Formular die Termine aus dem Stundenplan (Ticket 20) zum Ankreuzen, höchstens drei. Die alte Tabelle `trainingsslot` entfällt.

**Blocked by:** 20
**Siehe:** [ADR-0008](../../../docs/adr/0008-trainingstermine-als-wochenplan.md), `CONTEXT.md` → Trainingstermin, Trainingsfrequenz

## Acceptance Criteria

- [ ] `trainingsslot` ist durch eine Zuordnungstabelle `mitgliedschaft_trainingstermin` ersetzt
- [ ] Das Mitgliedsformular bietet die **nicht archivierten** Termine zur Auswahl, in Wochenreihenfolge
- [ ] Mehr als drei Termine werden abgewiesen — mit derselben Meldung wie bisher bei mehr als drei Slots
- [ ] Ein bereits zugeordneter **archivierter** Termin bleibt sichtbar und angekreuzt, erkennbar als archiviert. Er lässt sich abwählen, aber nicht neu vergeben
- [ ] `Trainingsfrequenz` zählt weiterhin die Anzahl, archivierte Zuordnungen **eingeschlossen**
- [ ] Mitgliederliste und Suche zeigen die Termine wie bisher; die Frequenzfilter (`Frequenzfilter`) funktionieren unverändert
- [ ] Ein Wiedereintritt (`Rejoin`) beginnt weiterhin **ohne** Termine, die alte Mitgliedschaft behält ihre
- [ ] Die Tests aus `service/trainingsslot_test.go` sind auf den Katalog umgeschrieben, nicht gelöscht
- [ ] `go test ./...` grün; `wails dev` und `wails build` unter Windows und Linux

## Notes

**Der Excel-Import ist zwischen diesem und Ticket 22 kaputt.** `service.Importsatz.Trainingsslots` trägt Freitext; sobald Slots Verweise sind, hat er kein Ziel mehr. Die drei Tickets 20–22 gehören zusammen und müssen zusammen fertig werden — in diesem Ticket darf der Import vorübergehend keine Termine übernehmen, muss aber übersetzen oder bauen.

**`trainingsslotsSchreiben` und `trainingsslotsLesen` verschwinden nicht, sie ändern ihre Währung.** Statt Strings gehen IDs hinein und heraus. Die Stelle, die beim Speichern alles löscht und neu schreibt (`member_service.go:622`), bleibt als Muster brauchbar.

**Benennung:** die Zuordnung hat keinen eigenen Fachbegriff. Im Code heißt sie nach der Tabelle, in der Oberfläche steht „Training" wie bisher. „Slot" verschwindet aus neuem Code (`CONTEXT.md` → Trainingsslot, abgelöst).
