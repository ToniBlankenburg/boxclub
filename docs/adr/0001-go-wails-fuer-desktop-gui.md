# ADR-0001: Go + Wails als Stack für die Desktop-GUI

**Status:** Accepted
**Datum:** 2026-09-09

## Kontext

Für die Boxclub-Mitgliederverwaltung wurde entschieden:

- **Form-Faktor:** Desktop-GUI (nicht CLI, nicht Web)
- **Ziel-Plattform in Produktion:** ausschließlich macOS (auf dem Rechner, auf dem auch MoneyMoney installiert ist)
- **Entwicklungs-Plattformen:** Windows + Linux — der Entwickler hat keinen täglichen Mac-Zugriff und muss auf seinen zwei Notebooks iterieren können
- **Nutzer:** ein einzelner Admin
- **Zusatzziel des Entwicklers:** Go lernen (Java ist als Alternativsprache verfügbar und beherrscht)

Die Kombination "Desktop-GUI + cross-platform + Ein-Personen-Projekt" ist die teuerste Ecke im Entscheidungsraum. Sprach- und Toolkit-Wahl ist deshalb nicht Formsache, sondern die zentrale Weiche.

## Entscheidung

Wir bauen die Anwendung mit **Go + Wails v2**. Das Frontend läuft als **Vanilla HTML + htmx + Tailwind CSS** in der nativen WebView, die Wails bereitstellt. Das Backend ist reines Go und exponiert Funktionen an das Frontend über Wails-Bindings.

Datenhaltung: **SQLite** in einer lokalen Datei.

## Betrachtete Alternativen

### Java + JavaFX

**Pro:** vertraute Sprache, ausgereiftes GUI-Framework, Tabellen und Formulare "gelöst", cross-platform via JAR.

**Contra:** kein Beitrag zum Lernziel Go. Verpasst die Gelegenheit, Go am realen Projekt zu üben.

### Go + Fyne

**Pro:** rein nativ, ein Sprachstack, echte Binaries pro Plattform.

**Contra:** Fyne ist funktional, aber optisch limitiert; das Layout-System hat Kanten; parallel wird Go-Sprache **und** ein weniger etabliertes GUI-Toolkit gelernt.

### Go + Wails (gewählt)

**Pro:** Go-Lernziel bleibt zentral. Native WebView liefert modernes UI-Rendering, ohne Electron mitzuschleppen. Wails erzeugt echte native Binaries pro Plattform. Tailwind + htmx halten den Frontend-Aufwand klein — kein SPA-Framework, kein Build-Zoo jenseits von Wails' eingebautem `npm`-Schritt.

**Contra:** ein zweiter Sprachstack (HTML/CSS/JS) im Projekt, auch wenn bewusst minimal gehalten. Wails-Ökosystem kleiner als Electron oder JavaFX.

## Konsequenzen

- **Positiv:** Go bleibt der primäre Lernstack. Die Entwicklung auf den Windows-/Linux-Notebooks funktioniert nativ (`wails dev`, `wails build`), ohne dass dafür ein Mac laufen muss. Der Mac-Release-Build wird separat auf einem realen Mac oder einem CI-Runner erzeugt. UI kann mit Tailwind ohne eigenes CSS "schön genug" werden. htmx passt konzeptionell zu Wails (Go-Funktion aufrufen → HTML-Fragment zurückliefern → im DOM ersetzen).
- **Negativ:** Grundwissen in HTML/CSS und htmx-Idiomatik nötig. Wails-spezifische Build- und Signing-Schritte für macOS/Windows-Releases müssen später gelernt werden.
- **Reversibilität:** Umstieg auf Java+JavaFX wäre effektiv ein Rewrite (Sprache, GUI-Framework, Buildkette). Umstieg auf reine Go-CLI mit gleichem SQLite-Schema wäre günstig — das Datenmodell bleibt neutral gegenüber der UI-Wahl.
