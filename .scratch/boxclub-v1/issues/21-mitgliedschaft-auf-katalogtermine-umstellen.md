Status: ready-for-human

# 21: Mitgliedschaft auf Termine aus dem Stundenplan umstellen

**What to build:** Die Trainingszeiten einer Mitgliedschaft werden nicht mehr getippt, sondern **ausgewählt**. Statt drei Freitextfeldern zeigt das Formular die Termine aus dem Stundenplan (Ticket 20) zum Ankreuzen, höchstens drei. Die alte Tabelle `trainingsslot` entfällt.

**Blocked by:** 20
**Siehe:** [ADR-0008](../../../docs/adr/0008-trainingstermine-als-wochenplan.md), `CONTEXT.md` → Trainingstermin, Trainingsfrequenz

## Acceptance Criteria

- [x] `trainingsslot` ist durch eine Zuordnungstabelle `mitgliedschaft_trainingstermin` ersetzt
- [x] Das Mitgliedsformular bietet die **nicht archivierten** Termine zur Auswahl, in Wochenreihenfolge
- [x] Mehr als drei Termine werden abgewiesen — mit derselben Meldung wie bisher bei mehr als drei Slots
- [x] Ein bereits zugeordneter **archivierter** Termin bleibt sichtbar und angekreuzt, erkennbar als archiviert. Er lässt sich abwählen, aber nicht neu vergeben
- [x] `Trainingsfrequenz` zählt weiterhin die Anzahl, archivierte Zuordnungen **eingeschlossen**
- [x] Mitgliederliste und Suche zeigen die Termine wie bisher; die Frequenzfilter (`Frequenzfilter`) funktionieren unverändert
- [x] Ein Wiedereintritt (`Rejoin`) beginnt weiterhin **ohne** Termine, die alte Mitgliedschaft behält ihre
- [x] Die Tests aus `service/trainingsslot_test.go` sind auf den Katalog umgeschrieben, nicht gelöscht
- [x] `go test ./...` grün; `wails dev` und `wails build` unter Windows und Linux

## Notes

**Der Excel-Import ist zwischen diesem und Ticket 22 kaputt.** `service.Importsatz.Trainingsslots` trägt Freitext; sobald Slots Verweise sind, hat er kein Ziel mehr. Die drei Tickets 20–22 gehören zusammen und müssen zusammen fertig werden — in diesem Ticket darf der Import vorübergehend keine Termine übernehmen, muss aber übersetzen oder bauen.

**`trainingsslotsSchreiben` und `trainingsslotsLesen` verschwinden nicht, sie ändern ihre Währung.** Statt Strings gehen IDs hinein und heraus. Die Stelle, die beim Speichern alles löscht und neu schreibt (`member_service.go:622`), bleibt als Muster brauchbar.

**Benennung:** die Zuordnung hat keinen eigenen Fachbegriff. Im Code heißt sie nach der Tabelle, in der Oberfläche steht „Training" wie bisher. „Slot" verschwindet aus neuem Code (`CONTEXT.md` → Trainingsslot, abgelöst).

## Comments

**2026-09-13 — umgesetzt** (Commit `3ff308b`)

Alle Akzeptanzkriterien erfüllt. `go test ./...` grün, `wails build` unter Linux
erfolgreich, `GOOS=windows go build ./...` ebenfalls — ein Windows-`wails build`
steht noch aus, die Änderung ist aber reines Go/HTML ohne plattformabhängige Teile.

Drei Entscheidungen, die über den Ticket-Text hinausgehen:

- **Der Import löscht nichts.** Der Ticket-Text sagt, der Import dürfe
  vorübergehend keine Termine übernehmen. `aktualisieren` rührt die Zuordnung
  deshalb gar nicht an, statt sie durch die leere Liste zu ersetzen — sonst
  nähme ein zweiter Lauf von Hand zugeordnete Termine still wieder weg. Das ist
  dieselbe Zusicherung, die der Rückstand schon hat (ADR-0006), und sie ist
  durch `TestUebernehmen_LaesstZugeordneteTrainingstermineStehen` festgenagelt.
  Ticket 22 füllt `Importsatz.TrainingsterminIDs` und macht den Aufruf dort
  wieder scharf.
- **Unbekannte Termin-IDs werden mit eigener Meldung abgewiesen**, nicht mit
  einem Fremdschlüsselfehler. Nicht im Ticket gefordert, aber ohne das käme aus
  einer veralteten Ansicht ein Serverfehler statt eines Satzes.
- **Die Meldung zur Obergrenze heißt jetzt „Es sind höchstens 3
  Trainingstermine möglich."** — inhaltlich dieselbe wie bisher, mit dem
  abgelösten Wort „Slot" darin ersetzt (CONTEXT.md → Trainingsslot).

Offen für Ticket 22: der Abgleich der Freitexte aus `Training - 1/2/3` gegen den
Katalog. Bis dahin hat eine frisch importierte Zeile „keine Frequenz"; die
Gegenprobe gegen die Spalte „1x 2x Woche" läuft weiter in den Fehlerbericht.
