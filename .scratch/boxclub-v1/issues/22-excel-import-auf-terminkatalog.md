Status: ready-for-human

# 22: Excel-Import auf den Terminkatalog umstellen

**What to build:** Der Import ordnet die Freitexte aus `Training - 1/2/3` **vorhandenen** Trainingsterminen zu und legt selbst keine an. Was er nicht zuordnen kann, geht in den Fehlerbericht aus Ticket 09 — dieselbe Stelle, an der heute schon unlesbare Beiträge und widersprüchliche Frequenzen landen.

**Blocked by:** 21
**Siehe:** [ADR-0008](../../../docs/adr/0008-trainingstermine-als-wochenplan.md)

## Acceptance Criteria

- [x] Der Importer ordnet per Textvergleich zu: Groß-/Kleinschreibung und Leerraum sind egal, alles andere muss stimmen
- [x] Ein **nicht** zuordenbarer Wert erzeugt einen Eintrag im Fehlerbericht mit Zeile, Spalte und dem gelesenen Text — und **legt keinen Termin an**
- [x] Ein nicht zuordenbarer Wert verhindert **nicht** die Übernahme des Mitglieds; es kommt ohne diesen Termin herein, wie heute bei anderen Feldfehlern
- [x] Der Wert `kein` bleibt, was er ist: kein Termin, keine Meldung (`importer.keinTraining`)
- [x] Die bestehende Prüfung „angegebene Frequenz widerspricht der Anzahl der Termine" bleibt und zählt die **zugeordneten** Termine
- [x] **Archivierte** Termine werden nicht zugeordnet — ein Import darf keine Zeiten vergeben, die es nicht mehr gibt. Der Wert geht in den Fehlerbericht
- [x] Der Fehlerbericht sagt dem Admin, was zu tun ist: erst den Stundenplan pflegen, dann erneut importieren
- [x] Tests in `importer/` und `service/import_test.go` auf den Katalog umgeschrieben
- [x] `go test ./...` grün; `wails build` unter Linux geprüft — Windows und `wails dev` stehen aus (kein Windows-Rechner, kein Display in dieser Umgebung), `GOOS=windows go build ./...` als Ersatz grün

## Notes

**Erwartungshaltung beim ersten Import:** die Excel enthält Schreibvarianten, die nicht treffen werden („Sa 10:30" gegen „Samstag 10:30"). Das ist eingepreist — eine Aufräumrunde nach dem ersten Lauf gehört zum Verfahren, nicht zu den Fehlern.

**Warum nicht automatisch anlegen:** ADR-0008 → Alternativen. Jede Schreibvariante würde ein Stundenplaneintrag, und aufgeräumt wird das nie.

**Die Import-Ansicht braucht einen Hinweis**, wenn der Stundenplan leer ist: sonst importiert jemand 200 Mitglieder und bekommt 400 Fehlermeldungen, ohne zu verstehen, warum.

## Comments

**2026-09-13 — umgesetzt**

Alle Akzeptanzkriterien erfüllt. `go test ./...` grün, `wails build` unter Linux
erfolgreich, `GOOS=windows go build ./...` ebenfalls — ein Windows-`wails build`
steht noch aus, die Änderung ist reines Go/HTML ohne plattformabhängige Teile.

Der Katalog kommt als `importer.Stundenplan` von außen herein
(`ExcelImporter.Lesen(datei, plan)`); der Importer holt ihn nicht selbst, die
Naht im `app/`-Orchestrator bleibt die einzige Stelle, an der beide Hälften
zusammenkommen.

Vier Entscheidungen, die über den Ticket-Text hinausgehen — die ersten beiden
vorab mit dir abgestimmt:

- **Ein Termin bietet drei Schreibweisen an:** „Samstag 10:30", „Samstag 10:30
  Uhr" und die volle Anzeige „Samstag 10:30 – 12:00 · Anfänger". Verglichen wird
  jede exakt (Groß-/Kleinschreibung und Leerraum egal). Ohne die Form mit „Uhr"
  träfe beim ersten echten Lauf praktisch keine Zeile, weil die Excel überall
  `<Wochentag> <Uhrzeit> Uhr` führt; ohne die kurze Form träfe kein Termin, für
  den eine Endzeit gepflegt ist.
- **Ein zweiter Lauf ersetzt die Zuordnung, sobald die Zeile Termine
  mitbringt**, und lässt sie sonst stehen. Der Weg aus dem Bericht heißt
  „Stundenplan pflegen, erneut importieren" und führte sonst ins Leere; „ersetze
  durch nichts" nähme umgekehrt genau die von Hand nachgetragenen Termine weg.
- **Der Bericht hat einen zweiten Abschnitt.** Damit ein nicht zuordenbarer Wert
  die Zeile nicht aufhält (AC 3), unterscheidet `importer.Ergebnis` jetzt
  `Fehler` (Zeile nicht übernommen) von `Hinweise` (Zeile übernommen, Nacharbeit
  offen). `Zeilenfehler` heißt deshalb `Zeilenmeldung`. Scheitert eine Zeile
  später doch am Service, wandern ihre Hinweise zu den Gründen — eine Zeile
  steht nie in beiden Listen. `Erfolgreich()` ist jetzt erst ohne Hinweise wahr.
- **Der Frequenz-*Widerspruch* ist ein Hinweis geworden.** Eine Zeile, deren
  Termin der Stundenplan nicht kennt, widerspräche der Spalte sonst zwangsläufig
  ein zweites Mal und wäre entgegen AC 3 nicht mehr übernehmbar. Die Meldung
  „keine lesbare Frequenz" bleibt dagegen ein Fehlerfall wie bisher — sie ist
  kein Widerspruch zwischen zwei Angaben, sondern eine undeutbare Zelle.

Ein **mehrdeutiger** Freitext (zwei Gruppen zur selben Zeit, die kurze Form
trifft beide) wird nicht geraten, sondern gemeldet — mit der Bitte, in der Excel
die vollständige Schreibweise einzutragen.

Die Import-Ansicht zeigt den Hinweis auf den leeren Stundenplan samt Knopf
dorthin — im Datei-Dialog **und** über dem Bericht.

**Nach dem Review** (`/code-review`, Achsen Standards und Spec) noch geändert:

- `service.Trainingstermin.KurzeAnzeige()` ist neu; der Importer setzt „Samstag
  10:30" nicht mehr selbst zusammen. Zwei Stellen, die denselben Text schreiben,
  wären beim ersten geänderten Trennzeichen auseinandergelaufen und hätten den
  Abgleich still gebrochen. `Anzeige()` baut jetzt darauf auf.
- `Zeilenmeldung.Gruende` heißt `Meldungen`: ein Hinweis ist kein Grund, und der
  Bericht beschriftet die Spalte auch so. `zeilenfehlerSortieren` heißt
  `nachZeileSortieren`.
- Der Warnkasten „Stundenplan ist leer" steht als `{{define "stundenplan-leer"}}`
  einmal da statt zweimal wortgleich — wie `navigation` und die übrigen
  Fragmente.
- **Zurückgenommen:** „keine lesbare Frequenz" hatte ich zum Hinweis
  herabgestuft. Der Spec-Review hat recht, dass AC 5 nur den Widerspruch meint;
  die Meldung ist wieder ein Grund und durch
  `TestLesen_UnlesbareFrequenzMachtDieZeileZumFehlerfall` festgenagelt.

Stehen gelassen: die Entdopplung der Termin-IDs im Importer (er braucht die Zahl
vor dem Service, der sie ohnehin normalisiert) und die Signatur
`zuordnen(spalte, text) (int64, string)` mit der leeren Meldung als Erfolgsfall —
ein eigener Ergebnistyp wäre hier Zeremonie; der Vertrag steht jetzt im
Doc-Kommentar.
