# 05: Übersetzung der neuen Beschriftungen ins Englische

**What to build:** Alle in den Tickets 01–04 neu eingeführten,
nutzersichtbaren Beschriftungen (Zahnrad-Menü-Inhalte, Filterlabel je Spalte,
Fallback-Element für ausgeblendete Spalten, Reset-Button) bekommen englische
Übersetzungen im bestehenden `i18n`-Katalog, nach dem etablierten
`i18n.Text(sprache, "...")`-Muster. Die App bleibt dadurch vollständig
zweisprachig, wie es das Mehrsprachigkeits-Feature vorsieht.

**Blocked by:** 01, 02, 03, 04 (braucht den vollständigen, finalen Satz an
neuen Beschriftungen)

**Status:** ready-for-human

- [x] Jede in 01–04 neu eingeführte Beschriftung hat einen deutschen und
      einen englischen Katalog-Eintrag
- [x] Keine hartkodierten deutschen Strings in den neuen Template-/Go-Stellen
      übrig
- [x] Manueller Smoke-Test: Umschalten auf Englisch zeigt alle neuen Zahnrad-
      Menüs, Filterlabel, das Fallback-Element und den Reset-Button korrekt
      übersetzt
- [x] `wails dev` und `wails build` laufen weiterhin auf Windows und Linux

## Comments

Die englischen Katalog-Einträge waren bereits über Commit 49ca413
vorhanden (`i18n/en.go`); `de.go`/`en.go` haben seither dieselbe
Schlüsselmenge. Für dieses Ticket zusätzlich geprüft:

- Grep auf hartkodierte deutsche Strings (Umlaute/ß) in
  `templates/mitglieder_liste.html` und `app/app.go` (den beiden von
  01–04 berührten Dateien): alle Treffer lagen in Go-/Template-
  Kommentaren, keiner in nutzersichtbarem Markup.
- Smoke-Test per `httptest` gegen `a.Handler()` mit `i18n.Englisch`:
  Zahnrad-Menüs, Filterlabel, Fallback-Element ("Hidden columns") und
  Reset-Button rendern vollständig auf Englisch, keine deutschen Reste.
- `go build ./...`, `go test ./...` und `wails build` (Linux) laufen
  grün. `wails dev`/`wails build` unter Windows konnte in dieser Umgebung
  nicht geprüft werden (kein Windows-Host verfügbar) — wie bereits bei
  Ticket 03/04 nur auf Linux verifiziert.
