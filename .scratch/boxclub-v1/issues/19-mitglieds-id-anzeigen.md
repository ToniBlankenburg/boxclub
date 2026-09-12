Status: ready-for-agent

# 19: Mitglieds-ID anzeigen

**What to build:** Der Vereinsadmin sieht zu jedem Mitglied dessen **Mitglieds-ID** — die laufende Nummer, unter der er es kennt und unter der es in seiner Excel steht. Bisher trägt die App die Nummer mit, zeigt sie aber nirgends: sie steckt nur in den htmx-Adressen der Listenzeile. Seit dem Excel-Import (Ticket 09) sind das die **echten Vereinsnummern** aus der Spalte `Mandatsreferenz`, nicht mehr irgendeine interne Zählung — damit wird sie zur Angabe, über die der Admin während der Umstellung beide Systeme vergleicht.

**Blocked by:** 09 (Excel-Import) — erst dadurch sind die IDs die Nummern des Vereins

## Acceptance Criteria

- [ ] Die Mitgliederliste zeigt die Mitglieds-ID pro Zeile
- [ ] Das Bearbeitungsformular zeigt sie ebenfalls — dort **nur lesend**: die Nummer ist die Identität eines Mitglieds über die Zeit und wird nie neu vergeben (`CONTEXT.md` → Mitglieds-ID)
- [ ] Die Spalte steht dort, wo der Admin sie sucht: vor dem Namen, nicht hinter den Zahlungsangaben
- [ ] Die Nummern sind untereinander vergleichbar gesetzt (`tabular-nums`), damit die Spalte beim Durchsehen als Spalte lesbar bleibt
- [ ] Die Suche findet ein Mitglied weiterhin über seine Nummer — das tut sie heute schon, es darf nur nicht kaputtgehen
- [ ] `go test ./...` grün; `wails dev` und `wails build` unter Windows und Linux

## Notes

**Kein Service-Ticket.** `service.Listeneintrag.MitgliedID` und `service.Mitglied.ID` existieren beide und werden in jeder htmx-Adresse der Liste benutzt — es fehlt allein die Anzeige. Das ist ein Template-Ticket, und es braucht keinen neuen Test am Service-Seam.

**Nicht zu verwechseln mit der Mandatsreferenz.** Die Excel-Spalte heißt so, enthält aber keine SEPA-Mandatsreferenz (`CONTEXT.md` → Mitglieds-ID). In der Oberfläche heißt die Angabe **Mitglieds-ID** oder kurz **Nr.**, nie „Mandatsreferenz".

**Offener Detailpunkt für die Umsetzung:** Die Listen-Kopfzeile führt heute sieben Spalten (Name, Status, Anschrift, Training, Beitrag, Rückstand, Eintritt) und ist damit schon breit. Eine achte Spalte kann sie über die Fensterbreite drücken. Beim Umsetzen prüfen, ob die Nummer als eigene schmale Spalte tragfähig ist oder besser in der Namenszelle über dem Namen steht — und die Entscheidung hier als Kommentar festhalten. Sie ist auch die Vorarbeit für Ticket zum Ein- und Ausblenden von Spalten, das aus demselben Platzproblem entsteht.
