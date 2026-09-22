Status: ready-for-agent

# 04: Trainingstermine übersetzt

**What to build:** `templates/trainingstermine.html` und
`app/trainingstermin.go` — Wochenplan, Archivieren/Reaktivieren,
Teilnehmerliste-Serienmail. Fachbegriffe: Trainingstermin → Training
session, Wochentag → Weekday, archiviert → archived.

**Blocked by:** 01 (i18n-Infrastruktur)

## Acceptance Criteria

- [ ] Alle sichtbaren Texte übersetzt, inklusive Wochentagsnamen
- [ ] `go test ./...` grün, `wails build` unter Linux grün
