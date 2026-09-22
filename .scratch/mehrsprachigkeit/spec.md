Status: ready-for-agent

# Spec: Mehrsprachige Oberfläche (Deutsch/Englisch)

## Problem Statement

Die App ist heute vollständig deutsch — jeder Text liegt entweder als
Literal in einem `templates/*.html`-Fragment oder als Go-String-Literal in
`app/*.go` (meist als `Beschriftung`-Feld). Der Nutzer möchte zwischen
Deutsch und Englisch umschalten können, ohne dass die App neu installiert
oder neu gebaut werden muss.

## Solution

Eine kleine Übersetzungsschicht (`i18n/`) mit zwei Katalogen (Deutsch,
Englisch), eine Einstellung, die den Neustart übersteht, und ein
Umschalter in der Oberfläche. Die Sprache gilt fürs ganze Programm — es
gibt kein Login und keinen zweiten Nutzer (CLAUDE.md → Out of scope), also
auch keine Sprache pro Nutzer.

**Umfang laut Klärung mit dem Nutzer:** Übersetzt wird **alles**,
einschließlich der Fachbegriffe aus `CONTEXT.md` (Mitglied → Member,
Mitgliedschaft → Membership, Rückstand → Arrears, …). Das gilt nur für
Texte, die der Nutzer sieht — Go-Bezeichner, DB-Spalten und die deutschen
Begriffe in `CONTEXT.md`/ADRs selbst ändern sich nicht; sie beschreiben das
Domänenmodell, nicht die Anzeige.

**Speicherort:** eine eigene JSON-Datei im Benutzer-Konfigurationsordner,
nicht die Datenbank — dieselbe Idee wie `BOXCLUB_DB` in `main.go`, nur für
eine Einstellung, die nichts mit den Mitgliederdaten zu tun hat und beim
Löschen der Dev-Datenbank nicht mit verschwinden soll.

**Umfang der Umsetzung:** wegen der Größe (≈9 Templates, ~1700 Zeilen
`app/app.go` mit Text-Literalen, Fehlermeldungen in `service/`) über
mehrere Sessions verteilt, siehe Tickets. Jedes Ticket liefert für sich
einen lauffähigen, getesteten Zwischenstand — kein Ticket hinterlässt eine
kaputte Build.

## Architektur

- `i18n.Sprache` (`"de"` / `"en"`), `i18n.Text(sprache, schluessel, args...) string`
  als reine Nachschlagefunktion, kein Zustand. Fehlt ein Schlüssel in einer
  Sprache, fällt sie auf Deutsch zurück, fehlt er auch dort, auf den
  Schlüssel selbst — eine fehlende Übersetzung darf nie zu einer leeren
  Stelle in der Oberfläche führen.
- Kataloge als Go-Maps (`i18n/de.go`, `i18n/en.go`), kein JSON/YAML: die
  App hat ohnehin keinen Übersetzungs-Workflow mit Fremdwerkzeug, und ein
  fehlender Schlüssel soll ein Compile-Fehler beim Zugriff über eine
  Konstante sein können, sobald Ticket 02 die Schlüssel typisiert (siehe
  dort) — mit einer Map aus Go-Literalen bleibt das möglich, mit JSON nicht.
- `i18n.Einstellungen{Sprache}` mit `Laden`/`Speichern` gegen eine
  JSON-Datei, Pfad überschreibbar via `BOXCLUB_EINSTELLUNGEN` (Pendant zu
  `BOXCLUB_DB`). Fehlt die Datei, gilt Deutsch als Standard — eine
  bestehende Installation soll nicht plötzlich englisch starten.
- `App` hält die aktuelle Sprache im Speicher (mutex-geschützt, der
  Assetserver bedient mehrere Goroutinen) und eine Methode `t` im
  Template-`FuncMap`, die bei jedem `Execute` die aktuelle Sprache
  nachschlägt — kein erneutes Parsen der Templates beim Sprachwechsel
  nötig.
- Text, den der Go-Code selbst zusammenbaut, bevor er ins Template geht
  (z. B. `navigationseintrag.Beschriftung`), löst seine Übersetzung im
  Go-Code auf (`i18n.Text(a.Sprache(), …)`), nicht im Template — sonst gäbe
  es zwei Übersetzungsstellen für denselben Text.

## Bekannte Lücke

`frontend/index.html` ist eine statische Datei (Vite-Build, kein
Go-Template) mit drei deutschen Strings (Fenstertitel, `aria-label`,
Ladehinweis). Sie wird nur für den Sekundenbruchteil bis zum ersten
`hx-trigger=load`-Aufruf sichtbar. Sie über das Backend auszuliefern statt
über den Wails-Assetserver wäre ein eigener Schnitt (siehe Ticket 08) und
ändert das Aufbau-Muster der App (ADR-0002); für v1 dieser Funktion bleibt
sie unübersetzt.

## Tickets

01. i18n-Infrastruktur: Katalog, Einstellungsdatei, Sprache im `App`,
    Umschalter in der Navigation, Navigation selbst vollständig übersetzt.
02. Mitgliederliste, Filterleiste, Mitglied-Formular
03. Dashboard
04. Trainingstermine
05. Excel-Import (inklusive Fehlerbericht)
06. Rechnung, Serienmail, Verein-Formular
07. Fehlermeldungen und Validierungstexte in `service/`
08. `frontend/index.html` über das Backend ausliefern (optional, s. o.)

Jedes Ticket 02–08 übersetzt zusätzlich die Fachbegriffe, die in seinem
Bereich vorkommen, in den englischen Katalog.
