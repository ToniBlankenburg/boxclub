Status: ready-for-human

# 23: Vereinsdaten pflegen

**What to build:** Eine Einstellungsseite für den Verein **selbst**: Name, Anschrift, Bankverbindung (IBAN, BIC, Kreditinstitut), Kontakt und eine freie **Fußzeile**. Es sind die ersten Daten der App, die kein Mitglied betreffen — gebraucht werden sie als Briefkopf der Rechnung (Ticket 25).

**Blocked by:** —
**Siehe:** `CONTEXT.md` → Vereinsdaten, [ADR-0009](../../../docs/adr/0009-rechnungen-und-monatssoll-ohne-zahlungsmodell.md)

## Acceptance Criteria

- [x] Tabelle `vereinsdaten` mit **genau einer** Zeile; die App legt sie beim Start leer an, falls sie fehlt
- [x] Eigener Menüpunkt „Verein", Formular zum Bearbeiten und Speichern
- [x] Alle Felder sind **freiwillig** — eine leere Angabe fehlt dann eben auf der Rechnung, sie blockiert nichts
- [x] Die **Fußzeile** ist mehrzeiliger Freitext. Dort steht, was der Verein steuerlich schreiben muss (etwa der Hinweis auf § 19 UStG) — die App kennt dazu keine Regel und prüft nichts
- [x] Die IBAN wird wie am Mitglied als **reiner Text** gespeichert: keine Validierung, kein Format (ADR-0006)
- [x] Service-Tests für Lesen und Schreiben, inklusive „Zeile fehlt und wird angelegt"
- [x] `go test ./...` grün; `wails build` unter Linux geprüft — Windows und `wails dev` stehen aus (kein Windows-Rechner, kein Display in dieser Umgebung), `GOOS=windows go build ./...` als Ersatz grün

## Notes

**Warum nicht im Code verdrahtet:** die App wäre dann für genau einen Verein gebaut, und ein Tippfehler im Namen wäre ein Entwicklungsauftrag. Es sind fünf Felder und eine Seite.

**Kein Logo in v1.** Die Fußzeile ist der Ort, an dem später auch ein Bild landen kann; solange niemand danach fragt, bleibt es bei Text.

## Comments

**2026-09-13 — umgesetzt**

Alle Akzeptanzkriterien erfüllt. `go test ./...` grün, `wails build` unter Linux
erfolgreich, `GOOS=windows go build ./...` ebenfalls; ein Windows-`wails build`
steht aus, die Änderung ist reines Go/HTML.

`service.Vereinsdaten` mit `GetVereinsdaten`/`SetVereinsdaten` liegt am
`MemberService` und nicht in einem eigenen Service: es ist der eine Service
dieser App, und ein zweiter für eine Zeile wäre Zeremonie (spec.md → Seams, wo
schon ein eigener `MitgliedschaftService` verworfen wurde).

Drei Entscheidungen, die über den Ticket-Text hinausgehen:

- **Die Anschrift ist derselbe Typ wie am Mitglied** (`service.Anschrift`, drei
  getrennte Angaben). Eine Briefanschrift ist gleich gebaut, ob sie einer Person
  oder einem Verein gehört — und `OrtZeile()`/`Einzeilig()` sind auf der
  Rechnung genau das, was gebraucht wird.
- **Umschließender Leerraum fällt bei jeder Angabe weg.** Diese Felder werden
  gedruckt; ein Leerzeichen hinter dem Vereinsnamen stünde im Briefkopf des PDF.
  Das ist keine Prüfung und kein Format — die IBAN bleibt Zeichen für Zeichen,
  wie sie getippt wurde (ADR-0006), nachgewiesen durch einen eigenen Test.
- **`bic` und `kreditinstitut` stehen als eigene Spalten** neben der IBAN, weil
  das Ticket die Bankverbindung so aufzählt und sie auf der Rechnung
  nebeneinanderstehen.

**Nach dem Review** (`/code-review`, Achsen Standards und Spec) noch geändert:

- **Das AC „Zeile fehlt und wird angelegt" war nur behauptet.** Der Spec-Review
  hat recht: solange `GetVereinsdaten` eine fehlende Zeile zu leeren Angaben
  glättete, konnte kein Test „angelegt" von „fehlt" unterscheiden. Der Zweig ist
  jetzt weg — die Zeile ist eine Zusicherung von `Open`, und dass das Lesen auf
  einer frischen Datenbank gelingt, ist der Nachweis. Dazu kommt
  `TestVereinsdaten_UeberlebenDenNeustart`: zwei Läufe auf derselben Datei, die
  Angaben bleiben stehen — das ist die andere Hälfte, die `INSERT OR IGNORE`
  zusichert.
- **Die Zeilen-Anlage steht nicht mehr in `const schema`**, sondern als eigene
  Anweisung mit eigener Fehlermeldung in `migrate()`. `schema` sind Tabellen;
  eine Daten-Anweisung darin hätte weder zum Namen noch zu „nothing is seeded"
  (CLAUDE.md) gepasst.
- **Der Meldungs-Kasten ist ein eigenes Fragment** (`templates/meldung.html`).
  Er stand wortgleich in Mitgliederliste und Stundenplan; die Vereinsansicht
  wäre die dritte Kopie geworden.
