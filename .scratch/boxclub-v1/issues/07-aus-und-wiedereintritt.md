Status: ready-for-human

# 07: Aus- und Wiedereintritt

**What to build:** Ein Mitglied kann mit einem Datum als ausgetreten markiert werden — es verschwindet aus der Standardansicht, der Datensatz bleibt aber erhalten. Ein ehemaliges Mitglied kann wieder eintreten; dabei entsteht eine **neue** `Mitgliedschaft`-Zeile, nicht ein neues `Mitglied`.

**Blocked by:** 03 (Mitgliederliste)

## Acceptance Criteria

- [x] Aktion "Austritt eintragen" fragt nach einem Austrittsdatum und setzt `austritt` auf der aktiven Mitgliedschaft
- [x] Nach dem Austritt verschwindet das Mitglied aus der Standardansicht "nur aktive", bleibt aber sichtbar unter dem Filter "auch ehemalige"
- [x] Aktion "Wiedereintritt" ist auf ehemaligen Mitgliedern verfügbar (nicht auf aktiven); sie fragt nach einem Eintrittsdatum und legt eine **neue** `mitgliedschaft`-Zeile an (`austritt` NULL)
- [x] Bei Wiedereintritt bleibt der `mitglied`-Datensatz derselbe — Stammdaten, ID, Historie
- [x] `MemberService.MarkExit(id, datum)` und `MemberService.Rejoin(id, datum)` am Seam
- [x] Test am Seam: Aus- gefolgt von Wiedereintritt → `mitglied` unverändert, zwei `mitgliedschaft`-Zeilen in der DB
- [x] Test am Seam: doppelter `MarkExit` auf dieselbe aktive Mitgliedschaft ist ein Fehler
- [x] Test am Seam: `Rejoin` auf ein bereits aktives Mitglied ist ein Fehler
- [x] Test am Seam: `MarkExit` mit Datum vor Eintrittsdatum ist ein Fehler

## Comments

**Umgesetzt.** Seam: `MemberService.MarkExit(id, datum)` und
`MemberService.Rejoin(id, datum)` in
[service/member_service.go](../../../service/member_service.go), Tests in
[service/lebenszyklus_test.go](../../../service/lebenszyklus_test.go) (10
Testfunktionen). Beide Methoden laufen in einer Transaktion: Zustand prüfen und
schreiben gehört zusammen, sonst könnten zwei laufende Zeiträume entstehen.

Fehlerarten am Seam: `ErrNichtAktiv` (Austritt ohne laufende Mitgliedschaft) und
`ErrBereitsAktiv` (Wiedereintritt eines aktiven Mitglieds) neben dem bestehenden
`ErrNichtGefunden`; Datumsregeln als `ValidierungsFehler`, weil die Oberfläche
dafür einen zeigbaren Satz braucht. Datumsvergleiche laufen über die ISO-Textform
— derselbe Grund wie bei `zahlungsstatusAm`: Kalendertage aus der Datenbank (UTC)
und aus der lokalen Uhr sind sonst je nach Zonenversatz einen Tag auseinander.

`service/export_test.go` ist weg. `AustrittFuerTest` war laut eigenem Kommentar
genau bis zu diesem Ticket geliehen; die vier Tests, die es benutzt haben, bauen
ihre Fixture jetzt über `MarkExit` auf.

**Oberfläche:** In der Eintritts-Spalte steht eine Schaltfläche, die je nach
Zustand "Austritt" oder "Wiedereintritt" heißt — nie beides, das erledigt
`{{if .Austritt}}`. Sie tauscht die Zeile gegen eine Datumseingabe (Muster der
`bezahlt_bis`-Zeile). Nach Erfolg kommt aber nicht die Zeile zurück, sondern die
ganze Liste per `HX-Retarget`: nach einem Austritt gibt es die Zeile in der
Standardansicht nicht mehr, und die Meldung sagt, wo das Mitglied jetzt steht.

**Drei Dinge über das Ticket hinaus** — bitte gegenlesen, ob gewollt:

1. **`Rejoin` lehnt ein Eintrittsdatum vor dem letzten Austritt ab.** Das Ticket
   fordert die Regel nur für `MarkExit`. Ohne die symmetrische Regel könnten sich
   zwei Zeiträume überlappen und "aktiv seit wann" hätte zwei Antworten.
2. **Leeres Datum ist bei beiden ein `ValidierungsFehler`.** Das `required` im
   Formular ist keine Regel, sondern eine Bequemlichkeit; die Pflichtangaben
   liegen laut `NeuesMitglied.validieren` im Service als einziger Quelle.
3. **`Mitglied.LetzteMitgliedschaft()` ist neu**, und das Bearbeitungsformular
   zeigt damit auch bei einem Ehemaligen dessen letzten Eintritt statt "—".
   Vorher war das nicht erreichbar, weil es ohne dieses Ticket keine ehemaligen
   Mitglieder aus der Oberfläche heraus gab.

**Bewusst nicht getan:** `Rejoin` bekommt keinen Sonderfall für ein Mitglied ganz
ohne Mitgliedschaft. Über `Create` kann das nicht entstehen (Mitglied und erster
Zeitraum werden in einer Transaktion angelegt), und eine Sperre für einen
unerreichbaren Zustand wäre Ballast.

**Rauchtest:** Der ganze Ablauf (Liste → Austritt → Ehemaligen-Filter →
Wiedereintritt, samt doppeltem Austritt, doppeltem Wiedereintritt, unbekannter ID
und ungültigem Datum) lief gegen `app.Handler()` per `httptest` durch. Der
Harnisch ist danach gelöscht — laut [CLAUDE.md](../../../CLAUDE.md) werden
`app/`-Handler nicht unit-getestet. Falls das doch bleiben soll, ist es billig
wiederherzustellen. `go test ./...`, `go vet`, `gofmt` und `wails build` sind
unter Linux sauber; Windows und macOS stehen noch aus.

**Nachtrag (Ticket 17):** `MarkExit(id, datum)` gibt es nicht mehr. Ersetzt durch
`SetKuendigung(id, service.Kuendigung{Datum, Austritt})` — dasselbe Verhalten,
zusätzlich mit dem Kündigungsdatum. Die Regeln oben gelten unverändert.
