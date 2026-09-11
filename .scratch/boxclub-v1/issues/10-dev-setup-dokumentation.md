Status: ready-for-agent

# 10: Dev-Setup-Dokumentation

**What to build:** Eine `README.md`, die einer frischen Klone-Sitzung erklärt, wie sie das Projekt zum Laufen bringt — inklusive plattform-spezifischer Voraussetzungen (WebView2 auf Windows, `libwebkit2gtk-4.0-dev` auf Linux, Go-Version), der `wails dev` / `wails build`-Kommandos und der häufigsten Startfehler.

**Blocked by:** 01 (Wails-Projekt-Skelett)

## Acceptance Criteria

- [ ] `README.md` im Repository-Root beschreibt die Voraussetzungen unter **Windows**: erforderliche Go-Version, WebView2-Runtime, Wails-CLI-Installation
- [ ] `README.md` beschreibt die Voraussetzungen unter **Linux**: erforderliche Go-Version, `libwebkit2gtk-4.0-dev` (Ubuntu/Debian) bzw. Distro-Äquivalente, Wails-CLI-Installation
- [ ] Kommando-Sequenz "Repo klonen → Prereqs installieren → `wails dev`" ist vollständig und wurde einmal auf einer Dev-Plattform durchgespielt
- [ ] Troubleshooting-Sektion listet mindestens die zwei häufigsten Startfehler pro Plattform mit Lösungshinweis
- [ ] Verweise auf [CONTEXT.md](CONTEXT.md), [ADR-0001](docs/adr/0001-go-wails-fuer-desktop-gui.md) und [spec.md](.scratch/boxclub-v1/spec.md) sind vorhanden
- [ ] Klarstellung: **macOS-Nutzung ist Produktions-Ziel, Windows/Linux sind Entwicklungs-Plattformen ohne Auslieferungs-Anspruch**
- [ ] Dokumentiert, dass es **keinen Migrationsmechanismus** gibt: das Schema entsteht über `CREATE TABLE IF NOT EXISTS`, und nach einer Schemaänderung muss die Entwicklungs-Datenbank (`BOXCLUB_DB`) gelöscht werden. Andernfalls startet die App gegen ein veraltetes Schema und die Fehlermeldung zeigt nicht auf die Ursache

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

