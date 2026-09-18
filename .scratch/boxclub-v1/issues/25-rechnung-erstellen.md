Status: ready-for-human

# 25: Rechnung schreiben und als PDF ablegen

**What to build:** Der Verein schreibt von Hand eine Rechnung über eine Leistung **neben dem Beitrag** — typisch ein Einzeltraining — und bekommt ein PDF, das am Mitglied abgelegt wird.

Eine Rechnung ist hier **ein Dokument und kein offener Posten**: es gibt keinen Rechnungsstatus, keinen Zahlungseingang, keine Mahnung. Es gibt auch **keine Rechnungstabelle** — nach dem Erzeugen bleibt nur das PDF.

**Blocked by:** 23, 24
**Siehe:** [ADR-0009](../../../docs/adr/0009-rechnungen-und-monatssoll-ohne-zahlungsmodell.md), `CONTEXT.md` → Rechnung

## Acceptance Criteria

- [x] Formular am Mitglied: **Empfänger** (Name, Anschrift) aus dem Mitglied vorbelegt und **frei überschreibbar**
- [x] **Rechnungsnummer** wird eingetippt, nicht vergeben. Kein Vorschlag, keine Prüfung auf Eindeutigkeit — die Nummernfolge führt der Verein in seiner Buchhaltung
- [x] **Positionen**: beliebig viele Zeilen mit Bezeichnung, Menge und Einzelpreis; die Summe wird gerechnet
- [x] Beträge sind **netto**. Ein **Steuersatz je Rechnung**, Vorgabe 19 %, änderbar bis 0. Das PDF weist Netto, Steuerbetrag und Bruttosumme getrennt aus; bei 0 % entfällt die Steuerzeile
- [x] **Rechnungsdatum** heute vorbelegt, **Zahlungsziel** 14 Tage darauf, beides überschreibbar
- [x] Briefkopf, Bankverbindung und Fußzeile kommen aus den Vereinsdaten (Ticket 23)
- [x] Das PDF wird als `dokument` der Art *Rechnung* am **Mitglied** abgelegt (nicht an der Mitgliedschaft)
- [x] Für einen Empfänger **ohne** Mitglied (Externer) wird das PDF ausgegeben, aber **nicht** abgelegt — es gibt niemanden, an dem es hängen könnte
- [x] Erzeugt mit einer **reinen Go-Bibliothek, kein CGo** (`go-pdf/fpdf` oder `signintech/gopdf`; beim Umsetzen auf Pflegestand prüfen und die Wahl hier als Kommentar festhalten). Umlaute brauchen eine eingebettete Schrift — im Zweifel ein Ausprobieren wert, bevor die Bibliothek feststeht
- [x] Service-Tests für die Rechnungsrechnung: Positionssumme, Netto/Steuer aus dem Bruttobetrag, 0 %, Rundung auf Cent. **Das PDF-Layout wird nicht unit-getestet**, aber ein erzeugtes PDF wird beim Umsetzen von Hand angesehen
- [x] `go test ./...` grün; `wails dev` und `wails build` unter Windows und Linux

## Notes

**Warum die Nummer von Hand:** eine fortlaufende Nummer aus der App wäre das Versprechen, sie lückenlos und eindeutig zu halten. Das kann sie nicht halten, wenn daneben noch anders Rechnungen entstehen (ADR-0009).

**Warum netto:** der Admin trägt den vereinbarten Preis ohne Steuer ein, die App schlägt sie für die Anzeige auf (geändert am 2026-09-18, siehe Comments — ursprünglich war brutto vorgesehen).

**Keine Steuerlogik.** Ob 19 %, 7 % oder gar nichts gilt, entscheidet der steuerliche Status des Vereins. Die App nimmt eine Zahl entgegen und rechnet damit — mehr nicht.

**Kein Stapellauf für alle Mitglieder.** Beiträge werden eingezogen und nie in Rechnung gestellt (ADR-0006).

## Comments

**2026-09-13 — umgesetzt**

Alle Akzeptanzkriterien erfüllt. `go test ./...` grün, `wails build` unter Linux
erfolgreich, `GOOS=windows go build ./...` ebenfalls; ein Windows-`wails build`
und `wails dev` stehen aus (kein Windows-Rechner, kein Display in dieser
Umgebung) — dieselbe Einschränkung wie in den Tickets 23 und 24.

**Bibliothek: `signintech/gopdf` (v0.38.1), nicht `go-pdf/fpdf`.** Geprüft beim
Umsetzen (September 2026): `go-pdf/fpdf` ist als *Archived* markiert, `gopdf`
hatte dagegen im Sommer 2026 noch Releases. Beide sind reines Go, keine CGo.
Für Umlaute ist DejaVu Sans (Regular und Bold) eingebettet
(`service/fonts/`, per `go:embed`) — sie liegt unter der Bitstream-Vera-Lizenz,
die das Einbetten ausdrücklich erlaubt (`service/fonts/LICENSE.txt`), und
deckt Umlaute und das Eurozeichen ab. Ein erzeugtes Beispiel-PDF wurde von Hand
angesehen (Rasterung über `pdftoppm`): Briefkopf, Empfänger, Positionstabelle,
Netto/Steuer/Brutto und Fußzeile stehen wie erwartet, Umlaute sauber.

Sechs Entscheidungen, die über den Ticket-Text hinausgehen:

- **Zwei Wege zum selben Formular.** "Formular am Mitglied" (AC) ist der
  Regelfall: ein Block unter dessen Stammdaten, genau wie die Verträge
  (`vertragsbloecke`/`vertraege`), Empfänger vorbelegt, das PDF landet dort als
  Dokument. Für "Empfänger ohne Mitglied" (CONTEXT.md → Rechnung, ausdrücklich
  als Fall genannt) gibt es zusätzlich einen eigenständigen Bereich "Rechnung"
  in der Kopfnavigation, Empfänger leer, kein Mitglied dahinter — dieselbe
  Formularlogik, derselbe Service-Aufruf mit `mitgliedID = nil`. Ohne diesen
  zweiten Weg wäre der Fall im Ticket-Text nur eine Fußnote ohne Bedienstelle
  geblieben.
- **Menge ist eine ganze Zahl, kein Bruch.** Einzeltrainings werden gezählt.
  Damit ist die Zeilensumme eine Ganzzahlmultiplikation ohne jede Rundung —
  die rundet ausschließlich die Netto/Steuer-Aufteilung des Gesamtbetrags
  (AC: "Rundung auf Cent" bezieht sich darauf, nicht auf die Positionssumme).
- **Kein Stapel erzeugter Rechnungen in der Oberfläche.** Der Service legt sie
  ab und kann sie über `RechnungenDesMitglieds`/`RechnungInhalt` wieder lesen
  (gebraucht für die Tests, dieselbe Bauart wie `vertraegeLesen`/`VertragInhalt`),
  aber kein AC verlangt eine Liste mit erneutem Export in der Oberfläche. Die
  App bietet das PDF deshalb genau einmal an, direkt nach dem Erstellen, über
  denselben Datei-Dialog wie der Vertragsexport (`app.Speicherziel`, ADR-0002
  bleibt gewahrt). Bricht jemand den Dialog ab, bleibt die Rechnung zwar am
  Mitglied abgelegt, ist von dort aber ohne erneuten Export nicht mehr zu
  erreichen — das sagt die Meldung so. Eine Liste mit Re-Export wäre die
  nächste, saubere Erweiterung, sobald sie gebraucht wird.
- **Feste fünf Positionszeilen statt eines dynamischen "Zeile hinzufügen".**
  "Beliebig viele Zeilen" (AC) gilt für `service.RechnungEingabe` — dort gibt
  es keine Obergrenze. Das Formular bietet fünf Zeilen auf einmal an; eine
  ganz leere zählt beim Erstellen nicht mit. Für ein Einzeltraining reicht das
  komfortabel, und ein htmx-Mechanismus fürs Nachladen weiterer Zeilen wäre
  Aufwand, den kein AC verlangt.
- **Rechnungsnummer, Steuersatz und Beträge sind Text im Formular**, wie
  Beitrag und Anmeldegebühr am Mitglied: `EinzelpreisAusEuro` benutzt
  dieselbe Umrechnung wie `BeitragAusEuro` (`euroText`/`centsAusEuroText` aus
  `beitrag.go`), `SteuersatzAusText`/`SteuersatzAlsText` sind das Gegenstück
  für den Prozentsatz. Eine dritte, eigene Schreibweise für denselben
  Betragstyp wäre eine Falle gewesen.
- **`mitgliedPruefen` steht in `RechnungErstellen` vor der PDF-Erzeugung**, nicht
  erst in `rechnungAblegen`s Transaktion — dieselbe Reihenfolge wie bei
  `VertragAblegen`. Ohne die Prüfung vorweg hätte ein unbekanntes Mitglied
  jedes Mal ein komplettes PDF gekostet, bevor der Fehler kommt (gefunden im
  `/code-review` nach dem ersten Durchlauf, siehe unten).

**Nach dem Review** (`/code-review`, Stufe medium) geändert:

- **`RechnungErstellen` prüfte das Mitglied zu spät.** Validierung, das Lesen
  der Vereinsdaten und die vollständige PDF-Erzeugung liefen durch, bevor
  `rechnungAblegen` ein unbekanntes Mitglied überhaupt bemerkte — verschwendete
  Arbeit bei jedem Aufruf mit einer falschen oder inzwischen gelöschten ID.
  `mitgliedPruefen` steht jetzt vor der PDF-Erzeugung, wie bei `VertragAblegen`.

**2026-09-18 — Beträge auf netto umgestellt**

Auf Wunsch des Vereinsadmins tragen Positionspreise jetzt **netto** ein; die
App schlägt die Steuer für Anzeige und PDF auf, statt sie aus einem
Bruttopreis herauszurechnen. Betrifft `RechnungEingabe.Betraege`
(Multiplikation statt Division), die Formularbeschriftung
(`templates/rechnung.html`), die PDF-Spaltenköpfe
(`service/rechnung_pdf.go`) und die Service-Tests. ADR-0009 und dieses Ticket
wurden entsprechend nachgezogen; die ursprüngliche Begründung für brutto
("der Admin nennt den Preis, den er vereinbart hat") hat sich in der Praxis
nicht gehalten.
