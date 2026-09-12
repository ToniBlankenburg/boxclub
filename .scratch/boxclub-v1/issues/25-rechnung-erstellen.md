Status: ready-for-agent

# 25: Rechnung schreiben und als PDF ablegen

**What to build:** Der Verein schreibt von Hand eine Rechnung über eine Leistung **neben dem Beitrag** — typisch ein Einzeltraining — und bekommt ein PDF, das am Mitglied abgelegt wird.

Eine Rechnung ist hier **ein Dokument und kein offener Posten**: es gibt keinen Rechnungsstatus, keinen Zahlungseingang, keine Mahnung. Es gibt auch **keine Rechnungstabelle** — nach dem Erzeugen bleibt nur das PDF.

**Blocked by:** 23, 24
**Siehe:** [ADR-0009](../../../docs/adr/0009-rechnungen-und-monatssoll-ohne-zahlungsmodell.md), `CONTEXT.md` → Rechnung

## Acceptance Criteria

- [ ] Formular am Mitglied: **Empfänger** (Name, Anschrift) aus dem Mitglied vorbelegt und **frei überschreibbar**
- [ ] **Rechnungsnummer** wird eingetippt, nicht vergeben. Kein Vorschlag, keine Prüfung auf Eindeutigkeit — die Nummernfolge führt der Verein in seiner Buchhaltung
- [ ] **Positionen**: beliebig viele Zeilen mit Bezeichnung, Menge und Einzelpreis; die Summe wird gerechnet
- [ ] Beträge sind **brutto**. Ein **Steuersatz je Rechnung**, Vorgabe 19 %, änderbar bis 0. Das PDF weist Netto, Steuerbetrag und Bruttosumme getrennt aus; bei 0 % entfällt die Steuerzeile
- [ ] **Rechnungsdatum** heute vorbelegt, **Zahlungsziel** 14 Tage darauf, beides überschreibbar
- [ ] Briefkopf, Bankverbindung und Fußzeile kommen aus den Vereinsdaten (Ticket 23)
- [ ] Das PDF wird als `dokument` der Art *Rechnung* am **Mitglied** abgelegt (nicht an der Mitgliedschaft)
- [ ] Für einen Empfänger **ohne** Mitglied (Externer) wird das PDF ausgegeben, aber **nicht** abgelegt — es gibt niemanden, an dem es hängen könnte
- [ ] Erzeugt mit einer **reinen Go-Bibliothek, kein CGo** (`go-pdf/fpdf` oder `signintech/gopdf`; beim Umsetzen auf Pflegestand prüfen und die Wahl hier als Kommentar festhalten). Umlaute brauchen eine eingebettete Schrift — im Zweifel ein Ausprobieren wert, bevor die Bibliothek feststeht
- [ ] Service-Tests für die Rechnungsrechnung: Positionssumme, Netto/Steuer aus dem Bruttobetrag, 0 %, Rundung auf Cent. **Das PDF-Layout wird nicht unit-getestet**, aber ein erzeugtes PDF wird beim Umsetzen von Hand angesehen
- [ ] `go test ./...` grün; `wails dev` und `wails build` unter Windows und Linux

## Notes

**Warum die Nummer von Hand:** eine fortlaufende Nummer aus der App wäre das Versprechen, sie lückenlos und eindeutig zu halten. Das kann sie nicht halten, wenn daneben noch anders Rechnungen entstehen (ADR-0009).

**Warum brutto:** der Admin nennt den Preis, den er vereinbart hat — „das Einzeltraining kostet 60 €". Netto einzugeben hieße, ihn jedes Mal rückwärts rechnen zu lassen.

**Keine Steuerlogik.** Ob 19 %, 7 % oder gar nichts gilt, entscheidet der steuerliche Status des Vereins. Die App nimmt eine Zahl entgegen und rechnet damit — mehr nicht.

**Kein Stapellauf für alle Mitglieder.** Beiträge werden eingezogen und nie in Rechnung gestellt (ADR-0006).
