Status: ready-for-human

# 28: Teilnehmerliste je Trainingstermin

**What to build:** Der Stundenplan ("Trainingstermine") zeigt heute nur die Termine selbst. Der Admin soll dort zu jedem Termin sehen, **wer dafür angemeldet ist** — die Frage "wer trainiert samstags um 10:30?" aus ADR-0008. Die Terminzeile wird aufklappbar und zeigt die *Teilnehmerliste* (siehe `CONTEXT.md` → Teilnehmerliste). Es kommt nichts Neues zu speichern: die Liste wird aus den vorhandenen Anmeldungen abgelesen.

**Blocked by:** —
**Siehe:** [ADR-0008](../../../docs/adr/0008-trainingstermine-als-wochenplan.md), [ADR-0004](../../../docs/adr/0004-suche-und-filter-im-speicher.md), `CONTEXT.md` → Teilnehmerliste, Trainingstermin, Status, Ruhend

## Acceptance Criteria

- [x] `MemberService` liefert je Trainingstermin die Teilnehmer — als Ableitung aus `mitgliedschaft_trainingstermin`, ohne neue Tabelle und ohne neue Spalte
- [x] Aufgenommen sind Mitglieder mit dem Status **Neu**, **Aktiv** oder **In Kündigungsfrist**, deren Mitgliedschaft **nicht ruhend** ist. Der Status kommt aus derselben Ablesung wie in der Mitgliederliste (`status.go`), es gibt keine zweite Regel dafür
- [x] **Ausgetretene** stehen nicht darin; ihre Anmeldung bleibt unberührt und zählt weiter zur Trainingsfrequenz
- [x] Maßgeblich ist die **maßgebliche Mitgliedschaft** (wie in der Mitgliederliste). Wer ausgetreten und wieder eingetreten ist, erscheint nur mit den Terminen des neuen Zeitraums; die des alten zählen nicht
- [x] Ein Mitglied steht je Termin höchstens einmal in der Liste
- [x] Die Namen stehen **alphabetisch nach Nachname, dann Vorname** (dieselbe Ordnung wie `nachNamenSortieren`)
- [x] Je Termin steht die **Anzahl** der Teilnehmer; es gibt keine Obergrenze und keine Warnung bei vielen Teilnehmern
- [x] Ein Termin ohne Teilnehmer zeigt "niemand angemeldet", ist aber genauso aufklappbar wie die anderen
- [x] In der Liste steht **nur der Name** — kein Beitrag, kein Rückstand, kein Kennzeichen für neu/Kündigungsfrist. Der Name führt ins Mitglied (wie in der Mitgliederliste)
- [x] Der Schalter "Archivierte anzeigen" gilt auch hier: archivierte Termine samt ihren Teilnehmern stehen nur bei gesetztem Schalter da. Ein Mitglied, das nur für einen archivierten Termin angemeldet ist, ist in der Standardansicht dort also nicht zu sehen
- [x] Ein Klick auf den Termin klappt die Teilnehmer auf (Tastatur: Enter/Leertaste). Bearbeiten und Archivieren/Zurückholen bleiben erreichbar, aber als eigene Schaltflächen: die Zeile öffnet das Formular nicht mehr, damit es niemand versehentlich öffnet
- [x] Es gibt **keine** Möglichkeit, in der Liste etwas abzuhaken, zu notieren oder zu ändern, und **keinen** PDF- oder Druckweg (Bildschirm reicht für v1)
- [x] Service-Tests (Muster: `service/member_service_test.go`, echte SQLite, kein Mocking) für: jeden Status (Neu / Aktiv / In Kündigungsfrist / Ausgetreten), ruhend, Wiedereintritt mit anderen Terminen im neuen Zeitraum, archivierten Termin mit bestehender Anmeldung, Termin ohne Teilnehmer, Sortierung, Anzahl
- [ ] `go test ./...` grün; `wails dev` und `wails build` unter Windows und Linux — `go test ./...` grün, `go vet` sauber, für Windows und macOS cross-kompiliert; `wails dev`, `wails build` und ein Klicktest im WebView stehen aus

## Notes

**Das ist keine Anwesenheit.** Die Liste sagt, wer *angemeldet* ist, nie, wer an einem Tag *da war*. Es gibt keine datierten Einheiten und kein Feld, in dem so etwas stehen könnte — das ist die Grenze aus ADR-0008 und `CLAUDE.md` (Out of scope: attendance tracking). Wer beim Bauen versucht ist, "Abhaken" oder "heute da" einzuführen, hat das Ticket verlassen.

**Warum ruhend draußen und Neu drin:** ohne Kennzeichen pro Name muss die Liste allein durch ihre Auswahl stimmen. Ein ruhendes Mitglied trainiert gerade nicht, also fehlt es; es taucht wieder auf, sobald `SetRuhend` zurückgenommen wird, an der Anmeldung ändert sich nichts. Ein Neuer ist verbindlich angemeldet und steht deshalb schon darin, obwohl der Eintritt noch bevorsteht — bewusste Entscheidung, nicht Versehen.

**Ansatz:** Suche und Filter laufen im Speicher (ADR-0004), und die Mitgliederliste liest die Anmeldungen ohnehin (`Trainingstermine` an der Listenzeile, `terminanmeldungenLesen`). Naheliegend ist, dieselben Zeilen einmal nach Termin zu gruppieren, statt eine zweite SQL-Abfrage mit eigener Statuslogik zu schreiben — dann kann die Teilnehmerliste dem Filter "Trainingstermin" der Mitgliederliste (Commit edc5f92) nie widersprechen, abgesehen von den hier festgelegten Ausschlüssen (Ausgetretene, Ruhende).

**Handler/Template:** `app/trainingstermin.go` und `templates/trainingstermine.html` — wie bisher dünner Adapter, nicht unit-getestet, manueller Smoke-Test genügt. Aufklappen darf ohne JS auskommen (`<details>`, wie das Spaltenmenü in Ticket 27) oder per `hx-get` nachladen; beides ist recht.

## Comments

**Service:** `MemberService.Teilnehmerlisten(auchArchivierte)` (`service/teilnehmerliste.go`)
liefert je Termin des Stundenplans eine `Teilnehmerliste` (Termin, Teilnehmer, `Anzahl()`), in
Wochenreihenfolge. Die Teilnehmer sind die Zeilen von `List()`, einmal nach Termin gruppiert;
einzig das Ruhend-Kennzeichen wird zusätzlich ausgeschlossen. Ausgetretene fehlen, weil `List()`
sie nicht liefert — es gibt bewusst keine zweite Statusregel daneben (ADR-0004);
`TestTeilnehmerlisten_NachStatus` fängt eine Änderung daran. Tests:
`service/teilnehmerliste_test.go` (Sortierung inkl. Umlaut, Anzahl, leerer Termin, Ruhend hin
und zurück, alle vier Status, Wiedereintritt, doppelter Termin in altem und neuem Zeitraum,
Archiv-Schalter).

**Oberfläche:** `terminzeile` trägt die Teilnehmer (als `teilnehmerchip` mit Initialen, reine
Darstellung im Adapter), `templates/trainingstermine.html` zeigt jeden Termin als `<details>`,
dessen `<summary>` die ganze Zeile ist: ein Klick auf den Termin klappt die Teilnehmer auf, ohne
Skript. Die Teilnehmer stehen als Chips mit Initialen-Kürzel, jeder führt ins Mitglied. Bearbeiten
und Archivieren/Zurückholen sind zwei Schaltflächen, die rechts *über* der Zeile liegen statt in
der `<summary>` — eine Schaltfläche darin bliebe nicht zuverlässig bei sich, und so öffnet ein
Klick auf den Termin nie das Formular. Vorher öffnete ein Klick auf die Zeile das Formular; das
ist bewusst abgeschafft. Nach Archivieren oder Zurückholen wird `#inhalt` ersetzt; aufgeklappte
Listen klappen dann zu. Nur gerendert, im Quelltext und im gebauten CSS geprüft, nicht im WebView
angeklickt.
