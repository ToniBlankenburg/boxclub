Status: ready-for-human

# 05: Excel-Import übersetzt

**What to build:** `templates/import.html`, `app/import.go` und der
Fehlerbericht aus `importer/` (siehe Ticket 09 der ursprünglichen Spec) —
inklusive der Zeilen- und Spaltenfehler, die der Importer je Excel-Zeile
meldet.

**Blocked by:** 01 (i18n-Infrastruktur)

## Acceptance Criteria

- [x] Formular- und Berichtstexte übersetzt
- [x] Fehlermeldungen des `ExcelImporter` übersetzt (Sprache muss vom
      Aufrufer übergeben werden — `importer/` kennt heute keine Sprache)
- [x] `go test ./...` grün, `wails build` unter Linux grün

## Comments

### Umsetzung

- Neue Schlüssel unter `import.*` in `i18n/de.go`/`i18n/en.go`: `import.*`
  ohne Unternamensraum für die statischen Seitentexte (Formular, Bericht,
  Stundenplan-leer-Hinweis), `import.fehler.*` für Gründe, an denen eine
  Zeile oder die ganze Datei scheitert, `import.hinweis.*` für Meldungen zu
  Zeilen, die trotzdem durchgehen (`Ergebnis.Hinweise`).
- **`importer.ExcelImporter.Lesen` nimmt jetzt `i18n.Sprache` als drittes
  Argument entgegen** (Kern des Tickets: `importer/` kennt selbst keine
  Sprache). Sie reist durch die interne Aufrufkette —
  `kopfLesen`/`zeilenLesen`/`zeileLesen` als Parameter, `zeilenleser` und
  `Stundenplan.zuordnen` mit einem eigenen Feld/Argument — bis zu jeder
  Stelle, die eine Meldung baut.
- `zeilenleser.melden`/`.hinweisen` nehmen jetzt einen Katalogschlüssel statt
  eines `fmt`-Formats entgegen und lösen ihn selbst über `i18n.Text` auf.
  Eine Ausnahme: `Stundenplan.zuordnen` liefert seine Meldung schon fertig
  übersetzt (sie entsteht dort, nicht im Zeilenleser) — dafür gibt es
  `hinweisenText`, das den Text unverändert anhängt, statt ihn ein zweites
  Mal (und falsch) durch den Katalog zu schicken.
- Die drei Konstanten `unbekannterFreitext`/`archivierterFreitext`/
  `mehrdeutigerFreitext` in `importer/stundenplan.go` sind aufgelöst — ihr
  Text steht jetzt unter `import.hinweis.termin_*`.
- „Verwaltung" (Blattname) und „Training - 1/2/3" (Spaltenüberschriften)
  bleiben in beiden Sprachkatalogen wörtlich gleich: das sind die
  tatsächlichen Namen aus der Excel-Vorlage des Vereins
  (`importer.blatt`, `importer.spalteTraining1` usw.) und keine
  Anzeigebeschriftung — dieselbe Grenze wie bei
  `service.Wochentag.Bezeichnung()` in Ticket 04.
- Der nicht-brechende Trenner in „Training&nbsp;-&nbsp;1/2/3" steht in den
  Katalogtexten als echtes U+00A0-Zeichen im Go-Quelltext, nicht als
  `&nbsp;`-Entity: `html/template` escapt den Text aus `i18n.Text` beim
  Einsetzen, und ein escaptes `&nbsp;` würde als sichtbarer Text
  „&amp;nbsp;" erscheinen statt als Leerzeichen.
- `app/import.go`s eigene drei Meldungen (Upload zu groß, keine Datei
  gewählt, falsche Dateiendung) lösen ihre Übersetzung jetzt über
  `i18n.Text(a.Sprache(), …)` auf, wie es Ticket 04 für Go-gebauten Text
  schon vorgesehen hat.
- Ein Review-Durchgang fand eine übersehene Stelle: `lebenszyklus`s
  Meldung für eine leere Status-Spalte hing noch am deutschen
  Literal als (nicht existierendem) Katalogschlüssel — behoben unter
  `import.fehler.status_fehlt`, mit Regressionstest.
- `go test ./...` grün, `wails build` unter Linux grün.
