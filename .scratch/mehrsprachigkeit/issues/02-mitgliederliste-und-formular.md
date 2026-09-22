Status: ready-for-human

# 02: Mitgliederliste, Filterleiste, Mitglied-Formular übersetzt

**What to build:** Alle Texte in `templates/mitglieder_liste.html`,
`templates/mitglied_formular.html` und die zugehörigen `Beschriftung`-Literale
in `app/app.go` (Spaltenköpfe, Filteroptionen, Feldbeschriftungen) wandern in
den Katalog, inklusive der Fachbegriffe (Mitglied → Member, Rückstand →
Arrears, Anmeldegebühr → Registration fee, Kündigung → Cancellation, Eintritt
→ Join date, Austritt → Exit date, Ruhend → Suspended, Trainingsfrequenz →
Training frequency, …).

**Blocked by:** 01 (i18n-Infrastruktur)

## Acceptance Criteria

- [x] Spaltenköpfe, Filteroptionen (`spalte*`, `filteroption`-Listen),
      Formularbeschriftungen übersetzt
- [x] Meldungstexte (`meldung{Text: "..."}`) an den betroffenen Stellen
      übersetzt
- [x] `go test ./...` grün, `wails build` unter Linux grün

## Comments

### Umsetzung

- `spalte` und `filteroption` trugen ihre Beschriftung bisher als deutsches
  Literal am Wert selbst. Beide sind jetzt sprachneutral (nur noch
  `Schluessel`/`Wert`) und lösen ihre Beschriftung erst in `listeDaten`
  auf, das dafür ein neues Feld `Sprache` trägt — dasselbe Muster wie
  `navigation()` in Ticket 01. `rueckstandsoptionen` und
  `frequenzoptionenBauen` sind deshalb von Paketvariablen zu Funktionen
  geworden, die die Sprache als Parameter nehmen.
- Feldbeschriftungen des Mitglied-Formulars stehen unter dem Schlüssel-Namensraum
  `feld.*` statt `mitglied.*`, weil das Teil-Template "feld"
  (`templates/mitglied_formular.html`) auch von `rechnung.html`,
  `trainingstermine.html` und `verein.html` verwendet wird — spätere
  Tickets sollen dieselben Schlüssel referenzieren statt sie zu verdoppeln
  (der Katalog-Kommentar in `i18n/de.go` weist darauf hin). Ebenso
  `allgemein.speichern`/`allgemein.abbrechen` für die beiden Formular-Knöpfe,
  die in praktisch jedem Formular der App gleich lauten.
- **Bewusst nicht übersetzt** in diesem Ticket (analog zu Ticket 01s
  Zwischenstand):
  - Die freien Fehlertexte, die `app.go` beim Parsen von Datumsfeldern
    selbst zusammenbaut (`"Kündigungsdatum ist kein gültiges Datum."` u. ä.,
    in `kuendigungLesen`, `alsAnmeldung`, `alsNeuesMitglied`,
    `wiedereintrittEintragen`). Das Ticket nennt explizit nur
    `meldung{Text: "..."}` als Umfang; diese Strings sind kein `meldung`,
    sondern Roh-`[]string`, dieselbe Kategorie wie die
    `fmt.Errorf`-Texte aus `service/`, die Ticket 07 zusammen mit den
    Validierungstexten behandelt. Sie bis dahin separat zu übersetzen hätte
    denselben Text an zwei Stellen im Katalog riskiert.
  - Alle `Bezeichnung()`-Ausgaben aus `service/` (Status, Rückstand-Kennzeichen,
    Trainingsfrequenz, Google-Bewertung, Termin-Anzeige) — sie entstehen
    dort als reine Fachlogik ohne Sprachbezug (ADR-0002) und bleiben bis
    Ticket 07 deutsch. Betroffen sind die Badges in `mitglied-zeile` sowie
    die Beschriftungen der Frequenz- und Google-Bewertungs-Kontrollkästchen.
  - `templates/dokument.html` ("vertraege") und `templates/rechnung.html`
    ("rechnung-bereich"), die im Mitglied-Formular als Registerkarten
    eingebettet sind — außerhalb der beiden im Ticket genannten Dateien,
    gehören zu Ticket 06.
- Smoke-getestet über `app.New` + `httptest` gegen beide Sprachen: Liste,
  Anlage- und Bearbeitungsformular, Rückstand- und Kündigungsformular,
  sortierte Kopfzeile — jeweils erwartetes Deutsch bzw. Englisch geprüft,
  keine Templatefehler.
