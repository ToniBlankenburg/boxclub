Status: ready-for-agent

# 01: i18n-Infrastruktur, Umschalter, Navigation übersetzt

**What to build:** Das Übersetzungs-Fundament, auf dem alle folgenden
Tickets aufbauen: Katalog-Package, persistierte Einstellung, Sprache im
`App`, ein Umschalter in der Oberfläche — und als erster vollständig
übersetzter Ausschnitt die Kopfzeilen-Navigation (`nav.*`).

**Blocked by:** —

## Acceptance Criteria

- [x] `i18n.Sprache` (`Deutsch`/`Englisch`), `i18n.Text(sprache, schluessel, args...) string`
      mit Rückfall auf Deutsch, dann auf den Schlüssel selbst
- [x] `i18n.Einstellungen{Sprache}`, `Laden(pfad)`/`(e) Speichern(pfad)` gegen
      eine JSON-Datei; fehlende Datei → Deutsch, kein Fehler
- [x] `BOXCLUB_EINSTELLUNGEN` überschreibt den Pfad, analog zu `BOXCLUB_DB`
- [x] `App` hält die aktuelle Sprache mutex-geschützt, `t` im
      Template-`FuncMap`, `navigation()` wird Methode und übersetzt
      `Beschriftung` je Bereich (`nav.<schluessel>`)
- [x] Route `POST /api/einstellungen/sprache`: setzt Sprache, schreibt die
      Einstellungsdatei, rendert die Verein-Seite neu (`App.vereinRendern`)
- [x] Umschalter ist ein `<select>` auf der Verein-Seite (`verein.html`),
      kein eigener Knopf/keine eigene Navigation-Erweiterung (Rücksprache mit
      Nutzer nach dem ersten Entwurf mit zwei Knöpfen in der Kopfzeile)
- [x] Test am Seam: `i18n.Text` für beide Sprachen, fehlender Schlüssel,
      fehlende Sprache
- [x] Test am Seam: `Einstellungen.Laden`/`Speichern` mit `t.TempDir()`,
      fehlende Datei, kaputte Datei
- [x] `go test ./...` grün, `wails build` unter Linux grün

## Comments

### Umsetzung

- `i18n.Text` fällt bei unbekanntem Schlüssel **innerhalb** einer Sprache auf
  den deutschen Katalog zurück (nicht auf eine leere Zeichenkette) und erst
  danach auf den Schlüssel selbst — ein Ticket, das einen Schlüssel im
  englischen Katalog vergisst, zeigt also deutschen Text statt eines
  sichtbaren Lochs.
- `navigation()` wurde von einer freien Funktion zu einer Methode auf `App`
  (7 Aufrufstellen angepasst) — sie braucht die aktuelle Sprache, die am
  `App` hängt. Die `Beschriftung`-Literale in `var bereiche` sind entfallen:
  sie wurden ohnehin bei jedem Aufruf überschrieben.
- Erster Entwurf hatte den Umschalter als zwei Pillen-Knöpfe (DE/EN) in der
  Kopfzeilen-Navigation, analog zu den Bereichs-Knöpfen, mit Rückkehr zur
  Mitgliederliste nach dem Umschalten. Der Nutzer wollte das nicht: keine
  eigenen Knöpfe für eine Einstellung, die man selten ändert. Jetzt steht sie
  als `<select>` auf der Verein-Seite (eigenes `<div>`, außerhalb des
  Vereinsdaten-Formulars — die Sprache ist keine `vereinsdaten`-Angabe,
  sondern die eigene Einstellungsdatei, siehe ADR-0017) und schickt sich bei
  `change` selbst ab. Die Rückkehr-Ansicht ist deshalb die Verein-Seite
  selbst (`App.vereinRendern`), nicht mehr die Mitgliederliste.
- Bewusst **nicht** übersetzt in diesem Ticket: alles außerhalb der
  Navigation (Listenansicht, Formulare, Dashboard, Rechnung, Import,
  Trainingstermine, Fehlermeldungen aus `service/`) — das sind die Tickets
  02–07. Bis dahin zeigt die englische Sprache eine übersetzte Navigation
  über sonst weiterhin deutschen Inhalten; das ist der bewusste
  Zwischenstand, kein Bug.
