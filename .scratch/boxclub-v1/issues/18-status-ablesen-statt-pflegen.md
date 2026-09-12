Status: ready-for-human

# 18: Status ablesen statt pflegen

**What to build:** Der Vereinsadmin sieht pro Mitglied dessen **Status** — *Neu*, *Aktiv*, *In Kündigungsfrist* oder *Ausgetreten* — ohne ihn zu pflegen: er wird aus den Datumsfeldern abgelesen. Damit kann er nicht mehr im Widerspruch zu den Daten stehen, wie es in der handgepflegten Excel regelmäßig passiert. Gleichzeitig ändert sich, was "aktiv" bedeutet: ein Mitglied in der Kündigungsfrist bleibt in der Standardansicht, weil es weiter trainiert und weiter zahlt.

**Blocked by:** 17 (Kündigung und ruhend) — der Status *In Kündigungsfrist* braucht das dort angelegte Kündigungsdatum

## Acceptance Criteria

- [x] Der Status wird **abgeleitet und nirgends gespeichert**:
  - *Neu* — Eintritt liegt in der Zukunft
  - *Aktiv* — Eintritt erreicht, Austritt nicht erreicht
  - *In Kündigungsfrist* — Kündigungsdatum gesetzt, Austritt liegt in der Zukunft
  - *Ausgetreten* — Austritt erreicht
- [x] *Ruhend* wird **zusätzlich** angezeigt, nicht anstelle des Status — es ist ein Merkmal, kein Lebenszyklus-Zustand
- [x] **"Aktiv" ändert seine Definition**: bisher "kein Austrittsdatum gesetzt", künftig "Austritt nicht erreicht"
- [x] Der Aktivitätsfilter zieht nach: ein Mitglied mit **zukünftigem** Austrittsdatum erscheint in der Standardansicht "nur aktive"; erst ab dem Austrittstag verschwindet es dort
- [x] Der Status ist pro Zeile in der Mitgliederliste sichtbar
- [x] Randfälle getestet: Eintritt genau heute, Austritt genau heute, Kündigungsdatum ohne Austritt, Austritt in der Zukunft, mehrere Mitgliedschaften nacheinander
- [x] Die Excel-Werte `Mitglied` und `Aktiv` ergeben denselben Status; `Stillgelegt` und `Inakiv` sind beide *ruhend* und kein eigener Status
- [ ] `go test ./...` grün; `wails dev` und `wails build` unter Windows und Linux

## Notes

Hier bereinigt sich die aktive Liste künftig **von selbst**: niemand muss ein Mitglied "auf ausgetreten setzen", das Austrittsdatum erledigt es am Stichtag. Das ist der eigentliche Gewinn gegenüber der Excel, in der der Status eine Spalte war, die jemand von Hand nachziehen musste.

## Comments

**Umgesetzt.** Der Status liegt in [`service/status.go`](../../../service/status.go)
als eigener Typ mit vier Werten; abgelesen wird er in `statusAus` aus Eintritt,
Kündigungsdatum und Austritt gegen den heutigen Tag. Gespeichert wird nichts —
`Mitgliedschaft.Status()` und `Listeneintrag.Status()` leiten beide über
dieselbe Funktion ab, damit sie nicht auseinanderlaufen können. Beide Ränder
zählen einschließend: am Eintrittstag ist das Mitglied aktiv, am Austrittstag
bereits ausgetreten.

**„Aktiv" bedeutet jetzt „Austritt nicht erreicht"** — überall, wo vorher
`austritt IS NULL` stand: `LaufendeMitgliedschaft`, `laufendeMitgliedschaftLesen`,
`eintraegeAbfrage` und `massgeblicheMitgliedschaft` teilen sich dafür die
SQL-Bedingung `laeuftNoch`. Daraus folgen drei Dinge, die das Ticket nicht
aufzählt, aber zwingend nach sich zieht und die Ticket 17 schon angekündigt
hatte: eine Kündigung lässt sich **während der Frist korrigieren** (statt in
`ErrNichtAktiv` zu laufen), ein Mitglied in der Frist lässt sich **ruhend
schalten**, und `Rejoin` meldet dort `ErrBereitsAktiv`. Weil die Korrektur
damit erreichbar wurde, füllt `app.kuendigungsformular` das Austrittsfeld jetzt
mit dem erfassten Termin vor — käme es leer herauf, löschte eine Korrektur am
Kündigungsdatum den Austritt gleich mit (`SetKuendigung` schreibt beide Spalten
zusammen).

**Eine Abweichung vom Ticketwortlaut**, bewusst und in
[CONTEXT.md](../../../CONTEXT.md) → Status nachgezogen: das Ticket definiert
*In Kündigungsfrist* als „Kündigungsdatum gesetzt, Austritt liegt in der
Zukunft". Umgesetzt ist „**eine** Kündigung ist erfasst, der Austritt ist nicht
erreicht" — es genügt also eines der beiden Daten. Grund: ein künftiger Austritt
ohne notiertes Kündigungsdatum ist der Normalfall der Altbestände, den
`SetKuendigung` ausdrücklich erlaubt; nach dem engen Wortlaut stünde so ein
Mitglied als schlicht *Aktiv* in der Liste und die einzige erfasste Angabe über
sein Ausscheiden bliebe unsichtbar. *Aktiv* bleibt damit der Zustand, über
dessen Ende nichts erfasst ist.

**In der Liste** steht der Status als eigene Spalte hinter dem Namen. *Ruhend*
sitzt als zweites Kennzeichen **daneben**, nicht an seiner Stelle; darunter das
Datum, aus dem der Status sich ergibt. Die alten Ad-hoc-Kennzeichen
(„ausgetreten am…", „gekündigt am…") sind darin aufgegangen. Auch die
Schaltflächen der Zeile entscheiden jetzt über `Status.Ausgetreten` statt über
das bloße Vorhandensein eines Austrittsdatums — in der Kündigungsfrist führen
sie zurück ins Kündigungsformular und nicht zum Wiedereintritt.

**Zur Excel-Zeile im Ticket:** der Importer (Ticket 09) existiert noch nicht.
Modellseitig ist das Kriterium erfüllt — es gibt genau vier Zustände, `Mitglied`
und `Aktiv` fallen darin zusammen, und für `Stillgelegt`/`Inakiv` gibt es keinen
Status, sondern das Feld `ruhend` daneben.

**Randfälle getestet** in [`service/status_test.go`](../../../service/status_test.go):
Eintritt genau heute, Eintritt in der Zukunft, Austritt genau heute, Austritt in
der Zukunft (mit und ohne Kündigungsdatum), Kündigungsdatum ohne Austritt,
Kündigung vor dem Eintritt, ruhend neben dem Status, zwei Mitgliedschaften
nacheinander (auch die Variante „ausgetreten + laufende in Kündigungsfrist",
bei der beide ein Austrittsdatum tragen), Standardansicht vor und ab dem
Austrittstag. Alle Fixtures liegen relativ zu heute — feste Kalendertage wären
ab morgen etwas anderes.

**Aus dem Review nachgezogen:** `massgeblicheMitgliedschaft` sortierte noch nach
`austritt IS NULL` und behauptete im Kommentar, es sei „dieselbe Regel wie in
`eintraegeAbfrage`" — beide teilen sich jetzt tatsächlich `laeuftNoch`, damit
eine Änderung denselben Zeitraum trifft, den die Liste zeigt. Dazu: die veraltete
Zusicherung „Austritt == nil bedeutet: die Mitgliedschaft läuft" am
`Mitgliedschaft`-Typ ersetzt, `Status.Bezeichnung` nennt alle vier Fälle
ausdrücklich (ein unbekannter Wert las sich sonst als „Aktiv"), das ungenutzte
`Status.Aktiv()` entfernt, und die Zeile liest den Status einmal in `$status`
statt viermal die Uhr zu fragen.

**Was geprüft ist:** `go test ./...`, `go vet`, `gofmt` und `wails build` unter
Linux; dazu kompilieren `GOOS=windows` und `GOOS=darwin` sauber durch. **Was
offen ist:** ein echter `wails build` bzw. `wails dev` unter Windows — der
braucht eine Windows-Maschine und ist von hier aus nicht zu machen. Das eine
offene Häkchen oben bleibt deshalb offen, bis das jemand nachholt.

**Keine Schemaänderung** — der Status wird abgelesen, nicht gespeichert. Die
Entwicklungs-Datenbank kann stehen bleiben.
