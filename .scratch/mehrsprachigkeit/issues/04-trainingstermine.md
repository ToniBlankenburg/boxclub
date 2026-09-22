Status: ready-for-human

# 04: Trainingstermine übersetzt

**What to build:** `templates/trainingstermine.html` und
`app/trainingstermin.go` — Wochenplan, Archivieren/Reaktivieren,
Teilnehmerliste-Serienmail. Fachbegriffe: Trainingstermin → Training
session, Wochentag → Weekday, archiviert → archived.

**Blocked by:** 01 (i18n-Infrastruktur)

## Acceptance Criteria

- [x] Alle sichtbaren Texte übersetzt, inklusive Wochentagsnamen
- [x] `go test ./...` grün, `wails build` unter Linux grün

## Comments

### Umsetzung

- Neue Schlüssel unter `trainingstermine.*` in `i18n/de.go`/`i18n/en.go` für
  Seite, Liste, leere Zustände und Formular. Feldbeschriftungen
  (Wochentag/Beginn/Ende/Bezeichnung) stehen unter `feld.*`, wie im
  Kommentar zu diesem Namensraum vorgesehen (Ticket 02).
- Die vier Rückmeldungen (angelegt/gespeichert/archiviert/reaktiviert)
  kommen paarweise als `_mit_name`/`_ohne_name`: `terminMeldung` nimmt jetzt
  zwei Schlüssel statt eines deutschen Textfragments entgegen und wählt
  `_ohne_name`, wenn der Termin zwischen Aktion und erneutem Lesen
  verschwunden ist.
- **Wochentagsnamen** (Acceptance Criteria) sitzen an zwei Stellen, die
  bewusst unterschiedlich behandelt sind:
  - Die Auswahlliste im Terminformular ist reine Formularbeschriftung und
    jetzt über neue Schlüssel `wochentag.1`–`wochentag.7` übersetzt
    (`wochentagsoptionenBauen` in `app/trainingstermin.go`, nimmt jetzt
    `i18n.Sprache` entgegen; `terminformularDaten` trägt dafür ein neues
    `Sprache`-Feld wie `listeDaten` es für die Filterlisten schon tut).
  - `service.Wochentag.Bezeichnung()` selbst — und damit
    `Trainingstermin.Anzeige()`/`KurzeAnzeige()` ("Samstag 10:30 – 12:00 ·
    Anfänger") — bleibt unverändert deutsch. Sie ist die Grundlage des
    Excel-Abgleichs (ADR-0008) und darf nicht von der Anzeigesprache
    abhängen; sie reiht sich damit in die in Ticket 02 dokumentierten,
    bis Ticket 07 unübersetzten `Bezeichnung()`-Ausgaben ein (Status,
    Rückstand, Trainingsfrequenz, Google-Bewertung — jetzt ergänzt um
    Termin-Anzeige). Betroffen sind die Zeilen im Stundenplan selbst, die
    zugehörigen `aria-label`s und die Terminoptionen im Mitgliederfilter
    (`app.go`, dort schon mit Verweis auf dieses Ticket kommentiert).
- `go test ./...` grün, `wails build` unter Linux grün.
