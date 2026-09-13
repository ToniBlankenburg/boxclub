# Boxclub Mitgliederverwaltung

Lokale Desktop-App zur Verwaltung der Vereinsmitglieder — ersetzt die bisherige
Excel-Tabelle. Go + [Wails v2](https://wails.io) mit nativem WebView-Frontend.
Hintergrund und Entscheidung für den Stack: [ADR-0001](docs/adr/0001-go-wails-fuer-desktop-gui.md).
Domain-Vokabular und fachlicher Zuschnitt: [CONTEXT.md](CONTEXT.md) und
[spec.md](.scratch/boxclub-v1/spec.md).

**Produktions-Zielplattform ist ausschließlich macOS** — dort läuft die App im
Alltag, auf dem Rechner, auf dem auch die Kontoauszüge landen. Windows und
Linux sind reine **Entwicklungs-Plattformen**: der Entwickler hat keinen
täglichen Mac-Zugriff und iteriert auf zwei Notebooks. Die dort erzeugten
Binaries sind Entwicklungs-Artefakte ohne Auslieferungsanspruch; es gibt keine
Windows- oder Linux-Releases (siehe „Out of scope" in [spec.md](.scratch/boxclub-v1/spec.md)).

## Voraussetzungen

Gemeinsam für beide Entwicklungs-Plattformen:

- **Go 1.25 oder neuer** (`go.mod` verlangt `go 1.25.0`; `go version` prüft die installierte Version)
- **Node.js + npm**, für den Frontend-Build-Schritt, den Wails automatisch aufruft (`frontend:install` / `frontend:build` in `wails.json`) — keine eigene npm-Pipeline, nur Wails' eingebauter Schritt
- **Wails-CLI**, installiert über Go selbst:

  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```

  Das Binary landet in `$(go env GOPATH)/bin` (Linux: meist `~/go/bin`,
  Windows: `%USERPROFILE%\go\bin`) — dieses Verzeichnis muss im `PATH` stehen,
  sonst meldet die Shell `wails: command not found` bzw. `'wails' is not
  recognized`.

### Windows

- Go 1.25+ und die Wails-CLI wie oben.
- **WebView2-Runtime.** Wails rendert das UI über Microsoft Edge WebView2, keine
  eigene Chromium-Kopie. Auf aktuellen Windows-10/11-Installationen ist die
  Runtime über Windows Update meist bereits vorhanden; falls nicht, installiert
  der [Evergreen Bootstrapper](https://developer.microsoft.com/microsoft-edge/webview2/)
  von Microsoft sie nach.
- Kein zusätzlicher C-Compiler nötig: SQLite läuft über `modernc.org/sqlite`
  (reines Go, kein CGo), und der native Fenster-Rahmen von Wails bindet
  WebView2 unter Windows über COM, nicht über CGo. Ein MinGW/gcc-Setup ist für
  dieses Projekt also nicht erforderlich.

### Linux (Ubuntu/Debian und Ableitungen)

- Go 1.25+ und die Wails-CLI wie oben.
- **GTK- und WebKit-Entwicklungspakete**, weil Wails unter Linux über CGo an
  GTK3/WebKitGTK bindet:

  ```bash
  sudo apt install libgtk-3-dev libwebkit2gtk-4.1-dev
  ```

  Ubuntu 24.04 und Linux Mint 22(.1) liefern nur noch `libwebkit2gtk-4.1-dev`,
  nicht mehr das ältere `libwebkit2gtk-4.0-dev`. Passend dazu steht in
  [`wails.json`](wails.json) bereits `"build:tags": "webkit2_41"` — das Tag
  wirkt sich nur auf Linux-Dateien aus und ist unter macOS/Windows wirkungslos,
  sodass `wails dev` / `wails build` ohne zusätzliche Flags funktionieren.
  Läuft die Entwicklung auf einer älteren Distribution (z. B. Ubuntu 22.04),
  die noch `libwebkit2gtk-4.0-dev` mitbringt, muss `build:tags` entsprechend
  zurückgesetzt werden.
- Ein C-Compiler (`build-essential`) für das CGo-Binding oben.

## Erste Schritte

```bash
git clone <repo-url>
cd boxclub

# Linux:
sudo apt install libgtk-3-dev libwebkit2gtk-4.1-dev build-essential
# Windows: WebView2-Runtime i. d. R. bereits vorhanden, siehe oben.

go install github.com/wailsapp/wails/v2/cmd/wails@latest

wails dev     # Hot-Reload-Entwicklungsmodus, öffnet das WebView-Fenster
```

Weitere Kommandos:

```bash
wails build   # natives Binary nach build/bin/
go test ./... # Go-Tests, ohne Wails-Toolchain lauffähig
go test ./service/...                              # nur ein Package
go test ./service/... -run TestMemberService_Create # nur ein Test
```

Diese Sequenz wurde auf Linux (Mint 22.1) einmal vollständig durchgespielt:
`wails dev` baut Frontend und Backend, startet den DevServer auf
`http://localhost:34115` und öffnet das native Fenster; `wails build` erzeugt
in wenigen Sekunden ein lauffähiges Binary unter `build/bin/`. Die
Windows-Verifikation steht noch aus (siehe Ticket 01) — sollte dabei etwas vom
Verhalten hier abweichen, bitte in diesem README ergänzen.

## Datenbank

Die SQLite-Datei legt die App beim ersten Start im Benutzer-Konfigurationsordner
an (macOS: `~/Library/Application Support/Boxclub/boxclub.db`). Für Entwicklung
und manuelle Tests überschreibt `BOXCLUB_DB=/pfad/zur/test.db wails dev` den Ort.

**Es gibt keinen Migrationsmechanismus.** Das Schema entsteht beim Start über
`CREATE TABLE IF NOT EXISTS` — bestehende Tabellen werden nie automatisch
verändert. Nach jeder Schemaänderung im Code muss die Entwicklungs-Datenbank
darum von Hand gelöscht werden:

```bash
rm "$BOXCLUB_DB"
# oder, ohne Override, am Default-Pfad der jeweiligen Plattform
```

Wird das versäumt, startet die App gegen ein veraltetes Schema. Die
resultierende Fehlermeldung (meist ein SQL-Fehler zu einer fehlenden Spalte
oder Tabelle) zeigt dabei nicht auf die eigentliche Ursache — an genau diesen
Startfehler denken, bevor man in der Anwendungslogik sucht.

## Troubleshooting

### Windows

1. **Leeres/weißes Fenster beim Start, kein Rendering.** Meist fehlt die
   WebView2-Runtime oder sie ist veraltet. Lösung: Evergreen Bootstrapper von
   Microsoft installieren (siehe oben) und `wails dev` neu starten.
2. **`wails: The term 'wails' is not recognized...` nach `go install`.** Das
   Wails-Binary liegt in `%USERPROFILE%\go\bin`, dieses Verzeichnis fehlt im
   `PATH`. Lösung: `PATH` um `%USERPROFILE%\go\bin` ergänzen (PowerShell-Profil
   oder Systemumgebungsvariablen) und die Shell neu starten.

### Linux

1. **`wails doctor` meldet `libwebkit ... Not Found`, obwohl
   `libwebkit2gtk-4.1-dev` installiert ist.** `wails doctor` kennt nur das
   Paket `webkit2gtk-4.0` und erkennt 4.1 nicht als vorhanden — das ist ein
   Fehlalarm dieses Kommandos, kein tatsächliches Problem. Prüfen, ob
   `pkg-config --exists webkit2gtk-4.1` erfolgreich ist und `wails.json` das
   `build:tags": "webkit2_41"`-Tag gesetzt hat; wenn ja, funktionieren
   `wails dev`/`wails build` trotz der Warnung.
2. **Compile-Fehler à la `gtk/gtk.h: No such file or directory` oder
   `Package webkit2gtk-4.1 was not found`.** Die GTK-/WebKit-Entwicklungspakete
   fehlen tatsächlich. Lösung: `sudo apt install libgtk-3-dev
   libwebkit2gtk-4.1-dev` (auf älteren Distributionen ohne 4.1-Paket:
   `libwebkit2gtk-4.0-dev` installieren und `build:tags` in `wails.json`
   entsprechend anpassen).
3. **`listen tcp 127.0.0.1:34115: bind: address already in use`.** Ein
   vorheriger `wails dev`-Prozess (oder der zugehörige Vite-/App-Prozess)
   läuft noch im Hintergrund, meist nach einem abgebrochenen Terminal statt
   sauberem Beenden. Lösung: laufende Prozesse beenden
   (`pkill -f "wails dev"`, ggf. zusätzlich den `frontend`-Vite-Prozess und das
   `build/bin/boxclub-dev-*`-Binary) und `wails dev` neu starten.

### Plattformübergreifend

- **Blankes Fenster / 404 auf alle Assets bei `go run .` statt `wails dev`.**
  `main.go` bindet `frontend/dist` per `//go:embed` ein; dieses Verzeichnis
  wird nur von Wails' Frontend-Build gefüllt. Direktes `go run .` oder
  `go build` embedded einen leeren (bzw. nur den `gitkeep`-Platzhalter
  enthaltenden) Ordner. Lösung: für die Anwendung selbst immer `wails dev`
  oder `wails build` verwenden; `go build ./...` / `go test ./...` bleiben für
  Backend-Arbeit ohne Wails-Toolchain trotzdem sinnvoll und funktionieren ohne
  Frontend-Build.

## Dokumentation

- Domain-Vokabular: [CONTEXT.md](CONTEXT.md)
- Architekturentscheidungen: [docs/adr/](docs/adr/)
- Spec und Tickets: [.scratch/boxclub-v1/](.scratch/boxclub-v1/)
