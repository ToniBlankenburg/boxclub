Status: ready-for-human

# 10: Dev-Setup-Dokumentation

**What to build:** Eine `README.md`, die einer frischen Klone-Sitzung erklärt, wie sie das Projekt zum Laufen bringt — inklusive plattform-spezifischer Voraussetzungen (WebView2 auf Windows, `libwebkit2gtk-4.0-dev` auf Linux, Go-Version), der `wails dev` / `wails build`-Kommandos und der häufigsten Startfehler.

**Blocked by:** 01 (Wails-Projekt-Skelett)

## Acceptance Criteria

- [x] `README.md` im Repository-Root beschreibt die Voraussetzungen unter **Windows**: erforderliche Go-Version, WebView2-Runtime, Wails-CLI-Installation
- [x] `README.md` beschreibt die Voraussetzungen unter **Linux**: erforderliche Go-Version, `libwebkit2gtk-4.0-dev` (Ubuntu/Debian) bzw. Distro-Äquivalente, Wails-CLI-Installation
- [x] Kommando-Sequenz "Repo klonen → Prereqs installieren → `wails dev`" ist vollständig und wurde einmal auf einer Dev-Plattform durchgespielt
- [x] Troubleshooting-Sektion listet mindestens die zwei häufigsten Startfehler pro Plattform mit Lösungshinweis
- [x] Verweise auf [CONTEXT.md](CONTEXT.md), [ADR-0001](docs/adr/0001-go-wails-fuer-desktop-gui.md) und [spec.md](.scratch/boxclub-v1/spec.md) sind vorhanden
- [x] Klarstellung: **macOS-Nutzung ist Produktions-Ziel, Windows/Linux sind Entwicklungs-Plattformen ohne Auslieferungs-Anspruch**
- [x] Dokumentiert, dass es **keinen Migrationsmechanismus** gibt: das Schema entsteht über `CREATE TABLE IF NOT EXISTS`, und nach einer Schemaänderung muss die Entwicklungs-Datenbank (`BOXCLUB_DB`) gelöscht werden. Andernfalls startet die App gegen ein veraltetes Schema und die Fehlermeldung zeigt nicht auf die Ursache

## Comments

### Linux-Prereq ist 4.1, nicht 4.0 (aus Ticket 01)

Die Acceptance Criteria oben nennen `libwebkit2gtk-4.0-dev`. Ubuntu 24.04 und
Linux Mint 22 liefern dieses Paket nicht mehr — dort gilt:

```
sudo apt install libgtk-3-dev libwebkit2gtk-4.1-dev
```

Dazu passend setzt `wails.json` `"build:tags": "webkit2_41"`. Ticket 01 hat als
Zwischenstand bereits eine knappe `README.md` angelegt; dieses Ticket baut sie zur
vollständigen Doku inklusive Troubleshooting aus.

### Umgesetzt

`README.md` auf die volle Doku ausgebaut: Voraussetzungen getrennt nach
Windows/Linux (inkl. der oben genannten 4.1-statt-4.0-Korrektur, mit Verweis
auf ältere Distros als Ausweichoption), Kommando-Sequenz vom Klonen bis
`wails dev`, Datenbank-/Migrations-Abschnitt und eine Troubleshooting-Sektion
mit je zwei Windows- und drei Linux-Fehlern plus einem plattformübergreifenden
(`go run .` statt `wails dev` → leeres Embed). Verweise auf `CONTEXT.md`,
ADR-0001 und `spec.md` sind gesetzt, die macOS/Windows/Linux-Klarstellung
steht in der Einleitung.

Auf Linux (Mint 22.1, diese Session) tatsächlich durchgespielt: `wails build`
erzeugt in ~6s ein lauffähiges Binary, `wails dev` baut Frontend und Backend,
startet den DevServer auf Port 34115 und öffnet das native Fenster — beides
fehlerfrei bis auf die dokumentierte `wails doctor`-Fehlmeldung zu
`libwebkit`, die keine echte Auswirkung hat. `go build ./...`, `go vet ./...`
und `go test ./...` bleiben sauber.

**Windows-Abschnitt ist inhaltlich, aber nicht auf einer echten
Windows-Maschine verifiziert** — hier stand wie in Ticket 01 keine zur
Verfügung. Die WebView2- und PATH-Hinweise stammen aus der Wails-Dokumentation
und allgemeinem Windows/Go-Wissen, nicht aus einem eigenen Testlauf. Bitte
einmal auf dem Windows-Notebook gegenlesen und bei Abweichungen im README
korrigieren, dann auf `ready-for-human` → erledigt setzen.

