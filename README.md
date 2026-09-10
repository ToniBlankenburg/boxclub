# Boxclub Mitgliederverwaltung

Lokale Desktop-App zur Verwaltung der Vereinsmitglieder — ersetzt die bisherige
Excel-Tabelle. Go + [Wails v2](https://wails.io) mit nativem WebView-Frontend.

Produktions-Zielplattform ist **macOS**; entwickelt wird unter Windows und Linux.
Die dort erzeugten Binaries sind Entwicklungs-Artefakte, kein Auslieferungsziel.

```bash
wails dev     # Hot-Reload-Entwicklungsmodus
wails build   # natives Binary nach build/bin/
go test ./... # Go-Tests, ohne Wails-Toolchain lauffähig
```

Die SQLite-Datei legt die App beim ersten Start im Benutzer-Konfigurationsordner
an (macOS: `~/Library/Application Support/Boxclub/boxclub.db`). Für Entwicklung
und manuelle Tests überschreibt `BOXCLUB_DB=/pfad/zur/test.db wails dev` den Ort.

Das vollständige Setup inklusive plattform-spezifischer Voraussetzungen und
Troubleshooting beschreibt
[Ticket 10](.scratch/boxclub-v1/issues/10-dev-setup-dokumentation.md).

## Dokumentation

- Domain-Vokabular: [CONTEXT.md](CONTEXT.md)
- Architekturentscheidungen: [docs/adr/](docs/adr/)
- Spec und Tickets: [.scratch/boxclub-v1/](.scratch/boxclub-v1/)
