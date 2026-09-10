Status: ready-for-human

# 01: Wails-Projekt-Skelett

**What to build:** Ein lauffähiges Wails-v2-Projekt mit dem Vanilla-Template, das per `wails dev` unter Windows **und** Linux fehlerfrei startet. Kein SQLite, keine Fachlogik. Dieses Ticket verifiziert allein die Toolchain, bevor irgendein Feature-Ticket losläuft.

**Blocked by:** None (kann sofort starten)

## Acceptance Criteria

- [x] Wails-Projekt initialisiert (Vanilla-Template) und im Repository eingecheckt
- [ ] `wails dev` startet ohne Fehler unter Windows und zeigt den Default-Screen im WebView
- [x] `wails dev` startet ohne Fehler unter Linux und zeigt den Default-Screen im WebView
- [x] `wails build` erzeugt auf mindestens einer Dev-Plattform ein natives Binary (Rauchtest)
- [x] `.gitignore` schließt Build-Artefakte, `frontend/node_modules`, `build/bin` etc. aus
- [x] Repository-Root enthält weiterhin die bestehenden Artefakte (`AGENTS.md`, `CONTEXT.md`, `docs/`, `.scratch/`) unverändert

## Notes

Siehe [CONTEXT.md](../../../CONTEXT.md) für Vokabular und [ADR-0001](../../../docs/adr/0001-go-wails-fuer-desktop-gui.md) für die Stack-Begründung. Dokumentation der Prereqs erfolgt separat in Ticket 10 — hier reicht ein funktionierendes Skelett.

## Comments

### Implementiert (Linux verifiziert, Windows offen)

Wails v2.15.0 Vanilla+Vite-Template ins Repository-Root initialisiert. Modulpfad
ist `github.com/ToniBlankenburg/boxclub` (passend zum Remote), damit die späteren
Packages `service/` und `importer/` sauber importierbar sind.

**Auf Linux (Mint 22.1) verifiziert:**

- `wails dev` startet fehlerfrei, WebKit-WebView kommt hoch, DevServer auf
  `http://localhost:34115` liefert den Default-Screen aus.
- `wails build` erzeugt `build/bin/boxclub` (ELF, 8,8 MB).
- `go vet ./...`, `go test ./...`, `gofmt -l .` sind sauber.

**Noch offen — `wails dev` / `wails build` unter Windows.** Konnte in dieser
Session nicht geprüft werden (kein Windows-Rechner verfügbar). Das ist das letzte
verbleibende Acceptance-Kriterium; bitte einmal auf dem Windows-Notebook laufen
lassen.

### webkit2gtk 4.1 statt 4.0

Ubuntu 24.04 / Mint 22 liefern kein `libwebkit2gtk-4.0-dev` mehr, nur noch 4.1.
Deshalb steht in `wails.json` jetzt `"build:tags": "webkit2_41"`. Das Tag wird
ausschließlich von Wails' Linux-Dateien ausgewertet und ist unter macOS und
Windows wirkungslos — `wails dev` / `wails build` funktionieren dadurch ohne
zusätzliche Kommandozeilen-Flags. Voraussetzung auf Linux ist jetzt
`sudo apt install libgtk-3-dev libwebkit2gtk-4.1-dev`.

**Ticket 10 anpassen:** dessen Acceptance Criteria nennen noch
`libwebkit2gtk-4.0-dev`.

### Zwei Fallstricke, die über das Template hinaus behoben wurden

1. **Zeilenenden.** Sämtliche 118 getrackten Dateien lagen im Working Tree als
   CRLF vor, committet waren sie als LF — jeder Commit hätte das komplette Repo
   umgeschrieben und damit das Kriterium "bestehende Artefakte unverändert"
   verletzt. Inhaltlich waren die Dateien nachweislich identisch (nur CR am
   Zeilenende). Auf LF zurückgesetzt und `.gitattributes` mit `* text=auto eol=lf`
   ergänzt, damit der Windows-Rechner das nicht erneut einschleppt.

2. **`frontend/dist/gitkeep`.** `main.go` bindet das Frontend über
   `//go:embed all:frontend/dist` ein; fehlt das Verzeichnis, kompiliert das
   Modul nicht. Das Template committet dafür ein `gitkeep`, das Vite aber bei
   jedem Build löscht — Working Tree dauerhaft dirty. Lösung ohne eigenen
   Build-Code: ein zusätzliches `frontend/public/gitkeep`. Vite kopiert
   `publicDir` nach dem Leeren von `outDir` hinein, und `public/` ist Vites
   Default — es braucht also keine `vite.config.js`. Verifiziert: ein Export nur
   der getrackten Dateien lässt sich ohne `npm install` und ohne Wails-CLI mit
   `go build ./...` und `go test ./...` übersetzen.

Keine Tests in diesem Ticket — es enthält per Definition keine Fachlogik und
damit keinen Seam, an dem sich testen ließe. Die Referenz-Testdatei
`service/member_service_test.go` entsteht in Ticket 02.

### Nach Code-Review angepasst

- `frontend/vite.config.js` wieder entfernt: das `frontend/public/gitkeep` oben
  liefert dieselbe Garantie ohne eigenen Build-Code. Damit bleibt auch die Zusage
  aus `CLAUDE.md` intakt, dass es "keine eigene npm-Pipeline über Wails'
  eingebauten Schritt hinaus" gibt.
- `README.md` auf Skelett-Umfang gekürzt. Die Prereq- und Troubleshooting-Doku
  gehört laut Notes dieses Tickets zu Ticket 10; die README verweist jetzt nur
  noch dorthin.
- `.gitattributes` auf Endungen reduziert, die tatsächlich im Repo vorkommen
  (`*.png`, `*.ico`, `*.woff2`). `*.db` wäre ohnehin wirkungslos gewesen, weil
  `.gitignore` `*.db` bereits ausschließt.

Offen gelassen: `"build:tags": "webkit2_41"` steht bewusst in `wails.json` und
nicht in einer lokalen Override-Datei, damit auf Linux das im Ticket geforderte
blanke `wails dev` funktioniert. Für Ticket 11 (macOS-Release) ist das Tag
wirkungslos.
