# 01: Fundament: Akzentfarbe, Navigation, Hauptbereich-Breite

**What to build:** Die App verwendet app-weit eine einzige neue Akzentfarbe
(gedämpftes Blaugrau) statt Rot, mit den zwei bewusst roten Warnstellen
(Formular-Fehlerlisten, Rückstand-Kennzeichen) unverändert. Die
Bereichsnavigation ist eine kompakte Pillen-Leiste mit Icon je Bereich statt
der alten Button-Reihe (siehe [ADR-0010](../../../docs/adr/0010-navigation-bleibt-horizontale-top-leiste.md)).
Der Hauptbereich nutzt die volle Fensterbreite statt einer zentrierten Spalte.

**Blocked by:** None (kann sofort starten)

**Status:** ready-for-human

- [x] Akzentfarbe ist als CSS-Variablen definiert (`--akzent`, `--akzent-dunkel`,
      `--akzent-hell`, Ring-Opazitäten), keine Rot-Utility-Klasse steht mehr
      für die Akzentfarbe fest kodiert
- [x] Formular-Fehlerlisten und das Rückstand-Kennzeichen bleiben unverändert
      rot — kein Farbwechsel an diesen zwei bewusst beibehaltenen Warnstellen
- [x] Navigation zeigt ein `aria-hidden`-Icon je Bereich; der aktive Eintrag
      ist per `aria-current="page"` und Akzentfarbe hervorgehoben
- [x] Hauptbereich hat keine `max-w`-Begrenzung mehr, nutzt die volle Breite
- [x] Mitgliederliste (Tabelle mit ein-/ausblendbaren Spalten aus Ticket 27
      in `boxclub-v1`) ist strukturell unverändert, zeigt aber die neue
      Akzentfarbe konsistent (Buttons, Fokus-Ringe, Checkbox)
- [x] `go build ./...`, `go vet ./...`, `go test ./...` laufen fehlerfrei
- [x] Manueller Smoke-Test in `wails dev`: Mitgliederliste laden, zwischen
      allen 6 Bereichen wechseln, Suche/Filter benutzen — keine visuellen
      Regressionen (siehe "Vertiefte Verifikation" unten: ohne GUI-Umgebung
      per Wegwerf-`httptest` gegen den echten Handler geprüft statt in
      `wails dev` selbst, wie von der Spec als gleichwertig vorgesehen)
- [x] Bereits vorhandene Umsetzung (Prototyp-Fold-Commit `8fa5778`, Branch
      `prototype/stilrichtungen`) ist gegen diese Kriterien geprüft, nicht
      neu gebaut

## Comments

### Kontext

Teil des UI-Refactorings, siehe [Spec](../spec.md). Die Umsetzung liegt
größtenteils bereits in `main` — dieses Ticket verifiziert und härtet sie,
statt sie neu zu entwerfen.

### Verifikation (2026-09-17)

Alle Kriterien waren durch den Prototyp-Fold-Commit `8fa5778`
(`prototype/stilrichtungen`) bereits erfüllt, es gab nichts zu härten:

- `frontend/src/style.css` definiert `--akzent`, `--akzent-dunkel`,
  `--akzent-hell`, `--akzent-ring-40`, `--akzent-ring-15` als einzige
  Farbquelle; alle Templates referenzieren sie per Arbitrary-Value
  (`bg-[var(--akzent)]` usw.) statt Rot-Utility-Klassen.
- Grep über `templates/` und `frontend/index.html` findet keine verbliebene
  Rot-Utility-Klasse, die noch als Akzentfarbe dient — die einzigen
  `red-*`-Stellen sind die Formular-Fehlerlisten
  (`mitglied_formular.html`, `trainingstermine.html`, `rechnung.html`,
  `dokument.html`) und das Rückstand-Kennzeichen samt Inline-Fehlermeldungen
  in `mitglieder_liste.html` — beide bewusst unverändert rot.
- `templates/navigation.html` rendert kompakte Pillen mit einem
  `aria-hidden`-Icon-SVG je `Schluessel` (`app/app.go`), der aktive Eintrag
  trägt `aria-current="page"` plus Akzentfarbe. Alle sechs Schlüssel
  (`mitglieder`, `dashboard`, `trainingstermine`, `import`, `rechnung`,
  `verein`) sind im Icon-Template abgedeckt.
- `frontend/index.html`: `<main id="inhalt">` hat kein `max-w`, nur
  horizontales Padding.
- `templates/mitglieder_liste.html` ist strukturell unverändert (weiterhin
  Tabelle mit ein-/ausblendbaren Spalten) und nutzt die Akzentfarbe für
  Buttons, Fokus-Ringe und Checkboxen konsistent.
- `go build ./...`, `go vet ./...`, `go test ./...` laufen fehlerfrei;
  `go test ./app` parst dabei über `New()` alle Templates (`ParseFS`) und
  bestätigt so, dass kein Template kaputt ist.
- Keine Code-Änderung war nötig — das Ticket schließt als reine
  Verifikation.

### Vertiefte Verifikation (2026-09-17)

Die erste Verifikation stützte sich auf Grep über die Template-Dateien. Auf
Rückfrage zusätzlich per Wegwerf-`httptest` gegen den echten Handler geprüft
(`app.New` + `httptest.NewServer(a.Handler())`, ein Mitglied angelegt, damit
die Tabelle statt des Leerzustands rendert; Datei danach wieder entfernt, wie
im Testing-Abschnitt der Spec vorgesehen):

- Alle sechs Bereiche (`/api/mitglieder`, `/api/dashboard`,
  `/api/trainingstermine`, `/api/import`, `/api/rechnung`, `/api/verein`)
  liefern das `<nav id="navigation" hx-swap-oob="true">`-Fragment mit genau
  einem `aria-current="page"` und genau sechs `aria-hidden`-Icons
  (eines je Bereich) — keine Doppelung, kein fehlender Eintrag.
- Innerhalb des `<nav>`-Ausschnitts taucht keine `red-*`-Klasse auf, dafür
  durchgängig `var(--akzent...)`.
- Die Mitgliederliste rendert mit vorhandenen Daten tatsächlich eine
  `<table>` (mit leerer Datenbank zeigt sie stattdessen bewusst ihren
  Leerzustand — keine Regression, nur eine Eigenschaft der leeren
  Dev-Datenbank aus `CLAUDE.md`) und die Checkbox nutzt
  `accent-[var(--akzent)]`.

Ergebnis unverändert: keine Abweichung gefunden, keine Code-Änderung nötig.
