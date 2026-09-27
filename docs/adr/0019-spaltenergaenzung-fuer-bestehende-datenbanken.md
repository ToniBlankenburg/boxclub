# ADR-0019: Fehlende Spalten an bestehenden Tabellen automatisch ergänzen

**Status:** Accepted
**Datum:** 2026-09-27

## Kontext

`migrate()` legt das Schema seit jeher über `CREATE TABLE IF NOT EXISTS` an.
Das deckt eine neue Tabelle vollständig ab, ändert aber eine bereits
bestehende Tabelle nicht — auch dann nicht, wenn das aktuelle Schema
inzwischen weitere Spalten kennt. CLAUDE.md hielt das bislang bewusst so fest:
*"there is no migration mechanism, so a schema change means deleting the dev
database"*.

Das trifft die Dev-Datenbank ohne Schaden — sie ist Wegwerfdaten. Es trifft
aber genauso eine Installation, die seit Monaten durchläuft: `vereinsdaten`
bekam mit ADR-0012 (`logo`, `logo_mime`) und ADR-0016
(`moneymoney_verwendungszweck`) neue Spalten, ohne dass eine bestehende Zeile
sie je bekäme. `GetVereinsdaten()` selektiert sie namentlich und scheitert an
`no such column: logo` — der Verein-Reiter ließ sich auf einer länger
laufenden Installation nicht mehr öffnen, ohne dass irgendwo eine sichtbare
Fehlermeldung erschien (der Handler antwortet mit HTTP 500, und htmx tauscht
bei einer Nicht-2xx-Antwort nichts in `#inhalt`). Dieselbe Lücke betrifft
`mitglied` (u. a. `rueckstand`, `postleitzahl`, `ort`, seit ADR-0006 und der
Aufteilung der Anschrift) und `mitgliedschaft` (u. a. `anmeldedatum`,
`kuendigungsdatum`, `ruhend`, `anmeldegebuehr_eingezogen`).

## Entscheidung

`migrate()` ergänzt nach dem `CREATE TABLE IF NOT EXISTS` jede in
`spaltenNachtraege` gelistete Spalte, die an ihrer Tabelle noch fehlt, per
`ALTER TABLE ... ADD COLUMN` — geprüft über `PRAGMA table_info`, damit ein
wiederholter Start (jeder Programmstart) keine `duplicate column`-Fehler
wirft.

**Nur additive, eindeutig unstrittige Spalten.** Jede Spalte in
`spaltenNachtraege` ist so gewählt, dass ihr Schema-Default für eine ältere
Zeile exakt "diese Angabe kannte die Zeile noch nicht" bedeutet — keine
geratene, sondern eine fehlende Information (CONTEXT.md bestätigt das je
Angabe: eine Anschrift ohne Postleitzahl ist eine gültige, ein Rückstand ohne
Notiz startet als "in Ordnung", eine Anmeldegebühr ohne Wert ist "keine
erhoben").

**`mitgliedschaft.beitrag_monatlich_cents` steht bewusst nicht in der
Liste.** Vor dessen Einführung gab es die inzwischen entfernte
`beitragsklasse`-Tabelle mit einem eigenen Satz pro Klasse (siehe ADR-0005).
0 € ist dort ein gültiger, aber anderer Beitrag als "unbekannt" — anders als
bei der Anmeldegebühr, wo der Verein "keine erhoben" und "0 €" bewusst
gleichsetzt (CONTEXT.md → Anmeldegebühr). Eine Datenbank aus dieser sehr
frühen Zeit bräuchte eine echte Datenübernahme aus der (längst entfernten)
`beitragsklasse`-Tabelle statt eines Defaults; `migrate()` bricht für sie
weiterhin kontrolliert mit einer klaren Fehlermeldung ab, statt Beiträge
stillschweigend auf 0 € zu setzen.

CLAUDE.md wird an dieser einen Stelle präzisiert: kein Migrationsmechanismus
für Schemawechsel, die Daten umdeuten oder zusammenführen — aber additive
Spalten mit eindeutigem Default werden ab jetzt automatisch ergänzt.

## Betrachtete Alternativen

### Versionsnummer + nummerierte Migrationsschritte

Ein klassisches Migrationsframework (`schema_version`-Tabelle, Schritt 1, 2,
3, …). **Contra:** deutlich mehr Maschinerie, als der tatsächliche Bedarf
rechtfertigt — es gab in der gesamten Historie genau eine Art Bruch
(Beitragsklasse-Ablösung, mit echter Datenübernahme) und sonst ausschließlich
additive Spalten. Eine Versionstabelle wäre Vorbereitung auf Migrationsarten,
die bislang nicht vorkamen.

### Nur die drei `vereinsdaten`-Spalten fest verdrahten

Der ursprüngliche, engere Zuschnitt (nur der konkrete Bug-Auslöser). **Contra:**
dieselbe Lücke besteht identisch für `mitglied` und `mitgliedschaft` seit
noch länger laufenden Installationen — ein Fix, der nur den einen
gemeldeten Fall schließt, ließe die nächste Installation beim nächsten
Formular erneut ins Leere laufen.

### Spaltenliste automatisch aus dem `schema`-String parsen

Statt einer gepflegten Liste die erwarteten Spalten je Tabelle per
Text-Parsing aus der `schema`-Konstante ableiten. **Contra:** verschleiert
genau die Unterscheidung, auf die es ankommt — dass jede Spalte einzeln
darauf geprüft ist, ob ihr Default eine fehlende Information oder eine
falsche Aussage wäre (siehe `beitrag_monatlich_cents` oben). Eine geparste
Liste hätte diese Spalte ungeprüft mitgenommen.

## Konsequenzen

- **Gut:** Eine bestehende Installation holt fehlende additive Spalten beim
  nächsten Start automatisch nach, ohne Datenverlust und ohne manuellen
  Eingriff.
- **Gut:** Der Mechanismus ist auf Lesbarkeit statt Cleverness angelegt —
  `spaltenNachtraege` ist eine flache, kommentierte Liste, kein SQL-Parser.
- **Schlecht:** Jede künftige additive Spalte an `mitglied`, `mitgliedschaft`
  oder `vereinsdaten` braucht einen bewussten Eintrag in `spaltenNachtraege`
  — vergisst man ihn, bricht die Spalte auf einer alten Installation weiter
  wie bisher mit `no such column`, nur eben unbemerkt bis zum nächsten
  Bug-Report.
- **Reversibilität:** hoch — der Mechanismus ist rein additiv (fügt nur
  Spalten hinzu, ändert nichts Bestehendes) und ließe sich durch Löschen der
  Liste jederzeit wieder abschalten, ohne dass eine bereits ergänzte Spalte
  Schaden anrichtet.
