Status: ready-for-human

# 19: Mitglieds-ID anzeigen

**What to build:** Der Vereinsadmin sieht zu jedem Mitglied dessen **Mitglieds-ID** — die laufende Nummer, unter der er es kennt und unter der es in seiner Excel steht. Bisher trägt die App die Nummer mit, zeigt sie aber nirgends: sie steckt nur in den htmx-Adressen der Listenzeile. Seit dem Excel-Import (Ticket 09) sind das die **echten Vereinsnummern** aus der Spalte `Mandatsreferenz`, nicht mehr irgendeine interne Zählung — damit wird sie zur Angabe, über die der Admin während der Umstellung beide Systeme vergleicht.

**Blocked by:** 09 (Excel-Import) — erst dadurch sind die IDs die Nummern des Vereins

## Acceptance Criteria

- [x] Die Mitgliederliste zeigt die Mitglieds-ID pro Zeile
- [x] Das Bearbeitungsformular zeigt sie ebenfalls — dort **nur lesend**: die Nummer ist die Identität eines Mitglieds über die Zeit und wird nie neu vergeben (`CONTEXT.md` → Mitglieds-ID)
- [x] Die Spalte steht dort, wo der Admin sie sucht: vor dem Namen, nicht hinter den Zahlungsangaben
- [x] Die Nummern sind untereinander vergleichbar gesetzt (`tabular-nums`), damit die Spalte beim Durchsehen als Spalte lesbar bleibt
- [x] Die Suche findet ein Mitglied weiterhin über seine Nummer — das tut sie heute schon, es darf nur nicht kaputtgehen
- [ ] `go test ./...` grün; `wails dev` und `wails build` unter Windows und Linux — grün und unter Linux gebaut, Windows steht aus (siehe Kommentar)

## Notes

**Kein Service-Ticket.** `service.Listeneintrag.MitgliedID` und `service.Mitglied.ID` existieren beide und werden in jeder htmx-Adresse der Liste benutzt — es fehlt allein die Anzeige. Das ist ein Template-Ticket, und es braucht keinen neuen Test am Service-Seam.

**Nicht zu verwechseln mit der Mandatsreferenz.** Die Excel-Spalte heißt so, enthält aber keine SEPA-Mandatsreferenz (`CONTEXT.md` → Mitglieds-ID). In der Oberfläche heißt die Angabe **Mitglieds-ID** oder kurz **Nr.**, nie „Mandatsreferenz".

**Offener Detailpunkt für die Umsetzung:** Die Listen-Kopfzeile führt heute sieben Spalten (Name, Status, Anschrift, Training, Beitrag, Rückstand, Eintritt) und ist damit schon breit. Eine achte Spalte kann sie über die Fensterbreite drücken. Beim Umsetzen prüfen, ob die Nummer als eigene schmale Spalte tragfähig ist oder besser in der Namenszelle über dem Namen steht — und die Entscheidung hier als Kommentar festhalten. Sie ist auch die Vorarbeit für Ticket zum Ein- und Ausblenden von Spalten, das aus demselben Platzproblem entsteht.

## Comments

**Entscheidung zum offenen Detailpunkt: eigene schmale Spalte, nicht in der
Namenszelle.** Die Nummer steht als achte Spalte ganz links vor dem Namen. Sie
ist dreistellig und bekommt mit `w-px` nur ihre eigene Breite — die sieben
bestehenden Spalten verlieren dadurch nichts, die Kopfzeile wird nicht über die
Fensterbreite gedrückt. Über dem Namen in dessen Zelle wäre sie zwar noch
sparsamer gewesen, aber dann keine Spalte mehr: die Nummern stünden untereinander
verschieden weit eingerückt, und `tabular-nums` nützt nichts, wenn schon die
linke Kante wandert. Vergleichbar untereinander zu stehen ist aber genau ihr
Zweck, solange der Verein beide Systeme nebeneinander führt.

Das Platzproblem ist damit **verschoben, nicht gelöst**. Die Reserve liegt jetzt
bei Anschrift und Training, die beide schon abschneiden statt umzubrechen
(`max-w-56` bzw. `max-w-48`) — die nächste Spalte zahlt der Nutzer dort. Das
bleibt die Vorarbeit für das Ticket zum Ein- und Ausblenden von Spalten.

**Die drei Inline-Formulare der Liste tragen die Nummer mit.** Kündigung,
Wiedereintritt und Rückstand ersetzen jeweils eine Zeile durch ein Formular und
bestanden aus Namenszelle plus `colspan="6"`. Ohne eine eigene Nummernzelle
stünde der Name dort unter „Nr." und die Zeile hätte eine Spalte zu wenig — die
Tabelle verrutschte, sobald man ein Formular öffnet. Sie haben jetzt dieselbe
erste Zelle wie die Listenzeile; der `colspan` bleibt bei 6.

**Im Formular steht die Nummer in der Überschrift, nicht als Feld.** „Mitglied bearbeiten"
und daneben „Nr. 42" — als Text, ohne Eingabefeld. Ein Feld, auch ein nur
lesendes wie Geburtsdatum und Eintritt, verspräche eine Angabe, die man pflegt;
die Mitglieds-ID wird aber einmal vergeben und nie neu (`CONTEXT.md` →
Mitglieds-ID). Beim Anlegen gibt es sie noch gar nicht, deshalb erscheint sie nur
im Bearbeiten-Modus. Nebeneffekt: das zweispaltige Raster bleibt unberührt — ein
Feld an erster Stelle hätte Vor- und Nachnamen auseinandergeschoben.

**Kein Service-Ticket, wie angekündigt** — `Listeneintrag.MitgliedID` und
`formularDaten.MitgliedID` gab es beide schon, Go-Code wurde nicht angefasst. Die
Suche über die Nummer läuft unverändert über `suchzeile` und ist unberührt.

**Geprüft:** `go test ./...` grün, `go vet` sauber, `wails build` unter Linux
erfolgreich. Ein Wegwerf-Test hat beim Umsetzen bestätigt, dass Kopf- und
Datenzeile je acht Zellen haben, die drei Inline-Formulare mit ihrem `colspan`
ebenfalls auf acht Spalten kommen, das Formular die Nummer ohne Eingabefeld zeigt
und die Suche nach der Nummer weiterhin trifft; er ist nicht im Baum geblieben
(CLAUDE.md: `app/` und Templates werden nicht unit-getestet). **Windows ist
offen** — dort wurde nicht gebaut.
