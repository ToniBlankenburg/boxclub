# 01: Fundament: Akzentfarbe, Navigation, Hauptbereich-Breite

**What to build:** Die App verwendet app-weit eine einzige neue Akzentfarbe
(gedämpftes Blaugrau) statt Rot, mit den zwei bewusst roten Warnstellen
(Formular-Fehlerlisten, Rückstand-Kennzeichen) unverändert. Die
Bereichsnavigation ist eine kompakte Pillen-Leiste mit Icon je Bereich statt
der alten Button-Reihe (siehe [ADR-0010](../../../docs/adr/0010-navigation-bleibt-horizontale-top-leiste.md)).
Der Hauptbereich nutzt die volle Fensterbreite statt einer zentrierten Spalte.

**Blocked by:** None (kann sofort starten)

**Status:** ready-for-agent

- [ ] Akzentfarbe ist als CSS-Variablen definiert (`--akzent`, `--akzent-dunkel`,
      `--akzent-hell`, Ring-Opazitäten), keine Rot-Utility-Klasse steht mehr
      für die Akzentfarbe fest kodiert
- [ ] Formular-Fehlerlisten und das Rückstand-Kennzeichen bleiben unverändert
      rot — kein Farbwechsel an diesen zwei bewusst beibehaltenen Warnstellen
- [ ] Navigation zeigt ein `aria-hidden`-Icon je Bereich; der aktive Eintrag
      ist per `aria-current="page"` und Akzentfarbe hervorgehoben
- [ ] Hauptbereich hat keine `max-w`-Begrenzung mehr, nutzt die volle Breite
- [ ] Mitgliederliste (Tabelle mit ein-/ausblendbaren Spalten aus Ticket 27
      in `boxclub-v1`) ist strukturell unverändert, zeigt aber die neue
      Akzentfarbe konsistent (Buttons, Fokus-Ringe, Checkbox)
- [ ] `go build ./...`, `go vet ./...`, `go test ./...` laufen fehlerfrei
- [ ] Manueller Smoke-Test in `wails dev`: Mitgliederliste laden, zwischen
      allen 6 Bereichen wechseln, Suche/Filter benutzen — keine visuellen
      Regressionen
- [ ] Bereits vorhandene Umsetzung (Prototyp-Fold-Commit `8fa5778`, Branch
      `prototype/stilrichtungen`) ist gegen diese Kriterien geprüft, nicht
      neu gebaut

## Comments

### Kontext

Teil des UI-Refactorings, siehe [Spec](../spec.md). Die Umsetzung liegt
größtenteils bereits in `main` — dieses Ticket verifiziert und härtet sie,
statt sie neu zu entwerfen.
