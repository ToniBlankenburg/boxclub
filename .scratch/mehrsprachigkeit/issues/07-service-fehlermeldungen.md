Status: ready-for-human

# 07: Fehlermeldungen und Validierungstexte in service/ übersetzt

**What to build:** `service.ValidierungsFehler`, Sentinel-Fehler
(`ErrNichtAktiv` u. ä.) und alle `fmt.Errorf`-Texte, die bis in die
Oberfläche durchgereicht werden (`app.fehlerAntwort`), bekommen
Übersetzungen. `service/` kennt heute keine Sprache — das ist der größte
Bruch mit dem bestehenden Seam (ADR-0002: `service/` ist reine Fachlogik,
keine Präsentation) und muss sorgfältig geschnitten werden: der Service
sollte weiterhin Sprache-unabhängige Fehlerwerte liefern (Sentinel-Fehler,
Fehlercodes), die `app/` erst am Rand in Text übersetzt — nicht der Service
selbst, der sonst eine Präsentationsabhängigkeit bekäme.

**Blocked by:** 01 (i18n-Infrastruktur)

## Acceptance Criteria

- [x] Klären: bekommt jeder Fehlerfall einen stabilen Schlüssel/Code, den
      `app/` übersetzt, oder bleibt der Fehlertext selbst deutsch und wird
      nur an bekannten Stellen ersetzt? (Architekturentscheidung, ggf. ADR)
- [x] `service/` bekommt **keine** Abhängigkeit auf `i18n` (ADR-0002 bleibt
      gültig) — Übersetzung passiert in `app/`
- [x] `go test ./...` grün, `wails build` unter Linux grün

## Comments

### Umsetzung

Architekturentscheidung als [ADR-0018](../../../docs/adr/0018-validierungsfehler-als-schluessel-nicht-als-text.md):
jeder Fehlerfall bekommt einen stabilen Schlüssel.

- `service.ValidierungsFehler.Meldungen` ist jetzt `[]service.Meldung`
  (`service/meldung.go`) statt `[]string` — `Meldung{Schluessel, Args}`
  spiegelt genau die Signatur von `i18n.Text(sprache, schluessel, args...)`.
  `service/` baut an allen ca. zwanzig Konstruktionsstellen über acht
  Dateien weiterhin genau die Werte, die vorher in `fmt.Sprintf` gingen
  (Dateiname, Positionsnummer, `MaxTrainingstermine`, …), jetzt als `Args`
  statt als fertigen Satz. Neue Schlüssel unter `validierung.<bereich>.<grund>`
  in `i18n/de.go`/`i18n/en.go`, etwa `validierung.mitglied.vorname_leer` oder
  `validierung.termin.archiviert`.
- `App.uebersetzeMeldungen(meldungen []service.Meldung) []string` (`app/app.go`)
  ist die einzige neue Übersetzungsstelle: sie ersetzt jedes bisherige direkte
  `fehler = validierung.Meldungen` an neun Aufrufstellen in `app/app.go`,
  `app/dokument.go`, `app/verein.go`, `app/trainingstermin.go`,
  `app/rechnung.go`. `app/rechnung.go`s eigener Helfer `validierungsMeldungen`
  bekommt dafür ein `sprache i18n.Sprache`-Argument — er reichte bisher
  ungeprüft deutschen Service-Text in einen eigenen `i18n.Text`-Aufruf
  (`rechnung.fehler_position_praefix`) durch, was in der englischen Ansicht
  einen deutschen Satzteil gezeigt hätte.
- `uebernehmen` (`app/import.go`, der Excel-Übernahme-Orchestrator) bekommt
  ebenfalls ein `sprache i18n.Sprache`-Argument: es ist eine freie Funktion
  ohne `*App`, die die `ValidierungsFehler`-Schlüssel einer abgewiesenen Zeile
  übersetzt, bevor sie neben die schon übersetzten Hinweise des
  `ExcelImporter` (Ticket 05) in den Bericht gehen.
- Sentinel-Fehler (`ErrNichtGefunden`, `ErrNichtAktiv`, `ErrBereitsAktiv`)
  waren bereits vor diesem Ticket sprachunabhängig gelöst (`errors.Is` in
  `app/`, Katalogschlüssel dort gewählt) — daran ändert sich nichts.
- Bewusst **nicht** übersetzt: die internen `fmt.Errorf`-Meldungen in
  `service/` (Transaktions- und DB-Fehler), die unverändert in
  `app.fehlerAntwort` landen — sie sind Diagnosetext für einen Programmfehler,
  keine Meldung zu einer abgelehnten Eingabe (Begründung in ADR-0018). Ebenso
  bewusst nicht übersetzt: `app/dokument.go` insgesamt — die Datei wurde von
  keinem Ticket 01–06 erfasst (kein `i18n`-Import), sie vollständig zu
  übersetzen wäre ein eigenes Ticket. Der `ValidierungsFehler`, den
  `VertragAblegen` liefert, läuft trotzdem korrekt durch
  `App.uebersetzeMeldungen`.
- Tests, die bisher den deutschen Wortlaut prüften
  (`validierung.Meldungen[0] == "Vorname darf nicht leer sein."`), prüfen
  jetzt Schlüssel und Argumente. Neuer Helfer `schluessel()` in
  `member_service_test.go` (dem Referenz-Testfile) für die häufige
  Schlüsselliste-vs-erwartet-Prüfung.
- `go test ./...` grün, `wails build` unter Linux grün.
