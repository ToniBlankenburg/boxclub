Status: ready-for-agent

# 11: Mac-Release-Build

**What to build:** Der erste ship-fähige `.app`-Bundle für macOS, den der Vereinsadmin auf seinem Mac starten kann. Enthält eine kleine Vorab-Entscheidung ("gelegentlicher Mac vs. `macos-latest`-GitHub-Actions-Runner") plus die eigentliche Umsetzung.

**Blocked by:** 09 (Excel-Import), 10 (Dev-Setup-Dokumentation) — und über 09 transitiv die Modell-Umbauten 12–18

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
