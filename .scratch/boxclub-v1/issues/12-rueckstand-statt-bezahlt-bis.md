Status: ready-for-agent

# 12: Rückstand statt `bezahlt_bis`

**What to build:** Der Vereinsadmin kann ein Mitglied als **im Rückstand** markieren, wenn dessen Lastschrift zurückgekommen ist, eine Notiz zum Vorgang hinterlegen und den Rückstand wieder aufheben. Die Liste zeigt pro Zeile ein zweiwertiges Kennzeichen und lässt sich darauf filtern. Das Datum `bezahlt_bis` verschwindet ersatzlos.

**Blocked by:** None (kann sofort starten)

Begründung und Kostenseite: [ADR-0006](../../../docs/adr/0006-rueckstand-statt-bezahlt-bis.md). Dieses Ticket **ersetzt Ticket 05**.

## Acceptance Criteria

- [ ] `bezahlt_bis` ist aus Schema, Service, Templates und Tests entfernt
- [ ] Am Mitglied gibt es ein Rückstands-Kennzeichen und eine freie Notiz dazu
- [ ] Der Rückstand lässt sich aus der Mitgliederliste heraus setzen und wieder aufheben
- [ ] Beim Setzen kann eine Notiz erfasst werden ("Rücklastschrift Oktober, angeschrieben am 05.10."), die später änderbar ist
- [ ] Die Listenzeile zeigt ein **zweiwertiges** Kennzeichen: grün für *in Ordnung*, rot für *im Rückstand*. Es gibt **keinen** dritten oder neutralen Zustand mehr
- [ ] Filter über der Liste: alle / nur im Rückstand / nur in Ordnung
- [ ] Ein Rückstand bleibt bestehen, wenn das Mitglied austritt — geprüft am Service-Seam
- [ ] Die Suche trifft zusätzlich die **Mitglieds-ID**, damit die Nummer direkt eingegeben werden kann (Story 5; das Suchprädikat wird in diesem Ticket ohnehin angefasst)
- [ ] Kein Rest von `bezahlt_bis`, "nicht bezahlt" oder "nicht gesetzt" in Code, Templates oder Tests
- [ ] `go test ./...` ist grün; `wails dev` und `wails build` laufen unter Windows und Linux

## Notes

Es gibt keinen Migrationsmechanismus — das Schema entsteht über `CREATE TABLE IF NOT EXISTS`. Die Entwicklungs-Datenbank (`BOXCLUB_DB`) muss vor dem ersten Start gelöscht werden. Genau deshalb passiert dieser Umbau jetzt und nicht nach dem ersten Mac-Release.
