Status: ready-for-agent

# 23: Vereinsdaten pflegen

**What to build:** Eine Einstellungsseite für den Verein **selbst**: Name, Anschrift, Bankverbindung (IBAN, BIC, Kreditinstitut), Kontakt und eine freie **Fußzeile**. Es sind die ersten Daten der App, die kein Mitglied betreffen — gebraucht werden sie als Briefkopf der Rechnung (Ticket 25).

**Blocked by:** —
**Siehe:** `CONTEXT.md` → Vereinsdaten, [ADR-0009](../../../docs/adr/0009-rechnungen-und-monatssoll-ohne-zahlungsmodell.md)

## Acceptance Criteria

- [ ] Tabelle `vereinsdaten` mit **genau einer** Zeile; die App legt sie beim Start leer an, falls sie fehlt
- [ ] Eigener Menüpunkt „Verein", Formular zum Bearbeiten und Speichern
- [ ] Alle Felder sind **freiwillig** — eine leere Angabe fehlt dann eben auf der Rechnung, sie blockiert nichts
- [ ] Die **Fußzeile** ist mehrzeiliger Freitext. Dort steht, was der Verein steuerlich schreiben muss (etwa der Hinweis auf § 19 UStG) — die App kennt dazu keine Regel und prüft nichts
- [ ] Die IBAN wird wie am Mitglied als **reiner Text** gespeichert: keine Validierung, kein Format (ADR-0006)
- [ ] Service-Tests für Lesen und Schreiben, inklusive „Zeile fehlt und wird angelegt"
- [ ] `go test ./...` grün; `wails dev` und `wails build` unter Windows und Linux

## Notes

**Warum nicht im Code verdrahtet:** die App wäre dann für genau einen Verein gebaut, und ein Tippfehler im Namen wäre ein Entwicklungsauftrag. Es sind fünf Felder und eine Seite.

**Kein Logo in v1.** Die Fußzeile ist der Ort, an dem später auch ein Bild landen kann; solange niemand danach fragt, bleibt es bei Text.
