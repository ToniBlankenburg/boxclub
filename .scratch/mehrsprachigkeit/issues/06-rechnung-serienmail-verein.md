Status: ready-for-human

# 06: Rechnung, Serienmail, Verein-Formular übersetzt

**What to build:** `templates/rechnung.html`, `templates/serienmail.html`,
`templates/verein.html` und die zugehörigen Handler
(`app/rechnung.go`, `app/verein.go`, `service/rechnung_pdf.go`). Die
erzeugte Rechnungs-PDF selbst (Briefkopf, Positionen, Fußzeile) bekommt
damit erstmals eine Sprache — bisher ist sie immer deutsch.

**Blocked by:** 01 (i18n-Infrastruktur)

## Acceptance Criteria

- [x] Formulartexte in Rechnung, Serienmail und Verein übersetzt
- [x] Die erzeugte Rechnungs-PDF trägt die zum Erstellzeitpunkt aktive
      Sprache (Betreff, Positionslabel, Steuerhinweis) — geklärt werden
      muss, ob das gewünscht ist oder die PDF unabhängig von der
      UI-Sprache deutsch bleiben soll (Rücksprache mit Nutzer nötig)
- [x] `go test ./...` grün, `wails build` unter Linux grün

## Comments

### Umsetzung

- Rücksprache mit dem Nutzer ergab: die PDF soll der UI-Sprache folgen,
  nicht fest deutsch bleiben.
- Neue Schlüssel unter `rechnung.*` (Formular), `rechnung.pdf.*` (Text auf
  dem PDF selbst), `serienmail.*` und `verein.*` in `i18n/de.go`/`i18n/en.go`,
  dazu ein paar `feld.*`-Ergänzungen für Felder, die es vorher noch nicht
  gab (Rechnungsnummer, Steuersatz, Zahlungsziel, Name des Vereins, BIC,
  Kreditinstitut …). `feld.bezeichnung` (vom Trainingstermin, englisch
  "Label") wird für die Rechnungsposition **nicht** wiederverwendet, obwohl
  der deutsche Text identisch wäre — ein Code-Review-Durchgang fand die
  Verwechslung: "Label" passt auf eine Termin-Bezeichnung, nicht auf eine
  Leistungsbeschreibung. Eigener Schlüssel `feld.position_bezeichnung`
  dafür, in Formular und PDF-Spaltenkopf gleichermaßen benutzt.
- **`service.RechnungBeschriftungen`** (neu, `service/rechnung_pdf.go`)
  bündelt alle Textbausteine des PDFs (Präfixe wie "Rechnungsnummer: ",
  die vier Spaltenköpfe der Positionstabelle, die Steuerzeile als
  `fmt.Sprintf`-Vorlage). `rechnungPDF` und `service.RechnungErstellen`
  nehmen sie jetzt als Parameter entgegen, statt die Texte fest zu
  verdrahten — `zeichner` trägt sie als Feld, `vereinKontaktzeilen`/
  `bankverbindungKontaktzeilen` bekommen sie als Argument. `service/`
  bekommt dabei **keinen** `i18n`-Import (ADR-0002/ADR-0017 bleiben
  gültig): `app/rechnung.go` löst die Beschriftungen über die neue
  Funktion `rechnungBeschriftungen(sprache)` auf und reicht den fertigen
  Text durch. `RechnungBeschriftungenDeutsch` ist der exportierte deutsche
  Standardsatz — Vorgabe für jeden Aufrufer ohne eigene Übersetzung, allen
  voran die sieben Aufrufstellen in `service/rechnung_test.go`.
- `app/rechnung.go`s eigene Meldungen (Datumsprüfung, Positionsfehler, der
  Datei-Dialog beim Anbieten, der Dateiname des Speicherziels) lösen ihre
  Übersetzung über `i18n.Text(a.Sprache(), …)` auf — derselbe Weg wie in
  Ticket 04/05. Meldungen aus `service.ValidierungsFehler` (etwa
  `EinzelpreisAusEuro`) bleiben bewusst deutsch: das ist Ticket 07.
- `app/verein.go` ebenso für die beiden Go-gebauten Meldungen (Logo zu
  groß, gespeichert). Die übrige Formularprüfung des Verein-Formulars
  läuft komplett über `service.LogoAktualisieren`/`ValidierungsFehler` und
  bleibt damit ebenfalls Ticket 07.
- `templates/verein.html`: der MoneyMoney-Hinweis mit eingebettetem
  `<strong>` ist in drei Katalogschlüssel (`_vor`/`_fett`/`_nach`) zerlegt,
  derselbe Kniff wie `import.stundenplan_leer_*` aus Ticket 05, damit
  `html/template` das `<strong>` nicht escapt. Der nicht-brechende
  Trenner in "§ 19 UStG" steht als ` `-Escape im Go-Quelltext
  (funktional identisch zum echten Zeichen aus Ticket 05, aber ohne
  unsichtbares Zeichen im Diff).
- `go test ./...` grün, `wails build` unter Linux grün.
