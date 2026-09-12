Status: ready-for-agent

# 20: Trainingstermine als Stundenplan pflegen

**What to build:** Der Verein pflegt seine wöchentlichen Trainingstermine in einer **eigenen Ansicht**: anlegen, bearbeiten, archivieren. Bisher existiert der Stundenplan nirgends — er steckt als Freitext in 200 einzelnen Mitgliederangaben und sonst im Kopf des Trainers.

Ein Trainingstermin hat **Wochentag** (feste Auswahl Mo–So), **Beginn**, wahlweise ein **Ende** und eine **Bezeichnung** ("Anfänger", "Wettkampf"). Angezeigt wird er als "Samstag 10:30 – 12:00 · Anfänger", bei fehlenden Angaben entsprechend kürzer.

Dieses Ticket baut nur den Katalog. Die Zuordnung zu Mitgliedschaften kommt in Ticket 21 — bis dahin steht die neue Ansicht für sich, und die Mitgliederliste ist unverändert.

**Blocked by:** —
**Siehe:** [ADR-0008](../../../docs/adr/0008-trainingstermine-als-wochenplan.md), `CONTEXT.md` → Trainingstermin

## Acceptance Criteria

- [ ] Eigener Menüpunkt „Trainingstermine" in `templates/navigation.html`
- [ ] Die Ansicht listet alle nicht archivierten Termine **in Wochenreihenfolge** — Montag vor Dienstag, innerhalb des Tages nach Beginn. Der Wochentag ist deshalb als Zahl zu speichern und nicht als Text, sonst sortiert das Alphabet
- [ ] Anlegen und Bearbeiten über ein Formular: Wochentag (Auswahl, Pflicht), Beginn (Pflicht), Ende (freiwillig), Bezeichnung (freiwillig)
- [ ] Ein Ende vor dem Beginn wird abgewiesen
- [ ] **Archivieren** statt Löschen: der Termin verschwindet aus der Liste und aus jeder Auswahl, bleibt aber in der Datenbank. Ein Löschpfad wird nicht gebaut
- [ ] Archivierte Termine sind auf Wunsch einblendbar und **reaktivierbar** — sonst ist ein Fehlgriff endgültig
- [ ] Service-Tests am `MemberService`-Seam für: anlegen, ändern, archivieren, reaktivieren, Sortierreihenfolge, Ende-vor-Beginn
- [ ] `go test ./...` grün; `wails dev` und `wails build` unter Windows und Linux

## Notes

**Keine Kapazität, kein Trainer, kein Ort.** Bewusst weggelassen (ADR-0008): eine Platzzahl verlangt sofort Antworten auf Warteliste und Überbuchung, und Trainer wie Ort passen in die Bezeichnung, solange es eine Halle gibt.

**Der Termin ist wiederkehrend, nicht datiert.** Es gibt keine Instanz je Kalenderwoche und keine Stelle, an der stehen könnte, wer da war. Das ist die Grenze zum ausgeschlossenen attendance tracking und darf in diesem Ticket nicht aufgeweicht werden.

**Schemaänderung heißt Dev-Datenbank löschen** (`CLAUDE.md`). Es gibt keinen Migrationsmechanismus und produktiv noch keine Daten.
