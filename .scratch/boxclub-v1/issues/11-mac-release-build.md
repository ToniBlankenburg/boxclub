Status: ready-for-agent

# 11: Mac-Release-Build

**What to build:** Der erste ship-fähige `.app`-Bundle für macOS, den der Vereinsadmin auf seinem Mac starten kann. Enthält eine kleine Vorab-Entscheidung ("gelegentlicher Mac vs. `macos-latest`-GitHub-Actions-Runner") plus die eigentliche Umsetzung.

**Blocked by:** 09 (Excel-Import), 10 (Dev-Setup-Dokumentation) — und über 09 transitiv die Modell-Umbauten 12–18. Das UI-Refactoring in `.scratch/ui-refactoring/` (eigener Feature-Slug) war zusätzlich ein Blocker — erledigt, siehe Comments.

## Acceptance Criteria

- [ ] Entscheidung dokumentiert (Kommentar in diesem Ticket): Mac-Build auf gelegentlich zugänglichem Mac vs. `macos-latest`-Runner auf GitHub Actions — mit Begründung
- [ ] `wails build` erzeugt ein `.app`-Bundle für macOS (arm64 oder universal)
- [ ] Das Bundle startet auf einem Mac ohne installierten Go-Compiler und ohne Node.js
- [ ] SQLite-Datei liegt an einem für macOS sinnvollen Ort (typischerweise `~/Library/Application Support/boxclub/`, nicht neben dem `.app`)
- [ ] Rauchtest im Bundle grün: Mitglied anlegen → in Liste sehen → Rückstand setzen → Kennzeichen wechselt auf rot → Rückstand aufheben → Kennzeichen wieder grün
- [ ] Signierung mindestens als Ad-hoc-Signatur; volle Notarisierung (Apple Developer ID) ist v1.5, nicht v1

## Notes

Ist explizit **Ship-Vorbereitung**, kein Kern-Feature — daher am Ende der Blockerkette.

**Blocker aktualisiert.** Ursprünglich hing dieses Ticket an 06/07/08 als "den letzten Kern-Feature-Tickets". Das trägt nicht mehr: der Modell-Umbau aus [ADR-0005](../../../docs/adr/0005-beitrag-individuell-statt-beitragsklasse.md) und [ADR-0006](../../../docs/adr/0006-rueckstand-statt-bezahlt-bis.md) ändert Schema und Bedienung nach 08 noch erheblich, und ein Release davor wäre sofort veraltet. Entscheidend ist außerdem: **ab dem ersten Release wird jede Schemaänderung zur echten Datenmigration**, weil es keinen Migrationsmechanismus gibt. Deshalb muss der Umbau vollständig vor dem Release liegen.

## Comments

### Zusätzlicher Blocker: UI-Refactoring (2026-09-17)

Vor diesem Ticket läuft jetzt noch ein UI-Refactoring (aufgeräumter,
übersichtlicher, im Stil von wiso mein Verein/WISO Steuer) als eigener
Feature-Slug `.scratch/ui-refactoring/`. Vier Prototyp-Runden sind
abgeschlossen (Branches `prototype/mitgliederliste-layout`,
`prototype/stilrichtungen`, `prototype/vier-restbereiche`,
`prototype/dokument-ablage`; Fold-Commits `8fa5778`, `f3a19fa`;
[ADR-0010](../../../docs/adr/0010-navigation-bleibt-horizontale-top-leiste.md)),
Spec und Tickets dafür stehen noch aus. Der erste Mac-Release soll das neue
UI zeigen, nicht das alte — deshalb wartet dieses Ticket zusätzlich zu den
Modell-Umbauten auf den Abschluss dieses Feature-Slugs.

### UI-Refactoring-Blocker erledigt (2026-09-19)

`.scratch/ui-refactoring/issues/07-freigabe-ticket-11-entsperren.md` hat die
Freigabe erteilt: alle Tickets 01–06 stehen auf `ready-for-human`,
`go build ./...`, `go vet ./...`, `go test ./...` sowie `wails build`
(Linux) laufen app-weit fehlerfrei, und der Durchlauf aller sechs Bereiche
(Mitglieder, Dashboard, Trainingstermine, Import, Rechnung, Verein) zeigt
keine vergessenen Rot-Reste außerhalb der zwei bewusst roten Warnstellen
(Formular-Fehlerlisten, Rückstand-Kennzeichen inkl. dessen Ableger auf
Dashboard und Import-Bericht). Dieses Ticket ist damit entsperrt und darf
starten.
