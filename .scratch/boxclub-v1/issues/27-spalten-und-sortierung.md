Status: ready-for-human

# 27: Spalten ein- und ausblenden, nach Spalte sortieren

**What to build:** Die Mitgliederliste führt seit Ticket 19 acht Spalten und ist an der Fensterbreite angekommen. Der Admin soll selbst entscheiden, **welche** Spalten er sieht — und die Liste **nach jeder** davon sortieren können, nicht nur nach Nachname.

**Blocked by:** —
**Siehe:** [ADR-0004](../../../docs/adr/0004-suche-und-filter-im-speicher.md), Kommentar in Ticket 19

## Acceptance Criteria

- [x] Klick auf eine Spaltenüberschrift sortiert danach; erneuter Klick dreht die Richtung um. Die aktive Spalte und die Richtung sind sichtbar
- [x] Sortiert wird **serverseitig im Speicher** (ADR-0004), nicht in SQL und nicht im Browser — wie Suche und Filter heute
- [x] Sortierung bleibt bei Suche und Filter erhalten und umgekehrt
- [x] Sinnvoll sortiert wird je Spaltentyp: Namen alphabetisch, Beträge und Nummern numerisch, Daten chronologisch, Status in der Reihenfolge des Lebenszyklus und nicht alphabetisch
- [x] Leere Werte landen **immer am Ende**, in beiden Richtungen. Ein Mitglied ohne Anschrift darf nicht die halbe erste Seite füllen
- [x] Spalten lassen sich über ein Menü ein- und ausblenden. **Name bleibt immer sichtbar** — eine Liste ohne Namen ist keine
- [x] Die Spaltenwahl überlebt den Neustart (`localStorage`, kein Go, keine Datenbank). Sie gilt pro Gerät; bei einer lokalen Einzelplatz-App ist das dasselbe
- [x] Eine ausgeblendete Spalte lässt sich nicht als Sortierspalte auswählen; ist die aktive Sortierspalte ausgeblendet, fällt die Sortierung auf Nachname zurück
- [x] Service-Tests für jede Sortierordnung inklusive Leerwerten
- [ ] `go test ./...` grün; `wails dev` und `wails build` unter Windows und Linux — grün und für Linux sowie für Windows cross-kompiliert (siehe Kommentar), `wails dev` und ein echter Windows-Lauf stehen aus

## Notes

**Warum die Spaltenwahl im Browser und die Sortierung auf dem Server:** die Sortierung ändert, *welche* Zeilen oben stehen, und gehört damit zu derselben Verarbeitung wie Suche und Filter (ADR-0004). Die Spaltenwahl ändert nur, was man von einer Zeile sieht — eine Ansichtsvorliebe eines einzigen Benutzers an einem einzigen Rechner. Dafür muss nichts durch Go.

**`nachNamenSortieren` bleibt der Standard** (`member_service.go:1515`) und wird zum Sonderfall einer allgemeineren Sortierung, nicht ersetzt.

**Das Platzproblem wird damit gelöst, nicht verschoben.** Ticket 19 hatte es an Anschrift und Training weitergereicht, die beide schon abschneiden statt umzubrechen.

## Comments

**Sortierung:** `service.Sortierung` (Spalte + Richtung, `service/sortierung.go`) ist ein
neues Feld an `Suchfilter` und reist damit im selben Aufruf wie Suchbegriff und die
übrigen Filter — genau das verlangt "bleibt erhalten und umgekehrt". `eintraegeSortieren`
ersetzt den bisherigen alleinigen Aufruf von `nachNamenSortieren` in `Search`;
`nachNamenSortieren` selbst ist unangetastet und wird für den Nullwert (Name,
aufsteigend) weiter direkt aufgerufen — der im Ticket angekündigte Sonderfall, nicht
ersetzt. Leere Werte (Anschrift, Training) landen unabhängig von der Richtung am Ende;
0 € Beitrag zählt ausdrücklich nicht als leer (ADR-0005). Anschrift sortiert nach Ort,
dann Straße; Rückstand nach "in Ordnung" vor "im Rückstand"; Status nach dem
Lebenszyklus (`status.go`-Reihenfolge), nicht alphabetisch. Service-Tests für alle acht
Spalten in beiden Richtungen plus Leerwert- und Filterkombination:
`service/sortierung_test.go`.

**Spaltenwahl:** localStorage-Schlüssel `boxclub.mitgliederliste.spalten`, Menü in der
Filterleiste (`<details>`, kein JS zum Auf-/Zuklappen nötig). Ausblendbar sind Status,
Anschrift, Training, Beitrag, Rückstand, Eintritt — Nr. und Name bleiben immer sichtbar
und stehen deshalb nicht im Menü, lassen sich aber weiterhin sortieren. Ausgeblendet
wird über `data-col`-Attribute an den betroffenen Zellen; die drei Inline-Formulare
(Kündigung, Wiedereintritt, Rückstand) tragen `data-colspan-optional`, damit ihr
`colspan` mitschrumpft und die Zeile nicht über den sichtbaren Rand der Tabelle
hinausreicht. Das ist der einzige Ort im Frontend mit eigener JS-Logik (`main.js`) —
alles andere bleibt bei reinen `hx-*`-Attributen, wie der Kopfkommentar dort jetzt auch
vermerkt.

**Sortierfallback bei ausgeblendeter Spalte:** läuft im Browser (`main.js`,
`sortierfallbackPruefen`), nicht in Go — der Server weiß nichts von der Sichtbarkeit.
Wird die gerade aktive Sortierspalte ausgeblendet, stößt `main.js` einen Request ohne
`sort`/`richtung` an, was serverseitig bereits der Standard "Name, aufsteigend" ist.

**Wie Sortierung und Filterleiste zusammenspielen:** ein Klick auf eine Kopfzeile
ersetzt nur `#mitglieder-ergebnis`, nicht die Filterleiste — deshalb reist die
Sortierung als verstecktes Feld *in* diesem Fragment mit, und die Filterleiste holt sie
per `hx-include="#mitglieder-ergebnis"` beim nächsten Suchen/Filtern wieder ab.
Umgekehrt nimmt jede Kopfzeile Suchbegriff und Filter per `hx-include` aus der
Filterleiste mit.

**Geprüft:** `go test ./...` grün, `go vet` sauber, `wails build` für linux/amd64 und
windows/amd64 erfolgreich (Cross-Kompilierung von Linux aus). Ein Wegwerf-Test hat
beim Umsetzen die gerenderte HTML-Ausgabe von `/api/mitglieder` geprüft (korrekte
`hx-vals`, `aria-sort`, `data-col`); er ist nicht im Baum geblieben (CLAUDE.md: `app/`
und Templates werden nicht unit-getestet). **Offen:** `wails dev` und ein echter Lauf
unter Windows.
