Status: ready-for-human

# 08: Beitragsklassen ansehen

**What to build:** Eine eigene Ansicht listet die verfügbaren **Beitragsklassen** mit Namen und Monatspreis. In v1 read-only — Pflege der Klassen selbst ist bewusst nicht enthalten.

**Blocked by:** 02 (SQLite-Bootstrap)

## Acceptance Criteria

- [x] Ein Navigationseintrag "Beitragsklassen" öffnet eine schlichte Tabelle
- [x] Jede Klasse wird mit Name und Preis pro Monat angezeigt (Preis lesbar formatiert, z. B. "60,00 €")
- [x] `MemberService.ListBeitragsklassen` liefert alle Klassen (unabhängig davon, ob Mitglieder zugeordnet sind)
- [x] Test am Seam: nach dem Seed sind genau die zwei erwarteten Klassen mit den definierten Preisen sichtbar

## Notes

Klein aus Absicht — die Ansicht ist eher eine Referenz für den Admin und liefert später den natürlichen Andockpunkt, wenn Klassen editierbar werden sollen (nicht v1).

## Comments

**Umgesetzt.** Seam: `MemberService.ListBeitragsklassen()` in
[service/member_service.go](../../../service/member_service.go), Tests in
[service/beitragsklassen_test.go](../../../service/beitragsklassen_test.go).
Die Methode liefert alle Zeilen ohne `WHERE` — `AktiveBeitragsklassen` und sie
teilen sich nur noch das Auslesen der Zeilen (`beitragsklassenAbfragen`), jeder
Aufrufer bringt seine vollständige Anweisung mit. Bewusst kein zusammengesetztes
SQL-Fragment: ADR-0004 führt genau das als Gewinn.

**Oberfläche:** `GET /api/beitragsklassen` rendert
[templates/beitragsklassen.html](../../../templates/beitragsklassen.html) — Name
und Preis, sonst nichts. Der Preis läuft über die schon vorhandene
Template-Funktion `euro` und steht als "60,00 €" da.

**Navigation:** AC 1 verlangt einen Navigationseintrag, und ein Eintrag ohne
Rückweg zu den Mitgliedern wäre kaputt — also gibt es beide. Sie steht in der
Kopfzeile, also außerhalb von `#inhalt`, und kommt deshalb per `hx-swap-oob` an
jeder Antwort mit, die eine ganze Bereichsansicht ersetzt. Antworten innerhalb
eines Bereichs (Formulare, einzelne Zeilen) lassen sie weg; die Markierung
bleibt dann stehen, wie sie ist. Damit bleibt die Regel aus ADR-0002 intakt:
kein eigenes JavaScript, die Markierung entsteht im Go-Template.

`frontend/index.html` hat dafür ein leeres `<nav id="navigation">` als Platz;
der erste Aufruf beim Laden füllt es.

**Zwei Dinge, die im Review rausgeflogen sind** — der Vollständigkeit halber:

1. Ein Abzeichen "nicht mehr angeboten" für Klassen mit `aktiv = 0`. In v1 gibt
   es keinen Weg, `aktiv` auf 0 zu setzen — das war UI auf Vorrat für einen
   unerreichbaren Zustand. `ListBeitragsklassen` liefert solche Klassen aber
   sehr wohl mit; **wenn Klassen editierbar werden, muss die Ansicht sie
   kenntlich machen**, sonst liest sie sich falsch.
2. Ein Leerzustand "Keine Beitragsklassen hinterlegt." — nach dem Seed nie
   erreichbar.

**Nicht getestet:** dass `ListBeitragsklassen` auch *deaktivierte* Klassen
liefert. Über die Service-API lässt sich keine Klasse deaktivieren, und
[CLAUDE.md](../../../CLAUDE.md) verbietet Tests, die direkt in SQL greifen. Der
Test deckt ab, was das Ticket fordert: die geseedeten Klassen mit Preis, und
eine Klasse ohne zugeordnetes Mitglied, die trotzdem in der Liste steht
(geprüft am konkreten Datensatz, nicht an der Anzahl — sonst hinge der Test an
derselben Quelle, die er absichern soll).

**Rauchtest:** Beide Ansichten, das Mitgliederformular und eine leere
Navigation liefen gegen `app.Handler()` per `httptest` durch — geprüft wurden
die Preisformatierung, dass das `<nav>` oberstes Element der Antwort ist (sonst
greift der oob-Austausch nicht), dass genau ein Bereich markiert ist, dass ein
Formular die Navigation nicht anfasst und dass eine leere Navigation gar kein
`<nav>` rendert (ein leeres würde die Kopfzeile löschen). Der Harnisch ist
danach gelöscht — laut CLAUDE.md werden `app/`-Handler nicht unit-getestet.
`go test ./...`, `go vet`, `gofmt` und `wails build` sind unter Linux sauber;
Windows und macOS stehen noch aus.

### Überholt: es gibt keine Beitragsklassen (2026-09-11)

Dieses Ticket ist umgesetzt und wird **ersatzlos zurückgebaut**. Die
Bestandsaufnahme der echten Excel-Tabelle
([excel-vorlage.md](../excel-vorlage.md)) führt zwei unabhängige Spalten:
`Beitrag` mit 18 gepflegten Werten von 0 bis 144 € und `1x 2x Woche` mit der
Trainingsfrequenz. Betrag und Frequenz variieren getrennt — zwei Mitglieder mit
derselben Frequenz zahlen unterschiedlich viel. Die Annahme "die Staffelung
ergibt sich ausschließlich aus der Trainingsfrequenz" und damit das ganze
Klassen-Konzept ist widerlegt.

Zurückgebaut in **Ticket 13** (Beitrag individuell). Begründung, Alternativen
und Kostenseite in
[ADR-0005](../../../docs/adr/0005-beitrag-individuell-statt-beitragsklasse.md) —
insbesondere, warum der naheliegende Kompromiss (Tabelle behalten, 18 Zeilen
seeden) an Story 13 scheitert: eine Preisänderung an der Klasse "60 €" würde
stillschweigend den Beitrag aller Mitglieder darin ändern, auch derjenigen,
deren 60 € eine individuelle Zusage sind.

Der Filter nach Beitragsklasse wird in **Ticket 15** durch einen Filter nach
Trainingsfrequenz ersetzt. Der Rückbau ist **kein Bug** — er ist die Entscheidung.

**Rückbau erledigt (Ticket 13).** Ansicht, Template, Route, Navigationseintrag,
Testdatei, Tabelle und Seeding sind weg; in der Mitgliederliste steht jetzt der
individuell vereinbarte Beitrag.
