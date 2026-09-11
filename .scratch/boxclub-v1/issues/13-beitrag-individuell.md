Status: ready-for-human

# 13: Beitrag individuell statt Beitragsklasse

**What to build:** Jedes Mitglied hat einen **individuell vereinbarten Monatsbeitrag**, der an seiner Mitgliedschaft hängt und frei gesetzt werden kann. Die Beitragsklassen verschwinden vollständig: Tabelle, Ansicht, Seeding und Filter. Eine frische Datenbank startet danach leer statt mit zwei vorgegebenen Klassen.

**Blocked by:** None (kann sofort starten)

Begründung und Kostenseite: [ADR-0005](../../../docs/adr/0005-beitrag-individuell-statt-beitragsklasse.md). Dieses Ticket **baut Ticket 08 zurück**.

## Acceptance Criteria

- [x] Die Tabelle `beitragsklasse` und der Fremdschlüssel am Mitglied sind entfernt
- [x] Es wird **nichts mehr geseedet** — eine frisch angelegte Datenbank enthält keine Zeilen
- [x] Der Beitrag hängt an der **Mitgliedschaft**, nicht am Mitglied, und wird in Cent gespeichert
- [x] Der Beitrag ist im Anlegen- und im Bearbeiten-Formular als Euro-Betrag eingebbar und wird verlustfrei in Cent umgerechnet
- [x] **0 € ist ein gültiger Beitrag** (Trainer zahlen nichts) und wird nicht als "nicht gesetzt" behandelt
- [x] Der Beitrag ist in der Mitgliederliste sichtbar
- [x] Die Beitragsklassen-Ansicht, ihr Template, der Navigationseintrag und die zugehörige Testdatei sind entfernt
- [x] Der Beitragsklassen-Filter ist entfernt; in der UI bleibt **kein toter Filter** stehen. Der Ersatz (Frequenzfilter) kommt in Ticket 15
- [x] Tests am Service-Seam: Beitrag setzen, ändern, 0 €, Euro-nach-Cent-Umrechnung
- [ ] `go test ./...` grün; `wails dev` und `wails build` unter Windows und Linux — `go test`/`go vet`/`wails build` unter Linux grün, **Windows ungeprüft**

## Notes

Entwicklungs-Datenbank vor dem ersten Start löschen (kein Migrationsmechanismus).

Achtung beim Rückbau: der Beitrag wandert **an die Mitgliedschaft**, nicht ans Mitglied. Ein Wiedereintritt bekommt dadurch seinen eigenen Beitrag, und der alte bleibt an der alten Mitgliedschaft stehen — das ist der Zweck der Aufteilung.

## Comments

**Umgesetzt.** Der Beitrag hängt an der **Mitgliedschaft**
(`mitgliedschaft.beitrag_monatlich_cents`), nicht am Mitglied — so wie dieses
Ticket und die [Spec](../spec.md) es verlangen. [ADR-0005](../../../docs/adr/0005-beitrag-individuell-statt-beitragsklasse.md)
sagte in der Entscheidung noch "Feld am Mitglied" und widersprach damit seinen
eigenen Konsequenzen; die Stelle ist nachgezogen, der Titel ebenso ("individuell
vereinbart" statt "individuell am Mitglied"). Inhaltlich ändert das nichts an der
Entscheidung.

**Seam:** `service.BeitragAusEuro` / `service.BeitragAlsEuro` in
[service/beitrag.go](../../../service/beitrag.go), Tests in
[service/beitrag_test.go](../../../service/beitrag_test.go). Die Umrechnung liegt
im Service und nicht im Adapter — das Ticket verlangt sie am Service-Seam
getestet, und dort gehört sie auch hin: gerechnet wird ganzzahlig, über `float64`
käme 0,29 € als 28,999… Cent an. Eingelesen werden "60", "60,50", "60.50" und
" 60 € ". Abgewiesen wird alles andere, **insbesondere der Tausenderpunkt**:
"1.234,56" ließe sich als 1234,56 € oder 1,23456 € lesen, und raten ist hier die
schlechteste Antwort — seit dem Wegfall des Fremdschlüssels fällt ein Tippfehler
im Beitrag durch nichts anderes mehr auf.

**0 € ist ein gültiger Betrag** und keine fehlende Angabe. Daraus folgt, dass ein
*leeres* Feld ein Fehler ist ("Bitte einen Beitrag angeben (0 für beitragsfrei).")
— die Null gehört hingeschrieben, sonst wäre "nicht ausgefüllt" stillschweigend
dasselbe wie "zahlt nichts".

**Update schreibt in zwei Tabellen.** Stammdaten und Beitrag liegen getrennt,
gehören aber zu einem Formular: `Update` läuft deshalb jetzt in einer
Transaktion. Der Beitrag landet auf der **maßgeblichen** Mitgliedschaft — der
laufenden, und wenn keine läuft, der zuletzt begonnenen. Das ist dieselbe
Sortierung, mit der `eintraegeAbfrage` die Zeile der Liste auswählt und die
`LetzteMitgliedschaft` liefert (als `massgeblicheMitgliedschaft` benannt und an
einer Stelle abgelegt). Damit ändert der Nutzer genau den Zeitraum, den er vor
sich sieht — auch bei einem Ehemaligen.

**Wiedereintritt:** die neue Mitgliedschaft startet mit dem zuletzt vereinbarten
Beitrag als Wert, änderbar wie jeder andere; die alte behält ihren eigenen. Die
Alternative — bei 0 € starten — wäre stiller Datenverlust gewesen, weil 0 € eben
ein gültiger Betrag ist und in der Liste nicht als "fehlt" auffiele. Das Ticket
fordert hier nichts, Story 23 der Spec verlangt genau das ("der alte Beitrag
bleibt erhalten, die neue Mitgliedschaft bekommt eigene").

**Zurückgebaut:** Tabelle `beitragsklasse`, `mitglied.beitragsklasse_id`, das
Seeding (eine frische Datenbank ist jetzt leer), `AktiveBeitragsklassen`,
`ListBeitragsklassen`, `Beitragsklasse(id)`, `Suchfilter.BeitragsklasseID`,
`GET /api/beitragsklassen`, `templates/beitragsklassen.html`,
`service/beitragsklassen_test.go` und der Navigationseintrag. Der
Klassen-Filter ist aus der Filterleiste verschwunden, ohne Platzhalter — der
Frequenzfilter kommt mit Ticket 15.

**Die Navigation bleibt stehen**, jetzt mit einem einzigen Eintrag. Sie samt
`hx-swap-oob`-Mechanik auszubauen, um sie in Ticket 17/18 wieder einzuziehen,
wäre teurer als der eine überflüssige Knopf. Falls sie bis zum Release allein
bleibt, gehört sie weg.

**Rauchtest:** gegen `app.Handler()` per `httptest` — Liste, Anlege- und
Bearbeiten-Formular, ein ungültiger Beitrag ("acht" → Formular kommt mit
Meldung zurück), Vorbelegung mit "65,50", Änderung auf 0 € ("0,00 €" in der
Liste), Suche/Filter mit allen Parametern, Aus- und Wiedereintritt samt
übernommenem Beitrag. Der Harnisch ist danach gelöscht: laut
[CLAUDE.md](../../../CLAUDE.md) werden `app/`-Handler nicht unit-getestet.

`go test ./...`, `go vet`, `gofmt` und `wails build` sind unter **Linux** sauber.
**Windows und macOS stehen aus** — die Änderung ist reines Go plus Templates ohne
plattformabhängigen Code, geprüft ist sie dort aber nicht.

**Vor dem ersten Start:** die Entwicklungs-Datenbank (`BOXCLUB_DB`) löschen. Es
gibt keinen Migrationsmechanismus, und `mitglied` hat eine Spalte weniger.
